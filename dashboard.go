package main

import (
	"fmt"
	"os/exec"

	libui "git.sr.ht/~rjarry/aerc/lib/ui"
	config "git.sr.ht/~rjarry/aerc/config"
	"github.com/google/shlex"
	"git.sr.ht/~rockorager/vaxis"
)

type Dashboard struct {
	focused libui.Interactive
	grid    *libui.Grid
	onExit  func()
	proxy   *Proxy
	status  *libui.Stack
	tabs    *libui.Tabs
	tabMap  map[*Client]*ClientView
}

func NewDashboard(proxy *Proxy) *Dashboard {
	clients := NewClientsView(proxy)

	config.Ui.LoadStyle()

	tabs := libui.NewTabs(func(d libui.Drawable) *config.UIConfig {
		return config.Ui
	})
	tabs.Add(clients, "Connections", true)

	status := libui.NewStack(config.Ui)
	status.Push(libui.NewText(
		fmt.Sprintf("WAYLAND_DISPLAY=%s -> %s",
			proxy.ProxyDisplay(), proxy.RemoteDisplay()),
		config.Ui.GetStyle(config.STYLE_DEFAULT)))
	/*.Reverse(true))*/

	grid := libui.NewGrid().Rows([]libui.GridSpec{
		{Strategy: libui.SIZE_EXACT, Size: libui.Const(1)},
		{Strategy: libui.SIZE_WEIGHT, Size: libui.Const(1)},
		{Strategy: libui.SIZE_EXACT, Size: libui.Const(1)},
	}).Columns([]libui.GridSpec{
		{Strategy: libui.SIZE_EXACT, Size: libui.Const(11)},
		{Strategy: libui.SIZE_WEIGHT, Size: libui.Const(1)},
	})
	grid.AddChild(libui.NewText("   wlhax   ", config.Ui.GetStyle(config.STYLE_HEADER))) //.Reverse(true))
	grid.AddChild(tabs.TabStrip).At(0, 1)
	grid.AddChild(tabs.TabContent).At(1, 0).Span(1, 2)
	grid.AddChild(status).At(2, 0).Span(1, 2)

	dash := &Dashboard{
		tabMap: make(map[*Client]*ClientView),
		grid:   grid,
		proxy:  proxy,
		tabs:   tabs,
		status: status,
	}
	dash.focus(nil)
	proxy.OnUpdate(func(c *Client) {
		clients.Invalidate()
		v := dash.tabMap[c]
		if v != nil {
			v.Invalidate()
		}
	})
	proxy.OnConnect(func(c *Client) {
		clients.Invalidate()
		v := dash.tabMap[c]
		if v != nil {
			// ???
			delete(dash.tabMap, c)
			tabs.Remove(v)
		}
		v = NewClientView(c)
		dash.tabMap[c] = v
		tabs.Add(v, fmt.Sprintf("Client %d", c.Pid()), false)
		// tabs.Select(len(tabs.Tabs) - 1)
	})
	proxy.OnDisconnect(func(c *Client) {
		clients.Invalidate()
		v := dash.tabMap[c]
		if v != nil {
			v.Invalidate()
		}
	})
	return dash
}

func (dash *Dashboard) Draw(ctx *libui.Context) {
	dash.grid.Draw(ctx)
}

func (dash *Dashboard) OnInvalidate(callback func(d libui.Drawable)) {
	//dash.grid.OnInvalidate(func(d libui.Drawable) {
	//	callback(dash)
	//})
}

func (dash *Dashboard) Invalidate() {
	dash.grid.Invalidate()
}

func (dash *Dashboard) Event(event vaxis.Event) bool {
	if dash.focused != nil {
		if dash.focused.Event(event) {
			return true
		}
	}
	/*
	interactive, ok := dash.tabs.Tabs[dash.tabs.Selected].
		Content.(libui.Interactive)
	if ok {
		if interactive.Event(event) {
			return true
		}
	}*/
	if key, ok := event.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyLeft):
			dash.tabs.PrevTab()
		case key.Matches(vaxis.KeyRight):
			dash.tabs.NextTab()
		case key.Matches('h'):
			dash.tabs.PrevTab()
		case key.Matches('l'):
			dash.tabs.NextTab()
		case key.Matches(':'):
			dash.BeginExCommand("")
		}
	}
	return false
}

func (dash *Dashboard) Focus(focus bool) {
	// This space deliberately left blank
}

func (dash *Dashboard) OnBeep(func() error) {
	// This space deliberately left blank
}

func (dash *Dashboard) OnExit(exit func()) {
	dash.onExit = exit
}

func (dash *Dashboard) focus(item libui.Interactive) {
	if dash.focused == item {
		return
	}
	if dash.focused != nil {
		dash.focused.Focus(false)
	}
	dash.focused = item
    /*
	interactive, ok := dash.tabs.tabs[dash.tabs.curIndex]
		Content.(libui.Interactive)
	if item != nil {
		item.Focus(true)
		if ok {
			interactive.Focus(false)
		}
	} else {
		if ok {
			interactive.Focus(true)
		}
	}*/
}

func (dash *Dashboard) BeginExCommand(cmd string) {
	previous := dash.focused
	exline := NewExLine(cmd, func(cmd string) {
		if len(cmd) == 0 {
			return
		}
		parts, err := shlex.Split(cmd)
		if err != nil {
			return
		}

		switch parts[0] {
		case "exec":
			if len(parts) < 2 {
				break
			}
			cmd := exec.Command(parts[1], parts[2:]...)
			cmd.Start()
		case "slow":
			dash.proxy.SlowMode = true
		case "fast":
			dash.proxy.SlowMode = false
		case "clear":
			var new_clients []*Client
			for _, client := range dash.proxy.Clients {
				if client.Err == nil {
					new_clients = append(new_clients, client)
				} else {
					v := dash.tabMap[client]
					if v != nil {
						delete(dash.tabMap, client)
						dash.tabs.Remove(v)
					}
				}
			}
			dash.proxy.Clients = new_clients
		case "block":
			dash.proxy.Block = true
		case "unblock":
			dash.proxy.Block = false
		case "closewrite":
			dash.proxy.CloseWrite()
		case "quit", "q":
			if dash.onExit != nil {
				dash.onExit()
			}
		}
	}, func() {
		dash.status.Pop()
		dash.focus(previous)
	})
	dash.status.Push(exline)
	dash.focus(exline)
}
