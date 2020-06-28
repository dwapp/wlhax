package main

import (
	"errors"
	"io"
)

// Before anyone asks about the arbitrary indexing with type asserts into deep structures:
//
//   Yes, I know. I regret everything.
//

type XdgConfigure struct {
	Serial int32
	Width  int32
	Height int32
}

type XdgSurfaceState struct {
	XdgSurface                                 *XdgSurface
	CurrentConfigure, PendingConfigure         XdgConfigure
	GeometryX, GeometryY, GeometryW, GeometryH int32
	XdgRole                                    interface{}
}

func (s *XdgSurfaceState) String() string {
	if s.XdgRole != nil {
		stringer := s.XdgRole.(interface {
			String() string
		})
		return stringer.String()
	}
	return s.XdgSurface.Object.String()
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
		s := &XdgSurface{
			Object:  obj,
			Surface: surface_obj_surface,
		}
		obj.Data = s
		surface_obj_surface.Next.Role = XdgSurfaceState{
			XdgSurface: s,
		}
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
		t := &XdgToplevel{
			Object:     obj,
			XdgSurface: xdg_surface,
		}
		obj.Data = t
		role := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
		role.XdgRole = XdgToplevelState{
			XdgToplevel: t,
		}
		xdg_surface.Surface.Next.Role = role
	case 2: // get_popup
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		pid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		parentobj, ok := r.client.ObjectMap[pid]
		if !ok {
			return errors.New("no such object")
		}
		parent := parentobj.Data.(*XdgSurface)
		obj := r.client.NewObject(oid, "xdg_popup")
		p := &XdgPopup{
			Object:     obj,
			XdgSurface: object.Data.(*XdgSurface),
			Parent:     parent,
		}
		obj.Data = p
		role := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
		role.XdgRole = XdgPopupState{
			XdgPopup: p,
		}
		xdg_surface.Surface.Next.Role = role
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
		xdg_surface.Surface.Next.Role = robj
	case 4: // ack_configure
		conf, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		if robj.PendingConfigure.Serial == conf {
			robj.CurrentConfigure = robj.PendingConfigure
		}
		xdg_surface.Surface.Next.Role = robj
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
		robj.PendingConfigure.Serial = conf
	}
	object.Data.(*XdgSurface).Surface.Next.Role = robj
	return nil
}

type XdgToplevelState struct {
	XdgToplevel *XdgToplevel
	Title       string
	AppId       string
	Parent      *XdgToplevel
}

func (s XdgToplevelState) String() string {
	return s.XdgToplevel.Object.String()
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
		oid, err := packet.ReadUint32()
		if err != nil && err != io.EOF {
			return err
		}
		if oid == 0 {
			toplevelstate.Parent = nil
			break
		}
		obj, ok := r.client.ObjectMap[oid]
		if !ok {
			return errors.New("no such object")
		}
		toplevel := obj.Data.(*XdgToplevel)
		toplevelstate.Parent = toplevel
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
	object := r.client.ObjectMap[packet.ObjectId]
	xdg_surface := object.Data.(*XdgToplevel).XdgSurface
	xdgstate := xdg_surface.Surface.Next.Role.(XdgSurfaceState)
	switch packet.Opcode {
	case 0: // configure
		width, err := packet.ReadInt32()
		if err != nil && err != io.EOF {
			return err
		}
		height, err := packet.ReadInt32()
		if err != nil && err != io.EOF {
			return err
		}
		xdgstate.PendingConfigure.Width = width
		xdgstate.PendingConfigure.Height = height
		xdg_surface.Surface.Next.Role = xdgstate
	case 1: // close
	}
	return nil
}

type XdgPopupState struct {
	XdgPopup *XdgPopup
}

func (s XdgPopupState) String() string {
	return s.XdgPopup.Object.String()
}

type XdgPopup struct {
	Object     *WaylandObject
	XdgSurface *XdgSurface
	Parent     *XdgSurface
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
