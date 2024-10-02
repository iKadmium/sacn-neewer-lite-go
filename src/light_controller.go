package main

import (
	"context"
	"fmt"
	"time"

	sacn "sacn_neewer_lite_go/sacn"

	"tinygo.org/x/bluetooth"
)

type LightController struct {
	sacnClient *sacn.SacnClient
	lights     map[string]*NeewerLight
}

func NewLightController() *LightController {
	return &LightController{
		sacnClient: nil,
		lights:     make(map[string]*NeewerLight),
	}
}

func (lc *LightController) Bind(config *Config) error {
	client, err := sacn.NewSacnClient(config.GetUniverses())
	if err != nil {
		return err
	}
	lc.sacnClient = client

	for _, lightConfig := range config.Lights {
		idBytes, err := bluetooth.ParseMAC(lightConfig.ID)
		if err != nil {
			return fmt.Errorf("invalid light ID %s: %v", lightConfig.ID, err)
		}
		light := NewLight(idBytes, lightConfig.Universe, lightConfig.Address)
		lc.lights[lightConfig.ID] = light
	}

	return nil
}

func (lc *LightController) handlePacket(packet *sacn.SacnDmxPacket) {
	for _, light := range lc.lights {
		if light.GetUniverse() == packet.Universe {
			red := packet.DmxData[light.GetAddress()]
			green := packet.DmxData[light.GetAddress()+1]
			blue := packet.DmxData[light.GetAddress()+2]
			light.SetColorRGB(red, green, blue)
		}
	}
}

func (lc *LightController) Listen(ctx context.Context) error {
	println("Listening for sACN packets...")

	select {
	case <-ctx.Done():
		fmt.Println("Received SIGTERM, shutting down...")
		return nil
	default:
		lc.sacnClient.Listen(lc.handlePacket)
	}

	return nil
}

func (lc *LightController) FindLightLoop(adapter bluetooth.Adapter) {
	println("scanning...")

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			l := lc.lights[device.Address.String()]
			if l != nil && !l.IsConnected() && device.Address.MAC == l.id {
				fmt.Println("Found device: ", l.id)
				l.Connect(device, adapter)

				println("connected to device:", device.Address.String())
			}
		})
		if err != nil {
			println("error scanning:", err.Error())
		}
	}()
}

func (lc *LightController) SendLoop() {
	for _, light := range lc.lights {
		go func(light *NeewerLight) {
			light.SendLoop(time.Millisecond * 50)
		}(light)
		go func(light *NeewerLight) {
			light.HeartbeatLoop(time.Second * 2)
		}(light)
	}
}

func (lc *LightController) Disconnect() error {
	for _, light := range lc.lights {
		if err := light.Disconnect(); err != nil {
			return err
		}
	}
	return lc.sacnClient.Disconnect()
}

func Scan(adapter bluetooth.Adapter) error {
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	})

	return err
}

// func main() {
// 	// Example usage
// 	config := LoadConfig("config.json")
// 	controller := NewLightController()
// 	if err := controller.Bind(config); err != nil {
// 		fmt.Fprintf(os.Stderr, "Error binding controller: %v\n", err)
// 		return
// 	}

// 	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
// 	defer stop()

// 	go func() {
// 		if err := controller.Listen(ctx); err != nil {
// 			fmt.Fprintf(os.Stderr, "Error listening: %v\n", err)
// 		}
// 	}()

// 	controller.FindLightLoop()

// 	if err := controller.Disconnect(); err != nil {
// 		fmt.Fprintf(os.Stderr, "Error disconnecting: %v\n", err)
// 	}
// }
