package locale

import (
	"fmt"
	"time"
)

var localeFormats = map[string]string{
	"en": "Jan 2, 2006 3:04 PM",
	"es": "2 ene 2006 15:04",
	"fr": "2 janv. 2006 15:04",
	"de": "2. Jan. 2006 15:04",
	"pt": "2 jan 2006 15:04",
	"zh": "2006年1月2日 15:04",
	"ja": "2006年1月2日 15:04",
}

var defaultFormat = "Jan 2, 2006 3:04 PM"

func FormatTimestamp(t time.Time, loc string, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}

	localized := t.In(location)

	format, ok := localeFormats[loc]
	if !ok {
		format = defaultFormat
	}

	return localized.Format(format)
}

func FormatDate(t time.Time, loc string, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}

	localized := t.In(location)

	switch loc {
	case "es":
		return localized.Format("2 ene 2006")
	case "fr":
		return localized.Format("2 janv. 2006")
	case "de":
		return localized.Format("2. Jan. 2006")
	default:
		return localized.Format("Jan 2, 2006")
	}
}

func FormatISO(t time.Time, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}
	return t.In(location).Format(time.RFC3339)
}

func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
}
