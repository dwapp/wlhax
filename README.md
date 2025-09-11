# wlhax

> Note: I forked this project for personal use, and may add non-standard protocol support in the future. The original project was https://git.sr.ht/~kennylevinsen/wlhax.

Fork of https://git.sr.ht/~sircmpwn/wlhax

Wayland proxy that monitors and displays various application state, such as the current surface tree, in a nice little TUI.

Useful for debugging Wayland applications and protocols.

## Features

- **Real-time Wayland protocol monitoring**: Intercepts and displays all Wayland messages
- **Interactive TUI**: Navigate through connected clients and their protocol state
- **Surface tree visualization**: Shows the hierarchy of Wayland surfaces
- **Protocol object tracking**: Monitor globals, surfaces, seats, keyboards, pointers, and more
- **Modern UI**: Built with pure [vaxis](https://git.sr.ht/~rockorager/vaxis) for fast terminal rendering

## How to build

```bash
go build
```

## How to use

1. Start wlhax (`./wlhax`). If you want to proxy an alternate wayland display, specify WAYLAND_DISPLAY when starting wlhax.
2. Follow the instructions, using either `:exec` to start an application, or starting an application externally while specifying `WAYLAND_DISPLAY=wlhax-0` in the environment.
3. Switch tabs to select the different available clients by using the left/right arrows.

### Keyboard shortcuts

- **Arrow keys / h,j,k,l**: Navigate through UI elements
- **Tab navigation**: Left/Right arrows or h/l to switch between client tabs
- **Space**: Toggle folding of protocol categories
- **:**: Enter command mode
- **:exec `<command>`**: Execute a new Wayland client
- **:quit**: Exit wlhax

### Protocol monitoring

wlhax intercepts all Wayland protocol messages between clients and the compositor, allowing you to:

- Monitor surface creation and destruction
- Track input events (keyboard, mouse, touch)
- Observe buffer management and damage tracking
- Debug protocol extensions and custom protocols
- Analyze performance issues in Wayland applications

## Technical details

### Architecture

- **Proxy server**: Creates a Unix socket that clients connect to
- **Message forwarding**: Transparently forwards messages to the real compositor
- **Protocol parsing**: Decodes Wayland messages for display
- **TUI interface**: Real-time visualization using vaxis terminal library

### Dependencies

- [vaxis](https://git.sr.ht/~rockorager/vaxis) - Modern terminal UI library
- Standard Go libraries for Wayland protocol handling
