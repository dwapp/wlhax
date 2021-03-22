package main

type WlOutput struct {
	Object *WaylandObject
}

func (*WlOutput) DashboardShouldDisplay() bool {
	return true
}

func (*WlOutput) DashboardCategory() string {
	return "Outputs"
}

func (output *WlOutput) DashboardPrint(printer func(string, ...interface{})) error {
	printer("%s - %s", Indent(0), output.Object)
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
	switch packet.Opcode {
	case 0: // geometry
	case 1: // mode
	case 2: // done
	case 3: // scale
	}
	return nil
}
