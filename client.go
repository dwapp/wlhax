package main

import (
	"fmt"
	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
	"github.com/gdamore/tcell"
)

type ClientView struct {
	libui.Invalidatable
	selected int
	client   *Client
}

func NewClientView(client *Client) *ClientView {
	return &ClientView{
		client: client,
	}
}

func (client *ClientView) showSurface(ctx *libui.Context, surface *WlSurface, y, depth int) int {
	style := tcell.StyleDefault
	prefix := ""
	for i := 0; i <= depth; i++ {
		prefix += "   "
	}
	prefix += " - "

	rolestr := "<unknown>"
	suffix := ""
	details := ""
	if surface.Current.Role != nil {
		switch role := surface.Current.Role.(type) {
		case WlSubSurfaceState:
			suffix = fmt.Sprintf(", desync: %t, x: %d, y: %d", role.Desync, role.X, role.Y)
			rolestr = role.String()
		case WlPointerSurfaceState:
			rolestr = role.String()
		case XdgSurfaceState:
			rolestr = role.String()
			switch xdg_role := role.XdgRole.(type) {
			case XdgToplevelState:
				suffix = fmt.Sprintf(", app_id: %s, title: %s", xdg_role.AppId, xdg_role.Title)
				if xdg_role.Parent != nil {
					suffix = fmt.Sprintf("%s, parent: %s", suffix, xdg_role.Parent.Object.String())
				}
				if role.CurrentConfigure.Serial == role.PendingConfigure.Serial {
					details = fmt.Sprintf("current: w=%d h=%d", role.CurrentConfigure.Width, role.CurrentConfigure.Height)
				} else {
					details = fmt.Sprintf("current: w=%d h=%d, pending: w=%d h=%d", role.CurrentConfigure.Width, role.CurrentConfigure.Height, role.PendingConfigure.Width, role.PendingConfigure.Height)
				}
			case XdgPopupState:
				suffix = fmt.Sprintf(", parent: %s", xdg_role.XdgPopup.Parent.Object.String())
			}
		}
	}
	if client.selected == y {
		style = style.Reverse(true)
	}
	ctx.Printf(0, y, style,
		"%s%s, role: %s, buffers: %d, frames: %d/%d%s", prefix, surface.Object, rolestr, surface.Current.BufferNum, surface.Frames, surface.RequestedFrames, suffix)
	y++
	if details != "" && y < ctx.Height() {
		prefix := ""
		for i := 0; i <= depth+3; i++ {
			prefix += "   "
		}

		if client.selected == y {
			style = style.Reverse(true)
		}
		ctx.Printf(0, y, style,
			"%s%s", prefix, details)
		y++
	}
	for _, child := range surface.Current.Children {
		if y >= ctx.Height() {
			return y
		}
		y = client.showSurface(ctx, child.Surface, y, depth+1)
	}
	return y
}

func (c *ClientView) Draw(ctx *libui.Context) {
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', tcell.StyleDefault)

	// TODO: Scrolling
	y := 0
	client := c.client
	status := "Connected"
	if client.Err != nil {
		status = client.Err.Error()
	}
	style := tcell.StyleDefault
	if c.selected == y {
		style = style.Reverse(true)
	}
	statusStyle := style.Reverse(false).Foreground(tcell.ColorGreen)
	if client.Err != nil {
		statusStyle = statusStyle.Foreground(tcell.ColorRed)
	}
	w := ctx.Printf(0, y, statusStyle, "%s  since %s  ",
		status, client.Timestamp.Format("15:04:05"))
	w += ctx.Printf(w, y, style,
		"rx: %-6d tx: %-6d globals: %-4d objects: %-4d",
		len(client.RxLog), len(client.TxLog),
		len(client.Globals), len(client.Objects))
	ctx.Fill(w, y, ctx.Width()-w, 1, ' ', style)
	y++
	statusStyle = style.Reverse(false).Foreground(tcell.ColorYellow)
	ctx.Printf(0, y, statusStyle, "Surfaces")
	y++
	for _, obj := range client.Objects {
		if y >= ctx.Height() {
			break
		}
		switch obj.Interface {
		case "wl_surface":
			surface := obj.Data.(*WlSurface)
			if surface.Current.Parent != nil {
				continue
			}
			y = c.showSurface(ctx, surface, y, 0)
		}
	}
	statusStyle = style.Reverse(false).Foreground(tcell.ColorYellow)
	ctx.Printf(0, y, statusStyle, "Seat objects")
	y++
	for _, obj := range client.Objects {
		if y >= ctx.Height() {
			break
		}
		switch obj.Interface {
		case "wl_seat", "wl_keyboard", "wl_pointer", "wl_touch":
			style := tcell.StyleDefault
			if c.selected == y {
				style = style.Reverse(true)
			}
			ctx.Printf(0, y, style, "    - %v", obj)
			y++
		}
	}
}

func (client *ClientView) Invalidate() {
	client.DoInvalidate(client)
}

// TODO: Scrolling
func (client *ClientView) SelectNext() {
	client.selected += 1
	client.Invalidate()
}

func (client *ClientView) SelectPrev() {
	client.selected -= 1
	if client.selected < 0 {
		client.selected = 0
	}
	client.Invalidate()
}

func (client *ClientView) Focus(focus bool) {
	// This space deliberately left blank
}

func (client *ClientView) Event(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyDown:
			client.SelectNext()
			return true
		case tcell.KeyUp:
			client.SelectPrev()
			return true
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				client.SelectNext()
				return true
			case 'k':
				client.SelectPrev()
				return true
			}
		}
	}
	return false
}
