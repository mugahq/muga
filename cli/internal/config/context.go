package config

import "context"

type ctxKey struct{}

// WithConfig stores a Config in a context.
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxKey{}, cfg)
}

// FromContext retrieves a Config from a context.
func FromContext(ctx context.Context) *Config {
	if c, ok := ctx.Value(ctxKey{}).(*Config); ok {
		return c
	}
	return &Config{APIURL: defaultAPIURL}
}
