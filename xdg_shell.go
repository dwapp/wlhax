package main

import (
	"errors"
	"io"
	"fmt"
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
	XdgRole                                    WlSurfaceRole
}

func (s XdgSurfaceState) String() string {
	if s.XdgRole != nil {
		return s.XdgRole.String()
	}
	return s.XdgSurface.Object.String()
}

func (s XdgSurfaceState) Details() []string {
	if s.XdgRole != nil {
		return s.XdgRole.Details()
	}
	return nil
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
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "xdg_positioner")
		p := &XdgPositioner{
			Object:  obj,
		}
		obj.Data = p
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
		posid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		parentobj, ok := r.client.ObjectMap[pid]
		if !ok {
			return errors.New("no such object")
		}
		parent := parentobj.Data.(*XdgSurface)
		posobj, ok := r.client.ObjectMap[posid]
		if !ok {
			return errors.New("no such object")
		}
		pos := posobj.Data.(*XdgPositioner)
		obj := r.client.NewObject(oid, "xdg_popup")
		p := &XdgPopup{
			Object:     obj,
			XdgSurface: xdg_surface,
			Positioner: pos,
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

type XdgPositioner struct {
	Object *WaylandObject
	Width, Height int32
	AnchorX, AnchorY, AnchorWidth, AnchorHeight int32
	Anchor uint32
	Gravity uint32
	ConstraintAdjustment uint32
	OffsetX, OffsetY int32
	Reactive bool
	ParentWidth, ParentHeight int32
	ParentConfigure uint32
}

func (t *XdgPositioner) Destroy() error {
	return nil
}

type XdgPositionerImpl struct {
	client *Client
}

func RegisterXdgPositioner(client *Client) {
	r := &XdgPositionerImpl{
		client: client,
	}
	client.Impls["xdg_positioner"] = r
}

func (r *XdgPositionerImpl) Request(packet *WaylandPacket) error {

	// What have I done.
	object := r.client.ObjectMap[packet.ObjectId]
	positioner := object.Data.(*XdgPositioner)
	switch packet.Opcode {
	case 0: // destroy
	case 1: // set_size
		w, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		h, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		positioner.Width = w
		positioner.Height = h
	case 2: // set_anchor_rect
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
		positioner.AnchorX = x
		positioner.AnchorY = y
		positioner.AnchorWidth = w
		positioner.AnchorHeight = h
	case 3: // set_anchor
		anchor, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		positioner.Anchor = anchor
	case 4: // set_gravity
		gravity, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		positioner.Gravity = gravity
	case 5: // set_constraint_adjustment
		ca, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		positioner.ConstraintAdjustment = ca
	case 6: // set_offset
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		positioner.OffsetX = x
		positioner.OffsetY = y
	case 7: // set_reactive
		positioner.Reactive = true
	case 8: // set_parent_size
		w, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		h, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		positioner.ParentWidth = w
		positioner.ParentHeight = h

	case 9: // set_parent_configure
		conf, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		positioner.ParentConfigure = conf
	}
	return nil
}

func (r *XdgPositionerImpl) Event(packet *WaylandPacket) error {
	return errors.New("no events expected for XdgPositioner")
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

func (s XdgToplevelState) Details() []string {
	var suffix, details string
	if s.Parent != nil {
		suffix = fmt.Sprintf("app_id: %s, title: %s, xdg_surface: %s, parent: %s", s.AppId, s.Title, s.XdgToplevel.XdgSurface.Object.String(), s.Parent.Object.String())
	} else {
		suffix = fmt.Sprintf("app_id: %s, title: %s, xdg_surface: %s", s.AppId, s.Title, s.XdgToplevel.XdgSurface.Object.String())
	}

	role := s.XdgToplevel.XdgSurface.Surface.Current.Role.(XdgSurfaceState)
	if role.CurrentConfigure.Serial == role.PendingConfigure.Serial {
		details = fmt.Sprintf("geom: x=%d y=%d w=%d h=%d, current: w=%d h=%d", role.GeometryX, role.GeometryY, role.GeometryW, role.GeometryH, role.CurrentConfigure.Width, role.CurrentConfigure.Height)
	} else {
		details = fmt.Sprintf("geom: x=%d y=%d w=%d h=%d, current: w=%d h=%d, pending: w=%d h=%d", role.GeometryX, role.GeometryY, role.GeometryW, role.GeometryH, role.CurrentConfigure.Width, role.CurrentConfigure.Height, role.PendingConfigure.Width, role.PendingConfigure.Height)
	}

	return []string{
		suffix,
		details,
	}
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

func (s XdgPopupState) Details() []string {
	p := s.XdgPopup.Positioner
	return []string{
		fmt.Sprintf("parent: %s", s.XdgPopup.Parent.Object.String()),
		fmt.Sprintf("positioner size: w=%d h=%d, anchor: %d, x=%d y=%d w=%d h=%d",
			p.Width, p.Height, p.Anchor, p.AnchorX, p.AnchorY, p.AnchorWidth, p.AnchorHeight),
		fmt.Sprintf("positioner gravity: %d, constraints: %d, offset: x=%d y=%d",
			p.Gravity, p.ConstraintAdjustment, p.OffsetX, p.OffsetY),
	}
}

type XdgPopup struct {
	Object     *WaylandObject
	XdgSurface *XdgSurface
	Parent     *XdgSurface
	Positioner *XdgPositioner
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
	switch packet.Opcode {
	case 0: // destroy
	case 1: // grab
	case 2: // reposition
	}
	return nil
}

func (r *XdgPopupImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // configure
	case 1: // popup_done
	case 2: // repositioned
	}

	return nil
}

