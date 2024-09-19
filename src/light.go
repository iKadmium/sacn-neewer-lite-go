package main

import (
	"fmt"

	"tinygo.org/x/bluetooth"
)

const WriteCharacteristicUuid = "69400002-B5A3-F393-E0A9-E50E24DCCA99"
const ServiceUuid = "69400001-b5a3-f393-e0a9-e50e24dcca99"

type Light struct {
	id         bluetooth.MAC
	universe   uint16
	address    uint16
	peripheral *bluetooth.Device
	cmd_char   *bluetooth.DeviceCharacteristic
}

func NewLight(id bluetooth.MAC, universe uint16, address uint16) *Light {
	return &Light{
		id:       id,
		universe: universe,
		address:  address,
	}
}

func getChecksum(sendValue []byte) byte {
	var checkSum byte
	for _, value := range sendValue {
		checkSum += value
	}
	return checkSum
}

func (l *Light) SetColorHSI(hue uint16, saturation, brightness byte) error {
	hueLSB := byte(hue & 0xFF)
	hueMSB := byte((hue >> 8) & 0xFF)

	colorCmd := []byte{120, 134, 4, hueLSB, hueMSB, saturation, brightness}
	colorCmd = append(colorCmd, getChecksum(colorCmd))

	fmt.Printf("Sending %v\n", colorCmd)

	if l.IsConnected() {
		l.cmd_char.WriteWithoutResponse(colorCmd)
	}
	return nil
}

func (l *Light) SetColorRGB(red, green, blue byte) error {
	hue, saturation, intensity := RgbToHsv(red, green, blue)
	return l.SetColorHSI(hue, saturation, intensity)
}

func (l *Light) Connect(peripheral bluetooth.Device) error {
	fmt.Printf("Connecting to %v\n", l.GetName())
	l.peripheral = &peripheral
	services, err := peripheral.DiscoverServices(nil)

	println("Services:", services)

	return err
}

func (l *Light) Disconnect() error {
	fmt.Printf("Disconnecting from %v\n", l.GetName())
	if l.peripheral != nil {
		return l.peripheral.Disconnect()
	}
	return nil
}

func (l *Light) GetName() string {
	if l.peripheral != nil {
		return "some name"
		//return l.peripheral.Name()
	}
	return ""
}

func (l *Light) GetAddress() uint16 {
	return l.address
}

func (l *Light) GetUniverse() uint16 {
	return l.universe
}

func (l *Light) IsConnected() bool {
	if l.peripheral != nil {
		return true
		//return l.peripheral.IsConnected()
	}
	return false
}

func (l *Light) FindLoop(adapter bluetooth.Adapter) {
	println("scanning...")
	writeCharacteristicUuid, err := bluetooth.ParseUUID(WriteCharacteristicUuid)
	if err != nil {
		println("error parsing UUID:", err.Error())
		return
	}
	serviceUuid, err := bluetooth.ParseUUID(ServiceUuid)
	if err != nil {
		println("error parsing UUID:", err.Error())
		return
	}

	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if !l.IsConnected() && device.Address.MAC == l.id {
			fmt.Println("Found device: ", l.id)
			address := device.Address
			device, err := adapter.Connect(address, bluetooth.ConnectionParams{})

			if err != nil {
				println("error connecting:", err.Error())
			}

			services, err := device.DiscoverServices([]bluetooth.UUID{serviceUuid})
			if err != nil {
				println("error discovering services:", err.Error())
			}

			if len(services) != 1 {
				println("no services found")
				return
			}
			characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeCharacteristicUuid})
			if err != nil {
				println("error discovering characteristics:", err.Error())
			}
			if len(characteristics) != 1 {
				println("no characteristics found")
				return
			}
			println("connected to device:", device.Address.String())
			l.cmd_char = &characteristics[0]
			l.peripheral = &device
		}
	})
	if err != nil {
		println("error scanning:", err.Error())
	}
}

func (l *Light) GetID() bluetooth.MAC {
	return l.id
}
