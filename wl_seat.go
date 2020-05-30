package main

type WlSeat struct {
	Object *WaylandObject
}

func (r *WlSeat) Destroy() error {
	return nil
}

type WlSeatImpl struct {
	client *Client
}

func RegisterWlSeat(client *Client) {
	r := &WlSeatImpl{
		client: client,
	}
	client.Impls["wl_seat"] = r
}

func (r *WlSeatImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // get_pointer
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_pointer")
		obj.Data = &WlPointer{
			Object: obj,
		}
	case 1: // get_keyboard
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		r.client.NewObject(oid, "wl_keyboard")
	case 2: // get_touch
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		r.client.NewObject(oid, "wl_touch")
	case 3: // release
	}
	return nil
}

func (r *WlSeatImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // capabilities
	case 1: // name
	}
	return nil
}
