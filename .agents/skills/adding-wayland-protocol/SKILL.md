---
name: adding-wayland-protocol
description: >-
  Use when adding support for a new Wayland protocol extension to wlhax —
  when asked to "add protocol", "implement protocol", "track protocol",
  "add interface", or any variant involving wlhax proxy protocol decoding.
---

# Adding Wayland Protocol Support

## Overview

wlhax is a transparent proxy — it does not implement protocol semantics,
it only decodes and records messages for the dashboard to display.

## 1. Read the Protocol XML

Locations (by priority):

```
/usr/share/wayland-protocols/stable/<name>/<name>.xml
/usr/share/wayland-protocols/unstable/<name>/<name>-unstable-vN.xml
/usr/share/wayland-protocols/staging/<name>/<name>-vN.xml
```

Key XML elements:

| XML element | Direction | Meaning |
|---|---|---|
| `<request>` | client → compositor | Decoded in `Request()` |
| `<event>` | compositor → client | Decoded in `Event()` |
| `<enum>` | — | Named constant values for fields |
| `<request type="destructor">` | — | The request destroys the object |

**Opcode order**: Assigned by the order `<request>` / `<event>` elements appear
in the XML — 0, 1, 2, …. Read args in the exact order listed per message.

**Arg type mapping**:

| XML type | Go read call |
|---|---|
| `uint` | `packet.ReadUint32()` |
| `int` | `packet.ReadInt32()` |
| `fixed` | `packet.ReadFixed()` |
| `string` | `packet.ReadString()` |
| `object` | `packet.ReadUint32()` → object ID |
| `new_id` | `packet.ReadUint32()` → must call `client.NewObject(oid, "iface")` and set `obj.Data` |
| `fd` | No read needed — FDs arrive via Unix socket ancillary data in `packet.Fds` |
| `array` | `ReadUint32()` for byte length, then loop reading elements |

## 2. File Naming

One `.go` file per protocol (or per tightly related group of interfaces).

File header comment must note the protocol version:
```go
// <interface_name> protocol version: <v>
package main
```

Naming: strip `_vN` suffix. Example: `zxdg_decoration_manager_v1` → `zxdg_decoration.go`

## 3. Implementation Steps

### 3.1 Object State Struct

```go
type ZxdgToplevelDecoration struct {
    Object         *WaylandObject
    Toplevel       *WaylandObject
    PreferredMode  *EnumZxdgDecorationMode
    ConfiguredMode *EnumZxdgDecorationMode
}

func (z *ZxdgToplevelDecoration) Destroy() error { return nil }
```

`Destroy()` implements the `Destroyable` interface. Return `nil` unless explicit cleanup is needed.

### 3.2 Enum Types

```go
type EnumZxdgDecorationMode uint32

const (
    EnumZxdgDecorationModeClientSide EnumZxdgDecorationMode = 1
    EnumZxdgDecorationModeServerSide EnumZxdgDecorationMode = 2
)

func (m EnumZxdgDecorationMode) String() string {
    switch m {
    case EnumZxdgDecorationModeClientSide:
        return "client-side"
    case EnumZxdgDecorationModeServerSide:
        return "server-side"
    default:
        return fmt.Sprintf("unknown(%d)", uint32(m))
    }
}
```

### 3.3 Impl Struct + Request/Event

```go
type ZxdgToplevelDecorationImpl struct { client *Client }

func RegisterZxdgToplevelDecoration(client *Client) {
    client.Impls["zxdg_toplevel_decoration_v1"] = &ZxdgToplevelDecorationImpl{client: client}
}

func (r *ZxdgToplevelDecorationImpl) Request(packet *WaylandPacket) error {
    object := r.client.ObjectMap[packet.ObjectId]
    dec := object.Data.(*ZxdgToplevelDecoration)
    switch packet.Opcode {
    case 0: // destroy
    case 1: // set_mode
        mode, err := packet.ReadUint32()
        if err != nil { return err }
        m := EnumZxdgDecorationMode(mode)
        dec.PreferredMode = &m
    case 2: // unset_mode
        dec.PreferredMode = nil
    }
    return nil
}

func (r *ZxdgToplevelDecorationImpl) Event(packet *WaylandPacket) error {
    object := r.client.ObjectMap[packet.ObjectId]
    dec := object.Data.(*ZxdgToplevelDecoration)
    switch packet.Opcode {
    case 0: // configure
        mode, err := packet.ReadUint32()
        if err != nil { return err }
        m := EnumZxdgDecorationMode(mode)
        dec.ConfiguredMode = &m
    }
    return nil
}
```

### 3.4 Manager Interfaces (Global Objects)

If the protocol exposes a manager obtained via `wl_registry.bind`, implement an
optional `Create` method:

```go
type ZxdgDecorationManagerImpl struct { client *Client }

func RegisterZxdgDecorationManager(client *Client) {
    client.Impls["zxdg_decoration_manager_v1"] = &ZxdgDecorationManagerImpl{client: client}
}

func (r *ZxdgDecorationManagerImpl) Create(obj *WaylandObject) Destroyable {
    return &ZxdgDecorationManager{Object: obj}
}

func (r *ZxdgDecorationManagerImpl) Request(packet *WaylandPacket) error {
    switch packet.Opcode {
    case 0: // destroy
    case 1: // get_toplevel_decoration
        oid, err := packet.ReadUint32() // new_id
        if err != nil { return err }
        tid, err := packet.ReadUint32() // object
        if err != nil { return err }
        toplevelObj := r.client.ObjectMap[tid]
        obj := r.client.NewObject(oid, "zxdg_toplevel_decoration_v1")
        obj.Data = &ZxdgToplevelDecoration{Object: obj, Toplevel: toplevelObj}
    }
    return nil
}

func (r *ZxdgDecorationManagerImpl) Event(packet *WaylandPacket) error {
    return errors.New("zxdg_decoration_manager_v1 has no events")
}
```

**Critical**: Every `new_id` arg must call `client.NewObject(oid, "interface_name")` and assign
`obj.Data`. Without this, subsequent messages to that object ID will not be dispatched correctly.

### 3.5 Dashboard Display (Optional)

Implement the `DashboardDisplayable` interface on the object state struct:

```go
func (*ZxdgToplevelDecoration) DashboardShouldDisplay() bool { return true }
func (*ZxdgToplevelDecoration) DashboardCategory() string    { return "XDG Decoration" }

func (z *ZxdgToplevelDecoration) DashboardPrint(printer func(string, ...interface{})) error {
    preferred := "unset"
    if z.PreferredMode != nil { preferred = z.PreferredMode.String() }
    configured := "not configured"
    if z.ConfiguredMode != nil { configured = z.ConfiguredMode.String() }
    printer("%s - %s, toplevel: %s, preferred: %s, configured: %s",
        Indent(0), z.Object, z.Toplevel, preferred, configured)
    return nil
}
```

`Indent(0)` for top-level entries, `Indent(1)` for nested children.
Only implement on live per-object structs, not on managers.

## 4. Registration

Add a `Register...` call for **every** new interface in `handleClient` (`proxy.go`):

```go
RegisterZxdgDecorationManager(client)
RegisterZxdgToplevelDecoration(client)
```

Every interface (including child objects) must be registered individually.
Convention: manager before child.

## 5. Build Verification

```sh
go build -o wlhax .
```

## Common Mistakes

| Symptom | Cause | Fix |
|---|---|---|
| Dashboard category missing | `Register...` not added | Add both `Register` calls in `handleClient` |
| Dashboard category missing | Struct missing `DashboardDisplayable` | Add `DashboardCategory`/`DashboardShouldDisplay`/`DashboardPrint` |
| `panic: interface conversion` | `obj.Data` is nil or wrong type | Ensure every `new_id` arg calls `client.NewObject` and sets `obj.Data` |
| Wrong decoded values | Opcode or arg order wrong | Re-read the XML; file order is the opcode order |
| Null object panic | Optional `object` arg can be 0 | Check ID for zero before `ObjectMap` lookup |
| Build error: duplicate key | Two `Register` functions write the same `client.Impls` key | Check for duplicate interface name strings |

## Full Example

See [`zxdg_decoration.go`](../zxdg_decoration.go)

Source protocol XML: `/usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml`
