package main

import (
	"context"
	"fmt"
	"sacn_neewer_lite_go/status"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
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

	status status.Status
}

func NewLight(id bluetooth.MAC, universe uint16, address uint16, resetContext context.Context) *NeewerLight {
	return &NeewerLight{
		id:             id,
		universe:       universe,
		address:        address,
		hue:            0,
		saturation:     0,
		brightness:     0,
		dirty:          true,
		last_send_time: time.Unix(0, 0),
		status:         status.NewStatus(true, true, resetContext),
	}
}

func getChecksum(sendValue []byte) byte {
	var checkSum byte
	for _, value := range sendValue {
		checkSum += value
	}
	return checkSum
}

func (l *NeewerLight) setColorHSI(hue uint16, saturation, brightness uint8) {
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
		l.status.Increment()
		return err
	}
	return nil
}

func (l *NeewerLight) SetColorRGB(red, green, blue byte) {
	hue, saturation, intensity := RgbToHsv(red, green, blue)
	// if intensity == 0 && l.brightness > 0 && l.brightness < 250 {
	// 	// hmm
	// 	l.status.Update("Flicker Detected", tcell.ColorRed)
	// } else {
	// 	l.status.Update("Receiving", tcell.ColorGreen)
	// }
	l.status.GetColorTracker().Update(tcell.NewRGBColor(int32(red), int32(green), int32(blue)))
	l.setColorHSI(hue, saturation, intensity)
}

func (l *NeewerLight) Connect(peripheral bluetooth.ScanResult, adapter *bluetooth.Adapter) error {
	l.status.Update("Connecting", tcell.ColorYellow)
	writeCharacteristicUuid, err := bluetooth.ParseUUID(WriteCharacteristicUuid)
	if err != nil {
		err = fmt.Errorf("error parsing UUID: %v", err)
		l.status.SetErr(err)
		return err
	}
	readCharecteristicUuid, err := bluetooth.ParseUUID(ReadCharecteristicUuid)
	if err != nil {
		err = fmt.Errorf("error parsing characteristic UUID: %v", err)
		l.status.SetErr(err)
		return err
	}
	serviceUuid, err := bluetooth.ParseUUID(ServiceUuid)
	if err != nil {
		err = fmt.Errorf("error parsing service UUID: %v", err)
		l.status.SetErr(err)
		return err
	}

	var device bluetooth.Device
	mainthread.Call(func() {
		device, err = adapter.Connect(peripheral.Address, bluetooth.ConnectionParams{})
	})
	if err != nil {
		err = fmt.Errorf("error connecting: %v", err)
		l.status.SetErr(err)
		return err
	}

	l.peripheral = &device

	var services []bluetooth.DeviceService
	mainthread.Call(func() {
		services, err = device.DiscoverServices([]bluetooth.UUID{serviceUuid})
	})
	if err != nil {
		err = fmt.Errorf("error discovering services: %v", err)
		l.status.SetErr(err)
		return err
	}

	if len(services) != 1 {
		err = fmt.Errorf("no services found")
		l.status.SetErr(err)
		return err
	}

	var characteristics []bluetooth.DeviceCharacteristic
	mainthread.Call(func() {
		characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{writeCharacteristicUuid})
	})
	if err != nil {
		err = fmt.Errorf("error discovering characteristics: %v", err)
		l.status.SetErr(err)
		return err
	}
	if len(characteristics) != 1 {
		err = fmt.Errorf("no characteristics found")
		l.status.SetErr(err)
		return err
	}
	l.write_char = &characteristics[0]

	mainthread.Call(func() {
		characteristics, err = services[0].DiscoverCharacteristics([]bluetooth.UUID{readCharecteristicUuid})
	})
	if err != nil {
		err = fmt.Errorf("error discovering characteristics: %v", err)
		l.status.SetErr(err)
		return err
	}
	if len(characteristics) != 1 {
		err = fmt.Errorf("no characteristics found")
		l.status.SetErr(err)
		return err
	}
	l.read_char = &characteristics[0]
	l.last_read_time = time.Now()
	mainthread.Call(func() {
		l.read_char.EnableNotifications(func(data []byte) {
			l.last_read_time = time.Now()
		})
	})

	l.status.Update("Connected", tcell.ColorGreen)

	return nil
}

func (l *NeewerLight) Disconnect() error {
	if l.peripheral != nil {
		l.status.Update("Disconnecting", tcell.ColorYellow)
		var err error
		mainthread.Call(func() {
			err = l.peripheral.Disconnect()
		})
		l.peripheral = nil
		l.write_char = nil
		l.read_char = nil

		l.status.Update("Disconnected", tcell.ColorRed)

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

func (l *NeewerLight) SendLoop(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				if l.IsConnected() {
					err := l.SendColor()
					if err != nil {
						dbusErr, isDbusErr := err.(dbus.Error)
						if !(isDbusErr && dbusErr.Name == "org.bluez.Error.InProgress") {
							err = fmt.Errorf("error sending color: %v", err)
							l.status.SetErr(err)
						}
					} else {
						l.status.Update("Sending", tcell.ColorGreen)
					}
				}
				time.Sleep(interval)
			}
		}
	}
}

func (l *NeewerLight) HeartbeatLoop(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
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
}

func (l *NeewerLight) GetLastReadTime() time.Time {
	return l.last_read_time
}

func (l *NeewerLight) GetID() bluetooth.MAC {
	return l.id
}

func (l *NeewerLight) GetStatus() *status.Status {
	return &l.status
}
