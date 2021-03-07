package main

import (
	"errors"
	"io"
)

type WlPointerSurfaceState struct {
	WlPointer *WlPointer
}

func (s WlPointerSurfaceState) String() string {
	return s.WlPointer.Object.String()
}

func (WlPointerSurfaceState) Details() []string {
	return nil
}

type WlPointer struct {
	Object         *WaylandObject
	Seat           *WlSeat
	PointerSurface *WlSurface

	EnteredSurface *WlSurface
	SurfaceX       float64
	SurfaceY       float64
	ButtonsHeld    int
}

func (pointer *WlPointer) dashboardPrint(printer func(string, ...interface{}), indent int) error {
	var surfaceObj *WaylandObject
	if pointer.EnteredSurface != nil {
		surfaceObj = pointer.EnteredSurface.Object
	}
	printer("%s - %s, focus: %s, x: %.02f, y: %.02f, buttons held: %d", Indent(indent), pointer.Object, surfaceObj, pointer.SurfaceX, pointer.SurfaceY, pointer.ButtonsHeld)
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
			if obj.PointerSurface != nil {
				obj.PointerSurface.Next.Role = nil
				obj.PointerSurface = nil
			}
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
	var err error
	switch packet.Opcode {
	case 0: // enter
		_, err = packet.ReadUint32()
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
		obj.ButtonsHeld = 0
	case 2: // motion
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		surfaceX, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		surfaceY, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		obj.SurfaceX = surfaceX.ToDouble()
		obj.SurfaceY = surfaceY.ToDouble()
	case 3: // button
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		state, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		switch state {
		case 0:
			obj.ButtonsHeld -= 1
		case 1:
			obj.ButtonsHeld += 1
		}
	case 4: // axis
	case 5: // frame
	case 6: // axis_source
	case 7: // axis_stop
	case 8: // axis_discrete
	}
	return nil
}
