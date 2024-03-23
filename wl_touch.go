package main

import (
	"errors"
)

type WlTouch struct {
	Object *WaylandObject
	Seat   *WlSeat
}

func (r *WlTouch) Destroy() error {
	for idx := range r.Seat.Children {
		if r.Seat.Children[idx].Data == r {
			r.Seat.Children = append(r.Seat.Children[:idx], r.Seat.Children[idx+1:]...)
			break
		}
	}
	return nil
}

type WlTouchImpl struct {
	client *Client
}

func RegisterWlTouch(client *Client) {
	r := &WlTouchImpl{
		client: client,
	}
	client.Impls["wl_touch"] = r
}

func (r *WlTouchImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	_, ok := object.Data.(*WlTouch)
	if !ok {
		return errors.New("object is not a wl_touch")
	}
	switch packet.Opcode {
	case 0: // release
	}
	return nil
}

func (r *WlTouchImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // down
	case 1: // up
	case 2: // motion
	case 3: // frame
	case 4: // cancel
	case 5: // shape
	case 6: // orientation
	}
	return nil
}
