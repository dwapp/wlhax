package main

import (
	"errors"
	"fmt"
)

type WpIdleInhibitor struct {
	Object         *WaylandObject
	Surface        *WaylandObject
	PreferredScale *uint32
}

func (w *WpIdleInhibitor) Destroy() error {
	return nil
}

func (*WpIdleInhibitor) DashboardShouldDisplay() bool {
	return true
}

func (*WpIdleInhibitor) DashboardCategory() string {
	return "Idle inhibitor"
}

func (w *WpIdleInhibitor) DashboardPrint(printer func(string, ...interface{})) error {
	printer("%s - %s, surface: %s", Indent(0), w.Object, w.Surface)
	return nil
}

type WpIdleInhibitorImpl struct {
	client *Client
}

func RegisterWpIdleInhibitor(client *Client) {
	r := &WpIdleInhibitorImpl{
		client: client,
	}
	client.Impls["zwp_idle_inhibitor_v1"] = r
}

func (w *WpIdleInhibitorImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	}
	return nil
}

func (w *WpIdleInhibitorImpl) Event(packet *WaylandPacket) error {
	return errors.New("zwp_idle_inhibitor_v1 has no events")
}

type WpIdleInhibitManager struct {
	Object *WaylandObject
}

func (w *WpIdleInhibitManager) Destroy() error {
	return nil
}

type WpIdleInhibitManagerImpl struct {
	client *Client
}

func RegisterWpIdleInhibitManager(client *Client) {
	r := &WpIdleInhibitManagerImpl{
		client: client,
	}
	client.Impls["zwp_idle_inhibit_manager_v1"] = r
}

func (w *WpIdleInhibitManagerImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // create_inhibitor
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
		obj := w.client.NewObject(oid, "zwp_idle_inhibitor_v1")
		obj.Data = &WpIdleInhibitor{
			Object:  obj,
			Surface: sobj,
		}
	}
	return nil
}

func (w *WpIdleInhibitManagerImpl) Event(packet *WaylandPacket) error {
	return errors.New("zwp_idle_inhibit_manager_v1 has no events")
}
