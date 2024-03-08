package main

import (
	"fmt"
)

type WlOutput struct {
	Object *WaylandObject
	Name   string
	Scale  int32
}

func (*WlOutput) DashboardShouldDisplay() bool {
	return true
}

func (*WlOutput) DashboardCategory() string {
	return "Outputs"
}

func (output *WlOutput) DashboardPrint(printer func(string, ...interface{})) error {
	s := output.Object.String()
	if output.Name != "" {
		s += fmt.Sprintf(" %q", output.Name)
	}
	if output.Scale != 0 {
		s += fmt.Sprintf(", scale: %d", output.Scale)
	}
	printer("%s - %s", Indent(0), s)
	return nil
}

func (*WlOutput) Destroy() error {
	return nil
}

type WlOutputImpl struct {
	client *Client
}

func RegisterWlOutput(client *Client) {
	r := &WlOutputImpl{
		client: client,
	}
	client.Impls["wl_output"] = r
}

func (r *WlOutputImpl) Create(obj *WaylandObject) Destroyable {
	return &WlOutput{
		Object: obj,
	}
}

func (r *WlOutputImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // release
	}
	return nil
}

func (r *WlOutputImpl) Event(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	output := object.Data.(*WlOutput)
	switch packet.Opcode {
	case 0: // geometry
	case 1: // mode
	case 2: // done
	case 3: // scale
		scale, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		output.Scale = scale
	case 4: // name
		name, err := packet.ReadString()
		if err != nil {
			return err
		}
		output.Name = name
	case 5: // description
	}
	return nil
}
