package main

import (
	"time"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

func main() {
	dash := NewDashboard()

	ui, err := libui.Initialize(dash)
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	dash.OnExit(ui.Exit)

	for !ui.ShouldExit() {
		if !ui.Tick() {
			time.Sleep(16 * time.Millisecond)
		}
	}
}
