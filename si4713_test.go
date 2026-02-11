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

	time.Sleep(time.Second * 2)

	t.Log(tx.powerDown())
}
