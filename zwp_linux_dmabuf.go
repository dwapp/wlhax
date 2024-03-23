package main

import (
	"errors"
	"fmt"
)

type ZwpLinuxDmabufBuffer struct {
	Width, Height int32
	Format, Flags uint32
}

func (b *ZwpLinuxDmabufBuffer) String() string {
	return fmt.Sprintf("linux-dmabuf, width: %d, height: %d, format: %d",
		b.Width, b.Height, b.Format)
}

type ZwpLinuxBufferParams struct {
	Object   *WaylandObject
	Creating *ZwpLinuxDmabufBuffer
}

func (r *ZwpLinuxBufferParams) Destroy() error {
	return nil
}

type ZwpLinuxBufferParamsImpl struct {
	client *Client
}

func RegisterZwpLinuxBufferParams(client *Client) {
	r := &ZwpLinuxBufferParamsImpl{
		client: client,
	}
	client.Impls["zwp_linux_buffer_params_v1"] = r
}

func (r *ZwpLinuxBufferParamsImpl) Request(packet *WaylandPacket) error {
	obj, ok := r.client.ObjectMap[packet.ObjectId]
	if !ok {
		return errors.New("no such zwp_linux_buffer_params_v1")
	}
	data := obj.Data.(*ZwpLinuxBufferParams)

	switch packet.Opcode {
	case 0: // destroy
	case 1: // add
	case 2: // create
		width, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		height, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		format, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		flags, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		data.Creating = &ZwpLinuxDmabufBuffer{
			Width:  width,
			Height: height,
			Format: format,
			Flags:  flags,
		}

	case 3: // create_immed
		oid, err := packet.ReadUint32()
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
		format, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		flags, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		data.Creating = &ZwpLinuxDmabufBuffer{
			Width:  width,
			Height: height,
			Format: format,
			Flags:  flags,
		}
		obj := r.client.NewObject(oid, "wl_buffer")

		obj.Data = &WlBuffer{
			Object:     obj,
			BufferType: data.Creating,
		}
	}
	return nil
}

func (r *ZwpLinuxBufferParamsImpl) Event(packet *WaylandPacket) error {
	obj, ok := r.client.ObjectMap[packet.ObjectId]
	if !ok {
		return errors.New("no such zwp_linux_buffer_params_v1")
	}
	data := obj.Data.(*ZwpLinuxBufferParams)

	switch packet.Opcode {
	case 0: // created
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_buffer")
		obj.Data = &WlBuffer{
			Object:     obj,
			BufferType: data.Creating,
		}
	case 1: // failed
	}
	return nil
}

type ZwpLinuxDmabuf struct {
	Object *WaylandObject
}

func (r *ZwpLinuxDmabuf) Destroy() error {
	return nil
}

type ZwpLinuxDmabufImpl struct {
	client *Client
}

func RegisterZwpLinuxDmabuf(client *Client) {
	r := &ZwpLinuxDmabufImpl{
		client: client,
	}
	client.Impls["zwp_linux_dmabuf_v1"] = r
}

func (r *ZwpLinuxDmabufImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // create_params
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "zwp_linux_buffer_params_v1")
		obj.Data = &ZwpLinuxBufferParams{
			Object: obj,
		}
	case 2: // get_default_feedback
	case 3: // get_surface_feedback
	}
	return nil
}

func (r *ZwpLinuxDmabufImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // format
	case 1: // modifier
	}
	return nil
}
