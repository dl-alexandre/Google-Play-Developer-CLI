package fastlane

import (
	"fmt"

	"github.com/dl-alexandre/gpd/internal/migrate"
)

// ValidateLocale validates fastlane metadata for a single locale.
func ValidateLocale(meta *LocaleMetadata) []migrate.ValidationError {
	var errs []migrate.ValidationError

	if meta == nil {
		return errs
	}

	if meta.TitleSet {
		if err := migrate.ValidateText("title", meta.Title); err != nil {
			err.Locale = meta.Locale
			errs = append(errs, *err)
		}
	}
	if meta.ShortDescriptionSet {
		if err := migrate.ValidateText("shortDescription", meta.ShortDescription); err != nil {
			err.Locale = meta.Locale
			errs = append(errs, *err)
		}
	}
	if meta.FullDescriptionSet {
		if err := migrate.ValidateText("fullDescription", meta.FullDescription); err != nil {
			err.Locale = meta.Locale
			errs = append(errs, *err)
		}
	}
	for key, text := range meta.Changelogs {
		if err := migrate.ValidateText("releaseNotes", text); err != nil {
			err.Locale = meta.Locale
			err.Message = fmt.Sprintf("release notes %s exceeds limit", key)
			errs = append(errs, *err)
		}
	}

	return errs
}
