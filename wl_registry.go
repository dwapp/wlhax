package main

import "github.com/pkg/errors"

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
		_, err = packet.ReadString()
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
		obj := r.client.NewObject(oid, global.Interface)
		if impl, ok := r.client.Impls[global.Interface]; ok {
			creatable, ok := impl.(interface {
				Create(*WaylandObject) Destroyable
			})
			if ok {
				obj.Data = creatable.Create(obj)
			}
		} else {
		}
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
		r.client.Globals = append(r.client.Globals, global)
		r.client.GlobalMap[gid] = global
	case 1: // global_remove
	}
	return nil
}
