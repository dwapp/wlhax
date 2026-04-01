# wlhax

Wayland proxy and terminal dashboard for inspecting clients, protocol objects, and surface state in real time.

This fork is based on:

- https://git.sr.ht/~sircmpwn/wlhax
- https://git.sr.ht/~kennylevinsen/wlhax

## What It Does

- Proxies a Wayland display socket
- Forwards protocol traffic between clients and the real compositor
- Tracks protocol objects, globals, buffers, and surfaces per client
- Renders the current state in a TUI built with `vaxis`

## Quick Start

```bash
go build
WAYLAND_DISPLAY=wayland-0 ./wlhax
```

Then start a client against the proxy socket shown in the status bar, or launch one from inside `wlhax` with `:exec <command>`.

## Controls

- `Left` / `Right`, `h` / `l`: switch tabs
- `Up` / `Down`, `j` / `k`: move selection
- `Space`: fold or unfold the current category
- `:`: command mode
- `:exec <command>`: launch a client
- `:slow`, `:fast`, `:block`, `:unblock`, `:clear`, `:quit`

## Documentation

- Technical design: [docs/technical.md](docs/technical.md)
