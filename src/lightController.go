package main

import (
	"context"
	"fmt"
	"time"

	sacn "sacn_neewer_lite_go/sacn"
	"sacn_neewer_lite_go/status"

	"github.com/gdamore/tcell/v2"
	"tinygo.org/x/bluetooth"
)

type LightController struct {
	sacnClient *sacn.SacnClient
	lights     map[string]*NeewerLight
	status     status.Status
}

func NewLightController(config *Config) (*LightController, error) {
	client, err := sacn.NewSacnClient(config.GetUniverses())
	if err != nil {
		return nil, err
	}

	lights := make(map[string]*NeewerLight)
	for _, lightConfig := range config.Lights {
		idBytes, err := bluetooth.ParseMAC(lightConfig.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid light ID %s: %v", lightConfig.ID, err)
		}
		light := NewLight(idBytes, lightConfig.Universe, lightConfig.Address)
		lights[lightConfig.ID] = light
	}

	return &LightController{
		sacnClient: client,
		lights:     lights,
		status:     status.NewStatus(false, false),
	}, err
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

func (lc *LightController) Listen(ctx context.Context) {
	lc.status.Update("Running", tcell.ColorGreen)
	go lc.sacnClient.Listen(ctx, lc.handlePacket)
}

func (lc *LightController) FindLightLoop(ctx context.Context, adapter bluetooth.Adapter) {
	for _, light := range lc.lights {
		light.GetStatus().Update("Searching", tcell.ColorYellow)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
					l := lc.lights[device.Address.String()]
					if l != nil && !l.IsConnected() && device.Address.MAC == l.id {
						l.Connect(device, adapter)
					}
				})
				if err != nil {
					err = fmt.Errorf("error scanning: %v", err)
					lc.status.SetErr(err)
				}
			}
		}
	}()
}

func (lc *LightController) SendLoop(ctx context.Context) {
	for _, light := range lc.lights {
		go light.SendLoop(ctx, time.Millisecond*80)
		go light.HeartbeatLoop(ctx, time.Second*2)
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

func (lc *LightController) GetLightStatuses() map[string]*status.Status {
	out := make(map[string]*status.Status)
	for id, light := range lc.lights {
		out[id] = light.GetStatus()
	}
	return out
}

func (lc *LightController) GetAppStatus() *status.Status {
	return &lc.status
}

func (lc *LightController) GetSacnStatus() *status.Status {
	return lc.sacnClient.GetStatus()
}
