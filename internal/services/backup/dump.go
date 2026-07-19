package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// manifestVersion is the archive format version.
const manifestVersion = 1

// goose keeps schema state; it must not be dumped or restored.
const gooseTable = "goose_db_version"

// TableDump describes one dumped table inside the manifest.
type TableDump struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Rows    int64    `json:"rows"`
}

// Manifest describes archive contents (stored as manifest.json in the tar root).
type Manifest struct {
	Version   int         `json:"version"`
	Type      string      `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	Tables    []TableDump `json:"tables,omitempty"`
	FileCount int         `json:"file_count,omitempty"`
}

// tableColumns returns orderable column names of a table, skipping
// GENERATED ... STORED columns (COPY cannot insert into them).
func tableColumns(ctx context.Context, q pgxQuerier, table string) ([]string, error) {
	rows, err := q.Query(ctx, `
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1 AND is_generated <> 'ALWAYS'
		ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, fmt.Errorf("columns of %s: %w", table, err)
	}
	defer rows.Close()
	cols := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

// publicTables returns user tables of the public schema (excluding goose).
func publicTables(ctx context.Context, q pgxQuerier) ([]string, error) {
	rows, err := q.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' ORDER BY tablename`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	tables := []string{}
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		if t != gooseTable {
			tables = append(tables, t)
		}
	}
	return tables, rows.Err()
}

// pgxQuerier is the subset of pgx query API used by dump helpers.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// quoteColumns renders a quoted, comma-separated column list.
func quoteColumns(cols []string) string {
	quoted := make([]string, len(cols))
	for i, c := range cols {
		quoted[i] = pgx.Identifier{c}.Sanitize()
	}
	return strings.Join(quoted, ", ")
}

// writeArchive builds a tar.gz backup archive of the given type into w.
// The database is dumped per table with COPY TO (text format) inside a single
// REPEATABLE READ transaction, so all tables share one consistent snapshot.
func writeArchive(ctx context.Context, w io.Writer, pool *pgxpool.Pool, uploadDir, backupType string, createdAt time.Time) (Manifest, error) {
	manifest := Manifest{Version: manifestVersion, Type: backupType, CreatedAt: createdAt.UTC()}
	withDB := backupType == TypeFull || backupType == TypeDB
	withFiles := backupType == TypeFull || backupType == TypeFiles

	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)

	var tmpDir string
	if withDB {
		var err error
		tmpDir, err = os.MkdirTemp("", "piplos-backup-*")
		if err != nil {
			return manifest, fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		manifest.Tables, err = dumpDatabase(ctx, pool, tmpDir)
		if err != nil {
			return manifest, err
		}
	}
	if withFiles {
		// FileCount заранее неизвестен без прохода; пройдём каталог дважды —
		// сначала посчитаем для манифеста, затем добавим файлы в архив.
		count, err := countFiles(uploadDir)
		if err != nil {
			return manifest, err
		}
		manifest.FileCount = count
	}

	// manifest.json первым: restore читает его до данных при потоковом чтении.
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return manifest, fmt.Errorf("marshal manifest: %w", err)
	}
	if err := writeTarEntry(tw, "manifest.json", manifestJSON); err != nil {
		return manifest, err
	}

	for _, t := range manifest.Tables {
		src := filepath.Join(tmpDir, t.Name+".copy")
		if err := addFileToTar(tw, src, "db/"+t.Name+".copy"); err != nil {
			return manifest, err
		}
	}
	if withFiles {
		if err := addUploadsToTar(tw, uploadDir); err != nil {
			return manifest, err
		}
	}

	if err := tw.Close(); err != nil {
		return manifest, fmt.Errorf("close tar: %w", err)
	}
	if err := gz.Close(); err != nil {
		return manifest, fmt.Errorf("close gzip: %w", err)
	}
	return manifest, nil
}

// dumpDatabase copies every public table into <tmpDir>/<table>.copy and
// returns table descriptors for the manifest.
func dumpDatabase(ctx context.Context, pool *pgxpool.Pool, tmpDir string) ([]TableDump, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin dump transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tables, err := publicTables(ctx, tx)
	if err != nil {
		return nil, err
	}

	dumps := make([]TableDump, 0, len(tables))
	for _, table := range tables {
		cols, err := tableColumns(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		if len(cols) == 0 {
			continue
		}
		dst := filepath.Join(tmpDir, table+".copy")
		f, err := os.Create(dst)
		if err != nil {
			return nil, fmt.Errorf("create dump file for %s: %w", table, err)
		}
		copySQL := fmt.Sprintf("COPY (SELECT %s FROM %s) TO STDOUT",
			quoteColumns(cols), pgx.Identifier{table}.Sanitize())
		tag, err := conn.Conn().PgConn().CopyTo(ctx, f, copySQL)
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			return nil, fmt.Errorf("dump table %s: %w", table, err)
		}
		dumps = append(dumps, TableDump{Name: table, Columns: cols, Rows: tag.RowsAffected()})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit dump transaction: %w", err)
	}
	return dumps, nil
}

// writeTarEntry writes an in-memory file into the tar.
func writeTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), ModTime: time.Now()}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("tar header %s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("tar write %s: %w", name, err)
	}
	return nil
}

// addFileToTar streams a file from disk into the tar under entryName.
func addFileToTar(tw *tar.Writer, src, entryName string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	hdr := &tar.Header{Name: entryName, Mode: 0o644, Size: st.Size(), ModTime: st.ModTime()}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("tar header %s: %w", entryName, err)
	}
	if _, err := io.Copy(tw, f); err != nil {
		return fmt.Errorf("tar write %s: %w", entryName, err)
	}
	return nil
}

// countFiles counts regular files under dir (0 when dir is missing).
func countFiles(dir string) (int, error) {
	count := 0
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("count files: %w", err)
	}
	return count, nil
}

// addUploadsToTar adds every regular file under uploadDir as files/<rel>.
func addUploadsToTar(tw *tar.Writer, uploadDir string) error {
	return filepath.WalkDir(uploadDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(uploadDir, p)
		if err != nil {
			return err
		}
		return addFileToTar(tw, p, "files/"+filepath.ToSlash(rel))
	})
}
