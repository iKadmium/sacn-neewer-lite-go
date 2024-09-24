package main

import (
	"context"
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
		err := adapter.Enable()
		if err != nil {
			println("error enabling adapter:", err.Error())
			return
		}

		println("scanning...")
		seenDevices := make(map[string]struct{})
		adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			deviceID := device.Address.String()
			if _, found := seenDevices[deviceID]; !found {
				seenDevices[deviceID] = struct{}{}

				if (len(os.Args) > 2 && os.Args[2] == "all") || device.LocalName() != "" {
					println("found new device:")
					println("\t", "ID:", deviceID)
					println("\t", "RSSI", device.RSSI)
					println("\t", "Name", device.LocalName())
					println("\t", "MAC", device.Address.MAC.String())

				}
			}
		})
		return
	} else {
		config, err := ConfigFromFile("data/config.json")
		if err != nil {
			println("error loading config:", err.Error())
			return
		}
		err = adapter.Enable()
		if err != nil {
			println("error enabling adapter:", err.Error())
			return
		}

		controller := NewLightController()
		ctx := context.Background()
		controller.Bind(config)
		controller.FindLightLoop(*adapter)
		controller.SendLoop()
		controller.Listen(ctx)

	}
}
