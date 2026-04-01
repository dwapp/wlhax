package main

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/dwapp/wlhax/ui"
)

const (
	intro string = `Waiting for clients...

Welcome to wlhax! Use the arrow keys to navigate between tabs and menu items
(or hjkl), and Enter to select. Quit by typing :quit or :q.

A Wayland display is running at the Unix socket shown at the bottom of the
screen. You can start Wayland clients pointing to this address manually, or use
:exec <command>... to have wlhax start one for you.

Commands: exec, slow, fast, clear, block, unblock, quit
`
)

type ClientsView struct {
	selected       int
	scroll         int
	viewportHeight int
	proxy          *Proxy
}

func NewClientsView(proxy *Proxy) *ClientsView {
	return &ClientsView{
		proxy: proxy,
	}
}

func (clients *ClientsView) Draw(ctx *ui.Context) {
	clients.viewportHeight = ctx.Height()
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', vaxis.Style{})

	proxy := clients.proxy
	if len(proxy.Clients) == 0 {
		ctx.Printf(0, 0, vaxis.Style{}, "%s", intro)
		return
	}

	y := 0
	for i := clients.scroll; i < len(proxy.Clients) && y < ctx.Height(); i++ {
		client := proxy.Clients[i]
		status := "Connected"
		if client.Err != nil {
			status = client.Err.Error()
		}
		style := vaxis.Style{}
		if clients.selected == i {
			style.Attribute = vaxis.AttrReverse
		}
		w := ctx.Printf(0, y, style,
			"Client %d: %s", client.Pid(), status)
		ctx.Fill(w, y, ctx.Width()-w, 1, ' ', style)
		y++
		statusStyle := style
		statusStyle.Attribute = vaxis.AttrNone
		statusStyle.Foreground = vaxis.RGBColor(0, 255, 0)
		if client.Err != nil {
			statusStyle.Foreground = vaxis.RGBColor(255, 0, 0)
		}
		w = ctx.Printf(0, y, statusStyle, "  since %s  ",
			client.Timestamp.Format("15:04:05"))
		w += ctx.Printf(w, y, style,
			"rx: %-6d tx: %-6d globals: %-4d objects: %-4d",
			len(client.RxLog), len(client.TxLog),
			len(client.Globals), len(client.Objects))
		ctx.Fill(w, y, ctx.Width()-w, 1, ' ', style)
		y++
	}
}

func (clients *ClientsView) Invalidate() {
	ui.Invalidate()
}

func (clients *ClientsView) SelectNext() {
	clients.selected += 1
	if clients.selected >= len(clients.proxy.Clients) {
		clients.selected = len(clients.proxy.Clients) - 1
	}
	clients.ensureSelectionVisible()
	clients.Invalidate()
}

func (clients *ClientsView) SelectPrev() {
	clients.selected -= 1
	if clients.selected < 0 {
		clients.selected = 0
	}
	clients.ensureSelectionVisible()
	clients.Invalidate()
}

func (clients *ClientsView) ScrollBy(delta int) {
	if len(clients.proxy.Clients) == 0 {
		return
	}

	clients.scroll += delta
	maxScroll := len(clients.proxy.Clients) - clients.visibleItems()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if clients.scroll < 0 {
		clients.scroll = 0
	}
	if clients.scroll > maxScroll {
		clients.scroll = maxScroll
	}
	clients.ensureSelectionVisible()
	clients.Invalidate()
}

func (clients *ClientsView) visibleItems() int {
	if clients.viewportHeight <= 0 {
		return 1
	}
	visible := clients.viewportHeight / 2
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (clients *ClientsView) ensureSelectionVisible() {
	visible := clients.visibleItems()
	if clients.selected < clients.scroll {
		clients.scroll = clients.selected
	}
	if clients.selected >= clients.scroll+visible {
		clients.scroll = clients.selected - visible + 1
	}
	if clients.scroll < 0 {
		clients.scroll = 0
	}
}

func (clients *ClientsView) Focus(focus bool) {
	// This space deliberately left blank
}

func (clients *ClientsView) Event(event vaxis.Event) bool {
	if key, ok := event.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyDown):
			clients.SelectNext()
			return true
		case key.Matches(vaxis.KeyUp):
			clients.SelectPrev()
			return true
		case key.Matches('j'):
			clients.SelectNext()
			return true
		case key.Matches('k'):
			clients.SelectPrev()
			return true
		}
	}
	return false
}

func (clients *ClientsView) MouseEvent(localX int, localY int, event vaxis.Event) {
	mouse, ok := event.(vaxis.Mouse)
	if !ok || mouse.EventType != vaxis.EventPress {
		return
	}

	switch mouse.Button {
	case vaxis.MouseWheelUp:
		clients.ScrollBy(-1)
		return
	case vaxis.MouseWheelDown:
		clients.ScrollBy(1)
		return
	case vaxis.MouseLeftButton:
		index := clients.scroll + localY/2
		if index < 0 || index >= len(clients.proxy.Clients) {
			return
		}
		clients.selected = index
		clients.ensureSelectionVisible()
		clients.Invalidate()
	}
}
