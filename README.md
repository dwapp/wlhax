# wlhax

> Note: I forked this project for personal use, and may add non-standard protocol support in the future. The original project was https://git.sr.ht/~kennylevinsen/wlhax.

Fork of https://git.sr.ht/~sircmpwn/wlhax

Wayland proxy that monitors and displays various application state, such as the current surface tree, in a nice little TUI.

Useful for debugging.


## How to build

```
go build
```

## How to use

1. Start wlhax (`./wlhax`). If you want to proxy an alternate wayland display, specify WAYLAND_DISPLAY when starting wlhax.
2. Follow the instructions, using either `:exec` to start an application, or starting an application externally while specifying `WAYLAND_DISPLAY=wlhax-0` in the environment.
3. Switch tabs to select the different availble clients by using the left/right arrows.
