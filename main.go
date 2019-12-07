package main

import (
	"time"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

func main() {
	proxy, err := NewProxy()
	if err != nil {
		panic(err)
	}
	go proxy.Run()
	defer proxy.Close()

	dash := NewDashboard(proxy)

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
