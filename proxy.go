package main

import (
	"errors"
	"net"
	"os"
	"path"
	"time"

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
		// TODO: Forward file descriptors
		// Shoutout to Golang for making that annoying
		// TODO: Read packet size and forward packets one at a time, so we can
		// decode & log them later
		// TODO: Share client/server code
		buf := make([]byte, 4096)
		for {
			n, err := remote.Read(buf)
			if err != nil {
				client.Close(err)
				return
			}

			conn.Write(buf[:n])
			proxy.onUpdate()
		}
	}()

	// Client loop
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				client.Close(err)
				return
			}
			remote.Write(buf[:n])
			proxy.onUpdate()
		}
	}()
}

func (client *Client) Close(err error) {
	client.conn.Close()
	client.remote.Close()
	client.Timestamp = time.Now()
	client.Err = err
}
