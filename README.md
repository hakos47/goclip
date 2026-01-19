# Goclip (Clipboard Manager)

A lightweight, persistent clipboard manager written in Go, designed specifically for Linux environments using BSPWM and Rofi. It supports text and image history with visual previews.

## Features

- **Multimedia Support:** Captures both text and images.
- **Persistence:** History is saved to disk (`~/.config/goclip/`), surviving reboots.
- **Visual Previews:** Shows text snippets and image thumbnails in Rofi.
- **Native Integration:** Automatically detects and applies your existing Rofi themes.
- **Auto-Paste:** Automatically pastes the selected item using `xdotool`.

## Requirements

Ensure you have the following installed on your system:

- **Go** (1.20+)
- **Rofi** (Menu interface)
- **xclip** (Clipboard manipulation)
- **xdotool** (Auto-paste simulation)
- **libx11-dev** (For Go clipboard bindings)

```bash
# Ubuntu/Debian/Kali
sudo apt install rofi xclip xdotool libx11-dev
```

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/yourusername/goclip.git
    cd goclip
    ```

2.  **Build:**
    ```bash
    go mod tidy
    go build -o goclip
    ```

3.  **Install:**
    Move the binary to a location in your PATH (optional but recommended):
    ```bash
    mkdir -p ~/bin
    mv goclip ~/bin/
    ```

## Configuration

### 1. Autostart (Daemon)
Add the following to your `~/.config/bspwm/bspwmrc` to start the watcher in the background:

```bash
# Start Goclip Daemon
~/bin/goclip -daemon &
```

### 2. Keyboard Shortcut
Add the trigger to `~/.config/sxhkd/sxhkdrc` (e.g., `Super + V`):

```bash
# Goclip Manager
super + v
    ~/bin/goclip
```

*Don't forget to reload sxhkd (`pkill -USR1 -x sxhkd`).*

## License

MIT License.