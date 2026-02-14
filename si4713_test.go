package si4713_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/jamest1234/si4713"
	"github.com/warthog618/go-gpiocdev"
)

const ResetPin = 4

func getI2CAndLine() (*i2c.I2C, *gpiocdev.Line, error) {
	var err error

	i2cDevice, err := i2c.NewI2C(0x63, 1)
	if err != nil {
		return nil, nil, err
	}

	resetLine, err := gpiocdev.RequestLine("gpiochip0", ResetPin, gpiocdev.AsOutput())
	return i2cDevice, resetLine, err
}

func TestNoiseMeasure(t *testing.T) {
	i2cDevice, resetLine, err := getI2CAndLine()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := si4713.New(i2cDevice, resetLine)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Close()

	err = tx.PowerUp()
	if err != nil {
		t.Fatal(err)
	}

	results := [][2]int{}

	for f := uint16(8750); f < 10800; f += 10 {
		err = tx.TuneMeasure(f)
		if err != nil {
			t.Fatal(err)
		}

		tuneStatus, err := tx.TuneStatus()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Measuring %d...", f)
		results = append(results, [2]int{int(f), int(tuneStatus.NoiseLevel)})
	}

	slices.SortFunc(results, func(a, b [2]int) int {
		return b[1] - a[1]
	})

	for _, r := range results {
		t.Log(r)
	}
}

func TestTransmission(t *testing.T) {
	i2cDevice, resetLine, err := getI2CAndLine()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := si4713.New(i2cDevice, resetLine)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Close()

	t.Log("Powering up")
	err = tx.PowerUp()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Setting power")
	err = tx.SetPower(115)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Setting frequency")
	err = tx.SetFrequency(9730)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Starting RDS")
	err = tx.BeginRDS(0x3333)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.SetRDSPSName("Test")
	if err != nil {
		t.Fatal(err)
	}

	err = tx.SetRDSRadioText("Testing testing testing testing 12345")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Sleeping for 10 seconds")
	time.Sleep(time.Second * 10)

	t.Log("Resetting")
	err = tx.Reset()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPowerUpPowerDown(t *testing.T) {
	i2cDevice, resetLine, err := getI2CAndLine()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := si4713.New(i2cDevice, resetLine)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Close()

	for i := range 5 {
		t.Log("Powering up")
		err = tx.PowerUp()
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Setting power")
		err = tx.SetPower(115)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Setting frequency")
		err = tx.SetFrequency(9730)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Starting RDS")
		err = tx.BeginRDS(0x3333)
		if err != nil {
			t.Fatal(err)
		}

		err = tx.SetRDSPSName("Hello")
		if err != nil {
			t.Fatal(err)
		}

		err = tx.SetRDSRadioText(fmt.Sprintf("Test #%d", i+1))
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Sleeping for 3 seconds")
		time.Sleep(time.Second * 3)

		t.Log("Resetting")
		err = tx.Reset()
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 3)
	}
}
