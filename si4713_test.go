package si4713

import (
	"slices"
	"testing"
	"time"
)

func TestNoiseMeasure(t *testing.T) {
	tx, err := New(0x63, 4)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Close()

	results := [][2]int{}

	for f := uint16(8750); f < 10800; f += 10 {
		tx.TuneMeasure(f)
		t.Logf("Measuring %d...", f)
		results = append(results, [2]int{int(f), int(tx.TuneStatus().NoiseLevel)})
	}

	slices.SortFunc(results, func(a, b [2]int) int {
		return a[1] - b[1]
	})

	for _, r := range results {
		t.Log(r)
	}
}

func TestTransmision(t *testing.T) {
	tx, err := New(0x63, 4)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Close()

	tx.SetPower(115)

	tx.SetFrequency(9930)

	tx.BeginRDS(0x3333)
	tx.SetRDSPS("Test")
	tx.SetRDSText("Testing testing testing testing 12345")

	time.Sleep(time.Second * 10)

	t.Log(tx.powerDown())
}
