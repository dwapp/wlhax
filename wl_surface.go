package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type WlSurfaceState struct {
	Buffer                             uint32
	BufferNum                          int
	BufferX, BufferY                   int32
	DamageX, DamageY, DamageW, DamageH int32
	Scale                              int32
	Transform                          int32
	Parent                             *WlSurface
	Children                           []*WlSubSurface
	Role                               interface{}
}

type WlSurface struct {
	Object          *WaylandObject
	Frames          uint32
	RequestedFrames uint32
	Current, Next   WlSurfaceState
}

func (r *WlSurface) Done() error {
	r.Frames += 1
	return nil
}

func (r *WlSurface) Destroy() error {
	if r.Next.Parent != nil {
		for idx := range r.Next.Parent.Next.Children {
			if r.Next.Parent.Next.Children[idx].Surface == r {
				r.Next.Parent.Next.Children = append(r.Next.Parent.Next.Children[:idx], r.Next.Parent.Next.Children[idx+1:]...)
				break
			}
		}
	}
	return nil
}

type WlSurfaceImpl struct {
	client *Client
}

func RegisterWlSurface(client *Client) {
	r := &WlSurfaceImpl{
		client: client,
	}
	client.Impls["wl_surface"] = r
}

func (r *WlSurfaceImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WlSurface)
	if !ok {
		return errors.New("object is not wl_surface")
	}
	switch packet.Opcode {
	case 0: // destroy
	case 1: // attach
		buffer, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.BufferNum = obj.Current.BufferNum + 1
		obj.Next.Buffer = buffer
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.attach(buffer: %d, x: %d, y: %d)\n", packet.ObjectId, buffer, x, y)
	case 2: // damage
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		w, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		h, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.DamageX = x
		obj.Next.DamageY = y
		obj.Next.DamageW = w
		obj.Next.DamageH = h
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.damage(%d, %d, %d, %d)\n", packet.ObjectId, x, y, w, h)
	case 3: // frame
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		cb := r.client.NewObject(oid, "wl_callback")
		cb.Data = &WlCallback{
			Origin:      object,
			Description: "frame",
			Subscriber:  obj,
		}
		obj.RequestedFrames += 1

		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.frame(callback: %s)\n", packet.ObjectId, cb)
	case 4: // set_opaque_region
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.set_opaque_region()\n", packet.ObjectId)
	case 5: // set_input_region
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.set_input_region()\n", packet.ObjectId)
	case 6: // commit
		// TODO: maybe we're messing up the children slice when we do things like this
		obj.Current = obj.Next
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.commit()\n", packet.ObjectId)
		if r.client.proxy.SlowMode {
			time.Sleep(250 * time.Millisecond)
		}
	case 7: // set_buffer_transform
		transform, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.Transform = transform
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.set_buffer_transform(transform: %d)\n", packet.ObjectId, transform)
	case 8: // set_buffer_scale
		scale, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.Scale = scale
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.set_buffer_scale(scale: %d)\n", packet.ObjectId, scale)
	case 9: // damage_buffer
		scale := obj.Current.Scale
		if scale == 0 {
			scale = 1
		}
		x, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		y, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		w, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		h, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.DamageX = x * scale
		obj.Next.DamageY = y * scale
		obj.Next.DamageW = w * scale
		obj.Next.DamageH = h * scale
		fmt.Fprintf(os.Stderr, "-> wl_surface@%d.damage_buffer(%d, %d, %d, %d)\n", packet.ObjectId, x, y, w, h)

	}
	return nil
}

func (r *WlSurfaceImpl) Event(packet *WaylandPacket) error {
	return nil
}
