package main

import (
	"errors"
	"fmt"
)

type WlShmPool struct {
	Object *WaylandObject
}

func (*WlShmPool) Destroy() error {
	return nil
}

type WlShmBuffer struct {
	Offset, Width, Height, Stride int32
	Format                        uint32
}

func (b *WlShmBuffer) String() string {
	return fmt.Sprintf("shm, width: %d, height: %d, format: %d",
		b.Width, b.Height, b.Format)
}

type WlShmPoolImpl struct {
	client *Client
}

func RegisterWlShmPool(client *Client) {
	r := &WlShmPoolImpl{
		client: client,
	}
	client.Impls["wl_shm_pool"] = r
}

func (r *WlShmPoolImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // create_buffer
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		offset, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		width, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		height, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		stride, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		format, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_buffer")
		b := &WlShmBuffer{
			Offset: offset,
			Width:  width,
			Height: height,
			Stride: stride,
			Format: format,
		}
		obj.Data = &WlBuffer{
			Object:     obj,
			BufferType: b,
		}
	case 1: // destroy
	case 2: // resize
	}
	return nil
}

func (r *WlShmPoolImpl) Event(packet *WaylandPacket) error {
	return errors.New("wl_shm_pool has no events")
}

type WlShm struct {
	Origin      *WaylandObject
	Description string
}

func (r *WlShm) Destroy() error {
	return nil
}

type WlShmImpl struct {
	client *Client
}

func RegisterWlShm(client *Client) {
	r := &WlShmImpl{
		client: client,
	}
	client.Impls["wl_shm"] = r
}

func (r *WlShmImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // create_pool
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_shm_pool")
		obj.Data = &WlShmPool{
			Object: obj,
		}
	}
	return nil
}

func (r *WlShmImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // format
	}
	return nil
}
