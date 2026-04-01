# Technical Notes

## Overview

`wlhax` is a transparent Wayland proxy with an attached terminal UI. It accepts client connections on a local Unix socket, opens a second connection to the real compositor, forwards packets in both directions, and updates an in-memory model that the dashboard renders.

The codebase is small, but it has three distinct layers:

1. Bootstrap and process lifecycle
2. Proxying and protocol state tracking
3. Terminal UI rendering

## Runtime Flow

### Startup

`main.go` is the entry point.

- Reads `WAYLAND_DISPLAY` to find the upstream compositor socket
- Tries to bind a proxy display named `wlhax-0` through `wlhax-9`
- Sets the process `WAYLAND_DISPLAY` to the proxy socket name in `NewProxy`
- Starts the proxy accept loop in a goroutine
- Builds the dashboard and enters the UI event loop

If a command is passed on the CLI, `main.go` launches it after the proxy is created.

### Connection Lifecycle

`Proxy.Run` accepts Unix socket connections from Wayland clients. For each client, `handleClient`:

- Creates a `Client` model with `wl_display@1`
- Registers protocol handlers into `client.Impls`
- Connects to the real Wayland display
- Starts one goroutine for compositor-to-client traffic
- Starts one goroutine for client-to-compositor traffic

Each direction reads a `WaylandPacket`, records it into client state, notifies the dashboard, and forwards the packet to the other socket.

## Core Types

### Proxy

Defined in `proxy.go`.

Responsibilities:

- Own the listening socket
- Track connected clients
- Expose coarse runtime controls such as `SlowMode`, `Block`, and `CloseWrite`
- Fan out UI callbacks through `OnUpdate`, `OnConnect`, and `OnDisconnect`

### Client

Also defined in `proxy.go`.

Responsibilities:

- Hold the pair of Unix socket connections
- Store packet logs: `RxLog` and `TxLog`
- Track live protocol objects in `Objects` and `ObjectMap`
- Track globals advertised by the compositor in `Globals` and `GlobalMap`
- Dispatch protocol decoding to interface-specific implementations in `Impls`

`Client` is the central state container that backs both protocol tracking and UI rendering.

### WaylandPacket

Defined in `protocol.go`.

Responsibilities:

- Represent a decoded Wayland message header plus payload
- Carry received file descriptors
- Provide typed readers such as `ReadUint32`, `ReadInt32`, `ReadFixed`, and `ReadString`
- Re-encode and forward the message with `WritePacket`

`ReadPacket` uses `ReadMsgUnix` to collect both the header and ancillary FD rights.

## Protocol Model

Protocol support is split by interface, usually one file per Wayland interface or extension:

- `wl_registry.go`
- `wl_surface.go`
- `wl_buffer.go`
- `wl_output.go`
- `wl_pointer.go`
- `xdg_shell.go`
- `zwp_linux_dmabuf.go`
- `wp_fractional_scale.go`
- `wp_cursor_shape_manager.go`

The common pattern is:

1. `Register...` installs an implementation into `client.Impls`
2. `Request` decodes client-to-compositor messages
3. `Event` decodes compositor-to-client messages
4. Optional `Create` builds typed object state when `wl_registry.bind` creates a new object
5. Optional `Destroy` releases references when an object is removed

The object model is incremental rather than authoritative. `wlhax` observes protocol traffic and derives state from the messages it understands. Unsupported interfaces still appear as protocol objects, but only with limited detail.

## Surface Tracking

`wl_surface.go` contains the richest state model in the repository.

It tracks:

- Current and pending surface state
- Attached buffers
- Damage region
- Output membership
- Frame callbacks
- Parent/child relationships for subsurfaces
- Surface roles such as xdg-shell roles

This state is what powers the surface tree shown in the dashboard.

## UI Structure

The UI layer lives in `ui/` and uses `vaxis`.

Important pieces:

- `ui/ui.go`: global event loop integration, redraw queue, suspend/resume
- `dashboard.go`: top-level screen composition and ex-command handling
- `clients.go`: connections overview tab
- `client.go`: per-client object/category browser
- `exline.go`: command line widget used for `:` commands

The dashboard is composed as:

- Header row with application title and tab strip
- Main content area for the selected tab
- Status line showing proxy and upstream display names

Redraws are global and asynchronous. `ui.Invalidate()` marks the UI dirty and schedules a render through a buffered channel.

## Extension Points

The codebase is easiest to extend in these places:

- Add a new protocol file with `Register...`, `Request`, `Event`, and optional object constructors
- Register the protocol in `handleClient`
- Implement `DashboardDisplayable` on typed objects that should appear in the client detail view
- Add new ex-commands in `Dashboard.BeginExCommand`

## Operational Notes

- The proxy socket is created under `$XDG_RUNTIME_DIR` when the provided display name is relative.
- The upstream display may also be given as an absolute socket path.
- The process mutates its own `WAYLAND_DISPLAY` to the proxy display name during startup.
- Packet forwarding includes passing received Unix FDs onward to the peer.

## Known Limitations

- No automated tests are present in the repository today.
- Access to shared state such as `Proxy.Clients` is largely unsynchronized.
- Unsupported protocols are only tracked as generic objects.
- The client list and object views still contain TODOs around scrolling and large-state ergonomics.

## Suggested Next Improvements

1. Add a small test suite around packet parsing and object lifecycle bookkeeping.
2. Introduce synchronization or event serialization around shared proxy and client state.
3. Separate packet transport from state mutation so protocol decoding is easier to test.
4. Add optional persistent logs or packet filtering for large sessions.
