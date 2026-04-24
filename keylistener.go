package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation

#include "keylistener.h"

extern void goKeyCallback(int keycode, int eventType);

static void bridgeCallback(int keycode, int eventType) {
    goKeyCallback(keycode, eventType);
}

static void startListenerBridge() {
    StartKeyListener(bridgeCallback);
}
*/
import "C"

import (
	"time"
)

// macOS virtual keycodes for special keys
const (
	KeycodeBackspace  = 51
	KeycodeReturn     = 36
	KeycodeSpace      = 49
	KeycodeEscape     = 53
	KeycodeTab        = 48
	KeycodeForwardDel = 117
)

// KeyEvent represents a single keystroke
type KeyEvent struct {
	Keycode   int   // Which key was pressed (macOS virtual keycode)
	EventType int   // 10 = keyDown, 11 = keyUp
	Timestamp int64 // When it happened (milliseconds since epoch)
}

// Channel where key events are sent from C to Go
var keyChannel chan KeyEvent

// goKeyCallback is called from C every time a key is pressed/released.
// The //export comment makes it visible to C code.
//
//export goKeyCallback
func goKeyCallback(keycode C.int, eventType C.int) {
	if keyChannel != nil {
		// Non-blocking send — if the channel is full, drop the event
		select {
		case keyChannel <- KeyEvent{
			Keycode:   int(keycode),
			EventType: int(eventType),
			Timestamp: time.Now().UnixMilli(),
		}:
		default:
			// Channel full, drop this event (rare, not a problem)
		}
	}
}

// StartListening begins capturing keystrokes.
// Returns a channel that receives KeyEvent values.
// The listener runs in a background goroutine.
func StartListening() <-chan KeyEvent {
	keyChannel = make(chan KeyEvent, 256)

	// Start the C event loop in the background.
	// CFRunLoopRun() blocks forever, so it must be in a goroutine.
	go func() {
		C.startListenerBridge()
	}()

	return keyChannel
}

// StopListening stops the event tap and cleans up.
func StopListening() {
	C.StopKeyListener()
}
