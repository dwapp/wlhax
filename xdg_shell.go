package main

import (
	"errors"
	"fmt"
	"os"
)

// Before anyone asks about the arbitrary indexing with type asserts into deep structures:
//
//   Yes, I know. I regret everything.
//

type XdgSurfaceState struct {
	CurrentConfigure, PendingConfigure         int32
	GeometryX, GeometryY, GeometryW, GeometryH int32
	XdgRole                                    interface{}
}

type XdgSurface struct {
	Object  *WaylandObject
	Surface *WlSurface
}

func (s *XdgSurface) Destroy() error {
	return nil
}

type XdgWmBaseImpl struct {
	client *Client
}

func RegisterXdgWmBase(client *Client) {
	r := &XdgWmBaseImpl{
		client: client,
	}
	client.Impls["xdg_wm_base"] = r
}

func (r *XdgWmBaseImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // create_positioner
	case 2: // get_xdg_surface
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		surface_obj, ok := r.client.ObjectMap[sid]
		if !ok {
			return errors.New("no such object")
		}
		surface_obj_surface := surface_obj.Data.(*WlSurface)

		obj := r.client.NewObject(oid, "xdg_surface")
		obj.Data = &XdgSurface{
			Surface: surface_obj_surface,
		}
		surface_obj_surface.Next.Role = XdgSurfaceState{}
		fmt.Fprintf(os.Stderr, "-> xdg_wm_base@%d.get_xdg_surface(xdg_surface: %s, surface: %s)\n", packet.ObjectId, obj, surface_obj)
	case 3: // pong
	}
	return nil
}

func (r *XdgWmBaseImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // ping
	}
	return nil
}

type XdgSurfaceImpl struct {
	client *Client
}

func RegisterXdgSurface(client *Client) {
	r := &XdgSurfaceImpl{
		client: client,
	}
	client.Impls["xdg_surface"] = r
}

func (r *XdgSurfaceImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	xdg_surface := object.Data.(*XdgSurface)
	robj := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_toplevel
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "xdg_toplevel")
		obj.Data = &XdgToplevel{
			Object:     obj,
			XdgSurface: xdg_surface,
		}
		role := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
		role.XdgRole = XdgToplevelState{}
		xdg_surface.Surface.Next.Role = role
		fmt.Fprintf(os.Stderr, "-> xdg_surface@%d.get_toplevel(xdg_stoplevel: %s)\n", packet.ObjectId, obj)
	case 2: // get_popup
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "xdg_popup")
		obj.Data = &XdgPopup{
			Object:     obj,
			XdgSurface: object.Data.(*XdgSurface),
		}
		role := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
		role.XdgRole = XdgPopupState{}
		xdg_surface.Surface.Next.Role = role
		fmt.Fprintf(os.Stderr, "-> xdg_surface@%d.get_popup(xdg_popup: %s)\n", packet.ObjectId, obj)
	case 3: // set_window_geometry
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		w, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		h, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		robj.GeometryX = x
		robj.GeometryY = y
		robj.GeometryW = w
		robj.GeometryH = h
	case 4: // ack_configure
		conf, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		robj.CurrentConfigure = conf
	}
	return nil
}

func (r *XdgSurfaceImpl) Event(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	robj := object.Data.(*XdgSurface).Surface.Next.Role.(XdgSurfaceState)
	switch packet.Opcode {
	case 0: // configure
		conf, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		robj.PendingConfigure = conf

	}
	object.Data.(*XdgSurface).Surface.Next.Role = robj
	return nil
}

type XdgToplevelState struct {
	Title string
	AppId string
}

type XdgToplevel struct {
	Object     *WaylandObject
	XdgSurface *XdgSurface
}

func (t *XdgToplevel) Destroy() error {
	return nil
}

type XdgToplevelImpl struct {
	client *Client
}

func RegisterXdgToplevel(client *Client) {
	r := &XdgToplevelImpl{
		client: client,
	}
	client.Impls["xdg_toplevel"] = r
}

func (r *XdgToplevelImpl) Request(packet *WaylandPacket) error {

	// What have I done.
	object := r.client.ObjectMap[packet.ObjectId]
	xdg_surface := object.Data.(*XdgToplevel).XdgSurface
	xdgstate := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
	toplevelstate := xdgstate.XdgRole.(XdgToplevelState)

	switch packet.Opcode {
	case 0: // destroy
	case 1: // set_parent
	case 2: // set_title
		str, err := packet.ReadString()
		if err != nil {
			return err
		}
		toplevelstate.Title = str
	case 3: // set_app_id
		str, err := packet.ReadString()
		if err != nil {
			return err
		}
		toplevelstate.AppId = str
	case 4: // show_window_menu
	case 5: // move
	case 6: // resize
	case 7: // set_max_size
	case 8: // set_min_size
	case 9: // set_maximized
	case 10: // unset_maximized
	case 11: // set_fullscreen
	case 12: // unset_fullscreen
	case 13: // set_minimized
	}
	xdgstate.XdgRole = toplevelstate
	xdg_surface.Surface.Next.Role = xdgstate
	return nil
}

func (r *XdgToplevelImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // configure
	case 1: // close
	}
	return nil
}

type XdgPopupState struct {
}

type XdgPopup struct {
	Object     *WaylandObject
	XdgSurface *XdgSurface
}

func (t *XdgPopup) Destroy() error {
	return nil
}

type XdgPopupImpl struct {
	client *Client
}

func RegisterXdgPopup(client *Client) {
	r := &XdgPopupImpl{
		client: client,
	}
	client.Impls["xdg_popup"] = r
}

func (r *XdgPopupImpl) Request(packet *WaylandPacket) error {
	return nil
}

func (r *XdgPopupImpl) Event(packet *WaylandPacket) error {
	return nil
}
