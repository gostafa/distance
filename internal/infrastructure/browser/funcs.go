// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func defaultBrowserRuntime() browserRuntime {
	return browserRuntime{
		startCommand: func(cmd *exec.Cmd) error { return cmd.Start() },
	}
}

// Open launches the platform browser for path.
func Open(path string) error {
	err := defaultBrowserRuntime().open(path)
	if err != nil {
		return fmt.Errorf("browser Open: %w", err)
	}

	return nil
}

func (browser browserRuntime) open(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)

	err := startNamed(browser, command(name, args))
	if err != nil {
		return fmt.Errorf("open %s in browser: %w", path, err)
	}

	return nil
}

func (browser browserRuntime) start(cmd *exec.Cmd) error {
	err := browser.startCommand(cmd)
	if err != nil {
		return fmt.Errorf("browser start: %w", err)
	}

	return nil
}

func startNamed(starter commandStarter, cmd *exec.Cmd) error {
	err := starter.start(cmd)
	if err != nil {
		return fmt.Errorf("browser startNamed: %w", err)
	}

	return nil
}

func command(name string, args []string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}

	resolved, lookErr := exec.LookPath(name)
	if lookErr == nil {
		cmd.Path = resolved
	}

	return cmd
}

// OpenCommand returns the OS-specific browser launcher name and arguments.
func OpenCommand(goos, path string) (name string, args []string) {
	switch goos {
	case "darwin":
		return "open", []string{path}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default:
		return "xdg-open", []string{path}
	}
}
