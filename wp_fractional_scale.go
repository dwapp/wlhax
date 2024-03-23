package main

import (
	"errors"
	"fmt"
)

type WpFractionalScale struct {
	Object         *WaylandObject
	Surface        *WaylandObject
	PreferredScale *uint32
}

func (w *WpFractionalScale) Destroy() error {
	return nil
}

func (*WpFractionalScale) DashboardShouldDisplay() bool {
	return true
}

func (*WpFractionalScale) DashboardCategory() string {
	return "Fractional scale"
}

func (w *WpFractionalScale) DashboardPrint(printer func(string, ...interface{})) error {

	if w.PreferredScale == nil {
		printer("%s - %s, surface: %s, scale: not set", Indent(0), w.Object, w.Surface)
	} else {
		printer("%s - %s, surface: %s, scale: %f (%d/120)", Indent(0), w.Object, w.Surface, float64(*w.PreferredScale)/120., *w.PreferredScale)
	}
	return nil
}

type WpFractionalScaleImpl struct {
	client *Client
}

func RegisterWpFractionalScale(client *Client) {
	r := &WpFractionalScaleImpl{
		client: client,
	}
	client.Impls["wp_fractional_scale_v1"] = r
}

func (w *WpFractionalScaleImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	}
	return nil
}

func (w *WpFractionalScaleImpl) Event(packet *WaylandPacket) error {
	object := w.client.ObjectMap[packet.ObjectId]
	obj, ok := object.Data.(*WpFractionalScale)
	if !ok {
		return errors.New("object is not wp_fractional_scale_v1")
	}

	switch packet.Opcode {
	case 0: // preferred_scale
		scale, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		obj.PreferredScale = &scale
	}
	return nil
}

type WpFractionalScaleManager struct {
	Object *WaylandObject
}

func (w *WpFractionalScaleManager) Destroy() error {
	return nil
}

type WpFractionalScaleManagerImpl struct {
	client *Client
}

func RegisterWpFractionalScaleManager(client *Client) {
	r := &WpFractionalScaleManagerImpl{
		client: client,
	}
	client.Impls["wp_fractional_scale_manager_v1"] = r
}

func (w *WpFractionalScaleManagerImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_fractional_scale
		oid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sid, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		sobj, ok := w.client.ObjectMap[sid]
		if !ok {
			return fmt.Errorf("no such surface object: %d", sid)
		}
		obj := w.client.NewObject(oid, "wp_fractional_scale_v1")
		obj.Data = &WpFractionalScale{
			Object:  obj,
			Surface: sobj,
		}
	}
	return nil
}

func (w *WpFractionalScaleManagerImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // error
	}
	return nil
}
