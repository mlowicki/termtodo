// +build raspbian

package main

import (
	"time"

	blinkt "github.com/alexellis/blinkt_go"
)

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
		r, g, b := 150, 0, 0
	outerloop:
		for {
			for _, pixel := range append(seq(0, 7), seq(6, 1)...) {
				bl.Clear()
				bl.SetPixel(pixel, r, g, b)
				bl.Show()
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ch:
					break outerloop
				}
			}
		}
		bl.Clear()
		bl.Show()
		close(ch)
	}()
	return &Blinkt{ch}
}
