package main

import (
	"github.com/gdamore/tcell"

	"git.sr.ht/~sircmpwn/aerc/lib/ui"
)

type ExLine struct {
	ui.Invalidatable
	commit      func(cmd string)
	finish      func()
	input       *ui.TextInput
}

func NewExLine(cmd string, commit func(cmd string),
	finish func()) *ExLine {

	input := ui.NewTextInput("").Prompt(":").Set(cmd)
	exline := &ExLine{
		commit:      commit,
		finish:      finish,
		input:       input,
	}
	input.OnInvalidate(func(d ui.Drawable) {
		exline.Invalidate()
	})
	return exline
}

func NewPrompt(prompt string, commit func(text string)) *ExLine {

	input := ui.NewTextInput("").Prompt(prompt)
	exline := &ExLine{
		commit:      commit,
		input:       input,
	}
	input.OnInvalidate(func(d ui.Drawable) {
		exline.Invalidate()
	})
	return exline
}

func (ex *ExLine) Invalidate() {
	ex.DoInvalidate(ex)
}

func (ex *ExLine) Draw(ctx *ui.Context) {
	ex.input.Draw(ctx)
}

func (ex *ExLine) Focus(focus bool) {
	ex.input.Focus(focus)
}

func (ex *ExLine) Event(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyCtrlJ:
			cmd := ex.input.String()
			ex.input.Focus(false)
			ex.commit(cmd)
			ex.finish()
		case tcell.KeyEsc, tcell.KeyCtrlC:
			ex.input.Focus(false)
			ex.finish()
		default:
			return ex.input.Event(event)
		}
	}
	return true
}

type nullHistory struct {
	input *ui.TextInput
}

func (*nullHistory) Add(string) {}

func (h *nullHistory) Next() string {
	return h.input.String()
}

func (h *nullHistory) Prev() string {
	return h.input.String()
}

func (*nullHistory) Reset() {}
