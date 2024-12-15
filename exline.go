package main

import (
	config "git.sr.ht/~rjarry/aerc/config"
	"git.sr.ht/~rjarry/aerc/lib/ui"
	"git.sr.ht/~rockorager/vaxis"
)

type ExLine struct {
	commit func(cmd string)
	finish func()
	input  *ui.TextInput
}

func NewExLine(cmd string, commit func(cmd string),
	finish func()) *ExLine {

	input := ui.NewTextInput("", config.Ui).Prompt(":").Set(cmd)
	exline := &ExLine{
		commit: commit,
		finish: finish,
		input:  input,
	}
	/*
	input.OnInvalidate(func(d ui.Drawable) {
		ui.Invalidate()
	})*/
	return exline
}

func NewPrompt(prompt string, commit func(text string)) *ExLine {

	input := ui.NewTextInput("", config.Ui).Prompt(prompt)
	exline := &ExLine{
		commit: commit,
		input:  input,
	}
	/*
	input.OnInvalidate(func(d ui.Drawable) {
		exline.Invalidate()
	})
	*/
	return exline
}

func (ex *ExLine) Invalidate() {
	//ex.DoInvalidate(ex)
}

func (ex *ExLine) Draw(ctx *ui.Context) {
	ex.input.Draw(ctx)
}

func (ex *ExLine) Focus(focus bool) {
	ex.input.Focus(focus)
}

func (ex *ExLine) Event(event vaxis.Event) bool {
	if key, ok := event.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyEnter), key.Matches('j', vaxis.ModCtrl):
			cmd := ex.input.String()
			ex.input.Focus(false)
			ex.commit(cmd)
			ex.finish()
		case key.Matches(vaxis.KeyEsc), key.Matches('c', vaxis.ModCtrl):
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
