package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/piplos/piplos.media/internal/utils"
)

// maxManifestSize guards against malformed archives.
const maxManifestSize = 10 << 20 // 10 MiB

// readManifest expects manifest.json to be the first tar entry.
func readManifest(tr *tar.Reader) (Manifest, error) {
	var m Manifest
	hdr, err := tr.Next()
	if err != nil {
		return m, fmt.Errorf("read archive: %w", err)
	}
	if hdr.Name != "manifest.json" {
		return m, fmt.Errorf("invalid archive: manifest.json must be the first entry, got %s", hdr.Name)
	}
	data, err := io.ReadAll(io.LimitReader(tr, maxManifestSize))
	if err != nil {
		return m, fmt.Errorf("read manifest: %w", err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("parse manifest: %w", err)
	}
	if m.Version != manifestVersion {
		return m, fmt.Errorf("unsupported archive version %d", m.Version)
	}
	return m, nil
}

// restoreArchive restores database tables and/or uploaded files from the
// archive in r. The archive is written by writeArchive: manifest.json first,
// then db/<table>.copy entries, then files/<rel> entries.
func restoreArchive(ctx context.Context, r io.Reader, pool *pgxpool.Pool, uploadDir string) (Manifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return Manifest{}, fmt.Errorf("open gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	manifest, err := readManifest(tr)
	if err != nil {
		return manifest, err
	}

	restoreDB := len(manifest.Tables) > 0
	restoreFiles := manifest.Type == TypeFull || manifest.Type == TypeFiles

	// Файлы распаковываются во временный каталог внутри uploadDir (тот же
	// filesystem/volume, поэтому rename атомарен); подмена происходит только
	// после успешного восстановления БД и полного чтения архива. Скрытые
	// каталоги (".restore-*") не видны в файловом менеджере админки.
	var tmpFilesDir string
	if restoreFiles {
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			return manifest, fmt.Errorf("create upload dir: %w", err)
		}
		tmpFilesDir, err = os.MkdirTemp(uploadDir, ".restore-*")
		if err != nil {
			return manifest, fmt.Errorf("create restore temp dir: %w", err)
		}
		defer os.RemoveAll(tmpFilesDir)
	}

	var dbRestore *dbRestoreState
	if restoreDB {
		dbRestore, err = beginDBRestore(ctx, pool, manifest.Tables)
		if err != nil {
			return manifest, err
		}
		defer dbRestore.rollback(ctx)
	}

	restoredTables := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return manifest, fmt.Errorf("read archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		switch {
		case strings.HasPrefix(hdr.Name, "db/"):
			if !restoreDB {
				continue
			}
			table := strings.TrimSuffix(strings.TrimPrefix(hdr.Name, "db/"), ".copy")
			if err := dbRestore.copyTable(ctx, table, tr); err != nil {
				return manifest, err
			}
			restoredTables[table] = true
		case strings.HasPrefix(hdr.Name, "files/"):
			if !restoreFiles {
				continue
			}
			if err := extractFile(tmpFilesDir, strings.TrimPrefix(hdr.Name, "files/"), tr); err != nil {
				return manifest, err
			}
		}
	}

	if restoreDB {
		for _, t := range manifest.Tables {
			if !restoredTables[t.Name] {
				return manifest, fmt.Errorf("archive is missing data for table %s", t.Name)
			}
		}
		if err := dbRestore.commit(ctx); err != nil {
			return manifest, err
		}
	}

	if restoreFiles {
		if err := swapUploadDir(uploadDir, tmpFilesDir); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

// dbRestoreState is an open transaction that truncates and refills tables.
type dbRestoreState struct {
	conn    *pgxpool.Conn
	tx      pgx.Tx
	columns map[string][]string
	done    bool
}

// beginDBRestore validates tables against the live schema, opens a transaction
// and truncates all target tables (single statement, consistent state).
func beginDBRestore(ctx context.Context, pool *pgxpool.Pool, tables []TableDump) (*dbRestoreState, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("begin restore transaction: %w", err)
	}
	st := &dbRestoreState{conn: conn, tx: tx, columns: map[string][]string{}}

	existing, err := publicTables(ctx, tx)
	if err != nil {
		st.rollback(ctx)
		return nil, err
	}
	existingSet := map[string]bool{}
	for _, t := range existing {
		existingSet[t] = true
	}

	names := make([]string, 0, len(tables))
	for _, t := range tables {
		if !existingSet[t.Name] {
			st.rollback(ctx)
			return nil, fmt.Errorf("table %s from backup does not exist in the database", t.Name)
		}
		cols, err := tableColumns(ctx, tx, t.Name)
		if err != nil {
			st.rollback(ctx)
			return nil, err
		}
		colSet := map[string]bool{}
		for _, c := range cols {
			colSet[c] = true
		}
		for _, c := range t.Columns {
			if !colSet[c] {
				st.rollback(ctx)
				return nil, fmt.Errorf("column %s.%s from backup does not exist in the database", t.Name, c)
			}
		}
		st.columns[t.Name] = t.Columns
		names = append(names, pgx.Identifier{t.Name}.Sanitize())
	}

	if _, err := tx.Exec(ctx, "TRUNCATE TABLE "+strings.Join(names, ", ")+" RESTART IDENTITY CASCADE"); err != nil {
		st.rollback(ctx)
		return nil, fmt.Errorf("truncate tables: %w", err)
	}
	return st, nil
}

// copyTable loads one table's COPY data from r.
func (st *dbRestoreState) copyTable(ctx context.Context, table string, r io.Reader) error {
	cols, ok := st.columns[table]
	if !ok {
		return fmt.Errorf("unexpected table %s in archive", table)
	}
	copySQL := fmt.Sprintf("COPY %s (%s) FROM STDIN",
		pgx.Identifier{table}.Sanitize(), quoteColumns(cols))
	if _, err := st.conn.Conn().PgConn().CopyFrom(ctx, r, copySQL); err != nil {
		return fmt.Errorf("restore table %s: %w", table, err)
	}
	return nil
}

func (st *dbRestoreState) commit(ctx context.Context) error {
	if err := st.tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit restore transaction: %w", err)
	}
	st.done = true
	st.conn.Release()
	return nil
}

func (st *dbRestoreState) rollback(ctx context.Context) {
	if st.done {
		return
	}
	st.done = true
	_ = st.tx.Rollback(ctx)
	st.conn.Release()
}

// extractFile writes one archive entry into dir, rejecting path escapes.
func extractFile(dir, rel string, r io.Reader) error {
	abs, err := utils.SafeJoin(dir, rel)
	if err != nil {
		return fmt.Errorf("invalid file path in archive: %q", rel)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", rel, err)
	}
	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("create %s: %w", rel, err)
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return fmt.Errorf("write %s: %w", rel, err)
	}
	return f.Close()
}

// swapUploadDir replaces the contents of uploadDir with the contents of
// tmpDir (a temp dir inside uploadDir). The directory itself stays in place
// (it may be a mounted volume, so it cannot be renamed).
func swapUploadDir(uploadDir, tmpDir string) error {
	tmpName := filepath.Base(tmpDir)
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return fmt.Errorf("read upload dir: %w", err)
	}
	for _, e := range entries {
		if e.Name() == tmpName {
			continue
		}
		if err := os.RemoveAll(filepath.Join(uploadDir, e.Name())); err != nil {
			return fmt.Errorf("clear upload dir: %w", err)
		}
	}
	newEntries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("read restore temp dir: %w", err)
	}
	for _, e := range newEntries {
		src := filepath.Join(tmpDir, e.Name())
		dst := filepath.Join(uploadDir, e.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("move restored files: %w", err)
		}
	}
	return nil
}
