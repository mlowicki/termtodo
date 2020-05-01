package main

import blinkt "github.com/alexellis/blinkt_go"

// Blinkt creates visual notification using Pimoroni Blinkt!.
type Blinkt struct {
	ch chan struct{}
}

// Stop disables visual notification.
func (b *Blinkt) Stop() {
	b.ch <- struct{}{}
	<-b.ch
}

// NewBlinkt returns active Blinkt notification.
func NewBlinkt() *Blinkt {
	ch := make(chan struct{})
	go func() {
		brightness := 0.5
		bl := blinkt.NewBlinkt(brightness)
		bl.Setup()
		r := 150
		g := 0
		b := 0
	outerloop:
		for {
			for pixel := 0; pixel < 8; pixel++ {
				select {
				case <-ch:
					break outerloop
				default:
				}
				bl.Clear()
				bl.SetPixel(pixel, r, g, b)
				bl.Show()
				blinkt.Delay(100)
			}
			for pixel := 7; pixel > 0; pixel-- {
				select {
				case <-ch:
					break outerloop
				default:
				}
				bl.Clear()
				bl.SetPixel(pixel, r, g, b)
				bl.Show()
				blinkt.Delay(100)
			}
		}
		bl.Clear()
		bl.Show()
		close(ch)
	}()
	return &Blinkt{ch}
}
