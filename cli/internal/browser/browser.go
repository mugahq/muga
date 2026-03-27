package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Open opens the given URL in the user's default browser.
//
// If the BROWSER environment variable is set to an empty string, the browser
// is not opened. This follows the convention used by gh, npm, and other CLI
// tools to allow headless / CI usage.
func Open(url string) error {
	if val, ok := os.LookupEnv("BROWSER"); ok && val == "" {
		return nil
	}
	cmd, args, err := command(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return exec.Command(cmd, args...).Start()
}

func command(goos, url string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "open", []string{url}, nil
	case "linux":
		return "xdg-open", []string{url}, nil
	case "windows":
		return "cmd", []string{"/c", "start", url}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}
