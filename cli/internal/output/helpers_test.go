package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"zero", time.Time{}, "never"},
		{"seconds ago", time.Now().Add(-30 * time.Second), "30s ago"},
		{"minutes ago", time.Now().Add(-5 * time.Minute), "5 min ago"},
		{"hours ago", time.Now().Add(-3 * time.Hour), "3h ago"},
		{"days ago", time.Now().Add(-48 * time.Hour), "2d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HumanDuration(tt.t)
			if got != tt.want {
				t.Errorf("HumanDuration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHumanInterval(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m"},
		{300, "5m"},
		{3600, "1h"},
		{86400, "24h"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := HumanInterval(tt.seconds)
			if got != tt.want {
				t.Errorf("HumanInterval(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestTruncateUUID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "550e8400"},
		{"abcd", "abcd"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TruncateUUID(tt.input)
			if got != tt.want {
				t.Errorf("TruncateUUID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderDetail(t *testing.T) {
	var buf bytes.Buffer
	rows := []DetailRow{
		{Key: "User", Value: "Alberto Bajo"},
		{Key: "Email", Value: "alberto@muga.sh"},
		{Key: "Tier", Value: "pro"},
	}

	if err := RenderDetail(&buf, rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")

	// Should have 2-space indent on every non-empty line.
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("expected 2-space indent, got %q", line)
		}
	}

	// Should contain all values.
	if !strings.Contains(out, "Alberto Bajo") {
		t.Error("expected 'Alberto Bajo' in output")
	}
	if !strings.Contains(out, "alberto@muga.sh") {
		t.Error("expected 'alberto@muga.sh' in output")
	}
	if !strings.Contains(out, "pro") {
		t.Error("expected 'pro' in output")
	}
}
