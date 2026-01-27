package fastlane

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndParseDirectory(t *testing.T) {
	dir := t.TempDir()
	meta := []LocaleMetadata{
		{
			Locale:              "en-US",
			Title:               "Title",
			TitleSet:            true,
			ShortDescription:    "Short",
			ShortDescriptionSet: true,
			FullDescription:     "Full",
			FullDescriptionSet:  true,
			Video:               "https://example.com",
			VideoSet:            true,
			Changelogs: map[string]string{
				"100": "notes",
			},
		},
	}
	if err := WriteDirectory(dir, meta); err != nil {
		t.Fatalf("WriteDirectory error: %v", err)
	}

	imagesDir := filepath.Join(dir, "en-US", "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatalf("mkdir images error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "icon.png"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write icon error: %v", err)
	}

	parsed, err := ParseDirectory(dir)
	if err != nil {
		t.Fatalf("ParseDirectory error: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 locale, got %d", len(parsed))
	}
	got := parsed[0]
	if !got.TitleSet || got.Title != "Title" {
		t.Fatalf("unexpected title: %+v", got)
	}
	if got.Changelogs["100"] != "notes" {
		t.Fatalf("unexpected changelog: %+v", got.Changelogs)
	}
	if len(got.Images) == 0 || len(got.Images["icon"]) != 1 {
		t.Fatalf("unexpected images: %+v", got.Images)
	}
}

func TestReadOptionalTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "title.txt")
	if err := os.WriteFile(path, []byte("hello\r\n"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	val, ok, err := readOptionalTextFile(path)
	if err != nil || !ok {
		t.Fatalf("expected ok, err=%v", err)
	}
	if val != "hello" {
		t.Fatalf("expected trimmed value, got %q", val)
	}
}

func TestReadImagesAndSorting(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	imagesDir := filepath.Join(localeDir, "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "icon.jpg"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "icon.png"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	screens := filepath.Join(imagesDir, "phoneScreenshots")
	if err := os.MkdirAll(screens, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	files := []string{"10.png", "2.jpg", "2.png", "1.jpg"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(screens, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	images, err := readImages(localeDir)
	if err != nil {
		t.Fatalf("readImages error: %v", err)
	}
	if len(images["icon"]) != 1 || filepath.Ext(images["icon"][0]) != ".png" {
		t.Fatalf("expected png preferred for icon, got %v", images["icon"])
	}
	screenFiles := images["phoneScreenshots"]
	if len(screenFiles) != 4 {
		t.Fatalf("expected 4 screenshot files, got %d", len(screenFiles))
	}
	base := filepath.Base
	if base(screenFiles[0]) != "1.jpg" || base(screenFiles[1]) != "2.png" || base(screenFiles[2]) != "2.jpg" || base(screenFiles[3]) != "10.png" {
		t.Fatalf("unexpected screenshot order: %v", screenFiles)
	}
}

func TestImageHelpers(t *testing.T) {
	if !isScreenshotDir("phoneScreenshots") || isScreenshotDir("other") {
		t.Fatalf("unexpected screenshot dir result")
	}
	if !isSingleImage("icon") || isSingleImage("other") {
		t.Fatalf("unexpected single image result")
	}
	if !isImageFile("a.PNG") || isImageFile("a.gif") {
		t.Fatalf("unexpected image file result")
	}
	if imageExtRank(".png") <= imageExtRank(".jpg") {
		t.Fatalf("expected png rank higher than jpg")
	}
	if !compareImageNames("2.png", "10.png") {
		t.Fatalf("expected numeric compare")
	}
}

func TestReadChangelogsErrors(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	changelogDir := filepath.Join(localeDir, "changelogs")
	if err := os.WriteFile(changelogDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if _, err := readChangelogs(localeDir); err == nil {
		t.Fatalf("expected error for non-dir changelogs")
	}
}

func TestReadImagesErrors(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	imagesDir := filepath.Join(localeDir, "images")
	if err := os.WriteFile(imagesDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if _, err := readImages(localeDir); err == nil {
		t.Fatalf("expected error for non-dir images")
	}
}

func TestWriteLocaleNil(t *testing.T) {
	if err := WriteLocale(t.TempDir(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestPreferSingleImageTie(t *testing.T) {
	a := preferSingleImage("featureGraphic", "/a.jpg", "/b.jpg")
	if filepath.Base(a) != "a.jpg" {
		t.Fatalf("expected a.jpg, got %s", filepath.Base(a))
	}
	if !compareByExtensionThenName("a.png", "b.jpg") {
		t.Fatalf("expected png before jpg")
	}
}

func TestParseDirectoryErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if _, err := ParseDirectory(file); err == nil {
		t.Fatalf("expected error for non-dir")
	}
}

func TestReadChangelogsMissing(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	changelogs, err := readChangelogs(localeDir)
	if err != nil || changelogs != nil {
		t.Fatalf("expected nil changelogs")
	}
}

func TestReadImagesMissing(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	images, err := readImages(localeDir)
	if err != nil || images != nil {
		t.Fatalf("expected nil images")
	}
}
