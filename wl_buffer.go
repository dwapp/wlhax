package main

import (
	"errors"
)

type WlBufferType interface {
	String() string
}

type BufferSubscriber interface {
	Release()
	Destroy()
	Object() *WaylandObject
}

type WlBuffer struct {
	Object      *WaylandObject
	Subscriber  BufferSubscriber
	BufferType  WlBufferType
	Attached    bool
	Committed   bool
}

func (*WlBuffer) DashboardShouldDisplay() bool {
	return true
}

func (*WlBuffer) DashboardCategory() string {
	return "Buffers"
}

func (b *WlBuffer) DashboardPrint(printer func(string, ...interface{})) error {
	if b.Committed {
		printer("%s - %s %s, active: %s", Indent(0), b.Object, b.BufferType.String(), b.Subscriber.Object())
	} else if b.Attached {
		printer("%s - %s %s, attach: %s", Indent(0), b.Object, b.BufferType.String(), b.Subscriber.Object())
	} else {
		printer("%s - %s %s", Indent(0), b.Object, b.BufferType.String())
	}
	return nil
}

func (r *WlBuffer) Destroy() error {
	if r.Subscriber != nil {
		r.Subscriber.Destroy()
		r.Attached = false
		r.Committed = false
	}
	return nil
}

type WlBufferImpl struct {
	client *Client
}

func (r *WlBufferImpl) Create() interface{} {
	return nil
}

func RegisterWlBuffer(client *Client) {
	r := &WlBufferImpl{
		client: client,
	}
	client.Impls["wl_buffer"] = r
}

func (r *WlBufferImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	}
	return nil
}

func (r *WlBufferImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // release
		obj, ok := r.client.ObjectMap[packet.ObjectId]
		if !ok {
			return errors.New("no such buffer")
		}
		data := obj.Data.(*WlBuffer)
		if data.Subscriber != nil {
			data.Subscriber.Release()
			data.Attached = false
			data.Committed = false
		}
	}
	return nil
}
