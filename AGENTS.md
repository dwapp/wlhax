# wlhax

## What It Is

Transparent Wayland proxy with terminal dashboard. Intercepts client↔compositor traffic,
decodes messages, tracks protocol state, and renders a live TUI.

## Architecture

Three layers:

1. **Bootstrap** — `main.go`: reads `WAYLAND_DISPLAY`, binds proxy socket `wlhax-0`..`9`,
   mutates process env to proxy name, launches CLI command if given
2. **Proxy + Protocol** — `proxy.go` (Proxy/Client types), `protocol.go` (WaylandPacket),
   per-interface `*.go` files
3. **UI** — `ui/` directory using vaxis; `dashboard.go` composes tabs, `client.go` per-client detail view

### Runtime Flow

```
client → proxy socket → handleClient() → [goroutine: client→compositor]
                                    → [goroutine: compositor→client]
                                    → in-memory Client model → dashboard render
```

`handleClient`: creates `Client` with `wl_display@1`, registers protocol handlers into
`client.Impls`, connects upstream, starts two forwarding goroutines.

## Core Types

| Type | File | Role |
|------|------|------|
| `Proxy` | `proxy.go` | Listening socket, client list, controls (`SlowMode`/`Block`/`CloseWrite`), UI callbacks |
| `Client` | `proxy.go` | Central state: socket pair, `RxLog`/`TxLog`, `Objects`/`ObjectMap`, `Globals`/`GlobalMap`, `Impls` dispatch |
| `WaylandPacket` | `protocol.go` | Decoded message header+payload, FDs, typed readers (`ReadUint32`/`ReadString`/`ReadFixed`), `WritePacket` for forwarding |

## Protocol Model

One file per interface (e.g. `wl_surface.go`, `xdg_shell.go`, `zwp_linux_dmabuf.go`).

Each file provides:
- `Register...(client)` — installs impl into `client.Impls`
- `Request(packet)` — decodes client→compositor
- `Event(packet)` — decodes compositor→client
- Optional `Create(obj)` — builds typed state on `wl_registry.bind`
- Optional `Destroy()` — cleanup on object removal

Model is **observational**: derives state from traffic, unsupported interfaces appear as generic objects.

### Adding a New Protocol

See the `adding-wayland-protocol` skill (`/skill:adding-wayland-protocol`).

## Key Conventions

- **Packet arg types**: `uint`→`ReadUint32`, `int`→`ReadInt32`, `fixed`→`ReadFixed`, `string`→`ReadString`, `object`→`ReadUint32` (lookup in `ObjectMap`), `new_id`→`ReadUint32` + `client.NewObject(oid, iface)` + set `obj.Data`, `fd`→no read (in `packet.Fds`), `array`→read length then loop
- **Opcodes**: assigned by XML element order (0, 1, 2, …)
- **Dashboard**: implement `DashboardDisplayable` (`DashboardShouldDisplay`, `DashboardCategory`, `DashboardPrint`) on object structs
- **Registration**: every interface (including children) must be `Register`'d in `handleClient`
- **File naming**: strip `_vN` suffix — `zxdg_decoration_manager_v1` → `zxdg_decoration.go`
- **File header**: `// <interface_name> protocol version: <v>`

## UI Structure

- `ui/ui.go` — event loop integration, redraw queue, suspend/resume
- `dashboard.go` — top-level composition, ex-command handling (`:commands`)
- `clients.go` — connections overview tab
- `client.go` — per-client object/category browser
- `exline.go` — command line widget

Redraws: `ui.Invalidate()` marks dirty, renders via buffered channel.

## Known Issues

- No tests
- Shared state (`Proxy.Clients`) largely unsynchronized
- Unsupported protocols tracked as generic objects only
- Scrolling/large-state ergonomics have TODOs

## Build

```sh
go build -o wlhax .
```

## Operational Notes

- Proxy socket created under `$XDG_RUNTIME_DIR` for relative display names
- Process mutates its own `WAYLAND_DISPLAY` to proxy name during startup
- FD forwarding included in packet relay via Unix socket ancillary data
