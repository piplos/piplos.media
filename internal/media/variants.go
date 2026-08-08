// Package media generates WebP size variants next to uploaded raster images.
// PNG/JPEG originals are removed after a successful encode; full .webp is the master.
package media

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
	gowebp "github.com/mayahiro/go-webp"
	_ "golang.org/x/image/webp"
)

// Card-oriented widths used by the public site (portfolio / articles previews).
var VariantWidths = []int{480, 960}

const (
	maxFullWidth = 1920
	webpQuality  = 80
)

// uploadRasterExtRE rewrites PNG/JPEG on /uploads/ paths (optional host, ?/# suffix).
var uploadRasterExtRE = regexp.MustCompile(
	`(?i)((?:https?://[^"'>\s]+)?/uploads/[^"'>\s?#]+)\.(?:png|jpe?g)([?#][^"'>\s]*)?`,
)

// Variants holds basenames (not paths) of generated WebP files.
type Variants struct {
	Full  string   // stem.webp
	Sized []string // stem-480.webp, stem-960.webp
}

// IsRasterExt reports whether ext is a raster format we convert (not SVG/GIF).
func IsRasterExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

// isDisposableOriginal reports PNG/JPEG sources we remove after WebP succeeds.
func isDisposableOriginal(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

// IsVariantFilename reports sized sidecars (skip as rebuild sources).
func IsVariantFilename(name string) bool {
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, ".webp") {
		return false
	}
	base := strings.TrimSuffix(lower, ".webp")
	for _, w := range VariantWidths {
		if strings.HasSuffix(base, fmt.Sprintf("-%d", w)) {
			return true
		}
	}
	return false
}

// Stem returns filename without extension.
func Stem(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// FullWebPName is the full-size WebP sibling of an original raster file.
func FullWebPName(filename string) string {
	return Stem(filename) + ".webp"
}

// SizedWebPName is stem-<width>.webp.
func SizedWebPName(filename string, width int) string {
	return fmt.Sprintf("%s-%d.webp", Stem(filename), width)
}

// NameConflicts reports whether name or its WebP family already exists in dir.
func NameConflicts(dir, name string) bool {
	if fileExists(filepath.Join(dir, name)) {
		return true
	}
	if !IsRasterExt(filepath.Ext(name)) || IsVariantFilename(name) {
		return false
	}
	if fileExists(filepath.Join(dir, FullWebPName(name))) {
		return true
	}
	for _, w := range VariantWidths {
		if fileExists(filepath.Join(dir, SizedWebPName(name, w))) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// PreferWebPURL rewrites a public/relative upload URL to the full WebP sibling.
func PreferWebPURL(raw string) string {
	if raw == "" {
		return raw
	}
	pathPart, query, ok := splitURLPath(raw)
	if !ok {
		return raw
	}
	ext := strings.ToLower(filepath.Ext(pathPart))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		pathPart = pathPart[:len(pathPart)-len(ext)] + ".webp"
	default:
		return raw
	}
	if query != "" {
		return pathPart + "?" + query
	}
	return pathPart
}

// PreferWebPInText rewrites /uploads/*.png|jpg|jpeg inside HTML/text to .webp.
func PreferWebPInText(s string) string {
	if s == "" || !strings.Contains(s, "/uploads/") {
		return s
	}
	return uploadRasterExtRE.ReplaceAllString(s, "${1}.webp${2}")
}

func splitURLPath(raw string) (pathPart, query string, ok bool) {
	if !strings.Contains(raw, "/uploads/") {
		return "", "", false
	}
	pathPart = raw
	if i := strings.IndexByte(raw, '?'); i >= 0 {
		pathPart, query = raw[:i], raw[i+1:]
	}
	return pathPart, query, true
}

// GenerateVariants writes full + sized WebP files next to absPath.
// After success, PNG/JPEG originals are deleted; the full .webp becomes the master.
// On mid-encode failure, partial WebP outputs from this call are rolled back.
func GenerateVariants(absPath string) (out Variants, err error) {
	name := filepath.Base(absPath)
	ext := strings.ToLower(filepath.Ext(name))
	if !IsRasterExt(ext) || IsVariantFilename(name) {
		return Variants{}, nil
	}

	img, openErr := imaging.Open(absPath, imaging.AutoOrientation(true))
	if openErr != nil {
		return Variants{}, fmt.Errorf("decode %s: %w", name, openErr)
	}

	dir := filepath.Dir(absPath)
	out = Variants{Full: FullWebPName(name)}
	written := make([]string, 0, 1+len(VariantWidths))

	defer func() {
		if err == nil {
			return
		}
		for _, p := range written {
			_ = os.Remove(p)
		}
	}()

	full := img
	if b := img.Bounds(); b.Dx() > maxFullWidth {
		full = imaging.Resize(img, maxFullWidth, 0, imaging.Lanczos)
	}
	fullPath := filepath.Join(dir, out.Full)
	if err = writeWebP(fullPath, full); err != nil {
		return Variants{}, err
	}
	written = append(written, fullPath)

	for _, w := range VariantWidths {
		sized := img
		if b := img.Bounds(); b.Dx() > w {
			sized = imaging.Resize(img, w, 0, imaging.Lanczos)
		}
		fname := SizedWebPName(name, w)
		sizedPath := filepath.Join(dir, fname)
		if err = writeWebP(sizedPath, sized); err != nil {
			return Variants{}, err
		}
		written = append(written, sizedPath)
		out.Sized = append(out.Sized, fname)
	}

	// Best-effort: WebP family is already the source of truth.
	if isDisposableOriginal(ext) {
		_ = os.Remove(absPath)
	}
	return out, nil
}

func writeWebP(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(path), err)
	}
	defer f.Close()
	opts := &gowebp.Options{
		Compression: gowebp.CompressionLossy,
		Quality:     webpQuality,
		Mode:        gowebp.ModeFast,
	}
	if err := gowebp.Encode(f, img, opts); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
	}
	return nil
}

// VariantPaths returns sized WebP sidecars for a WebP master.
// Disposable PNG/JPEG leftovers have no managed sidecars (master is independent).
func VariantPaths(absPath string) []string {
	name := filepath.Base(absPath)
	if strings.ToLower(filepath.Ext(name)) != ".webp" || IsVariantFilename(name) {
		return nil
	}
	dir := filepath.Dir(absPath)
	out := make([]string, 0, len(VariantWidths))
	for _, w := range VariantWidths {
		out = append(out, filepath.Join(dir, SizedWebPName(name, w)))
	}
	return out
}

// RemoveVariants deletes sized WebP sidecars for a WebP master (best-effort).
func RemoveVariants(absPath string) {
	for _, p := range VariantPaths(absPath) {
		_ = os.Remove(p)
	}
}

// RenameVariants renames sized sidecars when a WebP master is renamed/moved.
// The master itself is renamed by the caller first.
func RenameVariants(fromAbs, toAbs string) {
	fromVars := VariantPaths(fromAbs)
	toVars := VariantPaths(toAbs)
	if len(fromVars) == 0 || len(fromVars) != len(toVars) {
		return
	}
	for i := range fromVars {
		_ = os.Rename(fromVars[i], toVars[i])
	}
}

// RebuildDir walks root and regenerates variants for every raster master.
// Skips .webp when a disposable PNG/JPEG sibling with the same stem still exists
// (that sibling will produce the family and then be removed).
func RebuildDir(root string) (ok, failed int, err error) {
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		ext := filepath.Ext(name)
		if !IsRasterExt(ext) || IsVariantFilename(name) {
			return nil
		}
		if strings.EqualFold(ext, ".webp") && hasDisposableSibling(path) {
			return nil
		}
		if _, genErr := GenerateVariants(path); genErr != nil {
			failed++
			return nil
		}
		ok++
		return nil
	})
	return ok, failed, err
}

func hasDisposableSibling(webpAbs string) bool {
	dir := filepath.Dir(webpAbs)
	stem := Stem(filepath.Base(webpAbs))
	for _, e := range []string{".png", ".jpg", ".jpeg", ".PNG", ".JPG", ".JPEG"} {
		if fileExists(filepath.Join(dir, stem+e)) {
			return true
		}
	}
	return false
}
