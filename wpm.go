package main

import (
	"sync"
	"time"
)

// WPMEngine calculates typing speed using a rolling time window.
// It's thread-safe — the key listener goroutine writes to it,
// and the main loop goroutine reads from it.
type WPMEngine struct {
	mu          sync.Mutex
	timestamps  []int64       // millisecond timestamps of keyDown events
	windowSize  time.Duration // how far back we look (e.g. 5 seconds)
	currentWPM  float64
	everTyped   bool  // has the user typed since startup?
	lastKeyTime int64 // last keystroke time (persists even after window prune)
}

// NewWPMEngine creates a WPM calculator.
// window = how far back to look (5 seconds is the sweet spot).
func NewWPMEngine(window time.Duration) *WPMEngine {
	return &WPMEngine{
		windowSize: window,
		timestamps: make([]int64, 0, 1024),
	}
}

// RecordKeystroke records a key press at the given timestamp.
// Called every time a key is pressed.
func (w *WPMEngine) RecordKeystroke(ts int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Mark that the user has started typing
	w.everTyped = true
	w.lastKeyTime = ts

	// Add this timestamp
	w.timestamps = append(w.timestamps, ts)

	// Remove timestamps older than our window
	cutoff := ts - w.windowSize.Milliseconds()
	i := 0
	for i < len(w.timestamps) && w.timestamps[i] < cutoff {
		i++
	}
	// Keep only recent timestamps (copy to avoid memory leak)
	w.timestamps = append([]int64{}, w.timestamps[i:]...)

	// Recalculate WPM
	w.recalculate()
}

func (w *WPMEngine) recalculate() {
	n := len(w.timestamps)
	if n < 2 {
		w.currentWPM = 0
		return
	}

	// Time between first and last keystroke in the window
	elapsedSec := float64(w.timestamps[n-1]-w.timestamps[0]) / 1000.0
	if elapsedSec < 0.1 {
		return // too short to measure accurately
	}

	// Standard: 1 word = 5 characters
	chars := float64(n)
	words := chars / 5.0
	minutes := elapsedSec / 60.0
	w.currentWPM = words / minutes
}

// GetWPM returns the current words-per-minute.
func (w *WPMEngine) GetWPM() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.currentWPM
}

// IdleTime returns seconds since the last keystroke.
// Returns -1 if the user has never pressed a key since startup.
func (w *WPMEngine) IdleTime() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.everTyped {
		return -1 // never typed — signal to stay silent
	}
	return float64(time.Now().UnixMilli()-w.lastKeyTime) / 1000.0
}