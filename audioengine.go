package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// ─── EMBEDDED AUDIO FILES ───
// These directives bake all MP3 files into the binary at compile time.
// After building, the binary contains all the audio — no external files needed.

//go:embed audio/storm/*.mp3
var stormAudio embed.FS

// Target sample rate for all audio (CD quality)
const targetSampleRate = beep.SampleRate(44100)

// AudioEngine manages all audio playback:
// - 3 rain loops running continuously (volume-controlled)
// - Thunder one-shots fired on demand
type AudioEngine struct {
	mu sync.Mutex

	// The three rain loops — always playing, volume adjusted 20x/sec
	rainVolumes [3]*effects.Volume

	// Pre-loaded thunder sounds for instant playback
	thunderBuffers []*beep.Buffer

	// Timing
	lastThunderAt time.Time
}

// NewAudioEngine initializes the speaker and loads all audio files.
func NewAudioEngine() (*AudioEngine, error) {
	// Initialize the speaker
	// Buffer size = 1/30th of a second = low latency
	err := speaker.Init(targetSampleRate, targetSampleRate.N(time.Second/30))
	if err != nil {
		return nil, fmt.Errorf("speaker init failed: %w", err)
	}

	ae := &AudioEngine{
		lastThunderAt: time.Now().Add(-10 * time.Second),
	}

	// ── LOAD RAIN LOOPS ──
	rainFiles := []string{
		"audio/storm/rain_light.mp3",
		"audio/storm/rain_medium.mp3",
		"audio/storm/rain_heavy.mp3",
	}

	for i, path := range rainFiles {
		vol, err := ae.loadLoop(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", path, err)
		}
		ae.rainVolumes[i] = vol
		// Send to speaker — starts playing immediately (but is muted)
		speaker.Play(vol)
	}

	// ── LOAD THUNDER ONE-SHOTS ──
	ae.thunderBuffers = ae.loadAllBuffers("thunder")

	fmt.Printf("  Audio loaded: 3 rain loops, %d thunder sounds\n", len(ae.thunderBuffers))

	return ae, nil
}

// loadLoop loads an MP3, creates an infinite loop, wraps it in volume control.
// Returns muted — caller sends it to the speaker.
func (ae *AudioEngine) loadLoop(path string) (*effects.Volume, error) {
	data, err := stormAudio.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read %s: %w", path, err)
	}

	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return nil, fmt.Errorf("can't decode %s: %w", path, err)
	}

	// Load into memory buffer
	buf := beep.NewBuffer(format)
	buf.Append(streamer)
	streamer.Close()

	// Create infinite loop
	loopStreamer := beep.Loop(-1, buf.Streamer(0, buf.Len()))

	// Resample if needed (in case MP3 isn't exactly 44100 Hz)
	var finalStreamer beep.Streamer = loopStreamer
	if format.SampleRate != targetSampleRate {
		finalStreamer = beep.Resample(2, format.SampleRate, targetSampleRate, loopStreamer)
	}

	// Wrap in volume control, starting silent
	vol := &effects.Volume{
		Streamer: finalStreamer,
		Base:     2,
		Volume:   -10, // very quiet
		Silent:   true, // start muted
	}

	return vol, nil
}

// loadAllBuffers loads all MP3 files from audio/storm/ matching the prefix.
func (ae *AudioEngine) loadAllBuffers(prefix string) []*beep.Buffer {
	entries, err := stormAudio.ReadDir("audio/storm")
	if err != nil {
		return nil
	}

	var buffers []*beep.Buffer
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".mp3") {
			continue
		}

		path := "audio/storm/" + name
		data, err := stormAudio.ReadFile(path)
		if err != nil {
			continue
		}

		streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
		if err != nil {
			continue
		}

		buf := beep.NewBuffer(format)
		buf.Append(streamer)
		streamer.Close()
		_ = format // we handle resampling at playback if needed

		buffers = append(buffers, buf)
	}

	return buffers
}

// Update is called 20 times per second from the main loop.
// It adjusts rain volumes and maybe triggers thunder.
func (ae *AudioEngine) Update(level StormLevel) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// ── ADJUST RAIN VOLUMES ──
	// speaker.Lock ensures we don't modify audio mid-sample
	speaker.Lock()
	for i := 0; i < 3; i++ {
		targetVol := level.RainVolume[i]
		if targetVol < 0.01 {
			ae.rainVolumes[i].Silent = true
		} else {
			ae.rainVolumes[i].Silent = false
			// Convert linear 0-1 to logarithmic (decibels)
			// Human hearing is logarithmic:
			//   1.0 → 0 dB (full), 0.5 → -1.5 dB, 0.1 → -5 dB
			ae.rainVolumes[i].Volume = math.Log2(targetVol) * 1.5
		}
	}
	speaker.Unlock()

	// ── MAYBE FIRE THUNDER ──
	// Minimum 2-second gap between thunder sounds
	if level.CanThunder && time.Since(ae.lastThunderAt) > 2*time.Second {
		if rand.Float64() < level.ThunderProb {
			ae.fireThunder()
		}
	}
}

func (ae *AudioEngine) fireThunder() {
	ae.lastThunderAt = time.Now()

	if len(ae.thunderBuffers) == 0 {
		return
	}

	// Pick a random thunder sound and play it
	// This layers on top of rain — the speaker mixes automatically
	buf := ae.thunderBuffers[rand.Intn(len(ae.thunderBuffers))]
	speaker.Play(buf.Streamer(0, buf.Len()))
}
