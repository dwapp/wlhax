package main

import (
	"errors"
)

type WlKeyboardModifiers struct {
	Depressed uint32
	Latched   uint32
	Locked    uint32
	Group     uint32
}

type WlKeyboardRepeatInfo struct {
	Rate  int32
	Delay int32
}

type WlKeyboard struct {
	Object *WaylandObject
	Seat   *WlSeat

	EnteredSurface *WlSurface
	Modifiers      WlKeyboardModifiers
	RepeatInfo     WlKeyboardRepeatInfo
	KeysHeld       int
}

func (keyboard *WlKeyboard) dashboardPrint(printer func(string, ...interface{}), indent int) error {
	var surfaceObj *WaylandObject
	if keyboard.EnteredSurface != nil {
		surfaceObj = keyboard.EnteredSurface.Object
	}
	printer("%s - %s, entered: %s, keys held: %d", Indent(indent), keyboard.Object, surfaceObj, keyboard.KeysHeld)
	return nil
}

func (r *WlKeyboard) Destroy() error {
	for idx := range r.Seat.Children {
		if r.Seat.Children[idx].Data == r {
			r.Seat.Children = append(r.Seat.Children[:idx], r.Seat.Children[idx+1:]...)
			break
		}
	}
	return nil
}

type WlKeyboardImpl struct {
	client *Client
}

func RegisterWlKeyboard(client *Client) {
	r := &WlKeyboardImpl{
		client: client,
	}
	client.Impls["wl_keyboard"] = r
}

func (r *WlKeyboardImpl) Request(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	_, ok := object.Data.(*WlKeyboard)
	if !ok {
		return errors.New("object is not a wl_keyboard")
	}
	switch packet.Opcode {
	case 0: // release
	}
	return nil
}

func (r *WlKeyboardImpl) Event(packet *WaylandPacket) error {
	object := r.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WlKeyboard)
	if !ok {
		return errors.New("object is not a wl_keyboard")
	}
	var err error
	switch packet.Opcode {
	case 0: // keymap
	case 1: // enter
		_, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		surface_obj := r.client.ObjectMap[sid]
		if surface_obj == nil {
			return errors.New("no such object")
		}
		surface, ok := surface_obj.Data.(*WlSurface)
		if !ok {
			return errors.New("object is not surface")
		}
		obj.EnteredSurface = surface
	case 2: // leave
		obj.EnteredSurface = nil
		obj.KeysHeld = 0
	case 3: // key
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		_, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		state, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		switch state {
		case 0:
			obj.KeysHeld -= 1
		case 1:
			obj.KeysHeld += 1
		}
	case 4: // modifiers
		_, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj.Modifiers.Depressed, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		obj.Modifiers.Latched, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		obj.Modifiers.Locked, err = packet.ReadUint32()
		if err != nil {
			return err
		}
		obj.Modifiers.Group, err = packet.ReadUint32()
		if err != nil {
			return err
		}
	case 5: // repeat_info
		obj.RepeatInfo.Rate, err = packet.ReadInt32()
		if err != nil {
			return err
		}
		obj.RepeatInfo.Delay, err = packet.ReadInt32()
		if err != nil {
			return err
		}
	}
	return nil
}
