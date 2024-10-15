package main

import (
	"os"

	"golang.design/x/mainthread"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	mainthread.Init(realMain)
}

func realMain() {
	if len(os.Args) > 1 && os.Args[1] == "scan" {
		ScanMode()
	} else {
		ListenMode()
	}
}
