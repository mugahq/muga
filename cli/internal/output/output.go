package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

type ctxKey struct{}

// Opts holds output configuration resolved from flags and environment.
type Opts struct {
	JSON    bool
	NoColor bool
	Verbose bool
	Project string
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

// RenderTable writes an aligned table with headers and rows to w.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}

	table := tablewriter.NewTable(w,
		tablewriter.WithHeader(upper),
	)

	for _, row := range rows {
		cells := make([]any, len(row))
		for i, c := range row {
			cells[i] = c
		}
		if err := table.Append(cells...); err != nil {
			return fmt.Errorf("appending row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("rendering table: %w", err)
	}
	return nil
}
