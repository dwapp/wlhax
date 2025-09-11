package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dwapp/wlhax/ui"
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

	err = ui.Initialize(dash)
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	dash.OnExit(ui.Exit)

	// Main event loop
loop:
	for {
		select {
		case event := <-ui.Events:
			ui.HandleEvent(event)
		case callback := <-ui.Callbacks:
			callback()
		case <-ui.Redraw:
			ui.Render()
		case <-ui.SuspendQueue:
			if err := ui.Suspend(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to suspend UI: %v\n", err)
				break loop
			}
		case <-ui.Quit:
			break loop
		}
	}
}
