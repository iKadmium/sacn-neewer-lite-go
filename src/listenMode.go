package main

import "context"

func ListenMode() {
	config, err := ConfigFromFile("data/config.json")
	if err != nil {
		println("error loading config:", err.Error())
		println("To create a new config, run with the 'scan' argument")
		return
	}
	err = adapter.Enable()
	if err != nil {
		println("error enabling adapter:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	controller, err := NewLightController(config, ctx)
	if err != nil {
		println("error creating light controller:", err.Error())
		cancel()
		return
	}
	controller.FindLightLoop(ctx, *adapter)

	terminal_ui := NewTerminalUI()
	cleanup := func() {
		cancel()
		controller.Disconnect()
		terminal_ui.app.Stop()
	}

	terminal_ui.SetupSend(config, controller.GetAppStatus(), controller.GetSacnStatus(), controller.GetLightStatuses(), cleanup)
	terminal_ui.UpdateSend(ctx)

	controller.SendLoop(ctx)
	controller.Listen(ctx)

	if err := terminal_ui.app.Run(); err != nil {
		panic(err)
	}
}
