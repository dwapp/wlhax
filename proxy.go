package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/sys/unix"
	"github.com/kyoh86/xdg"
)

// TODO: Support synthetic servers without forwarding events, so users can
// manually send requests to clients
type Proxy struct {
	listener      net.Listener
	proxyDisplay  string
	remoteDisplay string
	remotePath    string
	onUpdate      func()

	Clients []*Client
}

type Client struct {
	conn   *net.UnixConn
	proxy  *Proxy
	remote *net.UnixConn

	Timestamp time.Time
	Err       error
}

type WaylandPacket struct {
	ObjectId  uint32
	Length    uint16
	Opcode    uint16
	Arguments []byte
	Fds       []uintptr
}

func NewProxy() (*Proxy, error) {
	// TODO: Allow multiple wlhax servers to be running? (Who cares?)
	proxyDisplay := "wlhax-0"

	proxyPath := path.Join(xdg.RuntimeDir(), proxyDisplay)
	l, err := net.Listen("unix", proxyPath)
	if err != nil {
		return nil, err
	}

	remoteDisplay, ok := os.LookupEnv("WAYLAND_DISPLAY")
	if !ok {
		return nil, errors.New("No WAYLAND_DISPLAY set, who do we proxy to?")
	}

	var remotePath string
	if !path.IsAbs(remoteDisplay) {
		remotePath = path.Join(xdg.RuntimeDir(), remoteDisplay)
	} else {
		remotePath = remoteDisplay
	}

	os.Setenv("WAYLAND_DISPLAY", proxyDisplay)

	return &Proxy{
		listener:      l,
		proxyDisplay:  proxyDisplay,
		remoteDisplay: remoteDisplay,
		remotePath:    remotePath,
	}, nil
}

func (proxy *Proxy) Run() {
	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			// XXX: Maybe we should tell someone that happened
			return
		}
		go proxy.handleClient(conn)
	}
}

func (proxy *Proxy) ProxyDisplay() string {
	return proxy.proxyDisplay
}

func (proxy *Proxy) RemoteDisplay() string {
	return proxy.remoteDisplay
}

func (proxy *Proxy) Close() {
	proxy.listener.Close()
}

func (proxy *Proxy) OnUpdate(onUpdate func()) {
	proxy.onUpdate = onUpdate
}

func ReadPacket(conn *net.UnixConn) (*WaylandPacket, error) {
	var fds []uintptr
	var buf [8]byte
	control := make([]byte, 24)

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

		ctrl, err := unix.ParseSocketControlMessage(control)
		if err != nil {
			return nil, err
		}
		for _, msg := range ctrl {
			_fds, err := unix.ParseUnixRights(&msg)
			if err != nil {
				return nil, errors.New("Unable to parse unix rights")
			}
			if len(_fds) != 1 {
				return nil, errors.New("Unexpectedly got >1 file descriptor")
			}
			fds = append(fds, uintptr(_fds[0]))
		}
	}

	packet := &WaylandPacket{
		ObjectId: binary.LittleEndian.Uint32(buf[0:4]),
		Opcode:   binary.LittleEndian.Uint16(buf[4:6]),
		Length:   binary.LittleEndian.Uint16(buf[6:8]),
		Fds:      fds,
	}

	packet.Arguments = make([]byte, packet.Length - 8)

	n, err = conn.Read(packet.Arguments)
	if err != nil {
		return nil, err
	}
	if int(packet.Length - 8) != len(packet.Arguments) {
		return nil, errors.New("Buffer is shorter than expected length")
	}

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

func (proxy *Proxy) handleClient(conn net.Conn) {
	client := &Client{
		conn: conn.(*net.UnixConn),

		Timestamp: time.Now(),
	}
	proxy.Clients = append(proxy.Clients, client)

	remote, err := net.Dial("unix", proxy.remotePath)
	if err != nil {
		proxy.onUpdate()
		client.Close(err)
		return
	}

	client.remote = remote.(*net.UnixConn)
	client.proxy = proxy
	proxy.onUpdate()

	// Remote loop
	go func() {
		for {
			packet, err := ReadPacket(client.remote)
			if err != nil {
				client.Close(err)
				return
			}
			err = packet.WritePacket(client.conn)
			if err != nil {
				client.Close(err)
				return
			}
			proxy.onUpdate()
		}
	}()

	// Client loop
	go func() {
		for {
			packet, err := ReadPacket(client.conn)
			if err != nil {
				client.Close(err)
				return
			}
			err = packet.WritePacket(client.remote)
			if err != nil {
				client.Close(err)
				return
			}
			proxy.onUpdate()
		}
	}()
}

func (client *Client) Close(err error) {
	client.conn.Close()
	client.remote.Close()
	client.Timestamp = time.Now()
	client.Err = err
	client.proxy.onUpdate()
}
