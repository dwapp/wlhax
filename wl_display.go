package main

import (
	"fmt"
	"os"
)

type WlDisplayImpl struct {
	client *Client
}

func RegisterWlDisplay(client *Client) {
	r := &WlDisplayImpl{
		client: client,
	}
	client.Impls["wl_display"] = r
}

func (r *WlDisplayImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	switch packet.Opcode {
	case 0: // sync
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_callback")
		obj.Data = &WlCallback{
			Origin:      object,
			Description: "sync",
		}
		fmt.Fprintf(os.Stderr, "-> wl_display@%d.sync(callback: %s)\n", packet.ObjectId, obj)
	case 1: // get_registry
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		r.client.NewObject(oid, "wl_registry")
	}
	return nil
}

func (r *WlDisplayImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // error
	case 1: // delete_id
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "wl_display@%d.delete(id: %d)\n", packet.ObjectId, oid)
		r.client.RemoveObject(oid)
	}
	return nil
}
