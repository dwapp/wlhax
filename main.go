package main

import (
	"os"
	"time"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

func main() {
	var proxyDisplay string
	if len(os.Args) > 1 {
		proxyDisplay = os.Args[1]
	} else {
		proxyDisplay = "wlhax-0"
	}
	remoteDisplay, ok := os.LookupEnv("WAYLAND_DISPLAY")
	if !ok {
		panic("No WAYLAND_DISPLAY set")
	}

	proxy, err := NewProxy(proxyDisplay, remoteDisplay)
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
