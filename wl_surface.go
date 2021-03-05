package main

import (
	"errors"
	"time"
	"strings"
)

type WlSurfaceRole interface {
	String() string
	Details() []string
}

type WlSurfaceState struct {
	Buffer                             uint32
	BufferNum                          int
	BufferX, BufferY                   int32
	DamageX, DamageY, DamageW, DamageH int32
	Scale                              int32
	Transform                          int32
	Parent                             *WlSurface
	Children                           []*WlSubSurface
	Role                               WlSurfaceRole
}

type WlSurface struct {
	Object          *WaylandObject
	Frames          uint32
	RequestedFrames uint32
	Current, Next   WlSurfaceState
	Outputs         []*WaylandObject
}

func (surface *WlSurface) dashboardOutput(printer func(string, ...interface{}), indent int) error {

	rolestr := "<unknown>"
	var details []string

	if surface.Current.Role != nil {
		rolestr = surface.Current.Role.String()
		details = surface.Current.Role.Details()
	}

	printer("%s - %s, role: %s", Indent(indent), surface.Object, rolestr)
	printer("%sbuffers: %d, frames: %d/%d", Indent(indent+3), surface.Current.BufferNum, surface.Frames, surface.RequestedFrames)


	if len(surface.Outputs) > 0 {
		var x []string
		for _, obj := range surface.Outputs {
			x = append(x, obj.String())
		}
		printer("%soutputs: %s", Indent(indent+3), strings.Join(x, ", "))
	}
	for _, d := range details {
		printer("%s%s", Indent(indent+3), d)
	}

	for _, child := range surface.Current.Children {
		if err := child.Surface.dashboardOutput(printer, indent+1); err != nil {
			return err
		}
	}
	return nil
}

func (*WlSurface) DashboardCategory() string {
	return "Surfaces"
}

func (surface *WlSurface) DashboardShouldDisplay() bool {
	return surface.Current.Parent == nil
}

func (surface *WlSurface) DashboardPrint(printer func(string, ...interface{})) error {
	return surface.dashboardOutput(printer, 0)
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
		obj.Next.BufferX = x
		obj.Next.BufferY = y
		obj.Next.Buffer = buffer
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

	case 4: // set_opaque_region
	case 5: // set_input_region
	case 6: // commit
		// TODO: maybe we're messing up the children slice when we do things like this
		obj.Current = obj.Next
		if r.client.proxy.SlowMode {
			time.Sleep(250 * time.Millisecond)
		}
	case 7: // set_buffer_transform
		transform, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.Transform = transform
	case 8: // set_buffer_scale
		scale, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.Next.Scale = scale
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
	}
	return nil
}

func (r *WlSurfaceImpl) Event(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WlSurface)
	if !ok {
		return errors.New("object is not wl_surface")
	}
	switch packet.Opcode {
	case 0: // enter
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		output_obj := r.client.ObjectMap[sid]
		if output_obj == nil {
			return errors.New("no such object")
		}
		obj.Outputs = append(obj.Outputs, output_obj)
	case 1: // leave
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		for idx := 0; idx < len(obj.Outputs); idx++ {
			if obj.Outputs[idx].ObjectId == sid {
				obj.Outputs = append(obj.Outputs[:idx], obj.Outputs[idx+1:]...)
				idx--
			}
		}
	}
	return nil
}
