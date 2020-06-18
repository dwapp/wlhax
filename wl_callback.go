package main

import (
	"errors"
)

type CallbackSubscriber interface {
	Done() error
}

type WlCallback struct {
	Origin      *WaylandObject
	Subscriber  CallbackSubscriber
	Description string
}

func (r *WlCallback) Destroy() error {
	return nil
}

type WlCallbackImpl struct {
	client *Client
}

func (r *WlCallbackImpl) Create() interface{} {
	return nil
}

func RegisterWlCallback(client *Client) {
	r := &WlCallbackImpl{
		client: client,
	}
	client.Impls["wl_callback"] = r
}

func (r *WlCallbackImpl) Request(packet *WaylandPacket) error {
	return errors.New("wl_callback has no requests")
}

func (r *WlCallbackImpl) Event(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // done
		obj, ok := r.client.ObjectMap[packet.ObjectId]
		if !ok {
			return errors.New("no such callback")
		}
		data := obj.Data.(*WlCallback)
		if data.Subscriber != nil {
			data.Subscriber.Done()
		}
	}
	return nil
}
