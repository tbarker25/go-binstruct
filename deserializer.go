package binstruct

import (
	"errors"
	"fmt"
	"reflect"
)

var customDeserializers = map[string]func([]byte) (interface{}, []byte){}
var errIndexOutOfBounds = errors.New("index out of bounds")

// Deserialize a binstruct structure, returning the unused bytes.
func Deserialize(v interface{}, buf []byte) ([]byte, error) {
	buf, err := deserialize(reflect.ValueOf(v), buf, 0)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Register a custom demarshaller function for an interface
func Register(value interface{}, fn func([]byte) (interface{}, []byte)) {
	rt := reflect.TypeOf(value).Elem()
	customDeserializers[rt.String()] = fn
}

func deserialize(v reflect.Value, b []byte, tags tags) (_b []byte, _ error) {
	if tags&tagLen32 != 0 && v.Type().Kind() != reflect.Slice {
		if len(b) < 4 {
			return nil, errIndexOutOfBounds
		}

		length := uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24
		b = b[4:]

		if len(b) < int(length) {
			return nil, errIndexOutOfBounds
		}

		defer func(b []byte) {
			_b = b
		}(b[length:])
		b = b[:length]
	}

	vt := v.Type()
	switch vt.Kind() {
	case reflect.Ptr:
		return deserialize(v.Elem(), b, tags)

	case reflect.Bool:
		if len(b) < 1 {
			return nil, errIndexOutOfBounds
		}
		val := b[0]
		b = b[1:]
		v.SetBool(val != 0)
		return b, nil

	case reflect.Uint8:
		if len(b) < 1 {
			return nil, errIndexOutOfBounds
		}
		val := b[0]
		b = b[1:]
		v.SetUint(uint64(val))
		return b, nil

	case reflect.Uint16:
		if len(b) < 2 {
			return nil, errIndexOutOfBounds
		}
		val := uint16(b[0]) |
			uint16(b[1])<<8
		b = b[2:]
		v.SetUint(uint64(val))
		return b, nil

	case reflect.Uint32:
		if len(b) < 4 {
			return nil, errIndexOutOfBounds
		}
		val := uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24
		b = b[4:]
		v.SetUint(uint64(val))
		return b, nil

	case reflect.Int32:
		if len(b) < 4 {
			return nil, errIndexOutOfBounds
		}
		val := uint32(b[0]) |
			uint32(b[1])<<8 |
			uint32(b[2])<<16 |
			uint32(b[3])<<24
		b = b[4:]
		v.SetInt(int64(val))
		return b, nil

	case reflect.Uint64:
		if len(b) < 8 {
			return nil, errIndexOutOfBounds
		}
		val := uint64(b[0])<<0 |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32 |
			uint64(b[5])<<40 |
			uint64(b[6])<<48 |
			uint64(b[7])<<56
		b = b[8:]
		v.SetUint(uint64(val))
		return b, nil

	case reflect.Int64:
		if len(b) < 8 {
			return nil, errIndexOutOfBounds
		}
		val := uint64(b[0])<<0 |
			uint64(b[1])<<8 |
			uint64(b[2])<<16 |
			uint64(b[3])<<24 |
			uint64(b[4])<<32 |
			uint64(b[5])<<40 |
			uint64(b[6])<<48 |
			uint64(b[7])<<56
		b = b[8:]
		v.SetInt(int64(val))
		return b, nil

	case reflect.String:
		if len(b) < 1 || len(b) < int(b[0])+1 {
			return nil, errIndexOutOfBounds
		}
		length := b[0]
		s := string(b[1 : length+1])
		b = b[1+length:]
		v.SetString(s)
		return b, nil

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := vt.Field(i)

			tags, err := decodeTags(fieldType.Tag.Get("binstruct"))
			if err != nil {
				return nil, err
			}

			if tags&tagSkip != 0 {
				continue
			}

			if len(b) == 0 && tags&tagOptionalEnd != 0 {
				continue
			}

			b, err = deserialize(field, b, tags)
			if err != nil {
				return nil, err
			}
		}
		return b, nil

	case reflect.Slice:
		var length int
		if tags&tagCount16 != 0 {
			if len(b) < 2 {
				return nil, errIndexOutOfBounds
			}
			length = int(uint16(b[0]) |
				uint16(b[1])<<8)
			b = b[2:]

		} else if tags&tagCount0 == 0 {
			if len(b) < 1 {
				return nil, errIndexOutOfBounds
			}
			length = int(b[0])
			b = b[1:]
		}

		slice := reflect.MakeSlice(vt, length, length)
		for i := 0; i < length; i++ {
			var err error
			b, err = deserialize(slice.Index(i), b, tags>>8)
			if err != nil {
				return nil, err
			}
		}
		v.Set(slice)
		return b, nil

	case reflect.Array:
		length := v.Len()
		for i := 0; i < length; i++ {
			var err error
			b, err = deserialize(v.Index(i), b, tags>>8)
			if err != nil {
				return nil, err
			}
		}
		return b, nil

	case reflect.Interface:
		fn, ok := customDeserializers[vt.String()]
		if !ok {
			return nil, fmt.Errorf("no demarshaller registered for interface: %s", vt.String())
		}

		val, b := fn(b)
		rv := reflect.ValueOf(val)
		v.Set(rv)
		return b, nil

	default:
		return nil, fmt.Errorf("unsupported type: " + v.Type().Kind().String())
	}
}
