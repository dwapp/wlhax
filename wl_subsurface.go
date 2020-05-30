package main

import (
	"errors"
	"fmt"
)

type WlSubSurfaceState struct {
	SubSurface *WlSubSurface
	X, Y       int32
	Desync     bool
}

func (s *WlSubSurfaceState) String() string {
	return s.SubSurface.String()
}

type WlSubSurface struct {
	Object  *WaylandObject
	Surface *WlSurface
}

func (r *WlSubSurface) String() string {
	return fmt.Sprintf("wl_subsurface@%d", r.Object.ObjectId)
}

func (r *WlSubSurface) Destroy() error {
	return nil
}

type WlSubSurfaceImpl struct {
	client *Client
}

func RegisterWlSubSurface(client *Client) {
	r := &WlSubSurfaceImpl{
		client: client,
	}
	client.Impls["wl_subsurface"] = r
}

func (r *WlSubSurfaceImpl) Destroy() error {
	return nil
}

func (r *WlSubSurfaceImpl) Request(packet *WaylandPacket) error {
	obj, ok := r.client.ObjectMap[packet.ObjectId].Data.(*WlSubSurface)
	if !ok {
		return errors.New("object is not wl_subsurface")
	}
	robj := obj.Surface.Next.Role.(WlSubSurfaceState)
	switch packet.Opcode {
	case 0: // destroy
	case 1: // set_position
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		robj.X, robj.Y = x, y
	case 2: // place_above

	case 3: // place_below
	case 4: // set_sync
		robj.Desync = false
	case 5: // set_desync
		robj.Desync = true
	}

	obj.Surface.Next.Role = robj
	return nil
}

func (r *WlSubSurfaceImpl) Event(packet *WaylandPacket) error {
	return nil
}
