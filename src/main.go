package main

import (
	"context"
	"os"
	"strconv"

	"github.com/rivo/tview"
	"golang.design/x/mainthread"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	mainthread.Init(realMain)
}

func ScanLoop() {
	err := adapter.Enable()
	if err != nil {
		println("error enabling adapter:", err.Error())
		return
	}

	// flex := tview.NewFlex()
	// flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Scanning...").SetTitleAlign(tview.AlignLeft)

	seenDevices := make(map[string]struct{})
	tree := tview.NewTreeView()

	root := tview.NewTreeNode("Devices")
	tree.SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true).SetTitle("Scanning...").SetTitleAlign(tview.AlignLeft)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		node.SetExpanded(!node.IsExpanded())
	})

	app := tview.NewApplication().SetRoot(tree, true)

	go func() {
		adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			deviceID := device.Address.String()
			if _, found := seenDevices[deviceID]; !found {
				seenDevices[deviceID] = struct{}{}

				if (len(os.Args) > 2 && os.Args[2] == "all") || device.LocalName() != "" {
					heading := deviceID
					if device.LocalName() != "" {
						heading = device.LocalName()
					}
					node := tview.NewTreeNode(heading)
					node.SetExpanded(false)
					node.SetSelectable(true)
					node.SetReference(device)

					node.AddChild(tview.NewTreeNode("ID: " + deviceID))
					if device.LocalName() != "" {
						node.AddChild(tview.NewTreeNode("Name: " + device.LocalName()))
					}
					node.AddChild(tview.NewTreeNode("RSSI: " + strconv.Itoa(int(device.RSSI))))
					node.AddChild(tview.NewTreeNode("MAC: " + device.Address.MAC.String()))

					app.QueueUpdate(func() {
						root.AddChild(node)
					})
				}
			}
		})
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func realMain() {
	if len(os.Args) > 1 && os.Args[1] == "scan" {
		ScanLoop()
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

		controller, err := NewLightController(config)
		if err != nil {
			println("error creating light controller:", err.Error())
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		controller.FindLightLoop(ctx, *adapter)

		terminal_ui := NewTerminalUI()
		cleanup := func() {
			cancel()
			controller.Disconnect()
			terminal_ui.app.Stop()
		}

		terminal_ui.Setup(config, controller.GetAppStatus(), controller.GetSacnStatus(), controller.GetLightStatuses(), cleanup)
		terminal_ui.Update(ctx)

		controller.SendLoop(ctx)
		controller.Listen(ctx)

		if err := terminal_ui.app.Run(); err != nil {
			panic(err)
		}

	}
}
