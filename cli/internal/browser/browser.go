package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open opens the given URL in the user's default browser.
// Returns an error if the browser could not be launched.
func Open(url string) error {
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
