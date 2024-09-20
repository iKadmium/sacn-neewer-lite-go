package main

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const WriteCharacteristicUuid = "69400002-B5A3-F393-E0A9-E50E24DCCA99"
const ReadCharecteristicUuid = "69400003-B5A3-F393-E0A9-E50E24DCCA99"
const ServiceUuid = "69400001-b5a3-f393-e0a9-e50e24dcca99"

type Light struct {
	id             bluetooth.MAC
	universe       uint16
	address        uint16
	hue            uint16
	saturation     byte
	brightness     byte
	peripheral     *bluetooth.Device
	write_char     *bluetooth.DeviceCharacteristic
	read_char      *bluetooth.DeviceCharacteristic
	last_read_time time.Time
}

func NewLight(id bluetooth.MAC, universe uint16, address uint16) *Light {
	return &Light{
		id:         id,
		universe:   universe,
		address:    address,
		hue:        0,
		saturation: 0,
		brightness: 0,
	}
}

func getChecksum(sendValue []byte) byte {
	var checkSum byte
	for _, value := range sendValue {
		checkSum += value
	}
	return checkSum
}

func (l *Light) SetColorHSI(hue uint16, saturation, brightness byte) {
	l.hue = hue
	l.saturation = saturation
	l.brightness = brightness
}

func (l *Light) SendColor() error {
	hueLSB := byte(l.hue & 0xFF)
	hueMSB := byte((l.hue >> 8) & 0xFF)

	colorCmd := []byte{120, 134, 4, hueLSB, hueMSB, l.saturation, l.brightness}
	colorCmd = append(colorCmd, getChecksum(colorCmd))

	println("Sending")

	if l.IsConnected() {
		l.write_char.WriteWithoutResponse(colorCmd)
	}
	return nil
}

func (l *Light) SetColorRGB(red, green, blue byte) {
	hue, saturation, intensity := RgbToHsv(red, green, blue)
	l.SetColorHSI(hue, saturation, intensity)
}

func (l *Light) Connect(peripheral bluetooth.ScanResult, adapter *bluetooth.Adapter) error {
	writeCharacteristicUuid, err := bluetooth.ParseUUID(WriteCharacteristicUuid)
	if err != nil {
		return fmt.Errorf("error parsing UUID: %v", err)
	}
	readCharecteristicUuid, err := bluetooth.ParseUUID(ReadCharecteristicUuid)
	if err != nil {
		return fmt.Errorf("error parsing UUID: %v", err)
	}
	serviceUuid, err := bluetooth.ParseUUID(ServiceUuid)
	if err != nil {
		return fmt.Errorf("error parsing UUID: %v", err)
	}

	device, err := adapter.Connect(peripheral.Address, bluetooth.ConnectionParams{})

	if err != nil {
		return fmt.Errorf("error connecting: %v", err)
	}

	l.peripheral = &device

	services, err := device.DiscoverServices([]bluetooth.UUID{serviceUuid})
	if err != nil {
		return fmt.Errorf("error discovering services: %v", err)
	}

	if len(services) != 1 {
		return fmt.Errorf("no services found")
	}
	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{writeCharacteristicUuid})
	if err != nil {
		return fmt.Errorf("error discovering characteristics: %v", err)
	}
	if len(characteristics) != 1 {
		return fmt.Errorf("no characteristics found")
	}
	l.write_char = &characteristics[0]

	characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{readCharecteristicUuid})
	if err != nil {
		return fmt.Errorf("error discovering characteristics: %v", err)
	}
	if len(characteristics) != 1 {
		return fmt.Errorf("no characteristics found")
	}
	l.read_char = &characteristics[0]
	l.last_read_time = time.Now()
	l.read_char.EnableNotifications(func(data []byte) {
		println("heartbeat from:", l.id.String())
		l.last_read_time = time.Now()
	})
	return err
}

func (l *Light) Disconnect() error {
	fmt.Printf("Disconnecting from %v\n", l.GetName())
	if l.peripheral != nil {
		err := l.peripheral.Disconnect()
		l.peripheral = nil
		l.write_char = nil
		l.read_char = nil

		return err
	}
	return nil
}

func (l *Light) GetName() string {
	if l.peripheral != nil {
		return "some name"
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
	if l.peripheral == nil || l.write_char == nil {
		return false
	}
	return true
}

func (l *Light) SendLoop(interval time.Duration) {
	for {
		if l.IsConnected() {
			err := l.SendColor()
			if err != nil {
				fmt.Println("Error sending color:", err)
			}
		}
		time.Sleep(interval)
	}
}

func (l *Light) HeartbeatLoop(interval time.Duration) {
	for {
		if l.IsConnected() {
			l.write_char.WriteWithoutResponse([]byte{120, 133, 0, 253})
			time.Sleep(interval)
			bytes := make([]byte, 20)
			_, _ = l.read_char.Read(bytes)
			if l.last_read_time.Add(interval).Before(time.Now()) {
				l.Disconnect()
			}
		}
		time.Sleep(interval)
	}
}

func (l *Light) GetLastReadTime() time.Time {
	return l.last_read_time
}

func (l *Light) GetID() bluetooth.MAC {
	return l.id
}
