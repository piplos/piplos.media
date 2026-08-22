package media_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/piplos/piplos.media/internal/media"
)

func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 80, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateVariants(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "cover.png")
	writeTestPNG(t, src, 1200, 800)

	got, err := media.GenerateVariants(src)
	if err != nil {
		t.Fatal(err)
	}
	if got.Full != "cover.webp" {
		t.Fatalf("full name = %q", got.Full)
	}
	for _, name := range []string{"cover.webp", "cover-480.webp", "cover-960.webp"} {
		st, err := os.Stat(filepath.Join(dir, name))
		if err != nil || st.Size() == 0 {
			t.Fatalf("missing/empty variant %s: %v", name, err)
		}
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("expected original PNG removed, got err=%v", err)
	}
}

func TestGenerateVariantsKeepsWebPMaster(t *testing.T) {
	dir := t.TempDir()
	srcPNG := filepath.Join(dir, "hero.png")
	writeTestPNG(t, srcPNG, 64, 48)
	if _, err := media.GenerateVariants(srcPNG); err != nil {
		t.Fatal(err)
	}
	master := filepath.Join(dir, "hero.webp")
	if _, err := media.GenerateVariants(master); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(master); err != nil {
		t.Fatalf("webp master must remain: %v", err)
	}
}

func TestPreferWebPURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://api.test/uploads/p/a.png", "https://api.test/uploads/p/a.webp"},
		{"/uploads/p/a.JPG", "/uploads/p/a.webp"},
		{"/uploads/a.png#frag", "/uploads/a.webp#frag"},
		{"https://api.test/uploads/p/a.webp", "https://api.test/uploads/p/a.webp"},
		{"https://cdn.test/other.png", "https://cdn.test/other.png"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := media.PreferWebPURL(tc.in); got != tc.want {
			t.Errorf("PreferWebPURL(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestPreferWebPInText(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`src="/uploads/a.png"`, `src="/uploads/a.webp"`},
		{`src='/uploads/a.PNG?v=2'`, `src='/uploads/a.webp?v=2'`},
		{`https://api.test/uploads/x.jpeg#frag`, `https://api.test/uploads/x.webp#frag`},
		{`no uploads here.png`, `no uploads here.png`},
	}
	for _, tc := range cases {
		if got := media.PreferWebPInText(tc.in); got != tc.want {
			t.Errorf("PreferWebPInText(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsVariantFilename(t *testing.T) {
	if !media.IsVariantFilename("x-480.webp") || !media.IsVariantFilename("x-960.webp") {
		t.Fatal("expected sized webp to be variants")
	}
	if media.IsVariantFilename("x.webp") || media.IsVariantFilename("x.png") {
		t.Fatal("full webp/png must not be treated as sized variants")
	}
}

func TestVariantPathsOnlySizedForWebPMaster(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "cover.png")
	if paths := media.VariantPaths(pngPath); len(paths) != 0 {
		t.Fatalf("png leftover must have no managed sidecars, got %v", paths)
	}
	webpPath := filepath.Join(dir, "cover.webp")
	paths := media.VariantPaths(webpPath)
	if len(paths) != 2 {
		t.Fatalf("webp master sidecars = %v", paths)
	}
	for _, p := range paths {
		base := filepath.Base(p)
		if base != "cover-480.webp" && base != "cover-960.webp" {
			t.Fatalf("unexpected sidecar %s", base)
		}
	}
}

func TestRemoveVariantsKeepsMaster(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "shot.png")
	writeTestPNG(t, src, 80, 60)
	if _, err := media.GenerateVariants(src); err != nil {
		t.Fatal(err)
	}
	master := filepath.Join(dir, "shot.webp")
	// Deleting a leftover PNG path must not wipe the WebP family.
	media.RemoveVariants(filepath.Join(dir, "shot.png"))
	if _, err := os.Stat(master); err != nil {
		t.Fatalf("master removed via png path: %v", err)
	}
	media.RemoveVariants(master)
	if _, err := os.Stat(master); err != nil {
		t.Fatalf("RemoveVariants must keep master file: %v", err)
	}
	for _, name := range []string{"shot-480.webp", "shot-960.webp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected sized %s removed", name)
		}
	}
}

func TestNameConflicts(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "tmp.png"), 8, 8)
	if _, err := media.GenerateVariants(filepath.Join(dir, "tmp.png")); err != nil {
		t.Fatal(err)
	}
	if !media.NameConflicts(dir, "tmp.png") {
		t.Fatal("existing webp family must conflict with tmp.png")
	}
	if media.NameConflicts(dir, "other.png") {
		t.Fatal("other.png should be free")
	}
}

func TestRebuildDirSkipsCompleteFamily(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "done.png")
	writeTestPNG(t, src, 40, 30)
	if _, err := media.GenerateVariants(src); err != nil {
		t.Fatal(err)
	}
	// Leftover PNG after a partial old deploy.
	writeTestPNG(t, src, 40, 30)
	ok, failed, err := media.RebuildDir(dir)
	if err != nil || failed != 0 || ok < 1 {
		t.Fatalf("RebuildDir ok=%d failed=%d err=%v", ok, failed, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("complete family should drop leftover PNG without re-encode path failing")
	}
}
