// keylistener.h
// Declares the C functions that talk to macOS CGEventTap.
// Go calls these through cgo.

#ifndef KEYLISTENER_H
#define KEYLISTENER_H

#include <CoreGraphics/CoreGraphics.h>

// Callback type: called with (keycode, eventType) on each key event
typedef void (*KeyCallback)(int keycode, int eventType);

// Start listening for global keystrokes (blocks — run in a goroutine)
void StartKeyListener(KeyCallback callback);

// Stop the listener and exit the run loop
void StopKeyListener(void);

#endif
