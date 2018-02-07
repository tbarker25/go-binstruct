package binstruct

import (
	"errors"
	"fmt"
	"reflect"
)

const epoch = -2203891200000

// Serialize a structure into the binstruct format.
func Serialize(v interface{}) ([]byte, error) {
	return serialize(reflect.ValueOf(v), []byte{}, 0)
}

func serialize(v reflect.Value, b []byte, tags tags) (_b []byte, _ error) {
	if tags&tagLen32 != 0 && v.Type().Kind() != reflect.Slice {
		offset := len(b)
		b = append(b, 0x00, 0x00, 0x00, 0x00)
		defer func() {
			val := uint32(len(_b) - offset - 4)
			_b[offset+0] = uint8(val >> 0)
			_b[offset+1] = uint8(val >> 8)
			_b[offset+2] = uint8(val >> 16)
			_b[offset+3] = uint8(val >> 24)
		}()
	}

	switch v.Type().Kind() {
	case reflect.Ptr:
		return serialize(v.Elem(), b, 0)

	case reflect.Bool:
		if v.Bool() {
			return append(b, 0x01), nil
		}
		return append(b, 0x00), nil

	case reflect.Uint8:
		return append(b, uint8(v.Uint())), nil

	case reflect.Uint16:
		val := uint16(v.Uint())
		return append(b,
			uint8(val),
			uint8(val>>8),
		), nil

	case reflect.Uint32:
		val := uint32(v.Uint())
		return append(b,
			uint8(val>>0),
			uint8(val>>8),
			uint8(val>>16),
			uint8(val>>24),
		), nil

	case reflect.Int32:
		val := uint32(v.Int())
		return append(b,
			uint8(val>>0),
			uint8(val>>8),
			uint8(val>>16),
			uint8(val>>24),
		), nil

	case reflect.Uint64:
		val := v.Uint()
		return append(b,
			uint8(val>>0),
			uint8(val>>8),
			uint8(val>>16),
			uint8(val>>24),
			uint8(val>>32),
			uint8(val>>40),
			uint8(val>>48),
			uint8(val>>56),
		), nil

	case reflect.Int64:
		val := uint64(v.Int())
		return append(b,
			uint8(val>>0),
			uint8(val>>8),
			uint8(val>>16),
			uint8(val>>24),
			uint8(val>>32),
			uint8(val>>40),
			uint8(val>>48),
			uint8(val>>56),
		), nil

	case reflect.String:
		s := v.String()
		if len(s) > 0xff {
			return nil, errors.New("string is longer than allowed")
		}

		b = append(b, uint8(len(s)))
		b = append(b, s...)
		return b, nil

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)
			tags, err := decodeTags(fieldType.Tag.Get("binstruct"))
			if err != nil {
				return nil, err
			}

			if tags&tagSkip != 0 {
				continue
			}

			b, err = serialize(field, b, tags)
			if err != nil {
				return nil, err
			}
		}
		return b, nil

	case reflect.Slice:
		length := v.Len()
		if tags&tagCount16 != 0 {
			if length > 0xffff {
				return nil, errors.New("slice is longer than allowed")
			}
			b = append(b, uint8(length), uint8(length>>8))

		} else if tags&tagCount0 == 0 {
			if length > 0xff {
				return nil, errors.New("slice is longer than allowed")
			}
			b = append(b, uint8(length))
		}

		for i := 0; i < length; i++ {
			var err error
			b, err = serialize(v.Index(i), b, tags>>8)
			if err != nil {
				return nil, err
			}
		}
		return b, nil

	case reflect.Array:
		length := v.Len()
		for i := 0; i < length; i++ {
			var err error
			b, err = serialize(v.Index(i), b, tags>>8)
			if err != nil {
				return nil, err
			}
		}
		return b, nil

	case reflect.Interface:
		return serialize(v.Elem(), b, 0)

	default:
		return nil, fmt.Errorf("unsupported type: " + v.Type().Kind().String())
	}
}
