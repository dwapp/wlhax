package main

import (
	"errors"
	"fmt"
)

type WpCursorShapeDevice struct {
	Object         	*WaylandObject
	Pointer        	*WlPointer
	serial          *uint32
	shape           *uint32 // wp_cursor_shape_device_v1.shape
}

func (w *WpCursorShapeDevice) Destroy() error {
	return nil
}

func (*WpCursorShapeDevice) DashboardShouldDisplay() bool {
	return true
}

func (*WpCursorShapeDevice) DashboardCategory() string {
	return "Cursor shape"
}

var shapeName = [...]string{"default", "context_menu", "help", "pointer", "progress", "wait", "cell",
	"crosshair", "text", "vertical_text", "alias", "copy", "move", "no_drop", "not_allowed",
	"grab", "grabbing", "e_resize", "n_resize", "ne_resize", "nw_resize", "s_resize", "se_resize",
	"sw_resize", "w_resize", "ew_resize", "ns_resize", "nesw_resize", "nwse_resize", "col_resize",
	"row_resize", "all_scroll", "zoom_in", "zoom_out" }

func (w *WpCursorShapeDevice) DashboardPrint(printer func(string, ...interface{})) error {
	if w.shape == nil {
		printer("%s - %s, shape: not set", Indent(0), w.Object)
    } else if int(*w.shape) > len(shapeName) {
		printer("%s - %s, shape(unknown id): %d", Indent(0), *w.shape)
	} else {
		printer("%s - %s, shape: %s", Indent(0), w.Object, shapeName[*w.shape-1])
	}

	return nil
}

type WpCursorShapeDeviceImpl struct {
	client *Client
}

func RegisterCursorShapeDevice(client *Client) {
	r := &WpCursorShapeDeviceImpl{
		client: client,
	}
	client.Impls["wp_cursor_shape_device_v1"] = r
}

func (w *WpCursorShapeDeviceImpl) Request(packet *WaylandPacket) error {
	object := w.client.ObjectMap[packet.ObjectId]
	shapeDevice, ok := object.Data.(*WpCursorShapeDevice)
	if !ok {
		return errors.New("object is not wp_cursor_shape_device_v1")
	}
	switch packet.Opcode {
	case 0: // destroy
	case 1: // set_shape
		serial, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		shape, err := packet.ReadUint32()
		if err != nil {
			return err
		}
		shapeDevice.serial = &serial
		shapeDevice.shape = &shape
	}
	return nil
}

func (w *WpCursorShapeDeviceImpl) Event(packet *WaylandPacket) error {
	return errors.New("wp_cursor_shape_device_v1 has no events")
}

type WpCursorShapeManager struct {
	Object *WaylandObject
}

func (w *WpCursorShapeManager) Destroy() error {
	return nil
}

type WpCursorShapeManagerImpl struct {
	client *Client
}

func RegisterCursorShapeManager(client *Client) {
	r := &WpCursorShapeManagerImpl{
		client: client,
	}
	client.Impls["wp_cursor_shape_manager_v1"] = r
}

func (w *WpCursorShapeManagerImpl) Request(packet *WaylandPacket) error {
	switch packet.Opcode {
	case 0: // destroy
	case 1: // get_pointer
		oid, err := packet.ReadUint32() // new_id<wp_cursor_shape_device_v1>
		if err != nil {
			return err
		}
		pid, err := packet.ReadUint32() // object<wl_pointer>
		if err != nil {
			return err
		}
		pobj, ok := w.client.ObjectMap[pid]
		if !ok {
			return fmt.Errorf("no such pointer object: %d", pid)
		}
		pointer, ok := pobj.Data.(*WlPointer)
		if !ok {
			return fmt.Errorf("object is not wl_pointer: %d", pid)
		}

		obj := w.client.NewObject(oid, "wp_cursor_shape_device_v1")
		obj.Data = &WpCursorShapeDevice {
			Object:  obj,
			Pointer: pointer,
		}

	case 2: // get_tablet_tool_v2
		// Not implemented
	}
	return nil
}

func (w *WpCursorShapeManagerImpl) Event(packet *WaylandPacket) error {
	return errors.New("wp_cursor_shape_manager_v1 has no events")
}
