package main

import (
	"fmt"
	"time"

	"golang.design/x/mainthread"
	"tinygo.org/x/bluetooth"
)

const WriteCharacteristicUuid = "69400002-B5A3-F393-E0A9-E50E24DCCA99"
const ReadCharecteristicUuid = "69400003-B5A3-F393-E0A9-E50E24DCCA99"
const ServiceUuid = "69400001-b5a3-f393-e0a9-e50e24dcca99"
const DirtyForceSendInterval = 5 * time.Second

type NeewerLight struct {
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
	dirty          bool
	last_send_time time.Time
}

func NewLight(id bluetooth.MAC, universe uint16, address uint16) *NeewerLight {
	return &NeewerLight{
		id:             id,
		universe:       universe,
		address:        address,
		hue:            0,
		saturation:     0,
		brightness:     0,
		dirty:          true,
		last_send_time: time.Unix(0, 0),
	}
}

func getChecksum(sendValue []byte) byte {
	var checkSum byte
	for _, value := range sendValue {
		checkSum += value
	}
	return checkSum
}

func (l *NeewerLight) setColorHSI(hue uint16, saturation, brightness byte) {
	if l.hue != hue || l.saturation != saturation || l.brightness != brightness {
		l.dirty = true
		l.hue = hue
		l.saturation = saturation
		l.brightness = brightness
	}
}

func (l *NeewerLight) SendColor() error {
	if l.IsConnected() && (l.dirty || l.last_send_time.Add(DirtyForceSendInterval).Before(time.Now())) {
		l.dirty = false
		l.last_send_time = time.Now()

		hueLSB := byte(l.hue & 0xFF)
		hueMSB := byte((l.hue >> 8) & 0xFF)

		colorCmd := []byte{120, 134, 4, hueLSB, hueMSB, l.saturation, l.brightness}
		colorCmd = append(colorCmd, getChecksum(colorCmd))

		_, err := l.write_char.WriteWithoutResponse(colorCmd)
		println("sent color to:", l.id.String())
		return err
	}
	return nil
}

func (l *NeewerLight) SetColorRGB(red, green, blue byte) {
	hue, saturation, intensity := RgbToHsv(red, green, blue)
	l.setColorHSI(hue, saturation, intensity)
}

func (l *NeewerLight) Connect(peripheral bluetooth.ScanResult, adapter *bluetooth.Adapter) error {
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

	var device bluetooth.Device
	mainthread.Call(func() {
		device, err = adapter.Connect(peripheral.Address, bluetooth.ConnectionParams{})
	})
	if err != nil {
		err = fmt.Errorf("error connecting: %v", err)
	}

	l.peripheral = &device

	var services []bluetooth.DeviceService
	mainthread.Call(func() {
		services, err = device.DiscoverServices([]bluetooth.UUID{serviceUuid})
	})
	if err != nil {
		return fmt.Errorf("error discovering services: %v", err)
	}

	if len(services) != 1 {
		return fmt.Errorf("no services found")
	}

	var characteristics []bluetooth.DeviceCharacteristic
	mainthread.Call(func() {
		characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{writeCharacteristicUuid})
	})
	if err != nil {
		return fmt.Errorf("error discovering characteristics: %v", err)
	}
	if len(characteristics) != 1 {
		return fmt.Errorf("no characteristics found")
	}
	l.write_char = &characteristics[0]

	mainthread.Call(func() {
		characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{readCharecteristicUuid})
	})
	characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{readCharecteristicUuid})
	if err != nil {
		return fmt.Errorf("error discovering characteristics: %v", err)
	}
	if len(characteristics) != 1 {
		return fmt.Errorf("no characteristics found")
	}
	l.read_char = &characteristics[0]
	l.last_read_time = time.Now()
	mainthread.Call(func() {
		l.read_char.EnableNotifications(func(data []byte) {
			println("heartbeat from:", l.id.String())
			l.last_read_time = time.Now()
		})
	})
	return err
}

func (l *NeewerLight) Disconnect() error {
	fmt.Printf("Disconnecting from %v\n", l.id.String())
	if l.peripheral != nil {
		var err error
		mainthread.Call(func() {
			err = l.peripheral.Disconnect()
		})
		l.peripheral = nil
		l.write_char = nil
		l.read_char = nil

		return err
	}
	return nil
}

func (l *NeewerLight) GetAddress() uint16 {
	return l.address
}

func (l *NeewerLight) GetUniverse() uint16 {
	return l.universe
}

func (l *NeewerLight) IsConnected() bool {
	if l.peripheral == nil || l.write_char == nil {
		return false
	}
	return true
}

func (l *NeewerLight) SendLoop(interval time.Duration) {
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

func (l *NeewerLight) HeartbeatLoop(interval time.Duration) {
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

func (l *NeewerLight) GetLastReadTime() time.Time {
	return l.last_read_time
}

func (l *NeewerLight) GetID() bluetooth.MAC {
	return l.id
}
