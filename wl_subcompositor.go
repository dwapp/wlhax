package main

import (
	"errors"
	"fmt"
	"os"
)

type WlSubCompositorImpl struct {
	client *Client
}

func RegisterWlSubCompositor(client *Client) {
	r := &WlSubCompositorImpl{
		client: client,
	}
	client.Impls["wl_subcompositor"] = r
}
func (r *WlSubCompositorImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_subsurface
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		source_obj := r.client.ObjectMap[sid]
		if source_obj == nil {
			return errors.New("no such object")
		}
		source_obj_surface, ok := source_obj.Data.(*WlSurface)
		if !ok {
			return errors.New("object is not surface")
		}
		pid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		parent_obj := r.client.ObjectMap[pid]
		if parent_obj == nil {
			return errors.New("no such object")
		}
		parent_obj_surface, ok := parent_obj.Data.(*WlSurface)
		if !ok {
			return errors.New("object is not surface")
		}
		obj := r.client.NewObject(oid, "wl_subsurface")
		d := &WlSubSurface{
			Object:  obj,
			Surface: source_obj_surface,
		}
		obj.Data = d
		parent_obj_surface.Next.Children = append(parent_obj_surface.Next.Children, d)
		source_obj_surface.Next.Role = WlSubSurfaceState{}
		source_obj_surface.Next.Parent = parent_obj_surface
		fmt.Fprintf(os.Stderr, "-> wl_subcompositor@%d.get_subsurface(subsurface: %s, surface: %s, parent: %s)\n", packet.ObjectId, obj, source_obj, parent_obj)
	}
	return nil
}

func (r *WlSubCompositorImpl) Event(packet *WaylandPacket) error {
	return nil
}
