package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	maxChannelLength = 16
	maxZoneLength    = 16
	maxContactLength = 16
)

type ContactType string

const (
	ContactTypeGroupCall   ContactType = "GroupCall"
	ContactTypePrivateCall ContactType = "PrivateCall"
)

type DMRContact struct {
	ID     string      `yaml:"id"`
	Name   string      `yaml:"name"`
	Number int         `yaml:"number"`
	Ring   bool        `yaml:"ring"`
	Type   ContactType `yaml:"type"`
}

func (c *DMRContact) Validate() error {
	if len(c.Name) > maxContactLength {
		return fmt.Errorf("contact name is too long")
	}
	return nil
}

type Contact struct {
	DMR DMRContact `yaml:"dmr"`
}

func (c *Contact) Validate() error {
	return c.DMR.Validate()
}

type Zone struct {
	ID   string   `yaml:"id"`
	Name string   `yaml:"name"`
	A    []string `yaml:"A"`
	B    []string `yaml:"B"`
}

func (z *Zone) Validate() error {
	if len(z.Name) > maxZoneLength {
		return fmt.Errorf("zone name is too long")
	}
	return nil
}

type Config struct {
	Count    int       `yaml:"-"`
	Contacts []Contact `yaml:"contacts"`
	Channels []Channel `yaml:"channels"`
	Zones    []Zone    `yaml:"zones"`
}

func (c *Config) ID() string {
	id := fmt.Sprintf("gen%d", c.Count)
	c.Count++
	return id
}

func (c *Config) AddContact(
	name string,
	number int,
	contactType ContactType,
) (Contact, error) {
	contact := Contact{
		DMR: DMRContact{
			ID:     c.ID(),
			Name:   name,
			Number: number,
			Ring:   false,
			Type:   contactType,
		},
	}
	if err := contact.Validate(); err != nil {
		return Contact{}, err
	}
	c.Contacts = append(c.Contacts, contact)
	return contact, nil
}

type AnalogSimplexChannelCfg struct {
	Name      string
	Frequency string
}

func (c *Config) AddAnalogSimplexZone(
	name string,
	power Power,
	bandwidth Bandwidth,
	channels []AnalogSimplexChannelCfg,
) error {
	channelIDs := []string{}
	for _, ch := range channels {
		channel := Channel{
			Analog: &AnalogChannel{
				ID:          c.ID(),
				Name:        fmt.Sprintf("%s %s", ch.Name, strings.TrimSuffix(ch.Frequency, " Mhz")),
				RXFrequency: ch.Frequency,
				TXFrequency: ch.Frequency,
				Power:       power,
				Bandwidth:   bandwidth,
			},
		}
		if err := channel.Validate(); err != nil {
			return err
		}
		c.Channels = append(c.Channels, channel)
		channelIDs = append(channelIDs, channel.Analog.ID)
	}
	zone := Zone{
		ID:   c.ID(),
		Name: name,
		A:    channelIDs,
	}
	if err := zone.Validate(); err != nil {
		return err
	}
	c.Zones = append(c.Zones, zone)
	return nil
}

func (c *Config) AddDigitalRepeaterZone(
	name string,
	colorCode int,
	rxFrequency string,
	txFrequency string,
	power Power,
	timeSlots map[TimeSlot][]Contact,
) error {
	channelIDs := []string{}
	for ts, contacts := range timeSlots {
		for _, contact := range contacts {
			channel := Channel{
				Digital: &DigitalChannel{
					ID:          c.ID(),
					Contact:     contact.DMR.ID,
					Name:        contact.DMR.Name,
					ColorCode:   colorCode,
					RXFrequency: rxFrequency,
					TXFrequency: txFrequency,
					Power:       power,
					TimeSlot:    ts,
				},
			}
			if err := channel.Validate(); err != nil {
				return err
			}
			c.Channels = append(c.Channels, channel)
			channelIDs = append(channelIDs, channel.Digital.ID)
		}
	}
	zone := Zone{
		ID:   c.ID(),
		Name: name,
		A:    channelIDs,
	}
	if err := zone.Validate(); err != nil {
		return err
	}
	c.Zones = append(c.Zones, zone)
	return nil
}

func main() {
	cfg := &Config{}

	parrot, err := cfg.AddContact(
		"B9990 PARROT",
		9990,
		ContactTypePrivateCall,
	)
	if err != nil {
		panic(err)
	}
	oarc, err := cfg.AddContact(
		"B2348479 OARC",
		2348479,
		ContactTypeGroupCall,
	)
	if err != nil {
		panic(err)
	}
	world, err := cfg.AddContact(
		"B91 WORLD",
		91,
		ContactTypeGroupCall,
	)
	if err != nil {
		panic(err)
	}
	disconnect, err := cfg.AddContact(
		"DISCONNECT",
		4000,
		ContactTypePrivateCall,
	)
	if err != nil {
		panic(err)
	}
	phWorld, err := cfg.AddContact(
		"P1 WORLD",
		1,
		ContactTypeGroupCall,
	)
	if err != nil {
		panic(err)
	}

	err = cfg.AddDigitalRepeaterZone(
		"D HOTSPOT",
		1,
		"434.45 Mhz",
		"439.45 Mhz",
		PowerMid,
		map[TimeSlot][]Contact{
			TimeSlotTS1: {parrot, world, disconnect},
			TimeSlotTS2: {oarc},
		},
	)
	if err != nil {
		panic(err)
	}

	err = cfg.AddDigitalRepeaterZone(
		"D LONDON GB7HH",
		3,
		"439.75 Mhz",
		"430.75 Mhz",
		PowerMax,
		map[TimeSlot][]Contact{
			TimeSlotTS1: {phWorld},
			TimeSlotTS2: {parrot, disconnect},
		},
	)
	if err != nil {
		panic(err)
	}

	// https://hamradiosouthernrepeaters.co.uk/images/PDF/Simplex_Channel_Frequency.pdf
	err = cfg.AddAnalogSimplexZone(
		"A VHF SIMPLEX",
		PowerMax,
		BandwidthNarrow,
		[]AnalogSimplexChannelCfg{
			{Name: "V40*", Frequency: "145.500 Mhz"},
			{Name: "V16", Frequency: "145.200 Mhz"},
			{Name: "V17", Frequency: "145.2125 Mhz"},
			{Name: "V18", Frequency: "145.225 Mhz"},
			{Name: "V19", Frequency: "145.2375 Mhz"},
			{Name: "V20", Frequency: "145.250 Mhz"},
			{Name: "V21", Frequency: "145.2625 Mhz"},
			{Name: "V22", Frequency: "145.275 Mhz"},
			{Name: "V23", Frequency: "145.2875 Mhz"},
			{Name: "V24", Frequency: "145.300 Mhz"},
			{Name: "V25", Frequency: "145.3125 Mhz"},
			{Name: "V26", Frequency: "145.325 Mhz"},
			{Name: "V27", Frequency: "145.3375 Mhz"},
			{Name: "V28", Frequency: "145.350 Mhz"},
			{Name: "V29", Frequency: "145.3625 Mhz"},
			{Name: "V30", Frequency: "145.375 Mhz"},
			{Name: "V31", Frequency: "145.3875 Mhz"},
			{Name: "V32", Frequency: "145.400 Mhz"},
			{Name: "V33", Frequency: "145.4125 Mhz"},
			{Name: "V34", Frequency: "145.425 Mhz"},
			{Name: "V35", Frequency: "145.4375 Mhz"},
			{Name: "V36", Frequency: "145.450 Mhz"},
			{Name: "V37", Frequency: "145.4625 Mhz"},
			{Name: "V38", Frequency: "145.475 Mhz"},
			{Name: "V39", Frequency: "145.4875 Mhz"},
			{Name: "V41", Frequency: "145.5125 Mhz"},
			{Name: "V42", Frequency: "145.525 Mhz"},
			{Name: "V43", Frequency: "145.5375 Mhz"},
			{Name: "V44", Frequency: "145.550 Mhz"},
			{Name: "V45", Frequency: "145.5625 Mhz"},
			{Name: "V46", Frequency: "145.575 Mhz"},
			{Name: "V47", Frequency: "145.5875 Mhz"},
		})
	if err != nil {
		panic(err)
	}
	err = cfg.AddAnalogSimplexZone(
		"A UHF SIMPLEX",
		PowerMax,
		BandwidthWide,
		[]AnalogSimplexChannelCfg{
			{Name: "U280*", Frequency: "433.500 Mhz"},
			{Name: "U272", Frequency: "433.400 Mhz"},
			{Name: "U273", Frequency: "433.4125 Mhz"},
			{Name: "U274", Frequency: "433.425 Mhz"},
			{Name: "U275", Frequency: "433.4375 Mhz"},
			{Name: "U276", Frequency: "433.450 Mhz"},
			{Name: "U277", Frequency: "433.4625 Mhz"},
			{Name: "U278", Frequency: "433.475 Mhz"},
			{Name: "U279", Frequency: "433.4875 Mhz"},
			{Name: "U281", Frequency: "433.5125 Mhz"},
			{Name: "U282", Frequency: "433.525 Mhz"},
			{Name: "U283", Frequency: "433.5375 Mhz"},
			{Name: "U284", Frequency: "433.550 Mhz"},
			{Name: "U285", Frequency: "433.5625 Mhz"},
			{Name: "U286", Frequency: "433.575 Mhz"},
		})
	if err != nil {
		panic(err)
	}

	out := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	err = enc.Encode(cfg)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("out.yaml", out.Bytes(), 0644); err != nil {
		panic(err)
	}

	return
}
