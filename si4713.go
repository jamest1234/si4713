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
	i2c       *i2c.I2C
	resetLine *gpiocdev.Line
}

func New(i2cDevice *i2c.I2C, resetLine *gpiocdev.Line) (*Si4713, error) {
	si4713 := &Si4713{
		i2c:       i2cDevice,
		resetLine: resetLine,
	}

	return si4713, si4713.Reset()
}

func (s *Si4713) sendCommand(buf []byte) (int, error) {
	n, err := s.i2c.WriteBytes(buf)
	time.Sleep(time.Millisecond * 50)
	return n, err
}

func (s *Si4713) setProperty(property, value uint16) error {
	_, err := s.sendCommand([]byte{
		CmdSetProperty,
		0,
		byte(property >> 8),
		byte(property & 0xFF),
		byte(value >> 8),
		byte(value & 0xFF),
	})

	return err
}

func (s *Si4713) getStatus() (uint8, error) {
	_, err := s.sendCommand([]byte{CmdGetIntStatus})
	if err != nil {
		return 0, err
	}

	response := make([]byte, 1)
	_, err = s.i2c.ReadBytes(response)
	return response[0], err
}

func (s *Si4713) getRev() (uint8, error) {
	_, err := s.sendCommand([]byte{CmdGetRev, 0})
	if err != nil {
		return 0, err
	}

	response := make([]byte, 9)
	_, err = s.i2c.ReadBytes(response)
	return response[1], err
}

func (s *Si4713) PowerUp() error {
	_, err := s.sendCommand([]byte{CmdPowerUp, 0x12, 0x50})
	if err != nil {
		return err
	}

	err = s.setProperty(PropRefClkFreq, 32768)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxPreemphasis, 1)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxACompGain, 10)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxACompEnable, 0)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 500)

	rev, err := s.getRev()
	if err != nil {
		return err
	}

	if rev != 13 {
		return fmt.Errorf("rev = %d, expected 13", rev)
	}

	return nil
}

// freq: Frequency in MHz * 100
func (s *Si4713) SetFrequency(freq uint16) error {
	_, err := s.sendCommand([]byte{CmdSetTxFreq, 0, byte(freq >> 8), byte(freq & 0xFF)})
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 50)

	for range 10 {
		status, err := s.getStatus()
		if err != nil {
			return err
		}

		if status&0x81 == 0x81 {
			return nil
		}

		if status&0x40 == 0x40 {
			return fmt.Errorf("error bit set: invalid argument")
		}

		time.Sleep(time.Millisecond * 10)
	}

	return fmt.Errorf("failed to get response status")
}

// Power must be from 88 to 115 dBuV
func (s *Si4713) SetPower(power uint8) error {
	_, err := s.sendCommand([]byte{CmdSetTxPower, 0, 0, power, 0})
	return err
}

func (s *Si4713) TuneMeasure(freq uint16) error {
	if freq%5 != 0 {
		freq -= freq % 5
	}

	_, err := s.sendCommand([]byte{CmdTxTuneMeasure, 0, byte(freq >> 8), byte(freq & 0xFF), 0})
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 50)

	for {
		status, err := s.getStatus()
		if err != nil {
			return err
		}

		if status == 0x81 {
			break
		}

		time.Sleep(time.Millisecond * 10)
	}

	return nil
}

type TuneStatus struct {
	Frequency        uint16
	Power            uint8 // dBuV
	AntennaCapacitor uint8
	NoiseLevel       uint8
}

func (s *Si4713) TuneStatus() (TuneStatus, error) {
	_, err := s.sendCommand([]byte{CmdTxTuneStatus, 1})
	if err != nil {
		return TuneStatus{}, err
	}

	response := make([]byte, 8)
	_, err = s.i2c.ReadBytes(response)

	return TuneStatus{
		Frequency:        (uint16(response[2]) << 8) | uint16(response[3]),
		Power:            response[5],
		AntennaCapacitor: response[6],
		NoiseLevel:       response[7],
	}, err
}

type ASQStatus struct {
	ASQ     uint8
	InLevel uint8
}

func (s *Si4713) ReadASQ() (ASQStatus, error) {
	_, err := s.sendCommand([]byte{CmdTxASQStatus, 1})
	if err != nil {
		return ASQStatus{}, err
	}

	response := make([]byte, 5)
	_, err = s.i2c.ReadBytes(response)

	return ASQStatus{
		ASQ:     response[1],
		InLevel: response[4],
	}, err
}

func (s *Si4713) BeginRDS(programID uint16) error {
	err := s.setProperty(PropTxAudioDeviation, 6625)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSDeviation, 200)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSInterruptSource, 1)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSPI, programID)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSPsMix, 3)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSPsMisc, 0x1008)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSPsRepeatCount, 3)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSMessageCount, 1)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSPsAF, 0xE0E0)
	if err != nil {
		return err
	}

	err = s.setProperty(PropTxRDSFIFOSize, 0)
	if err != nil {
		return err
	}

	return s.setProperty(PropTxComponentEnable, 7)
}

func (s *Si4713) SetRDSPS(str string) error {
	if len(str) > 8 {
		str = str[:8]
	} else if len(str) < 8 {
		str += strings.Repeat(" ", 8-len(str))
	}

	buf := make([]byte, 6)
	buf[0] = CmdTxRDSPs

	copy(buf[2:], str[:4])
	_, err := s.sendCommand(buf)
	if err != nil {
		return err
	}

	buf[1] = 1
	copy(buf[2:], str[4:])
	_, err = s.sendCommand(buf)
	return err
}

func (s *Si4713) SetRDSText(str string) error {
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
		_, err := s.sendCommand(buf)
		if err != nil {
			return err
		}

		if i == 0 {
			buf[1] = 0x04
		}
		i++
	}

	return nil
}

func (s *Si4713) SetGPIOCtrl(x uint8) error {
	_, err := s.sendCommand([]byte{CmdGPOCtl, x})
	return err
}

func (s *Si4713) SetGPIO(x uint8) error {
	_, err := s.sendCommand([]byte{CmdGPOSet, x})
	return err
}

func (s *Si4713) Reset() error {
	err := s.resetLine.SetValue(1)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)

	err = s.resetLine.SetValue(0)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)

	err = s.resetLine.SetValue(1)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)
	return nil
}

func (s *Si4713) Close() error {
	err := s.Reset()
	if err != nil {
		return err
	}

	err = s.resetLine.Close()
	if err != nil {
		return err
	}

	return s.i2c.Close()
}
