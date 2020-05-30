package main

import (
	"fmt"
	"os"
)

type WlCompositorImpl struct {
	client *Client
}

func RegisterWlCompositor(client *Client) {
	r := &WlCompositorImpl{
		client: client,
	}
	client.Impls["wl_compositor"] = r
}

func (r *WlCompositorImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // create_surface
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_surface")
		obj.Data = &WlSurface{
			Object: obj,
		}
		fmt.Fprintf(os.Stderr, "-> wl_compositor@%d.create_surface(surface: %s)\n", packet.ObjectId, obj)
	case 1: // create_region
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_region")
		fmt.Fprintf(os.Stderr, "-> wl_compositor@%d.create_region(region: %s)\n", packet.ObjectId, obj)
	}
	return nil
}

func (r *WlCompositorImpl) Event(packet *WaylandPacket) error {
	return nil
}
