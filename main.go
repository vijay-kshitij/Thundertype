package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Version is set at build time via ldflags
var version = "dev"

func main() {
	fmt.Printf("\n")
	fmt.Printf("  ⛈  thundertype %s\n", version)
	fmt.Printf("  Your keyboard sounds like a thunderstorm.\n")
	fmt.Printf("\n")

	// ── STEP 1: Start the key listener ──
	fmt.Println("  Starting key listener...")
	keys := StartListening()
	// Runs in background goroutine. If this fails, the C code
	// prints an error about Input Monitoring permission.

	// Give the event tap a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// ── STEP 2: Create the WPM calculator ──
	// 5-second rolling window — responsive but not jittery
	wpmEngine := NewWPMEngine(5 * time.Second)

	// ── STEP 3: Create the audio engine ──
	fmt.Println("  Loading audio...")
	audioEngine, err := NewAudioEngine()
	if err != nil {
		fmt.Printf("\n  ERROR: %v\n\n", err)
		os.Exit(1)
	}

	// ── STEP 4: Set up the main loop ──
	// Ticker fires every 50ms (20 times per second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	// Catch Ctrl+C for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println()
	fmt.Println("  Listening... type anywhere and hear the storm.")
	fmt.Println("  Press Ctrl+C to quit.")
	fmt.Println()

	var backspaceCount int

	// ── STEP 5: The main event loop ──
	// "select" waits for whichever event happens first:
	// - a key was pressed
	// - 50ms passed (time to update audio)
	// - user pressed Ctrl+C
	for {
		select {

		case key := <-keys:
			// A key event came through the channel
			if key.EventType == 10 { // 10 = keyDown
				// Record for WPM calculation
				wpmEngine.RecordKeystroke(key.Timestamp)

				// Handle backspace comedy
				if key.Keycode == KeycodeBackspace {
					backspaceCount++
					// You can add comedy SFX here later:
					// if backspaceCount >= 5 { playComedySound() }
				} else {
					backspaceCount = 0
				}
			}

		case <-ticker.C:
			// 50ms passed — update audio based on current typing speed
			wpm := wpmEngine.GetWPM()
			idle := wpmEngine.IdleTime()
			level := MapWPMToStorm(wpm, idle)
			audioEngine.Update(level)

		case <-sigCh:
			// User pressed Ctrl+C
			fmt.Println("\n  Storm subsiding...")
			StopListening()
			// Brief pause so the last sounds can fade
			time.Sleep(300 * time.Millisecond)
			fmt.Println("  Goodbye.")
			return
		}
	}
}
