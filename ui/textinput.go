package ui

import (
	"strings"

	"git.sr.ht/~rockorager/vaxis"
)

type TextInput struct {
	ctx      *Context
	text     []rune
	index    int
	password bool
	prompt   string
	change   []func(ti *TextInput)
	focus    bool
	style    vaxis.Style
}

func NewTextInput(text string, style vaxis.Style) *TextInput {
	return &TextInput{
		text:  []rune(text),
		index: len([]rune(text)),
		style: style,
	}
}

func (ti *TextInput) Prompt(prompt string) *TextInput {
	ti.prompt = prompt
	return ti
}

func (ti *TextInput) Set(value string) *TextInput {
	ti.text = []rune(value)
	ti.index = len(ti.text)
	return ti
}

func (ti *TextInput) String() string {
	return string(ti.text)
}

func (ti *TextInput) StringLeft() string {
	return string(ti.text[:ti.index])
}

func (ti *TextInput) StringRight() string {
	return string(ti.text[ti.index:])
}

func (ti *TextInput) insertAtCursor(r rune) {
	left := ti.text[:ti.index]
	right := ti.text[ti.index:]
	ti.text = append(left, append([]rune{r}, right...)...)
	ti.index++
}

func (ti *TextInput) deleteAtCursor() {
	if ti.index > 0 {
		ti.text = append(ti.text[:ti.index-1], ti.text[ti.index:]...)
		ti.index--
	}
}

func (ti *TextInput) deleteForward() {
	if ti.index < len(ti.text) {
		ti.text = append(ti.text[:ti.index], ti.text[ti.index+1:]...)
	}
}

func (ti *TextInput) Draw(ctx *Context) {
	ti.ctx = ctx
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', ti.style)

	text := ti.prompt + string(ti.text)
	if ti.password {
		text = ti.prompt + strings.Repeat("*", len(ti.text))
	}

	ctx.Printf(0, 0, ti.style, "%s", text)

	if ti.focus {
		cursorPos := len([]rune(ti.prompt)) + ti.index
		if cursorPos < ctx.Width() {
			ctx.SetCursor(cursorPos, 0, vaxis.CursorBeamBlinking)
		}
	}
}

func (ti *TextInput) Focus(focus bool) {
	ti.focus = focus
	if !focus && ti.ctx != nil {
		ti.ctx.HideCursor()
	}
	ti.Invalidate()
}

func (ti *TextInput) Event(event vaxis.Event) bool {
	if !ti.focus {
		return false
	}

	if key, ok := event.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyBackspace):
			ti.deleteAtCursor()
		case key.Matches(vaxis.KeyDelete):
			ti.deleteForward()
		case key.Matches(vaxis.KeyLeft):
			if ti.index > 0 {
				ti.index--
			}
		case key.Matches(vaxis.KeyRight):
			if ti.index < len(ti.text) {
				ti.index++
			}
		case key.Matches(vaxis.KeyHome):
			ti.index = 0
		case key.Matches(vaxis.KeyEnd):
			ti.index = len(ti.text)
		default:
			if key.Text != "" && len(key.Text) == 1 {
				r := []rune(key.Text)[0]
				if r >= 32 { // printable character
					ti.insertAtCursor(r)
				}
			}
		}

		for _, change := range ti.change {
			change(ti)
		}

		ti.Invalidate()
		return true
	}

	return false
}

func (ti *TextInput) OnChange(onChange func(ti *TextInput)) {
	ti.change = append(ti.change, onChange)
}

func (ti *TextInput) Invalidate() {
	Invalidate()
}
