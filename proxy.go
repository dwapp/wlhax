package main

import (
	"net"
	"os"
	"path"
	"time"

	"github.com/kyoh86/xdg"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
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

	Err       error
	Timestamp time.Time

	RxLog []*WaylandPacket
	TxLog []*WaylandPacket

	Objects   []*WaylandObject
	ObjectMap map[uint32]*WaylandObject
	Globals   []*WaylandGlobal
	GlobalMap map[uint32]*WaylandGlobal
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

func (proxy *Proxy) handleClient(conn net.Conn) {
	wl_display := &WaylandObject{
		Interface: "wl_display",
		ObjectId: 1,
	}

	client := &Client{
		conn: conn.(*net.UnixConn),

		Timestamp: time.Now(),

		Objects: []*WaylandObject{wl_display},
		ObjectMap: map[uint32]*WaylandObject{
			1: wl_display,
		},

		Globals: nil,
		GlobalMap: make(map[uint32]*WaylandGlobal),
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
			client.RecordRx(packet)
			err = packet.WritePacket(client.conn)
			if err != nil {
				client.Close(err)
				return
			}
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
			client.RecordTx(packet)
			err = packet.WritePacket(client.remote)
			if err != nil {
				client.Close(err)
				return
			}
			for _, fd := range packet.Fds {
				// TODO: We don't necessarily want to close these always
				unix.Close(int(fd))
			}
		}
	}()
}

func (client *Client) Close(err error) {
	if client.Err == nil {
		client.Err = err
	}
	client.conn.Close()
	client.remote.Close()
	client.Timestamp = time.Now()
	client.proxy.onUpdate()
}

func (client *Client) NewObject(objectId uint32, iface string) *WaylandObject {
	object := &WaylandObject{
		Interface: iface,
		ObjectId:  objectId,
	}
	client.ObjectMap[objectId] = object
	client.Objects = append(client.Objects, object)
	return object
}

func (client *Client) RecordRx(packet *WaylandPacket) {
	client.RxLog = append(client.TxLog, packet)

	// Fallback for objects with unknown interfaces
	if object, ok := client.ObjectMap[packet.ObjectId]; !ok {
		client.NewObject(packet.ObjectId, "(unknown)")
	} else {
		// Interpret events we understand
		// TODO: Generalize this based on protocol XML
		packet.Reset()
		switch object.Interface {
		case "wl_registry":
			switch packet.Opcode {
			case 0:
				gid, err := packet.ReadUint32()
				if err != nil {
					client.Close(errors.Wrap(err, "wl_registry decode gid"))
					break
				}
				iface, err := packet.ReadString()
				if err != nil {
					client.Close(errors.Wrap(err, "wl_registry decode iface"))
					break
				}
				ver, err := packet.ReadUint32()
				if err != nil {
					client.Close(errors.Wrap(err, "wl_registry decode version"))
					break
				}
				global := &WaylandGlobal{
					GlobalId: gid,
					Interface: iface,
					Version: ver,
				}
				client.Globals = append(client.Globals, global)
				client.GlobalMap[gid] = global
			}
		}
	}

	client.proxy.onUpdate()
}

func (client *Client) RecordTx(packet *WaylandPacket) {
	client.TxLog = append(client.TxLog, packet)

	// Fallback for objects with unknown interfaces
	if object, ok := client.ObjectMap[packet.ObjectId]; !ok {
		client.NewObject(packet.ObjectId, "(unknown)")
	} else {
		// Interpret events we understand
		// TODO: Generalize this based on protocol XML
		packet.Reset()
		switch object.Interface {
		case "wl_display":
			switch packet.Opcode {
			case 1: // get_registry
				oid, err := packet.ReadUint32()
				if err != nil {
					client.Close(err)
					break
				}
				client.NewObject(oid, "wl_registry")
			}
		}
	}

	client.proxy.onUpdate()
}
