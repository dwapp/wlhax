package main

import (
	"github.com/gdamore/tcell"
	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

type Dashboard struct {
	onExit func()
	grid   *libui.Grid
}

func NewDashboard() *Dashboard {
	grid := libui.NewGrid().Rows([]libui.GridSpec{
		{libui.SIZE_WEIGHT, 1},
	}).Columns([]libui.GridSpec{
		{libui.SIZE_WEIGHT, 1},
	})
	grid.AddChild(libui.NewText("Hello world!"))

	return &Dashboard{
		grid: grid,
	}
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
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyESC:
			if dash.onExit != nil {
				dash.onExit()
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
