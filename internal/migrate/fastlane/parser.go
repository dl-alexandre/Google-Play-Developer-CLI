// Package fastlane provides parsing and writing for fastlane metadata.
package fastlane

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// LocaleMetadata represents fastlane metadata for a single locale.
type LocaleMetadata struct {
	Locale              string
	Title               string
	TitleSet            bool
	ShortDescription    string
	ShortDescriptionSet bool
	FullDescription     string
	FullDescriptionSet  bool
	Video               string
	VideoSet            bool
	Changelogs          map[string]string
	Images              map[string][]string
}

const (
	extPNG  = ".png"
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
)

// ParseDirectory reads a fastlane metadata directory into structured metadata.
func ParseDirectory(dir string) ([]LocaleMetadata, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	var metas []LocaleMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		locale := entry.Name()
		meta, err := parseLocaleDir(filepath.Join(absDir, locale), locale)
		if err != nil {
			return nil, err
		}
		metas = append(metas, meta)
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Locale < metas[j].Locale
	})
	return metas, nil
}

// WriteDirectory writes locale metadata to a fastlane metadata directory.
func WriteDirectory(dir string, metadata []LocaleMetadata) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for i := range metadata {
		if err := WriteLocale(dir, &metadata[i]); err != nil {
			return err
		}
	}
	return nil
}

// WriteLocale writes metadata for a single locale.
func WriteLocale(dir string, meta *LocaleMetadata) error {
	if meta == nil {
		return nil
	}
	localeDir := filepath.Join(dir, meta.Locale)
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		return err
	}

	if meta.TitleSet {
		if err := writeTextFile(filepath.Join(localeDir, "title.txt"), meta.Title); err != nil {
			return err
		}
	}
	if meta.ShortDescriptionSet {
		if err := writeTextFile(filepath.Join(localeDir, "short_description.txt"), meta.ShortDescription); err != nil {
			return err
		}
	}
	if meta.FullDescriptionSet {
		if err := writeTextFile(filepath.Join(localeDir, "full_description.txt"), meta.FullDescription); err != nil {
			return err
		}
	}
	if meta.VideoSet {
		if err := writeTextFile(filepath.Join(localeDir, "video.txt"), meta.Video); err != nil {
			return err
		}
	}
	if len(meta.Changelogs) > 0 {
		changelogDir := filepath.Join(localeDir, "changelogs")
		if err := os.MkdirAll(changelogDir, 0o755); err != nil {
			return err
		}
		keys := make([]string, 0, len(meta.Changelogs))
		for key := range meta.Changelogs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := writeTextFile(filepath.Join(changelogDir, key+".txt"), meta.Changelogs[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseLocaleDir(localeDir, locale string) (LocaleMetadata, error) {
	meta := LocaleMetadata{
		Locale: locale,
	}

	title, ok, err := readOptionalTextFile(filepath.Join(localeDir, "title.txt"))
	if err != nil {
		return meta, err
	}
	if ok {
		meta.Title = title
		meta.TitleSet = true
	}

	shortDesc, ok, err := readOptionalTextFile(filepath.Join(localeDir, "short_description.txt"))
	if err != nil {
		return meta, err
	}
	if ok {
		meta.ShortDescription = shortDesc
		meta.ShortDescriptionSet = true
	}

	fullDesc, ok, err := readOptionalTextFile(filepath.Join(localeDir, "full_description.txt"))
	if err != nil {
		return meta, err
	}
	if ok {
		meta.FullDescription = fullDesc
		meta.FullDescriptionSet = true
	}

	video, ok, err := readOptionalTextFile(filepath.Join(localeDir, "video.txt"))
	if err != nil {
		return meta, err
	}
	if ok {
		meta.Video = video
		meta.VideoSet = true
	}

	changelogs, err := readChangelogs(localeDir)
	if err != nil {
		return meta, err
	}
	if len(changelogs) > 0 {
		meta.Changelogs = changelogs
	}

	images, err := readImages(localeDir)
	if err != nil {
		return meta, err
	}
	if len(images) > 0 {
		meta.Images = images
	}

	return meta, nil
}

func readOptionalTextFile(path string) (value string, ok bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimRight(string(data), "\r\n"), true, nil
}

func writeTextFile(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o644)
}

func readChangelogs(localeDir string) (map[string]string, error) {
	changelogDir := filepath.Join(localeDir, "changelogs")
	entries, err := os.ReadDir(changelogDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	changelogs := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".txt" {
			continue
		}
		key := strings.TrimSuffix(name, ".txt")
		value, ok, err := readOptionalTextFile(filepath.Join(changelogDir, name))
		if err != nil {
			return nil, err
		}
		if ok {
			changelogs[key] = value
		}
	}
	return changelogs, nil
}

func readImages(localeDir string) (map[string][]string, error) {
	imagesDir := filepath.Join(localeDir, "images")
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	images := map[string][]string{}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(imagesDir, name)
		if entry.IsDir() {
			if !isScreenshotDir(name) {
				continue
			}
			files, err := readImageFiles(fullPath)
			if err != nil {
				return nil, err
			}
			if len(files) > 0 {
				images[name] = files
			}
			continue
		}

		if !isImageFile(name) {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if isSingleImage(base) {
			if singleImageExtRank(base, filepath.Ext(name)) == 0 {
				continue
			}
			existing := images[base]
			if len(existing) == 0 {
				images[base] = []string{fullPath}
				continue
			}
			images[base] = []string{preferSingleImage(base, existing[0], fullPath)}
		}
	}

	return images, nil
}

func readImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isImageFile(entry.Name()) {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Slice(files, func(i, j int) bool {
		return compareImageNames(files[i], files[j])
	})
	return files, nil
}

func isScreenshotDir(name string) bool {
	switch name {
	case "phoneScreenshots", "tabletScreenshots", "sevenInchScreenshots", "tenInchScreenshots", "tvScreenshots", "wearScreenshots":
		return true
	default:
		return false
	}
}

func isSingleImage(name string) bool {
	switch name {
	case "icon", "featureGraphic", "promoGraphic", "tvBanner":
		return true
	default:
		return false
	}
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case extPNG, extJPG, extJPEG:
		return true
	default:
		return false
	}
}

func compareImageNames(a, b string) bool {
	aBase := filepath.Base(a)
	bBase := filepath.Base(b)
	aName := strings.TrimSuffix(aBase, filepath.Ext(aBase))
	bName := strings.TrimSuffix(bBase, filepath.Ext(bBase))
	aNum, aErr := strconv.Atoi(aName)
	bNum, bErr := strconv.Atoi(bName)
	if aErr == nil && bErr == nil {
		if aNum == bNum {
			return compareByExtensionThenName(aBase, bBase)
		}
		return aNum < bNum
	}
	if aErr == nil {
		return true
	}
	if bErr == nil {
		return false
	}
	return aBase < bBase
}

func preferSingleImage(base, current, candidate string) string {
	currentRank := singleImageExtRank(base, filepath.Ext(current))
	candidateRank := singleImageExtRank(base, filepath.Ext(candidate))
	if candidateRank > currentRank {
		return candidate
	}
	if candidateRank < currentRank {
		return current
	}
	if filepath.Base(candidate) < filepath.Base(current) {
		return candidate
	}
	return current
}

func singleImageExtRank(base, ext string) int {
	ext = strings.ToLower(ext)
	if base == "icon" {
		if ext == extPNG {
			return 2
		}
		return 0
	}
	switch ext {
	case extPNG:
		return 2
	case extJPG, extJPEG:
		return 1
	default:
		return 0
	}
}

func compareByExtensionThenName(aBase, bBase string) bool {
	aRank := imageExtRank(filepath.Ext(aBase))
	bRank := imageExtRank(filepath.Ext(bBase))
	if aRank != bRank {
		return aRank > bRank
	}
	return aBase < bBase
}

func imageExtRank(ext string) int {
	switch strings.ToLower(ext) {
	case extPNG:
		return 2
	case extJPG, extJPEG:
		return 1
	default:
		return 0
	}
}
