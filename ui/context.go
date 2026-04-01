package ui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
)

// A context allows you to draw in a sub-region of the terminal
type Context struct {
	window vaxis.Window
	x, y   int
}

func (ctx *Context) Offset() (int, int) {
	return ctx.x, ctx.y
}

func (ctx *Context) Width() int {
	width, _ := ctx.window.Size()
	return width
}

func (ctx *Context) Height() int {
	_, height := ctx.window.Size()
	return height
}

// returns the vaxis Window for this context
func (ctx *Context) Window() vaxis.Window {
	return ctx.window
}

func NewContext(vx *vaxis.Vaxis) *Context {
	win := vx.Window()
	return &Context{win, 0, 0}
}

func (ctx *Context) Subcontext(x, y, width, height int) *Context {
	if x < 0 || y < 0 {
		panic(fmt.Errorf("Attempted to create context with negative offset"))
	}
	win := ctx.window.New(x, y, width, height)
	return &Context{win, x, y}
}

func (ctx *Context) SetCell(x, y int, ch rune, style vaxis.Style) {
	width, height := ctx.window.Size()
	if x >= width || y >= height {
		// no-op when dims are inadequate
		return
	}
	ctx.window.SetCell(x, y, vaxis.Cell{
		Character: vaxis.Character{
			Grapheme: string(ch),
		},
		Style: style,
	})
}

func (ctx *Context) Printf(x, y int, style vaxis.Style, format string, a ...interface{}) int {
	width, height := ctx.window.Size()

	if x >= width || y >= height {
		// no-op when dims are inadequate
		return 0
	}

	str := fmt.Sprintf(format, a...)
	runes := []rune(str)

	for i, r := range runes {
		if x+i >= width {
			return i
		}
		ctx.SetCell(x+i, y, r, style)
	}

	return len(runes)
}

func (ctx *Context) Fill(x, y, width, height int, rune rune, style vaxis.Style) {
	win := ctx.window.New(x, y, width, height)
	win.Fill(vaxis.Cell{
		Character: vaxis.Character{
			Grapheme: string(rune),
			Width:    1,
		},
		Style: style,
	})
}

func (ctx *Context) SetCursor(x, y int, style vaxis.CursorStyle) {
	ctx.window.ShowCursor(x, y, style)
}

func (ctx *Context) HideCursor() {
	ctx.window.Vx.HideCursor()
}

func (ctx *Context) Size() (int, int) {
	return ctx.window.Size()
}
