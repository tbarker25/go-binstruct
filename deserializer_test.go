package binstruct

import (
	"testing"
	"time"
)

func TestDeserialize(t *testing.T) {
	t.Parallel()
	v := struct {
		TimeUTC time.Time
	}{}

	input := []byte{0xa0, 0x5f, 0xfc, 0x0c, 0x5d, 0x03, 0x00, 0x00}

	Deserialize(&v, input)

	want := "2017-05-09T01:31:48Z"
	if v.TimeUTC.Format(time.RFC3339Nano) != want {
		t.Errorf("got: %s, want: %s", v.TimeUTC.Format(time.RFC3339Nano), want)
	}
}
