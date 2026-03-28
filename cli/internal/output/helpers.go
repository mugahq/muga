package output

import (
	"fmt"
	"io"
	"math"
	"time"
)

// DetailRow represents a key-value pair for detail views.
type DetailRow struct {
	Key   string
	Value string
}

// RenderDetail writes key-value pairs with 2-space indent and fixed-width keys.
func RenderDetail(w io.Writer, rows []DetailRow) error {
	keyWidth := 0
	for _, r := range rows {
		if len(r.Key) > keyWidth {
			keyWidth = len(r.Key)
		}
	}

	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "  %-*s  %s\n", keyWidth, r.Key, r.Value); err != nil {
			return fmt.Errorf("writing detail row: %w", err)
		}
	}
	return nil
}

// HumanDuration formats the time elapsed since t as a human-readable string.
// Returns "never" for zero time.
func HumanDuration(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)
	if d < 0 {
		d = -d
	}

	switch {
	case d < time.Minute:
		s := int(math.Round(d.Seconds()))
		if s <= 0 {
			s = 1
		}
		return fmt.Sprintf("%ds ago", s)
	case d < time.Hour:
		m := int(math.Round(d.Minutes()))
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(math.Round(d.Hours()))
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(math.Round(d.Hours() / 24))
		return fmt.Sprintf("%dd ago", days)
	}
}

// HumanInterval formats seconds into a human-readable interval.
func HumanInterval(seconds int) string {
	switch {
	case seconds <= 0:
		return "0s"
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		m := seconds / 60
		return fmt.Sprintf("%dm", m)
	default:
		h := seconds / 3600
		return fmt.Sprintf("%dh", h)
	}
}

// TruncateUUID returns the first 8 characters of a UUID string.
func TruncateUUID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
