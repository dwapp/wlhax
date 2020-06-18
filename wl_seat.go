package main

type WlSeat struct {
	Object *WaylandObject
	Children []*WaylandObject
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

func (r *WlSeatImpl) Create() Destroyable {
	return &WlSeat{}
}

func (r *WlSeatImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	seat := object.Data.(*WlSeat)
	switch packet.Opcode {
	case 0: // get_pointer
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_pointer")
		obj.Data = &WlPointer{
			Object: obj,
			Seat:   seat,
		}
		seat.Children = append(seat.Children, obj)
	case 1: // get_keyboard
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_keyboard")
		obj.Data = &WlKeyboard{
			Object: obj,
			Seat:   seat,
		}
		seat.Children = append(seat.Children, obj)
	case 2: // get_touch
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj := r.client.NewObject(oid, "wl_touch")
		obj.Data = &WlTouch{
			Object: obj,
			Seat:   seat,
		}
		seat.Children = append(seat.Children, obj)
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
