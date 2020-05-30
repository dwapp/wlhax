package main

import (
	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
	"github.com/gdamore/tcell"
)

const (
	intro string = `Waiting for clients...

Welcome to wlhax! Use the arrow keys to navigate between tabs and menu items
(or hjkl), and Enter to select. Quit with escape or 'q'.

A Wayland display is running at the Unix socket shown at the bottom of the
screen. You can start Wayland clients pointing to this address manually, or use
:exec <command>... to have wlhax start one for you.
`
)

type ClientsView struct {
	libui.Invalidatable
	selected int
	proxy    *Proxy
}

func NewClientsView(proxy *Proxy) *ClientsView {
	return &ClientsView{
		proxy: proxy,
	}
}

func (clients *ClientsView) showSurface(ctx *libui.Context, surface *WlSurface, y, depth int) int {
	prefix := ""
	for i := 0; i <= depth; i++ {
		prefix += "   "
	}
	prefix += " - "
	ctx.Printf(0, y, tcell.StyleDefault,
		"%ssurface: %s, frames: %d, pending frames: %d", prefix, surface, surface.Frames, surface.PendingFrames)
	y++
	if y >= ctx.Height() {
		return y
	}
	for _, child := range surface.Current.Children {
		y = clients.showSurface(ctx, child.Surface, y, depth+1)
		if y >= ctx.Height() {
			return y
		}
	}
	return y
}

func (clients *ClientsView) Draw(ctx *libui.Context) {
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', tcell.StyleDefault)

	proxy := clients.proxy
	if len(proxy.Clients) == 0 {
		ctx.Printf(0, 0, tcell.StyleDefault, "%s", intro)
		return
	}

	// TODO: Scrolling
	y := 0
	for i := 0; i < len(proxy.Clients) && y < ctx.Height(); i++ {
		client := proxy.Clients[i]
		status := "Connected"
		if client.Err != nil {
			status = client.Err.Error()
		}
		style := tcell.StyleDefault
		if clients.selected == i {
			style = style.Reverse(true)
		}
		w := ctx.Printf(0, y, style,
			"%p: %s", client, status)
		ctx.Fill(w, y, ctx.Width()-w, 1, ' ', style)
		y++
		statusStyle := style.Reverse(false).Foreground(tcell.ColorGreen)
		if client.Err != nil {
			statusStyle = statusStyle.Foreground(tcell.ColorRed)
		}
		surfaces := 0
		w = ctx.Printf(0, y, statusStyle, "  since %s  ",
			client.Timestamp.Format("15:04:05"))
		w += ctx.Printf(w, y, style,
			"rx: %-6d tx: %-6d globals: %-4d objects: %-4d surfaces: %-4d",
			len(client.RxLog), len(client.TxLog),
			len(client.Globals), len(client.Objects), surfaces)
		ctx.Fill(w, y, ctx.Width()-w, 1, ' ', style)
		y++
		for _, obj := range client.Objects {
			if obj.Interface == "wl_surface" {
				surface := obj.Data.(*WlSurface)
				if surface.Current.Parent != nil {
					continue
				}
				y = clients.showSurface(ctx, surface, y, 0)
				if y >= ctx.Height() {
					break
				}
			}
		}
	}
}

func (clients *ClientsView) Invalidate() {
	clients.DoInvalidate(clients)
}

// TODO: Scrolling
func (clients *ClientsView) SelectNext() {
	clients.selected += 1
	if clients.selected >= len(clients.proxy.Clients) {
		clients.selected = len(clients.proxy.Clients) - 1
	}
	clients.Invalidate()
}

func (clients *ClientsView) SelectPrev() {
	clients.selected -= 1
	if clients.selected < 0 {
		clients.selected = 0
	}
	clients.Invalidate()
}

func (clients *ClientsView) Focus(focus bool) {
	// This space deliberately left blank
}

func (clients *ClientsView) Event(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyDown:
			clients.SelectNext()
			return true
		case tcell.KeyUp:
			clients.SelectPrev()
			return true
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				clients.SelectNext()
				return true
			case 'k':
				clients.SelectPrev()
				return true
			}
		}
	}
	return false
}
