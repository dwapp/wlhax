package main

import (
	"sort"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/dwapp/wlhax/ui"
)

type ClientView struct {
	selected        int
	currentCategory string
	client          *Client
	viewportHeight  int
	currentLines    int
	scroll          int
	folded          map[string]bool
	lineCategories  map[int]string
}

func NewClientView(client *Client) *ClientView {
	return &ClientView{
		client:         client,
		folded:         make(map[string]bool),
		lineCategories: make(map[int]string),
	}
}

func Indent(depth int) string {
	var indent strings.Builder
	for i := 0; i <= depth; i++ {
		indent.WriteString("   ")
	}
	return indent.String()
}

type DashboardDisplayable interface {
	DashboardCategory() string
	DashboardShouldDisplay() bool
	DashboardPrint(func(string, ...interface{})) error
}

func (c *ClientView) Draw(ctx *ui.Context) {
	c.viewportHeight = ctx.Height()
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', vaxis.Style{})

	y := 0

	printerWithStyle := func(style vaxis.Style, formatter string, v ...interface{}) {
		if y < c.scroll || y-c.scroll >= ctx.Height() {
			y++
			return
		}
		if c.selected == y {
			style.Attribute = vaxis.AttrReverse
		}
		w := ctx.Printf(0, y-c.scroll, style, formatter, v...)
		ctx.Fill(w, y-c.scroll, ctx.Width()-w, 1, ' ', style)
		y++
	}
	printer := func(formatter string, v ...interface{}) {
		style := vaxis.Style{}
		printerWithStyle(style, formatter, v...)
	}

	// TODO: Scrolling
	client := c.client
	status := "Connected"
	if client.Err != nil {
		status = client.Err.Error()
	}

	statusStyle := vaxis.Style{
		Foreground: vaxis.RGBColor(0, 255, 0),
	}
	if client.Err != nil {
		statusStyle.Foreground = vaxis.RGBColor(255, 0, 0)
	}
	printerWithStyle(statusStyle, "%s  since %s  rx: %-6d tx: %-6d globals: %-4d objects: %-4d",
		status, client.Timestamp.Format("15:04:05"), len(client.RxLog), len(client.TxLog),
		len(client.Globals), len(client.Objects))

	var categories []string
	sorted := make(map[string][]DashboardDisplayable)
	for _, obj := range client.Objects {
		if obj.Data == nil {
			continue
		}
		displayable, ok := obj.Data.(DashboardDisplayable)
		if !ok || !displayable.DashboardShouldDisplay() {
			continue
		}

		category := displayable.DashboardCategory()
		arr := sorted[category]
		if arr == nil {
			categories = append(categories, category)
		}
		sorted[category] = append(arr, displayable)
	}
	sort.Sort(sort.StringSlice(categories))
	c.currentCategory = ""
	c.lineCategories = make(map[int]string, len(categories))
	for _, category := range categories {
		if y == c.selected {
			c.currentCategory = category
		}
		c.lineCategories[y] = category
		printerWithStyle(vaxis.Style{Foreground: vaxis.IndexColor(226)}, category)

		if c.folded[category] {
			continue
		}
		children := sorted[category]
		for _, child := range children {
			child.DashboardPrint(printer)
		}
	}
	c.currentLines = y
}

func (client *ClientView) Invalidate() {
	ui.Invalidate()
}

func (client *ClientView) SelectNext(inc int) {
	client.selected += inc
	if client.selected >= client.currentLines {
		client.selected = client.currentLines - 1
	}
	if client.selected >= client.scroll+client.viewportHeight {
		client.scroll = client.selected - client.viewportHeight + 1
	}
	client.Invalidate()
}

func (client *ClientView) SelectPrev(inc int) {
	client.selected -= inc
	if client.selected < 0 {
		client.selected = 0
	}
	if client.selected < client.scroll {
		client.scroll = client.selected
	}
	client.Invalidate()
}

func (client *ClientView) Toggle() {
	if client.currentCategory == "" {
		return
	}
	client.folded[client.currentCategory] = !client.folded[client.currentCategory]
	client.Invalidate()
}

func (client *ClientView) ScrollBy(delta int) {
	client.scroll += delta
	maxScroll := client.currentLines - client.viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if client.scroll < 0 {
		client.scroll = 0
	}
	if client.scroll > maxScroll {
		client.scroll = maxScroll
	}
	if client.selected < client.scroll {
		client.selected = client.scroll
	}
	if client.viewportHeight > 0 && client.selected >= client.scroll+client.viewportHeight {
		client.selected = client.scroll + client.viewportHeight - 1
	}
	if client.selected >= client.currentLines {
		client.selected = client.currentLines - 1
	}
	if client.selected < 0 {
		client.selected = 0
	}
	client.Invalidate()
}

func (client *ClientView) Focus(focus bool) {
	// This space deliberately left blank
}

func (client *ClientView) Event(event vaxis.Event) bool {
	if key, ok := event.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyDown):
			client.SelectNext(1)
			return true
		case key.Matches(vaxis.KeyUp):
			client.SelectPrev(1)
			return true
		case key.Matches(vaxis.KeyPgDown):
			client.SelectNext(client.viewportHeight)
			return true
		case key.Matches(vaxis.KeyPgUp):
			client.SelectPrev(client.viewportHeight)
			return true
		case key.Matches('j'):
			client.SelectNext(1)
			return true
		case key.Matches('k'):
			client.SelectPrev(1)
			return true
		case key.Matches(' '):
			client.Toggle()
			return true
		}
	}
	return false
}

func (client *ClientView) MouseEvent(localX int, localY int, event vaxis.Event) {
	mouse, ok := event.(vaxis.Mouse)
	if !ok || mouse.EventType != vaxis.EventPress {
		return
	}

	switch mouse.Button {
	case vaxis.MouseWheelUp:
		client.ScrollBy(-3)
		return
	case vaxis.MouseWheelDown:
		client.ScrollBy(3)
		return
	case vaxis.MouseLeftButton:
	default:
		return
	}

	line := client.scroll + localY
	if line < 0 || line >= client.currentLines {
		return
	}

	client.selected = line
	if category, ok := client.lineCategories[line]; ok {
		client.currentCategory = category
		client.folded[category] = !client.folded[category]
	}
	client.Invalidate()
}
