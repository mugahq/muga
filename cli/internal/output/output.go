package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type ctxKey struct{}

// Opts holds output configuration resolved from flags and environment.
type Opts struct {
	JSON    bool
	NoColor bool
	Verbose bool
	Project string
	Tier    string
	IsTTY   bool
}

// DetectTTY sets IsTTY and applies environment overrides for color.
func (o *Opts) DetectTTY() {
	o.IsTTY = term.IsTerminal(int(os.Stdout.Fd()))

	if os.Getenv("NO_COLOR") != "" || os.Getenv("CI") == "true" {
		o.NoColor = true
	}
}

// WithOpts stores Opts in a context.
func WithOpts(ctx context.Context, opts *Opts) context.Context {
	return context.WithValue(ctx, ctxKey{}, opts)
}

// FromContext retrieves Opts from a context.
func FromContext(ctx context.Context) *Opts {
	if o, ok := ctx.Value(ctxKey{}).(*Opts); ok {
		return o
	}
	return &Opts{}
}

// RenderJSON encodes v as indented JSON to w.
func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

// RenderTable writes an aligned table with uppercase headers, 2-space indent,
// and no borders to w.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	cols := len(headers)
	if cols == 0 {
		return nil
	}

	// Compute column widths from headers and rows.
	widths := make([]int, cols)
	for i, h := range headers {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	// Print header row.
	var b strings.Builder
	b.WriteString("  ")
	for i, h := range headers {
		if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "%-*s", widths[i], strings.ToUpper(h))
	}
	fmt.Fprintln(w, b.String())

	// Print data rows.
	for _, row := range rows {
		b.Reset()
		b.WriteString("  ")
		for i := 0; i < cols; i++ {
			if i > 0 {
				b.WriteString("  ")
			}
			val := ""
			if i < len(row) {
				val = row[i]
			}
			fmt.Fprintf(&b, "%-*s", widths[i], val)
		}
		fmt.Fprintln(w, b.String())
	}

	return nil
}
