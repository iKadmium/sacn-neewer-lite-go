package main

import (
	"context"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {

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

	// println("scanning...")
	// err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	// 	println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	// })
	// if err != nil {
	// 	println("error scanning:", err.Error())
	// 	return
	// }

}
