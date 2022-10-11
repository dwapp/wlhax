package main

import (
	"os"
	"time"
	"os/exec"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

const proxyDisplay = "wlhax-0"

func main() {
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

	if len(os.Args) > 1 {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Start()
	}

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
