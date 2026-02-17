package fastlane

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestReadChangelogsSkipsNonText(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	changelogDir := filepath.Join(localeDir, "changelogs")
	if err := os.MkdirAll(changelogDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changelogDir, "notes.md"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changelogDir, "100.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	changelogs, err := readChangelogs(localeDir)
	if err != nil {
		t.Fatalf("readChangelogs error: %v", err)
	}
	if len(changelogs) != 1 || changelogs["100"] != "keep" {
		t.Fatalf("unexpected changelogs: %+v", changelogs)
	}
}

func TestReadImagesIgnoresUnknownDirsAndFiles(t *testing.T) {
	dir := t.TempDir()
	localeDir := filepath.Join(dir, "en-US")
	imagesDir := filepath.Join(localeDir, "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "featureGraphic.jpg"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "featureGraphic.png"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "README.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(imagesDir, "unknownDir"), 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	images, err := readImages(localeDir)
	if err != nil {
		t.Fatalf("readImages error: %v", err)
	}
	if len(images["featureGraphic"]) != 1 || filepath.Ext(images["featureGraphic"][0]) != ".png" {
		t.Fatalf("expected png preferred for featureGraphic, got %v", images["featureGraphic"])
	}
}

func TestParseDirectorySkipsNonDirEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	localeDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localeDir, "title.txt"), []byte("Title"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	metas, err := ParseDirectory(dir)
	if err != nil {
		t.Fatalf("ParseDirectory error: %v", err)
	}
	if len(metas) != 1 || metas[0].Locale != "en-US" {
		t.Fatalf("unexpected metas: %+v", metas)
	}
}

func TestWriteLocaleNilMetadata(t *testing.T) {
	dir := t.TempDir()
	err := WriteLocale(dir, nil)
	if err != nil {
		t.Errorf("WriteLocale(nil) should not error, got: %v", err)
	}
}

func TestWriteDirectoryEmpty(t *testing.T) {
	dir := t.TempDir()
	err := WriteDirectory(dir, []LocaleMetadata{})
	if err != nil {
		t.Errorf("WriteDirectory with empty metadata should not error, got: %v", err)
	}
}

func TestWriteDirectoryNil(t *testing.T) {
	dir := t.TempDir()
	err := WriteDirectory(dir, nil)
	if err != nil {
		t.Errorf("WriteDirectory with nil metadata should not error, got: %v", err)
	}
}

func TestWriteLocaleAllFieldsSet(t *testing.T) {
	dir := t.TempDir()
	meta := &LocaleMetadata{
		Locale:              "en-US",
		Title:               "Test Title",
		TitleSet:            true,
		ShortDescription:    "Short desc",
		ShortDescriptionSet: true,
		FullDescription:     "Full description",
		FullDescriptionSet:  true,
		Video:               "https://video.example.com",
		VideoSet:            true,
		Changelogs: map[string]string{
			"100": "Version 100 notes",
			"200": "Version 200 notes",
		},
	}

	err := WriteLocale(dir, meta)
	if err != nil {
		t.Fatalf("WriteLocale error: %v", err)
	}

	localeDir := filepath.Join(dir, "en-US")
	titlePath := filepath.Join(localeDir, "title.txt")
	if _, err := os.Stat(titlePath); err != nil {
		t.Errorf("title.txt not created: %v", err)
	}

	shortPath := filepath.Join(localeDir, "short_description.txt")
	if _, err := os.Stat(shortPath); err != nil {
		t.Errorf("short_description.txt not created: %v", err)
	}

	fullPath := filepath.Join(localeDir, "full_description.txt")
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("full_description.txt not created: %v", err)
	}

	videoPath := filepath.Join(localeDir, "video.txt")
	if _, err := os.Stat(videoPath); err != nil {
		t.Errorf("video.txt not created: %v", err)
	}

	changelogDir := filepath.Join(localeDir, "changelogs")
	changelog100 := filepath.Join(changelogDir, "100.txt")
	if _, err := os.Stat(changelog100); err != nil {
		t.Errorf("changelog 100.txt not created: %v", err)
	}
}

func TestCompareImageNames(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"numeric 1 less than 10", "1.png", "10.png", true},
		{"numeric 10 not less than 1", "10.png", "1.png", false},
		{"alphabetic ordering", "a.png", "b.png", true},
		{"same name png vs jpg", "image.png", "image.jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareImageNames(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareImageNames(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestImageExtRank(t *testing.T) {
	tests := []struct {
		ext      string
		expected int
	}{
		{".png", 2},
		{".jpg", 1},
		{".jpeg", 1},
		{".webp", 0},
		{".unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := imageExtRank(tt.ext)
			if result != tt.expected {
				t.Errorf("imageExtRank(%q) = %d, want %d", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestSingleImageExtRank(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		ext      string
		expected int
	}{
		{"icon png", "icon", ".png", 2},
		{"icon jpg", "icon", ".jpg", 0},
		{"icon jpeg", "icon", ".jpeg", 0},
		{"feature png", "featureGraphic", ".png", 2},
		{"feature jpg", "featureGraphic", ".jpg", 1},
		{"unknown ext", "icon", ".gif", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := singleImageExtRank(tt.base, tt.ext)
			if result != tt.expected {
				t.Errorf("singleImageExtRank(%q, %q) = %d, want %d", tt.base, tt.ext, result, tt.expected)
			}
		})
	}
}

func TestPreferSingleImage(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		current   string
		candidate string
		expected  string
	}{
		{
			name:      "prefer png over jpg for icon",
			base:      "icon",
			current:   "icon.jpg",
			candidate: "icon.png",
			expected:  "icon.png",
		},
		{
			name:      "keep current if better rank",
			base:      "icon",
			current:   "icon.png",
			candidate: "icon.jpg",
			expected:  "icon.png",
		},
		{
			name:      "prefer png for feature graphic",
			base:      "featureGraphic",
			current:   "featureGraphic.jpg",
			candidate: "featureGraphic.png",
			expected:  "featureGraphic.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preferSingleImage(tt.base, tt.current, tt.candidate)
			if result != tt.expected {
				t.Errorf("preferSingleImage(%q, %q, %q) = %q, want %q", tt.base, tt.current, tt.candidate, result, tt.expected)
			}
		})
	}
}

func TestParseDirectoryEmptyDir(t *testing.T) {
	dir := t.TempDir()
	metas, err := ParseDirectory(dir)
	if err != nil {
		t.Fatalf("ParseDirectory on empty dir should not error, got: %v", err)
	}
	if len(metas) != 0 {
		t.Errorf("expected 0 locales, got %d", len(metas))
	}
}

func TestWriteTextFileError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: path handling differs")
	}

	err := writeTextFile("/nonexistent/path/file.txt", "content")
	if err == nil {
		t.Error("writeTextFile should error when directory doesn't exist")
	}
}
