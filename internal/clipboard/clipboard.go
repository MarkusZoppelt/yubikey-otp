package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// WriteAll copies text to the system clipboard using available utilities
func WriteAll(text string) error {
	switch runtime.GOOS {
	case "linux":
		return writeLinux(text)
	case "darwin":
		return writeDarwin(text)
	case "windows":
		return writeWindows(text)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func writeLinux(text string) error {
	// Try clipboard utilities in order of preference
	utilities := [][]string{
		{"wl-copy"},                    // Wayland
		{"xclip", "-selection", "clipboard"}, // X11
		{"xsel", "--clipboard", "--input"},   // X11 alternative
	}

	for _, util := range utilities {
		if _, err := exec.LookPath(util[0]); err == nil {
			cmd := exec.Command(util[0], util[1:]...)
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no clipboard utilities available. Please install xsel, xclip, or wl-clipboard")
}

func writeDarwin(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func writeWindows(text string) error {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}