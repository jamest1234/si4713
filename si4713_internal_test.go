package si4713

import (
	"log"
	"testing"

	"github.com/warthog618/go-gpiocdev"
)

const ResetPin = 4

func TestResetPin(t *testing.T) {
	resetLine, err := gpiocdev.RequestLine("gpiochip0", ResetPin, gpiocdev.AsOutput())
	if err != nil {
		t.Fatal(err)
	}

	defer resetLine.Close()

	si4713 := Si4713{resetLine: resetLine}

	err = si4713.Reset()
	if err != nil {
		log.Fatal(err)
	}
}
