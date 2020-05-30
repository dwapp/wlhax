package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type WlRegistryImpl struct {
	client *Client
}

func RegisterWlRegistry(client *Client) {
	r := &WlRegistryImpl{
		client: client,
	}
	client.Impls["wl_registry"] = r
}

func (r *WlRegistryImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // bind
		gid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		global, ok := r.client.GlobalMap[gid]
		if !ok {
			return errors.New("no such global")
		}
		name, err := packet.ReadString()
		if err != nil {
			return err
		}
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "-> wl_registry@%d.bind(global: %d (%s), id: %d)\n", packet.ObjectId, gid, name, oid)
		r.client.NewObject(oid, global.Interface)

	}
	return nil
}

func (r *WlRegistryImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // global
		gid, err := packet.ReadUint32()
		if err != nil {
			return errors.Wrap(err, "wl_registry decode gid")
		}
		iface, err := packet.ReadString()
		if err != nil {
			return errors.Wrap(err, "wl_registry decode iface")
		}
		ver, err := packet.ReadUint32()
		if err != nil {
			return errors.Wrap(err, "wl_registry decode version")
		}
		global := &WaylandGlobal{
			GlobalId:  gid,
			Interface: iface,
			Version:   ver,
		}
		fmt.Fprintf(os.Stderr, "new global: %d, type: %s, version: %d\n", gid, strconv.Quote(iface), ver)
		r.client.Globals = append(r.client.Globals, global)
		r.client.GlobalMap[gid] = global
	case 1: // global_remove
		gid, err := packet.ReadUint32()
		if err != nil {
			return errors.Wrap(err, "wl_registry decode gid")
		}
		fmt.Fprintf(os.Stderr, "removed global: %d\n", gid)
	}
	return nil
}
