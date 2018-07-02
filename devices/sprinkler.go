package devices

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

// SprinklerMCP23017Driver -
type SprinklerMCP23017Driver struct {
	name      string
	address   int
	connector i2c.Connector
	mcp23017  *i2c.MCP23017Driver
	pinStatus map[string]uint8
	status    bool
	gobot.Commander
}

// NewSprinklerMCP23017Driver -
func NewSprinklerMCP23017Driver(a i2c.Connector, addr int) *SprinklerMCP23017Driver {
	s := &SprinklerMCP23017Driver{
		name:      gobot.DefaultName("SprinklerMCP23017"),
		address:   addr,
		connector: a,
		mcp23017:  i2c.NewMCP23017Driver(a, i2c.WithBus(1), i2c.WithAddress(addr)),
		pinStatus: map[string]uint8{"A": uint8(0), "B": uint8(0)},
		Commander: gobot.NewCommander(),
	}

	s.AddCommand("On", func(params map[string]interface{}) interface{} {
		port := string(params["port"].(string))
		pin := uint8(params["pin"].(uint8))
		return s.On(port, pin)
	})
	s.AddCommand("Off", func(params map[string]interface{}) interface{} {
		port := string(params["port"].(string))
		pin := uint8(params["pin"].(uint8))
		return s.Off(port, pin)
	})
	return s
}

// On -
func (s *SprinklerMCP23017Driver) Off(port string, pin uint8) (err error) {
	// if err = s.mcp23017.PinMode(pin, 0, port); err != nil {
	// 	return err
	// }
	if err = s.mcp23017.WriteGPIO(pin, 1, port); err != nil {
		return err
	}
	s.pinStatus[port] &= ^(1 << pin)
	s.status = true
	return
}

// Off -
func (s *SprinklerMCP23017Driver) On(port string, pin uint8) (err error) {
	// if err = s.mcp23017.PinMode(pin, 0, port); err != nil {
	// 	return err
	// }
	if err = s.mcp23017.WriteGPIO(pin, 0, port); err != nil {
		return err
	}
	s.pinStatus[port] |= 1 << pin
	s.status = true
	return
}

// Name returns the SprinklerMCP23017Driver name
func (s *SprinklerMCP23017Driver) Name() string { return s.name }

// SetName sets the SprinklerMCP23017Driver name
func (s *SprinklerMCP23017Driver) SetName(n string) { s.name = n }

// Start implements the Driver interface
func (s *SprinklerMCP23017Driver) Start() (err error) {
	var i uint8
	if err = s.mcp23017.Start(); err != nil {
		return err
	}
	for i = 0; i < 8; i++ {
		s.Off("A", i)
		s.Off("B", i)
	}
	return
}

// ReadPin -
func (s *SprinklerMCP23017Driver) ReadPin(pin uint8, port string) (state uint8) {
	return (s.pinStatus[port] & (1 << pin))
}

// Stop to clean up all pin
func (s *SprinklerMCP23017Driver) Stop() (err error) {
	return s.Halt()
}

// Halt implements the Driver interface
func (s *SprinklerMCP23017Driver) Halt() (err error) {
	var i uint8
	for i = 0; i < 8; i++ {
		if err = s.mcp23017.WriteGPIO(i, 1, "A"); err != nil {
			return
		}
		if err = s.mcp23017.WriteGPIO(i, 1, "B"); err != nil {
			return
		}
	}
	return
}

// Connection returns the SprinklerMCP23017Driver Connection
func (s *SprinklerMCP23017Driver) Connection() gobot.Connection {
	return s.connector.(gobot.Connection)
}
