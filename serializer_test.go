package binstruct

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"
)

func TestSerialize(t *testing.T) {
	t.Parallel()
	v := struct {
		TimeUTC time.Time
	}{}

	var err error
	v.TimeUTC, err = time.Parse(time.RFC3339Nano, "2017-05-09T01:31:48Z")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	buf, err := Serialize(&v)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !bytes.Equal(buf, []byte{0xa0, 0x5f, 0xfc, 0x0c, 0x5d, 0x03, 0x00, 0x00}) {
		t.Errorf("%s", hex.Dump(buf))
	}
}
