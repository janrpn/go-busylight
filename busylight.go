package led

import (
	"encoding/hex"
	"fmt"
	"github.com/baaazen/go-hid"
	"image/color"
	"time"
)

// Device type: BusyLight UC
var BusyLightUC DeviceType

// Device type: BusyLight Lync
var BusyLightLync DeviceType

func init() {
	BusyLightUC = addDriver(usbDriver{
		Name:      "BusyLight UC",
		Type:      &BusyLightUC,
		VendorId:  0x27BB,
		ProductId: 0x3BCB,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLight(d, func(c color.Color) {
				r, g, b, _ := c.RGBA()
				d.Write([]byte{0x00, 0x00, 0x00, byte(r >> 8), byte(g >> 8), byte(b >> 8), 0x00, 0x00, 0x00})
			}), nil
		},
	})
	BusyLightLync = addDriver(usbDriver{
		Name:      "BusyLight Lync",
		Type:      &BusyLightLync,
		VendorId:  0x04D8,
		ProductId: 0xF848,
		Open: func(d hid.Device) (Device, error) {
			return newBusyLight(d, func(c color.Color) {
				r, g, b, _ := c.RGBA()
				buff := []byte{0x01}
				buff = append(buff, 00)
				buff = append(buff, byte(r>>8))
				buff = append(buff, byte(g>>8))
				buff = append(buff, byte(b>>8))
				buff = append(buff, 00, 00)
				buff = append(buff, 00, 00) //Possible sound
				buff = append(buff, 00, 00)
				buff = append(buff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x04, 0xab)

				d.Write([]byte{0x00, 0x00, 0x00, byte(r >> 8), byte(g >> 8), byte(b >> 8), 0x00, 0x00, 0x00})
				encodedStr := hex.EncodeToString(buff)
				fmt.Printf("%s\n", encodedStr)
			}), nil
		},
	})
}

type busylightDev struct {
	closeChan chan<- struct{}
	colorChan chan<- color.Color
	closed    *bool
}

func newBusyLight(d hid.Device, setcolorFn func(c color.Color)) *busylightDev {
	closeChan := make(chan struct{})
	colorChan := make(chan color.Color)
	ticker := time.NewTicker(20 * time.Second) // If nothing is send after 30 seconds the device turns off.
	closed := false
	go func() {
		var curColor color.Color = color.Black
		for !closed {
			select {
			case <-ticker.C:
				setcolorFn(curColor)
			case col := <-colorChan:
				curColor = col
				setcolorFn(curColor)
			case <-closeChan:
				ticker.Stop()
				setcolorFn(color.White) // turn off device
				d.Close()
				closed = true
			}
		}
	}()
	return &busylightDev{closeChan: closeChan, colorChan: colorChan, closed: &closed}
}

func (d *busylightDev) SetKeepActive(v bool) error {
	return ErrKeepActiveNotSupported
}

func (d *busylightDev) SetColor(c color.Color) error {
	d.colorChan <- c
	return nil
}

func (d *busylightDev) Close() {
	d.closeChan <- struct{}{}
}

func (d *busylightDev) IsClosed() bool {
	return *d.closed
}
