package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"tinygo.org/x/bluetooth"
)

type LightController struct {
	sacnClient *SacnClient
	lights     []*Light
}

func NewLightController() *LightController {
	return &LightController{
		sacnClient: nil,
		lights:     []*Light{},
	}
}

func (lc *LightController) Bind(config *Config) error {
	client, err := NewSacnClient(config.GetUniverses())
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
		//light := NewLight(lightConfig.ID, lightConfig.Address, lightConfig.Universe)
		lc.lights = append(lc.lights, light)
	}

	return nil
}

func (lc *LightController) handlePacket(packet *SacnDmxPacket) error {
	for _, light := range lc.lights {
		if light.IsConnected() && light.GetUniverse() == packet.Universe {
			red := packet.DmxData[light.GetAddress()]
			green := packet.DmxData[light.GetAddress()+1]
			blue := packet.DmxData[light.GetAddress()+2]
			if err := light.SetColorRGB(red, green, blue); err != nil {
				return err
			}
		}
	}
	return nil
}

func (lc *LightController) Listen(ctx context.Context) error {
	buf := make([]byte, 1024)
	socket := lc.sacnClient.GetConn()

	println("Listening for sACN packets...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Received SIGTERM, shutting down...")
			return nil
		default:
			n, err := socket.Read(buf)
			if err != nil {
				return err
			}
			packet := buf[:n]
			if IsDataPacket(packet) {
				sacnPacket, err := FromBytes(packet)
				if err != nil {
					return err
				}
				if err := lc.handlePacket(sacnPacket); err != nil {
					fmt.Fprintf(os.Stderr, "Error handling packet: %v\n", err)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (lc *LightController) FindLightLoop(adapter bluetooth.Adapter) {
	for _, light := range lc.lights {
		go func(light *Light) {
			light.FindLoop(adapter)
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
