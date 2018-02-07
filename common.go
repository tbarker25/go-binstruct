package binstruct

import (
	"fmt"
	"strings"
)

type tags uint16

const (
	// tagSkip:
	// when this is set, skip this field when serializing and deserializing.
	tagSkip tags = 0x01

	// tagCount0:
	// when this bit is set for a slice field, don't use a length byte before
	// the values data.
	tagCount0 tags = 0x02

	// tagCount16:
	// when this bit is set for a slice field, use a 16-bit length field
	// instead of the default 8-bit field.
	tagCount16 tags = 0x04

	// tagLen16:
	// when this bit is set, prepend the total bytes this field uses.
	tagLen32 tags = 0x08

	// tagOptionalEnd:
	// when this bit is set, if there's no bytes remaining then
	// stop processing at this field. If this tag is not set then
	// an error would be thrown.
	tagOptionalEnd = 0x10
)

func decodeTags(s string) (tags, error) {
	var tags tags
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' })
	for _, t := range fields {
		switch t {
		case "count0":
			tags |= tagCount0
		case "count16":
			tags |= tagCount16
		case "len32":
			tags |= tagLen32
		case "-":
			tags |= tagSkip
		case "end":
			tags |= tagOptionalEnd

		// same tags but apply to slice and array elements
		case "[count0]":
			tags |= tagCount0 << 8
		case "[count16]":
			tags |= tagCount16 << 8
		case "[len32]":
			tags |= tagLen32 << 8
		case "[end]":
			tags |= tagOptionalEnd << 8

		default:
			return 0, fmt.Errorf("unknown tag: " + t)
		}
	}
	return tags, nil
}
