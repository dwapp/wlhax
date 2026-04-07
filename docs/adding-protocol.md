# Adding Support for a New Wayland Protocol

This document explains how to add tracking support for a Wayland protocol extension in wlhax.
wlhax works as a transparent proxy: it does not implement protocol semantics — it only decodes
and records messages so the dashboard can display them.

---

## 1. Reading the Protocol XML

Protocol XML files are typically located at:

- `/usr/share/wayland-protocols/stable/<name>/<name>.xml`
- `/usr/share/wayland-protocols/unstable/<name>/<name>-unstable-vN.xml`
- `/usr/share/wayland-protocols/staging/<name>/<name>-vN.xml`

A protocol consists of one or more `<interface>` elements. Each interface contains:

| XML element | Direction | Meaning |
|---|---|---|
| `<request>` | client → compositor | Decoded in `Request()` |
| `<event>` | compositor → client | Decoded in `Event()` |
| `<enum>` | — | Named constant values for fields |

`<request type="destructor">` means the request destroys the object.

Common `<arg>` types and their corresponding read calls:

| XML type | Go read call |
|---|---|
| `uint` | `packet.ReadUint32()` |
| `int` | `packet.ReadInt32()` |
| `fixed` | `packet.ReadFixed()` |
| `string` | `packet.ReadString()` |
| `object` | `packet.ReadUint32()` → object ID |
| `new_id` | `packet.ReadUint32()` → new object ID; call `client.NewObject(oid, "interface_name")` |
| `fd` | No read needed — FDs arrive via Unix socket ancillary data in `packet.Fds` |
| `array` | `ReadUint32()` for the byte length, then loop reading elements |

---

## 2. File Naming and Conventions

Create one `.go` file per protocol (or per tightly related group of interfaces).

**Important**: Start the file with a comment indicating the target version of the implemented protocol:
```go
// <interface_name> protocol version: <v>
package main
```

File naming convention:

| Interface name | File name |
|---|---|
| `zxdg_decoration_manager_v1` | `zxdg_decoration.go` |
| `wp_fractional_scale_manager_v1` | `wp_fractional_scale.go` |
| `zwp_idle_inhibit_manager_v1` | `zwp_idle_inhibit.go` |

If a protocol XML defines only one interface, or a manager together with its child objects, put
them all in the same file.

---

## 3. Implementation Pattern

### 3.1 Object State Struct

Each interface gets a Go struct that holds the tracked state for one live object.

```go
type ZxdgToplevelDecoration struct {
    Object         *WaylandObject
    Toplevel       *WaylandObject
    PreferredMode  *EnumZxdgDecorationMode // set by set_mode request
    ConfiguredMode *EnumZxdgDecorationMode // set by configure event
}

func (z *ZxdgToplevelDecoration) Destroy() error { return nil }
```

`Destroy()` implements the `Destroyable` interface. It is called when the object is removed.
Returning `nil` is fine unless you need explicit cleanup.

### 3.2 Enums

Map XML `<enum>` entries to a Go type with a `String()` method for readable dashboard output:

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

### 3.3 The Implementation Struct

Each interface needs an `Impl` struct with `Request` and `Event` methods:

```go
type ZxdgToplevelDecorationImpl struct {
    client *Client
}

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
        if err != nil {
            return err
        }
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
        if err != nil {
            return err
        }
        m := EnumZxdgDecorationMode(mode)
        dec.ConfiguredMode = &m
    }
    return nil
}
```

**Opcode ordering**: opcodes are assigned by the order `<request>` (or `<event>`) elements appear
inside the interface in the XML — 0, 1, 2, … Read the args in the exact order they are listed
in the XML for each message.

### 3.4 Manager Interfaces (Global Objects)

If the protocol exposes a manager that clients obtain via `wl_registry.bind`, implement an
optional `Create` method. The registry implementation calls `Create` automatically when the
client binds the interface:

```go
type ZxdgDecorationManagerImpl struct {
    client *Client
}

func RegisterZxdgDecorationManager(client *Client) {
    client.Impls["zxdg_decoration_manager_v1"] = &ZxdgDecorationManagerImpl{client: client}
}

// Create is called by the registry implementation when the client binds this interface.
func (r *ZxdgDecorationManagerImpl) Create(obj *WaylandObject) Destroyable {
    return &ZxdgDecorationManager{Object: obj}
}

func (r *ZxdgDecorationManagerImpl) Request(packet *WaylandPacket) error {
    switch packet.Opcode {
    case 0: // destroy
    case 1: // get_toplevel_decoration
        oid, err := packet.ReadUint32() // new_id arg
        if err != nil {
            return err
        }
        tid, err := packet.ReadUint32() // object arg (the associated xdg_toplevel)
        if err != nil {
            return err
        }
        toplevelObj := r.client.ObjectMap[tid]
        obj := r.client.NewObject(oid, "zxdg_toplevel_decoration_v1")
        obj.Data = &ZxdgToplevelDecoration{
            Object:   obj,
            Toplevel: toplevelObj,
        }
    }
    return nil
}

func (r *ZxdgDecorationManagerImpl) Event(packet *WaylandPacket) error {
    return errors.New("zxdg_decoration_manager_v1 has no events")
}
```

**Critical**: whenever a request argument has type `new_id`, you must call
`client.NewObject(oid, "interface_name")` and assign `obj.Data`. Without this, any subsequent
messages targeting that object ID will not be dispatched to the correct implementation.

### 3.5 Dashboard Display (Optional)

Implement the `DashboardDisplayable` interface on the object state struct to make it appear in
the per-client dashboard view:

```go
func (*ZxdgToplevelDecoration) DashboardShouldDisplay() bool { return true }
func (*ZxdgToplevelDecoration) DashboardCategory() string    { return "XDG Decoration" }

func (z *ZxdgToplevelDecoration) DashboardPrint(printer func(string, ...interface{})) error {
    preferred := "unset"
    if z.PreferredMode != nil {
        preferred = z.PreferredMode.String()
    }
    configured := "not configured"
    if z.ConfiguredMode != nil {
        configured = z.ConfiguredMode.String()
    }
    printer("%s - %s, toplevel: %s, preferred: %s, configured: %s",
        Indent(0), z.Object, z.Toplevel, preferred, configured)
    return nil
}
```

`Indent(depth int)` returns an indentation string. Use `Indent(0)` for top-level entries and
`Indent(1)` for nested children.

Only implement this interface on the live per-object structs (e.g. the decoration object itself),
not on the manager — managers do not typically need a dashboard entry.

---

## 4. Register in proxy.go

Add a `Register...` call for every new interface inside the `handleClient` function in `proxy.go`:

```go
// proxy.go — inside handleClient
RegisterZxdgDecorationManager(client)
RegisterZxdgToplevelDecoration(client)
```

Notes:
- **Every interface must be registered individually**, including child objects such as
  `zxdg_toplevel_decoration_v1`.
- Order does not affect correctness, but follow the convention of manager before child.

---

## 5. Build

```sh
go build -o wlhax .
```

---

## 6. Common Mistakes

| Symptom | Cause | Fix |
|---|---|---|
| Dashboard category missing | `Register...` call not added | Add both `Register` calls in `handleClient` |
| Dashboard category missing | Struct does not implement `DashboardDisplayable` | Add `DashboardCategory`, `DashboardShouldDisplay`, `DashboardPrint` |
| `panic: interface conversion` | `obj.Data` is nil or wrong type | Ensure every `new_id` arg calls `client.NewObject` and sets `obj.Data` |
| Wrong values decoded | Opcode or arg order wrong | Re-read the XML carefully; order in the file is the opcode order |
| Null object panic | Optional `object` arg can be 0 | Check the ID for zero before looking it up in `ObjectMap` |
| Build error: duplicate key | Two `Register` functions write the same `client.Impls` key | Check for duplicate interface name strings |

---

## 7. Full Example

See [`zxdg_decoration.go`](../zxdg_decoration.go) for a complete implementation covering:

- `zxdg_decoration_manager_v1` — global manager bound via `wl_registry`
- `zxdg_toplevel_decoration_v1` — per-toplevel decoration object created by the manager

Source protocol XML:
`/usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml`
