package main

import (
	"errors"
	"io"
)

type WlPointerSurfaceState struct {
	WlPointer *WlPointer
}

func (s *WlPointerSurfaceState) String() string {
	return s.WlPointer.Object.String()
}

type WlPointer struct {
	Object         *WaylandObject
	Seat           *WlSeat
	PointerSurface *WlSurface

	EnteredSurface *WlSurface
}

func (pointer *WlPointer) dashboardPrint(printer func(string, ...interface{}), indent int) error {
	if pointer.EnteredSurface != nil {
		printer("%s - %s, entered: %s", Indent(indent), pointer.Object, pointer.EnteredSurface.Object)
	} else {
		printer("%s - %s", Indent(indent), pointer.Object)
	}
	return nil
}

func (r *WlPointer) Destroy() error {
	for idx := range r.Seat.Children {
		if r.Seat.Children[idx].Data == r {
			r.Seat.Children = append(r.Seat.Children[:idx], r.Seat.Children[idx+1:]...)
			break
		}
	}
	return nil
}

type WlPointerImpl struct {
	client *Client
}

func RegisterWlPointer(client *Client) {
	r := &WlPointerImpl{
		client: client,
	}
	client.Impls["wl_pointer"] = r
}

func (r *WlPointerImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WlPointer)
	if !ok {
		return errors.New("object is not a wl_pointer")
	}
	switch packet.Opcode {
	case 0: // set_cursor
		_, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err == io.EOF {
			sid = 0
		} else if err != nil {
			return err
		}
		if sid == 0 {
			obj.PointerSurface.Next.Role = nil
			obj.PointerSurface = nil
			return nil
		}

		source_obj := r.client.ObjectMap[sid]
		if source_obj == nil {
			return errors.New("no such object")
		}
		_ = source_obj
		source_obj_surface, ok := source_obj.Data.(*WlSurface)
		if !ok {
			return errors.New("object is not surface")
		}
		source_obj_surface.Next.Role = WlPointerSurfaceState{
			WlPointer: obj,
		}
		obj.PointerSurface = source_obj_surface
	case 1: // release
	}
	return nil
}

func (r *WlPointerImpl) Event(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WlPointer)
	if !ok {
		return errors.New("object is not a wl_pointer")
	}
	switch packet.Opcode {
	case 0: // enter
		_, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		surface_obj := r.client.ObjectMap[sid]
		if surface_obj == nil {
			return errors.New("no such object")
		}
		surface, ok := surface_obj.Data.(*WlSurface)
		if !ok {
			return errors.New("object is not surface")
		}
		obj.EnteredSurface = surface
	case 1: // leave
		obj.EnteredSurface = nil
	case 2: // motion
	case 3: // button
	case 4: // axis
	case 5: // frame
	case 6: // axis_source
	case 7: // axis_stop
	case 8: // axis_discrete
	}
	return nil
}
