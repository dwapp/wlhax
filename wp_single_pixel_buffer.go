package main

import (
	"errors"
	"fmt"
)

type WpSinglePixelBuffer struct {
	Red, Green, Blue, Alpha uint32
}

func (b *WpSinglePixelBuffer) String() string {
	return fmt.Sprintf("single-pixel-buffer, r: %d, g: %d, b: %d, a: %d",
		b.Red, b.Green, b.Blue, b.Alpha)
}

type WpSinglePixelBufferManager struct {
	Object *WaylandObject
}

func (r *WpSinglePixelBufferManager) Destroy() error {
	return nil
}

type WpSinglePixelBufferManagerImpl struct {
	client *Client
}

func RegisterWpSinglePixelBufferManager(client *Client) {
	r := &WpSinglePixelBufferManagerImpl{
		client: client,
	}
	client.Impls["wp_single_pixel_buffer_manager_v1"] = r
}

func (r *WpSinglePixelBufferManagerImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // create_u32_rgba_buffer
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		red, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		green, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		blue, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		alpha, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_buffer")
		b := &WpSinglePixelBuffer{
			Red:   red,
			Green: green,
			Blue:  blue,
			Alpha: alpha,
		}
		obj.Data = &WlBuffer{
			Object:     obj,
			BufferType: b,
		}
	}
	return nil
}

func (r *WpSinglePixelBufferManagerImpl) Event(packet *WaylandPacket) error {
	return errors.New("wp_single_pixel_buffer_manager has no events")
}
