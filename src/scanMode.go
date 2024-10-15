package main

import (
	"context"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tinygo.org/x/bluetooth"
)

func ScanMode() {
	err := adapter.Enable()
	if err != nil {
		println("error enabling adapter:", err.Error())
		return
	}

	seenDevices := make(map[string]struct{})
	config := ConfigFromFileOrNew("data/config.json")

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	app := tview.NewApplication().SetRoot(flex, true)

	tree := tview.NewTreeView()
	treeRoot := tview.NewTreeNode("Devices")
	tree.SetRoot(treeRoot).SetCurrentNode(treeRoot)
	tree.SetBorder(true).SetTitle("Scanning...").SetTitleAlign(tview.AlignLeft)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		node.SetExpanded(!node.IsExpanded())
	})
	app.SetFocus(tree)
	flex.AddItem(tree, 0, 1, true)

	instructions := tview.NewTextView()
	instructions.SetText("Press 'a' to add a light, 'r' to remove a light, 's' to save and quit, 'q' to quit without saving")
	flex.AddItem(instructions, 1, 1, false)

	// Auto-add lights from the config
	for _, light := range config.Lights {
		deviceID := light.ID
		heading := deviceID + " -> Universe: " + strconv.Itoa(int(light.Universe)) + ", Address: " + strconv.Itoa(int(light.Address))
		node := tview.NewTreeNode(heading)
		node.SetExpanded(false)
		node.SetSelectable(true)
		node.SetColor(tcell.ColorGreen)
		node.AddChild(tview.NewTreeNode("ID: " + deviceID))
		node.AddChild(tview.NewTreeNode("Universe: " + strconv.Itoa(int(light.Universe))))
		node.AddChild(tview.NewTreeNode("Address: " + strconv.Itoa(int(light.Address))))
		treeRoot.AddChild(node)
	}

	_, cancel := context.WithCancel(context.Background())

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			cancel()
			app.Stop()
		}
		if event.Rune() == 'a' {
			addLight(config, flex, tree, app)
		}
		if event.Rune() == 'r' {
			removeLight(config, flex, tree, app)
		}
		if event.Rune() == 's' {
			save(config)
			cancel()
			app.Stop()
		}
		return event
	})

	go adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		deviceID := device.Address.String()
		if _, found := seenDevices[deviceID]; !found {
			seenDevices[deviceID] = struct{}{}

			if (len(os.Args) > 2 && os.Args[2] == "all") || device.LocalName() != "" {
				inConfig, universe, address := lightIsInConfig(config, device)
				heading := getDeviceName(device, inConfig, universe, address)

				// find the node in the tree
				app.QueueUpdate(func() {
					var node *tview.TreeNode
					for _, child := range treeRoot.GetChildren() {

						childText := child.GetText()
						if len(childText) >= len(deviceID) {
							id := childText[:len(deviceID)]
							if id == deviceID {
								node = child
								node.SetText(heading)
								break
							}
						}
					}
					if node == nil {
						node = tview.NewTreeNode(heading)
					}
					node.SetExpanded(false)
					node.SetSelectable(true)
					node.SetReference(device)
					if inConfig {
						node.SetColor(tcell.ColorGreen)
					}

					node.AddChild(tview.NewTreeNode("ID: " + deviceID))
					if device.LocalName() != "" {
						node.AddChild(tview.NewTreeNode("Name: " + device.LocalName()))
					}
					node.AddChild(tview.NewTreeNode("RSSI: " + strconv.Itoa(int(device.RSSI))))
					node.AddChild(tview.NewTreeNode("MAC: " + device.Address.MAC.String()))

					if !inConfig {
						treeRoot.AddChild(node)
					}
				})
			}
		}
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func save(config *Config) {
	err := config.Save("data/config.json")

	if err != nil {
		println("error saving config:", err.Error())
	}
}

func lightIsInConfig(config *Config, device bluetooth.ScanResult) (bool, uint16, uint16) {
	for _, light := range config.Lights {
		if light.ID == device.Address.String() {
			return true, light.Universe, light.Address
		}
	}
	return false, 0, 0
}

func getDeviceName(device bluetooth.ScanResult, inConfig bool, universe uint16, address uint16) string {
	heading := device.Address.String()
	if device.LocalName() != "" {
		heading = device.LocalName()
	}
	if inConfig {
		heading += " -> Universe: " + strconv.Itoa(int(universe)) + ", Address: " + strconv.Itoa(int(address))
	}
	return heading
}

func addLight(config *Config, flex *tview.Flex, tree *tview.TreeView, app *tview.Application) {
	form := tview.NewForm().
		AddInputField("Universe", "", 20, nil, nil).
		AddInputField("Address", "", 20, nil, nil)
	form.
		AddButton("Save", func() {
			universeStr := form.GetFormItemByLabel("Universe").(*tview.InputField).GetText()
			addressStr := form.GetFormItemByLabel("Address").(*tview.InputField).GetText()

			selected := tree.GetCurrentNode()
			selectedReference := selected.GetReference().(bluetooth.ScanResult)

			universe, err := strconv.Atoi(universeStr)
			if err != nil {
				println("Invalid universe:", universeStr)
				return
			}
			address, err := strconv.Atoi(addressStr)
			if err != nil {
				println("Invalid address:", addressStr)
				return
			}

			config.Lights = append(config.Lights, LightConfig{
				ID:       selectedReference.Address.String(),
				Universe: uint16(universe),
				Address:  uint16(address),
			})

			selected.SetColor(tcell.ColorGreen)
			selected.SetText(selected.GetText() + " -> Universe: " + universeStr + ", Address:" + addressStr)
			app.SetRoot(flex, true).SetFocus(tree)
		}).
		AddButton("Cancel", func() {
			app.SetRoot(flex, true).SetFocus(tree)
		})

	form.SetBorder(true).SetTitle("Enter Details").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true).SetFocus(form)
}

func removeLight(config *Config, flex *tview.Flex, tree *tview.TreeView, app *tview.Application) {
	selected := tree.GetCurrentNode()
	selectedReference := selected.GetReference().(bluetooth.ScanResult)

	for i, light := range config.Lights {
		if light.ID == selectedReference.Address.String() {
			config.Lights = append(config.Lights[:i], config.Lights[i+1:]...)
			selected.SetColor(tcell.ColorWhite)
			selected.SetText(getDeviceName(selectedReference, false, 0, 0))
			app.SetRoot(flex, true).SetFocus(tree)
			return
		}
	}
}
