// keylistener.c
// Talks to macOS CGEventTap to observe global keyboard events.
// This is passive (listen-only) — it never blocks or modifies keystrokes.
// Requires Input Monitoring permission on macOS 10.15+.

#include "keylistener.h"
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static KeyCallback globalCallback = NULL;

// Called by macOS for every keyboard event
static CGEventRef eventCallback(
    CGEventTapProxy proxy,
    CGEventType type,
    CGEventRef event,
    void *refcon
) {
    // macOS disables taps that are slow — re-enable if that happens
    if (type == kCGEventTapDisabledByTimeout) {
        CGEventTapEnable(eventTap, true);
        return event;
    }

    // Only process key down and key up events
    if (type == kCGEventKeyDown || type == kCGEventKeyUp) {
        CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(
            event, kCGKeyboardEventKeycode
        );
        if (globalCallback) {
            globalCallback((int)keycode, (int)type);
        }
    }

    // Pass the event through — we're a listener, not a blocker
    return event;
}

void StartKeyListener(KeyCallback callback) {
    globalCallback = callback;

    // We want to hear about keyDown and keyUp
    CGEventMask eventMask = (
        CGEventMaskBit(kCGEventKeyDown) |
        CGEventMaskBit(kCGEventKeyUp)
    );

    // Create the event tap
    eventTap = CGEventTapCreate(
        kCGSessionEventTap,            // tap the whole user session
        kCGHeadInsertEventTap,         // insert at head of chain
        kCGEventTapOptionListenOnly,   // PASSIVE — observe only
        eventMask,
        eventCallback,
        NULL
    );

    if (!eventTap) {
        fprintf(stderr,
            "\n"
            "  ERROR: Cannot create event tap.\n"
            "\n"
            "  thundertype needs Input Monitoring permission.\n"
            "  Grant it in:\n"
            "    System Settings > Privacy & Security > Input Monitoring\n"
            "\n"
            "  Add your terminal app (Terminal.app or iTerm) to the list,\n"
            "  then restart thundertype.\n"
            "\n"
        );
        return;
    }

    // Wire the tap into the macOS run loop
    runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault, eventTap, 0
    );
    CFRunLoopAddSource(
        CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes
    );
    CGEventTapEnable(eventTap, true);

    // This blocks forever — that's why Go runs it in a goroutine
    CFRunLoopRun();
}

void StopKeyListener(void) {
    if (eventTap) {
        CGEventTapEnable(eventTap, false);
        CFRunLoopStop(CFRunLoopGetCurrent());
    }
}
