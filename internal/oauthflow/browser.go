package oauthflow

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type BrowserOpener func(rawURL string) error

func OpenBrowser(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("login URL is empty")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	case "darwin":
		cmd = exec.Command("open", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}
