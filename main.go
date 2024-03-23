package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	libui "git.sr.ht/~sircmpwn/aerc/lib/ui"
)

func main() {
	remoteDisplay, ok := os.LookupEnv("WAYLAND_DISPLAY")
	if !ok {
		panic("No WAYLAND_DISPLAY set")
	}

	var (
		err   error
		path  string
		proxy *Proxy
	)
	for idx := 0; idx < 10; idx++ {
		path = fmt.Sprintf("wlhax-%d", idx)
		if proxy, err = NewProxy(path, remoteDisplay); err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	defer os.Remove(path)
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
