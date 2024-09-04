package main

import "fmt"

type Bandwidth string

const (
	BandwidthNarrow = "Narrow"
	BandwidthWide   = "Wide"
)

type TimeSlot string

const (
	TimeSlotTS1 TimeSlot = "TS1"
	TimeSlotTS2 TimeSlot = "TS2"
)

type Power string

const (
	PowerLow  Power = "Low"
	PowerMid  Power = "Mid"
	PowerHigh Power = "High"
	PowerMax  Power = "Max"
)

type DigitalChannel struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Power       Power    `yaml:"power"`
	RXFrequency string   `yaml:"rxFrequency"`
	TXFrequency string   `yaml:"txFrequency"`
	TimeSlot    TimeSlot `yaml:"timeSlot"`
	ColorCode   int      `yaml:"colorCode"`
	Contact     string   `yaml:"contact"`
}

func (c *DigitalChannel) Validate() error {
	if len(c.Name) > maxChannelLength {
		return fmt.Errorf("channel name %q is too long", c.Name)
	}
	return nil
}

type AnalogChannel struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	RXFrequency string    `yaml:"rxFrequency"`
	TXFrequency string    `yaml:"txFrequency"`
	Power       Power     `yaml:"power"`
	Bandwidth   Bandwidth `yaml:"bandwidth"`
}

func (c *AnalogChannel) Validate() error {
	if len(c.Name) > maxChannelLength {
		return fmt.Errorf("channel name %q is too long", c.Name)
	}
	return nil
}

type Channel struct {
	Digital *DigitalChannel `yaml:"digital,omitempty"`
	Analog  *AnalogChannel  `yaml:"analog,omitempty"`
}

func (c *Channel) Validate() error {
	switch {
	case c.Digital != nil && c.Analog != nil:
		return fmt.Errorf("channel cannot be both digital and analog")
	case c.Digital == nil && c.Analog == nil:
		return fmt.Errorf("channel must be either digital or analog")
	case c.Digital != nil:
		return c.Digital.Validate()
	case c.Analog != nil:
		return c.Analog.Validate()
	}
	return nil
}
