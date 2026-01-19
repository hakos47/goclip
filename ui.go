package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UIHandler manages interactions with the user interface and system clipboard.
type UIHandler struct {
	ThemePath string
}

// NewUIHandler creates a handler, trying to locate the user's preferred theme.
func NewUIHandler() *UIHandler {
	// Priority: User's specific Docker theme -> Standard Rofi config -> Default
	home, _ := os.UserHomeDir()
	preferredTheme := filepath.Join(home, ".config/polybar/shapes/scripts/rofi/dockermenu.rasi")

	finalTheme := ""
	if _, err := os.Stat(preferredTheme); err == nil {
		finalTheme = preferredTheme
	}

	return &UIHandler{
		ThemePath: finalTheme,
	}
}

// ShowMenu displays the history in Rofi and returns the selected item index.
func (u *UIHandler) ShowMenu(items []Item) (int, bool) {
	if len(items) == 0 {
		return 0, false
	}

	var rofiInput strings.Builder
	for _, item := range items {
		// Clean text for display
		rofiInput.WriteString(item.Preview)
		
		if item.Type == TypeImage {
			rofiInput.WriteString("\x00icon\x1f" + item.Content)
		}
		rofiInput.WriteString("\n")
	}

	args := []string{
		"-dmenu",
		"-i",                // Case insensitive
		"-p", "Clipboard",   // Prompt
		"-format", "i",      // Output index (0-based)
		"-show-icons",
		"-mesg", "<b>Historial</b>",
	}

	if u.ThemePath != "" {
		args = append(args, "-theme", u.ThemePath)
	}

	cmd := exec.Command("rofi", args...)
	cmd.Stdin = strings.NewReader(rofiInput.String())

	output, err := cmd.Output()
	if err != nil {
		// Exit code 1 usually means user cancelled (ESC)
		return 0, false
	}

	indexStr := strings.TrimSpace(string(output))
	var index int
	_, err = fmt.Sscanf(indexStr, "%d", &index)
	if err != nil || index < 0 || index >= len(items) {
		return 0, false
	}

	return index, true
}

// PasteItem puts the item into the system clipboard and simulates Ctrl+V.
func (u *UIHandler) PasteItem(item Item) error {
	// 1. Load content into Clipboard (Persistent via xclip)
	var cmd *exec.Cmd

	if item.Type == TypeText {
		// Pipe text to xclip
		cmd = exec.Command("xclip", "-selection", "clipboard", "-in")
		cmd.Stdin = strings.NewReader(item.Content)
	} else {
		// Pipe image file path to xclip as image/png
		cmd = exec.Command("xclip", "-selection", "clipboard", "-t", "image/png", "-in", item.Content)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xclip failed: %s, stderr: %s", err, stderr.String())
	}

	// 2. Simulate Paste (Ctrl+V)
	// Add a small delay to ensure focus returns to the target window
	time.Sleep(150 * time.Millisecond)
	
	pasteCmd := exec.Command("xdotool", "key", "ctrl+v")
	if err := pasteCmd.Run(); err != nil {
		return fmt.Errorf("xdotool failed: %w", err)
	}

	return nil
}
