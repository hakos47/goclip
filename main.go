package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.design/x/clipboard"
)

const (
	MaxHistoryItems = 20
	PreviewLength   = 60
)

func main() {
	daemonMode := flag.Bool("daemon", false, "Run in background watch mode")
	flag.Parse()

	// Initialize history manager
	history, err := NewHistory(MaxHistoryItems)
	if err != nil {
		log.Fatalf("Failed to initialize history: %v", err)
	}

	// Initialize system clipboard access
	if err := clipboard.Init(); err != nil {
		// Log warning but proceed, as UI mode might not need the watcher init strictly depending on platform,
		// though strictly speaking for 'clipboard' package it is needed.
		log.Printf("Warning: Clipboard init failed: %v", err)
	}

	if *daemonMode {
		runDaemon(history)
	} else {
		runUI(history)
	}
}

// runDaemon starts the background process to monitor clipboard changes.
func runDaemon(h *History) {
	fmt.Println("Starting Clipboard Manager Daemon...")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down daemon...")
		cancel()
	}()

	// Watch channels
	textCh := clipboard.Watch(ctx, clipboard.FmtText)
	imgCh := clipboard.Watch(ctx, clipboard.FmtImage)

	fmt.Println("Listening for clipboard events...")

	for {
		select {
		case <-ctx.Done():
			return

		case data := <-textCh:
			content := string(data)
			if strings.TrimSpace(content) == "" {
				continue
			}

			// Generate Preview
			preview := strings.ReplaceAll(content, "\n", " ")
			if len(preview) > PreviewLength {
				preview = preview[:PreviewLength] + "..."
			}

			err := h.Add(Item{
				Type:      TypeText,
				Content:   content,
				Preview:   preview,
				Timestamp: time.Now(),
			})
			if err != nil {
				log.Printf("Error adding text item: %v", err)
			} else {
				log.Println("Captured Text")
			}

		case data := <-imgCh:
			path, err := SaveImage(data)
			if err != nil {
				log.Printf("Error saving image: %v", err)
				continue
			}

			err = h.Add(Item{
				Type:      TypeImage,
				Content:   path,
				Preview:   "[Image] " + filepath.Base(path),
				Timestamp: time.Now(),
			})
			if err != nil {
				log.Printf("Error adding image item: %v", err)
			} else {
				log.Println("Captured Image")
			}
		}
	}
}

// runUI triggers the visual menu.
func runUI(h *History) {
	items := h.GetItems()
	if len(items) == 0 {
		return
	}

	ui := NewUIHandler()
	
	// Show menu and get selection
	index, selected := ui.ShowMenu(items)
	if !selected {
		return
	}

	// Execute paste
	item := items[index]
	if err := ui.PasteItem(item); err != nil {
		log.Printf("Error pasting item: %v", err)
	}
}
