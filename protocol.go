package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"net"

	"golang.org/x/sys/unix"
)

type WaylandPacket struct {
	ObjectId  uint32
	Length    uint16
	Opcode    uint16
	Arguments []byte
	Fds       []uintptr

	buffer *bytes.Buffer
}

type WaylandGlobal struct {
	Interface string
	GlobalId  uint32
	Version   uint32
}

type WaylandFixed int32

func (f WaylandFixed) ToInt32() int32 {
	return int32(f) / 256
}

func (f WaylandFixed) ToDouble() float64 {
	var i int64
	i = ((1023 + 44) << 52) + (1 << 51) + int64(f)
	d := math.Float64frombits(uint64(i))
	return d - (3 << 43)
}

func ReadPacket(conn *net.UnixConn) (*WaylandPacket, error) {
	var fds []uintptr
	var buf [8]byte
	control := make([]byte, 128)

	n, oobn, _, _, err := conn.ReadMsgUnix(buf[:], control)
	if err != nil {
		return nil, err
	}
	if n != 8 {
		return nil, errors.New("Unable to read message header")
	}
	if oobn > 0 {
		if oobn > len(control) {
			return nil, errors.New("Control message buffer undersized")
		}

		ctrl, err := unix.ParseSocketControlMessage(control[:oobn])
		if err != nil {
			return nil, err
		}
		for _, msg := range ctrl {
			_fds, err := unix.ParseUnixRights(&msg)
			if err != nil {
				return nil, errors.New("Unable to parse unix rights")
			}
			for _, fd := range _fds {
				fds = append(fds, uintptr(fd))
			}
		}
	}

	packet := &WaylandPacket{
		ObjectId: binary.LittleEndian.Uint32(buf[0:4]),
		Opcode:   binary.LittleEndian.Uint16(buf[4:6]),
		Length:   binary.LittleEndian.Uint16(buf[6:8]),
		Fds:      fds,
	}

	packet.Arguments = make([]byte, packet.Length-8)

	n, err = conn.Read(packet.Arguments)
	if err != nil {
		return nil, err
	}
	if int(packet.Length-8) != len(packet.Arguments) {
		return nil, errors.New("Buffer is shorter than expected length")
	}

	packet.Reset()
	return packet, nil
}

func (packet WaylandPacket) WritePacket(conn *net.UnixConn) error {
	var header bytes.Buffer
	size := uint32(len(packet.Arguments) + 8)

	binary.Write(&header, binary.LittleEndian, uint32(packet.ObjectId))
	binary.Write(&header, binary.LittleEndian,
		uint32(size<<16|uint32(packet.Opcode)&0x0000ffff))

	var oob []byte
	for _, fd := range packet.Fds {
		rights := unix.UnixRights(int(fd))
		oob = append(oob, rights...)
	}

	body := append(header.Bytes(), packet.Arguments...)
	d, c, err := conn.WriteMsgUnix(body, oob, nil)
	if err != nil {
		return err
	}
	if c != len(oob) || d != len(body) {
		return errors.New("WriteMsgUnix failed")
	}

	return nil
}

func (packet *WaylandPacket) Reset() {
	packet.buffer = bytes.NewBuffer(packet.Arguments)
}

func szup(n uint32) uint32 {
	if n%4 == 0 {
		return n
	}
	return n + (4 - (n % 4))
}

func (packet *WaylandPacket) ReadInt32() (int32, error) {
	var out int32
	err := binary.Read(packet.buffer, binary.LittleEndian, &out)
	return out, err
}

func (packet *WaylandPacket) ReadUint32() (uint32, error) {
	var out uint32
	err := binary.Read(packet.buffer, binary.LittleEndian, &out)
	return out, err
}

func (packet *WaylandPacket) ReadFixed() (WaylandFixed, error) {
	var out int32
	err := binary.Read(packet.buffer, binary.LittleEndian, &out)
	return WaylandFixed(out), err
}

func (packet *WaylandPacket) ReadString() (string, error) {
	var l uint32
	err := binary.Read(packet.buffer, binary.LittleEndian, &l)
	if err != nil {
		return "", err
	}
	var pl uint32 = szup(l)
	buf := make([]byte, pl)
	n, err := packet.buffer.Read(buf)
	if err != nil {
		return "", err
	}
	if n != int(pl) {
		return "", errors.New("ReadString underread")
	}
	return string(buf[:l-1]), nil
}

func (packet *WaylandPacket) Data() []byte {
	return packet.buffer.Bytes()
}
