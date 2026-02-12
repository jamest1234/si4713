package si4713

import (
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	t.Log("Started")

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
