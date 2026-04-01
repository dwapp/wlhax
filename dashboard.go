package main

import (
	"fmt"
	"os/exec"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/dwapp/wlhax/ui"
)

type Dashboard struct {
	focused ui.Interactive
	grid    *ui.Grid
	onExit  func()
	proxy   *Proxy
	status  *ui.Stack
	tabs    *ui.Tabs
	tabMap  map[*Client]*ClientView
}

func NewDashboard(proxy *Proxy) *Dashboard {
	clients := NewClientsView(proxy)

	tabs := ui.NewTabs()
	tabs.Add(clients, "Connections", true)

	status := ui.NewStack()
	status.Push(ui.NewText(
		fmt.Sprintf("WAYLAND_DISPLAY=%s -> %s",
			proxy.ProxyDisplay(), proxy.RemoteDisplay()),
		vaxis.Style{Foreground: vaxis.RGBColor(0, 255, 0)}))

	grid := ui.NewGrid().Rows([]ui.GridSpec{
		{Strategy: ui.SIZE_EXACT, Size: ui.Const(1)},
		{Strategy: ui.SIZE_WEIGHT, Size: ui.Const(1)},
		{Strategy: ui.SIZE_EXACT, Size: ui.Const(1)},
	}).Columns([]ui.GridSpec{
		{Strategy: ui.SIZE_EXACT, Size: ui.Const(11)},
		{Strategy: ui.SIZE_WEIGHT, Size: ui.Const(1)},
	})
	grid.AddChild(ui.NewText("   wlhax   ", vaxis.Style{
		Foreground: vaxis.RGBColor(245, 248, 250),
		Background: vaxis.RGBColor(0, 96, 128),
	}))
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

func (dash *Dashboard) Draw(ctx *ui.Context) {
	dash.grid.Draw(ctx)
}

func (dash *Dashboard) OnInvalidate(callback func(d ui.Drawable)) {
	// Note: OnInvalidate handling changed in new UI system
	// Direct invalidation is now handled globally
	ui.Invalidate()
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

	interactive, ok := dash.tabs.Selected().Content.(ui.Interactive)
	if ok {
		if interactive.Event(event) {
			return true
		}
	}
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

func (dash *Dashboard) MouseEvent(localX int, localY int, event vaxis.Event) {
	dash.grid.MouseEvent(localX, localY, event)
}

func (dash *Dashboard) OnBeep(func() error) {
	// This space deliberately left blank
}

func (dash *Dashboard) OnExit(exit func()) {
	dash.onExit = exit
}

func (dash *Dashboard) focus(item ui.Interactive) {
	if dash.focused == item {
		return
	}
	if dash.focused != nil {
		dash.focused.Focus(false)
	}
	dash.focused = item

	interactive, ok := dash.tabs.Selected().Content.(ui.Interactive)
	if item != nil {
		item.Focus(true)
		if ok {
			interactive.Focus(false)
		}
	} else {
		if ok {
			interactive.Focus(true)
		}
	}
}

func (dash *Dashboard) BeginExCommand(cmd string) {
	previous := dash.focused
	exline := NewExLine(cmd, func(cmd string) {
		if len(cmd) == 0 {
			return
		}
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
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
