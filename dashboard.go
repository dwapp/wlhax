package main

import (
	"fmt"
	"os/exec"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
	"github.com/gdamore/tcell"
	"github.com/google/shlex"
)

type Dashboard struct {
	focused libui.Interactive
	grid    *libui.Grid
	onExit  func()
	proxy   *Proxy
	status  *libui.Stack
	tabs    *libui.Tabs
}

func NewDashboard(proxy *Proxy) *Dashboard {
	clients := NewClientsView(proxy)

	tabs := libui.NewTabs()
	tabs.Add(clients, "Connections")

	status := libui.NewStack()
	status.Push(libui.NewText(
		fmt.Sprintf("WAYLAND_DISPLAY=%s -> %s",
			proxy.ProxyDisplay(), proxy.RemoteDisplay())).
		Reverse(true))

	grid := libui.NewGrid().Rows([]libui.GridSpec{
		{libui.SIZE_EXACT, 1},
		{libui.SIZE_WEIGHT, 1},
		{libui.SIZE_EXACT, 1},
	}).Columns([]libui.GridSpec{
		{libui.SIZE_EXACT, 11},
		{libui.SIZE_WEIGHT, 1},
	})
	grid.AddChild(libui.NewText("   wlhax   ").Reverse(true))
	grid.AddChild(tabs.TabStrip).At(0, 1)
	grid.AddChild(tabs.TabContent).At(1, 0).Span(1, 2)
	grid.AddChild(status).At(2, 0).Span(1, 2)

	dash := &Dashboard{
		grid:   grid,
		proxy:  proxy,
		tabs:   tabs,
		status: status,
	}
	dash.focus(nil)
	proxy.OnUpdate(func() {
		clients.Invalidate()
	})
	return dash
}

func (dash *Dashboard) Draw(ctx *libui.Context) {
	dash.grid.Draw(ctx)
}

func (dash *Dashboard) OnInvalidate(callback func(d libui.Drawable)) {
	dash.grid.OnInvalidate(func(d libui.Drawable) {
		callback(dash)
	})
}

func (dash *Dashboard) Invalidate() {
	dash.grid.Invalidate()
}

func (dash *Dashboard) Event(event tcell.Event) bool {
	if dash.focused != nil {
		if dash.focused.Event(event) {
			return true
		}
	}
	interactive, ok := dash.tabs.Tabs[dash.tabs.Selected].
		Content.(libui.Interactive)
	if ok {
		if interactive.Event(event) {
			return true
		}
	}
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyESC:
			if dash.onExit != nil {
				dash.onExit()
			}
		case tcell.KeyLeft:
			dash.tabs.PrevTab()
		case tcell.KeyRight:
			dash.tabs.NextTab()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				if dash.onExit != nil {
					dash.onExit()
				}
			case 'h':
				dash.tabs.PrevTab()
			case 'l':
				dash.tabs.NextTab()
			case ':':
				dash.BeginExCommand("")
			}
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
	interactive, ok := dash.tabs.Tabs[dash.tabs.Selected].
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
	}
}

func (dash *Dashboard) BeginExCommand(cmd string) {
	previous := dash.focused
	exline := NewExLine(cmd, func(cmd string) {
		parts, _ := shlex.Split(cmd)
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
		}
	}, func() {
		dash.status.Pop()
		dash.focus(previous)
	})
	dash.status.Push(exline)
	dash.focus(exline)
}
