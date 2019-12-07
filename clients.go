package main

import (
	"github.com/gdamore/tcell"
	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
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
			"%p: %s since %s",
			client, status, client.Timestamp.Format("15:04:05"))
		ctx.Fill(w, y, ctx.Width() - w, 1, ' ', style)
		y++
	}
}

func (clients *ClientsView) Invalidate() {
	clients.DoInvalidate(clients)
}
