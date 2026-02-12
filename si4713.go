package si4713

import (
	"fmt"
	"strings"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
	"github.com/warthog618/go-gpiocdev"
)

func init() {
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
}

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

func (s *Si4713) TuneMeasure(freq uint16) {
	if freq%5 != 0 {
		freq -= freq % 5
	}

	s.i2c.WriteBytes([]byte{CmdTxTuneMeasure, 0, byte(freq >> 8), byte(freq & 0xFF), 0})

	for s.getStatus() != 0x81 {
		time.Sleep(time.Millisecond * 10)
	}
}

type TuneStatus struct {
	Frequency        uint16
	Power            uint8 // dBuV
	AntennaCapacitor uint8
	NoiseLevel       uint8
}

func (s *Si4713) TuneStatus() TuneStatus {
	s.i2c.WriteBytes([]byte{CmdTxTuneStatus, 1})

	response := make([]byte, 8)
	s.i2c.ReadBytes(response)

	return TuneStatus{
		Frequency:        (uint16(response[2]) << 8) | uint16(response[3]),
		Power:            response[5],
		AntennaCapacitor: response[6],
		NoiseLevel:       response[7],
	}
}

type ASQStatus struct {
	ASQ     uint8
	InLevel uint8
}

func (s *Si4713) ReadASQ() ASQStatus {
	s.i2c.WriteBytes([]byte{CmdTxASQStatus, 1})

	response := make([]byte, 5)
	s.i2c.ReadBytes(response)

	return ASQStatus{
		ASQ:     response[1],
		InLevel: response[4],
	}
}

func (s *Si4713) BeginRDS(programID uint16) {
	s.setProperty(PropTxAudioDeviation, 6625)
	s.setProperty(PropTxRDSDeviation, 200)
	s.setProperty(PropTxRDSInterruptSource, 1)
	s.setProperty(PropTxRDSPI, programID)
	s.setProperty(PropTxRDSPsMix, 3)
	s.setProperty(PropTxRDSPsMisc, 0x1008)
	s.setProperty(PropTxRDSPsRepeatCount, 3)
	s.setProperty(PropTxRDSMessageCount, 1)
	s.setProperty(PropTxRDSPsAF, 0xE0E0)
	s.setProperty(PropTxRDSFIFOSize, 0)
	s.setProperty(PropTxComponentEnable, 7)
}

func (s *Si4713) SetRDSPS(str string) {
	if len(str) > 8 {
		str = str[:8]
	} else if len(str) < 8 {
		str += strings.Repeat(" ", 8-len(str))
	}

	buf := make([]byte, 6)
	buf[0] = CmdTxRDSPs

	copy(buf[2:], str[:4])
	s.i2c.WriteBytes(buf)

	buf[1] = 1
	copy(buf[2:], str[4:])
	s.i2c.WriteBytes(buf)
}

func (s *Si4713) SetRDSText(str string) {
	if len(str) > 64 {
		str = str[:64]
	}

	buf := make([]byte, 8)

	buf[0] = CmdTxRDSBuff
	buf[1] = 0x06
	buf[2] = 0x20

	i := uint8(0)
	for len(str) > 0 {
		for j := 4; j < 8; j++ {
			buf[j] = ' '
		}

		n := copy(buf[4:], str[:min(4, len(str))])
		str = str[n:]

		buf[3] = i
		s.i2c.WriteBytes(buf)

		if i == 0 {
			buf[1] = 0x04
		}
		i++
	}
}

func (s *Si4713) SetGPIOCtrl(x uint8) {
	s.i2c.WriteBytes([]byte{CmdGPOCtl, x})
}

func (s *Si4713) SetGPIO(x uint8) {
	s.i2c.WriteBytes([]byte{CmdGPOSet, x})
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
