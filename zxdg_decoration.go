package main

import (
	"errors"
	"fmt"
)

// EnumZxdgDecorationMode represents the decoration mode enum from zxdg_toplevel_decoration_v1.
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

// ---------------------------------------------------------------------------
// zxdg_toplevel_decoration_v1
// ---------------------------------------------------------------------------

type ZxdgToplevelDecoration struct {
	Object   *WaylandObject
	Toplevel *WaylandObject
	// PreferredMode is set by the client via set_mode; nil means unset_mode.
	PreferredMode *EnumZxdgDecorationMode
	// ConfiguredMode is the mode sent by the compositor via configure.
	ConfiguredMode *EnumZxdgDecorationMode
}

func (z *ZxdgToplevelDecoration) Destroy() error {
	return nil
}

func (*ZxdgToplevelDecoration) DashboardShouldDisplay() bool {
	return true
}

func (*ZxdgToplevelDecoration) DashboardCategory() string {
	return "XDG Decoration"
}

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

type ZxdgToplevelDecorationImpl struct {
	client *Client
}

func RegisterZxdgToplevelDecoration(client *Client) {
	r := &ZxdgToplevelDecorationImpl{
		client: client,
	}
	client.Impls["zxdg_toplevel_decoration_v1"] = r
}

func (r *ZxdgToplevelDecorationImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	dec, ok := object.Data.(*ZxdgToplevelDecoration)
	if !ok {
		return errors.New("object is not zxdg_toplevel_decoration_v1")
	}

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
	dec, ok := object.Data.(*ZxdgToplevelDecoration)
	if !ok {
		return errors.New("object is not zxdg_toplevel_decoration_v1")
	}

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

// ---------------------------------------------------------------------------
// zxdg_decoration_manager_v1
// ---------------------------------------------------------------------------

type ZxdgDecorationManager struct {
	Object *WaylandObject
}

func (z *ZxdgDecorationManager) Destroy() error {
	return nil
}

type ZxdgDecorationManagerImpl struct {
	client *Client
}

func RegisterZxdgDecorationManager(client *Client) {
	r := &ZxdgDecorationManagerImpl{
		client: client,
	}
	client.Impls["zxdg_decoration_manager_v1"] = r
}

func (r *ZxdgDecorationManagerImpl) Create(obj *WaylandObject) Destroyable {
	return &ZxdgDecorationManager{Object: obj}
}

func (r *ZxdgDecorationManagerImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_toplevel_decoration
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		tid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		toplevelObj, ok := r.client.ObjectMap[tid]
		if !ok {
			return fmt.Errorf("zxdg_decoration_manager_v1: no such toplevel object: %d", tid)
		}
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
