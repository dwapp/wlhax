package main

import (
	"fmt"
	"os"
	"os/exec"

	libui "git.sr.ht/~rjarry/aerc/lib/ui"
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

	err = libui.Initialize(dash)
	if err != nil {
		panic(err)
	}
	defer libui.Close()

	dash.OnExit(libui.Exit)

loop:
	for {
		select {
		case event := <-libui.Events:
			libui.HandleEvent(event)
		case callback := <-libui.Callbacks:
			callback()
		case <-libui.Redraw:
			libui.Render()
		case <-libui.SuspendQueue:
			err = libui.Suspend()
				if err != nil {
					//app.PushError(fmt.Sprintf("suspend: %s", err))
				}
		case <-libui.Quit:
			    // Do someing
				break loop
		}
	}
}
