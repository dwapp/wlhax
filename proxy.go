package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/kyoh86/xdg"
	"golang.org/x/sys/unix"
)

// TODO: Support synthetic servers without forwarding events, so users can
// manually send requests to clients
type Proxy struct {
	listener      net.Listener
	proxyDisplay  string
	remoteDisplay string
	remotePath    string
	onUpdate      func(*Client)
	onConnect     func(*Client)
	onDisconnect  func(*Client)

	Clients []*Client

	SlowMode bool
	Block    bool
}

type Implementation interface {
	Request(packet *WaylandPacket) error
	Event(packet *WaylandPacket) error
}

type Destroyable interface {
	Destroy() error
}

type Client struct {
	conn   *net.UnixConn
	proxy  *Proxy
	remote *net.UnixConn
	pid    int32

	Err       error
	Timestamp time.Time

	RxLog []*WaylandPacket
	TxLog []*WaylandPacket

	Objects   []*WaylandObject
	ObjectMap map[uint32]*WaylandObject
	Globals   []*WaylandGlobal
	GlobalMap map[uint32]*WaylandGlobal

	lock sync.RWMutex

	Impls map[string]Implementation
}

func (c *Client) String() string {
	return fmt.Sprintf("Client { pid: %d }", c.pid)
}

func (c *Client) Pid() int32 {
	return c.pid
}

type WaylandObject struct {
	ObjectId  uint32
	Interface string
	Data      Destroyable
}

func (wo *WaylandObject) String() string {
	if wo == nil {
		return "<nil object>"
	}
	return fmt.Sprintf("%s@%d", wo.Interface, wo.ObjectId)
}

func NewProxy(proxyDisplay, remoteDisplay string) (*Proxy, error) {
	var proxyPath string
	if !path.IsAbs(proxyDisplay) {
		proxyPath = path.Join(xdg.RuntimeDir(), proxyDisplay)
	} else {
		proxyPath = proxyDisplay
	}

	l, err := net.Listen("unix", proxyPath)
	if err != nil {
		return nil, err
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

func (proxy *Proxy) CloseWrite() {
	for _, client := range proxy.Clients {
		client.conn.CloseWrite()
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

func (proxy *Proxy) OnUpdate(onUpdate func(*Client)) {
	proxy.onUpdate = onUpdate
}

func (proxy *Proxy) OnConnect(onConnect func(*Client)) {
	proxy.onConnect = onConnect
}

func (proxy *Proxy) OnDisconnect(onDisconnect func(*Client)) {
	proxy.onDisconnect = onDisconnect
}

func (proxy *Proxy) handleClient(conn net.Conn) {
	wl_display := &WaylandObject{
		Interface: "wl_display",
		ObjectId:  1,
	}

	client := &Client{
		conn: conn.(*net.UnixConn),

		Timestamp: time.Now(),

		Objects: []*WaylandObject{wl_display},
		ObjectMap: map[uint32]*WaylandObject{
			1: wl_display,
		},

		Globals:   nil,
		GlobalMap: make(map[uint32]*WaylandGlobal),

		Impls: make(map[string]Implementation),
	}

	pid, _ := getPidOfConn(client.conn)
	client.pid = pid

	proxy.Clients = append(proxy.Clients, client)

	RegisterWlDisplay(client)
	RegisterWlRegistry(client)
	RegisterWlOutput(client)
	RegisterWlBuffer(client)
	RegisterWlShm(client)
	RegisterWlShmPool(client)
	RegisterWlCallback(client)
	RegisterWlSeat(client)
	RegisterWlKeyboard(client)
	RegisterWlPointer(client)
	RegisterWlTouch(client)
	RegisterWlCompositor(client)
	RegisterWlSubCompositor(client)
	RegisterWlSurface(client)
	RegisterWlSubSurface(client)
	RegisterXdgWmBase(client)
	RegisterXdgPositioner(client)
	RegisterXdgSurface(client)
	RegisterXdgToplevel(client)
	RegisterXdgPopup(client)
	RegisterZwpLinuxDmabuf(client)
	RegisterZwpLinuxBufferParams(client)
	RegisterWpSinglePixelBufferManager(client)
	RegisterWpViewporter(client)
	RegisterWpViewport(client)
	RegisterWpFractionalScaleManager(client)
	RegisterWpFractionalScale(client)
	RegisterWpIdleInhibitManager(client)
	RegisterWpIdleInhibitor(client)

	remote, err := net.Dial("unix", proxy.remotePath)
	if err != nil {
		proxy.onUpdate(nil)
		client.Close(err)
		return
	}

	client.remote = remote.(*net.UnixConn)
	client.proxy = proxy
	proxy.onConnect(client)

	// Remote loop
	go func() {
		for {
			packet, err := ReadPacket(client.remote)
			if err != nil {
				client.Close(err)
				return
			}
			client.lock.Lock()
			client.RecordRx(packet)
			client.lock.Unlock()
			client.proxy.onUpdate(client)
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
			client.lock.Lock()
			client.RecordTx(packet)
			client.lock.Unlock()
			for proxy.Block {
				time.Sleep(500 * time.Millisecond)
			}
			client.proxy.onUpdate(client)
			err = packet.WritePacket(client.remote)
			if err != nil {
				client.Close(err)
				return
			}
			for _, fd := range packet.Fds {
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
	client.proxy.onDisconnect(client)
}

func (client *Client) RemoveObject(objectId uint32) {
	if o, ok := client.ObjectMap[objectId]; ok {
		for idx := range client.Objects {
			if client.Objects[idx] == o {
				client.Objects = append(client.Objects[:idx], client.Objects[idx+1:]...)
				break
			}
		}
		if o.Data != nil {
			o.Data.Destroy()
		}
		delete(client.ObjectMap, objectId)
	}
}

func (client *Client) NewObject(objectId uint32, iface string) *WaylandObject {
	object := &WaylandObject{
		Interface: iface,
		ObjectId:  objectId,
	}
	client.RemoveObject(objectId)
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
		if impl, ok := client.Impls[object.Interface]; ok {
			if err := impl.Event(packet); err != nil {
				client.Close(err)
			}
		}
	}
}

func (client *Client) RecordTx(packet *WaylandPacket) {
	client.TxLog = append(client.TxLog, packet)

	// Fallback for objects with unknown interfaces
	if object, ok := client.ObjectMap[packet.ObjectId]; !ok {
		client.NewObject(packet.ObjectId, "(unknown)")
	} else {
		if impl, ok := client.Impls[object.Interface]; ok {
			if err := impl.Request(packet); err != nil {
				client.Close(err)
			}
		}
	}
}
