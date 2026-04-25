# ⛈ Thundertype

**Your keyboard sounds like a thunderstorm.**

Type slowly — gentle rain. Type faster — the storm builds. Stop typing — silence returns. Thunder cracks at peak speed.

## Install

```bash
# Clone the repo
git clone https://github.com/YOUR_GITHUB_USERNAME/thundertype.git
cd thundertype

# Download dependencies
go mod tidy

# Build and run
make run
```

### Requirements

- **macOS** (uses CGEventTap for keystroke detection)
- **Go 1.22+** (`brew install go`)
- **Xcode CLI tools** (`xcode-select --install`)

### Permissions

On first run, macOS will ask for **Input Monitoring** permission. Grant it in:

> System Settings → Privacy & Security → Input Monitoring

Add your terminal app (Terminal.app or iTerm) to the list, then restart thundertype.

## Usage

After building with `make run`, start typing anywhere on your Mac — in your editor, browser, terminal, anywhere. The storm responds to your typing speed globally.

To run it again later:

```bash
cd thundertype
make run
```

**Ctrl+C** to quit.

## How It Works

1. **Key Listener** — Passively observes global keystrokes via macOS CGEventTap
2. **WPM Engine** — Calculates typing speed using a 5-second rolling window
3. **Storm Mapper** — Converts WPM to storm intensity (rain volumes + thunder probability)
4. **Audio Engine** — Crossfades 3 rain loops and layers thunder one-shots

| Typing Speed | Experience |
|---|---|
| Idle (4s+) | Silence |
| 1-20 WPM | Light rain patter |
| 20-40 WPM | Rain picks up |
| 40-60 WPM | Medium rain, wind |
| 60-90 WPM | Heavy rain, thunder starts |
| 90-120+ WPM | Full storm |

## Audio

Place your MP3 files in `audio/storm/`:

```
audio/storm/
├── rain_light.mp3     ← gentle rain loop
├── rain_medium.mp3    ← steady rain loop
├── rain_heavy.mp3     ← heavy rain loop
├── thunder_01.mp3     ← distant thunder
├── thunder_02.mp3     ← distant thunder
├── thunder_03.mp3     ← heavy thunder
└── thunder_04.mp3     ← heavy thunder
```

Rain files loop seamlessly. Thunder files play as one-shots at high typing speeds.

## License

MIT