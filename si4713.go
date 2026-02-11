package si4713

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/warthog618/go-gpiocdev"
)

type Si4713 struct {
	resetLine *gpiocdev.Line
	i2c       *i2c.I2C
}

func New(i2cAddr uint8, resetPin int) (*Si4713, error) {
	si4713 := &Si4713{}

	var err error
	si4713.i2c, err = i2c.NewI2C(i2cAddr, 1)
	if err != nil {
		return nil, err
	}

	si4713.resetLine, err = gpiocdev.RequestLine("gpiochip0", resetPin)
	if err != nil {
		return nil, err
	}

	si4713.Reset()

	err = si4713.powerUp()
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	rev := si4713.getRev()
	if rev != 13 {
		return nil, fmt.Errorf("rev = %d, expected 13", rev)
	}

	return si4713, nil
}

func (s *Si4713) setProperty(property, value uint16) {
	s.i2c.WriteBytes([]byte{
		CmdSetProperty,
		0,
		byte(property >> 8),
		byte(property & 0xFF),
		byte(value >> 8),
		byte(value & 0xFF),
	})
}

func (s *Si4713) powerUp() error {
	_, err := s.i2c.WriteBytes([]byte{CmdPowerUp, 0x12, 0x50})
	if err != nil {
		return err
	}

	s.setProperty(PropRefClkFreq, 32768)
	s.setProperty(PropTxPreemphasis, 0)
	s.setProperty(PropTxACompGain, 10)
	s.setProperty(PropTxACompEnable, 0)
	return nil
}

func (s *Si4713) powerDown() byte {
	s.i2c.WriteBytes([]byte{CmdPowerDown})

	response := make([]byte, 1)
	s.i2c.ReadBytes(response)

	return response[0]
}

func (s *Si4713) getStatus() byte {
	s.i2c.WriteBytes([]byte{CmdGetIntStatus})

	response := make([]byte, 1)
	s.i2c.ReadBytes(response)

	return response[0]
}

func (s *Si4713) getRev() byte {
	s.i2c.WriteBytes([]byte{CmdGetRev, 0})

	response := make([]byte, 9)
	s.i2c.ReadBytes(response)

	return response[1]
}

// Frequency in MHz * 100
func (s *Si4713) SetFrequency(freq uint16) {
	s.i2c.WriteBytes([]byte{CmdSetTxFreq, 0, byte(freq >> 8), byte(freq & 0xFF)})

	for (s.getStatus() & 0x81) != 0x81 {
		time.Sleep(time.Millisecond * 10)
	}
}

// Power must be from 88 to 115 dBuV
func (s *Si4713) SetPower(power uint8) {
	s.i2c.WriteBytes([]byte{CmdSetTxPower, 0, 0, power, 0})
}

func (s *Si4713) Reset() {
	s.resetLine.SetValue(1)
	time.Sleep(time.Millisecond * 10)
	s.resetLine.SetValue(0)
	time.Sleep(time.Millisecond * 10)
	s.resetLine.SetValue(1)
	time.Sleep(time.Millisecond * 10)
}

func (s *Si4713) Close() {
	s.resetLine.Close()
	s.i2c.Close()
}
