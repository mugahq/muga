package output

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDetectTTYNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	opts := &Opts{}
	opts.DetectTTY()

	if !opts.NoColor {
		t.Error("expected NoColor=true when NO_COLOR is set")
	}
}

func TestDetectTTYCI(t *testing.T) {
	t.Setenv("CI", "true")

	opts := &Opts{}
	opts.DetectTTY()

	if !opts.NoColor {
		t.Error("expected NoColor=true when CI=true")
	}
}

func TestContextRoundTrip(t *testing.T) {
	opts := &Opts{JSON: true, Project: "test"}
	ctx := WithOpts(context.Background(), opts)

	got := FromContext(ctx)
	if got != opts {
		t.Error("expected same opts from context")
	}
}

func TestFromContextDefault(t *testing.T) {
	got := FromContext(context.Background())
	if got == nil {
		t.Fatal("expected non-nil default opts")
	}
	if got.JSON {
		t.Error("expected JSON=false in default opts")
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "muga", "status": "ok"}

	if err := RenderJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["name"] != "muga" {
		t.Errorf("expected name=muga, got %q", result["name"])
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"api", "up"},
		{"web", "down"},
	}

	if err := RenderTable(&buf, headers, rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected uppercased header NAME in output")
	}
	if !strings.Contains(out, "api") {
		t.Error("expected row value 'api' in output")
	}
	if !strings.Contains(out, "down") {
		t.Error("expected row value 'down' in output")
	}
}

func TestRenderTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"Name"}
	rows := [][]string{}

	if err := RenderTable(&buf, headers, rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderJSONNested(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"monitors": []map[string]string{
			{"name": "api", "status": "up"},
		},
	}

	if err := RenderJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Error("expected valid JSON output")
	}
}

func TestDetectTTYDefault(t *testing.T) {
	// Without NO_COLOR or CI set, NoColor should remain as initialized.
	t.Setenv("NO_COLOR", "")
	t.Setenv("CI", "")

	opts := &Opts{}
	opts.DetectTTY()

	// In test, stdout is not a TTY.
	if opts.IsTTY {
		t.Error("expected IsTTY=false in test environment")
	}
	if opts.NoColor {
		t.Error("expected NoColor=false when NO_COLOR and CI are not set")
	}
}
