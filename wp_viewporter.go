package main

import (
	"errors"
	"fmt"
	"strings"
)

type WpViewport struct {
	Object *WaylandObject
	Surface *WaylandObject
	SourceX, SourceY, SourceWidth, SourceHeight WaylandFixed
	DestWidth, DestHeight int32
	SourceSet, DestSet bool
}

func (w *WpViewport) Destroy() error {
	return nil
}

func (*WpViewport) DashboardShouldDisplay() bool {
	return true
}

func (*WpViewport) DashboardCategory() string {
	return "Viewports"
}

func (w *WpViewport) DashboardPrint(printer func(string, ...interface{})) error {
	var viewportStr []string
	viewportStr = append(viewportStr, fmt.Sprintf("surface: %s",
		w.Surface.String()))
	if w.SourceSet {
		viewportStr = append(viewportStr, fmt.Sprintf("source: x=%f y=%f w=%f h=%f",
			w.SourceX.ToDouble(), w.SourceY.ToDouble(), w.SourceWidth.ToDouble(), w.SourceHeight.ToDouble()))
	}
	if w.DestSet {
		viewportStr = append(viewportStr, fmt.Sprintf("dest: w=%d h=%d", w.DestWidth, w.DestHeight))
	}
	printer("%s - %s, %s", Indent(0), w.Object, strings.Join(viewportStr, ", "))
	return nil
}

type WpViewportImpl struct {
	client *Client
}

func RegisterWpViewport(client *Client) {
	r := &WpViewportImpl{
		client: client,
	}
	client.Impls["wp_viewport"] = r
}

func (w *WpViewportImpl) Request(packet *WaylandPacket) error {
	obj, ok := w.client.ObjectMap[packet.ObjectId]
	if !ok {
		return errors.New("no such viewport")
	}
	data := obj.Data.(*WpViewport)

	// State is not double-buffered yet :(
	switch packet.Opcode {
	case 0: // destroy
	case 1: // set_source
		x, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		y, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		width, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		height, err := packet.ReadFixed()
		if err != nil {
			return err
		}
		data.SourceX = x
		data.SourceY = y
		data.SourceWidth = width
		data.SourceHeight = height
		data.SourceSet = true
	case 2: // set_destination
		width, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		height, err := packet.ReadInt32()
		if err != nil {
			return err
		}
		data.DestWidth = width
		data.DestHeight = height
		data.DestSet = true
	}
	return nil
}

func (w *WpViewportImpl) Event(packet *WaylandPacket) error {
	return errors.New("wp_viewport has no events")
}

type WpViewporter struct {
	Object *WaylandObject
}

func (w *WpViewporter) Destroy() error {
	return nil
}

type WpViewporterImpl struct {
	client *Client
}

func RegisterWpViewporter(client *Client) {
	r := &WpViewporterImpl{
		client: client,
	}
	client.Impls["wp_viewporter"] = r
}

func (w *WpViewporterImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_viewport
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
		obj := w.client.NewObject(oid, "wp_viewport")
		obj.Data = &WpViewport{
			Object: obj,
			Surface: sobj,
		}
	}
	return nil
}

func (w *WpViewporterImpl) Event(packet *WaylandPacket) error {
	return errors.New("wp_viewporter has no events")
}

