// Stub for Blinkt! integration to allow build on non-Raspbian platforms.
// +build !raspbian

package main

// Blinkt creates visual notification using Pimoroni Blinkt!.
type Blinkt struct {
}

// Stop disables visual notification.
func (b *Blinkt) Stop() {
}

// NewBlinkt returns active Blinkt notification.
func NewBlinkt() *Blinkt {
	return &Blinkt{}
}
