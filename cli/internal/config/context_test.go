package config

import (
	"context"
	"testing"
)

func TestWithConfigRoundTrip(t *testing.T) {
	cfg := &Config{APIURL: "https://custom.api", Project: "test"}
	ctx := WithConfig(context.Background(), cfg)

	got := FromContext(ctx)
	if got.APIURL != "https://custom.api" {
		t.Errorf("APIURL = %q, want https://custom.api", got.APIURL)
	}
	if got.Project != "test" {
		t.Errorf("Project = %q, want test", got.Project)
	}
}

func TestFromContextDefault(t *testing.T) {
	got := FromContext(context.Background())
	if got.APIURL != "https://api.muga.sh" {
		t.Errorf("APIURL = %q, want default", got.APIURL)
	}
}
