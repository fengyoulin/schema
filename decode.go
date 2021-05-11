package schema

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// Decoder in binary mode
type Decoder struct {
	io.Reader
	Extend map[reflect.Type]func(reflect.Value, *Decoder) error
}

// Decode the data
func (d *Decoder) Decode(a interface{}) (err error) {
	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not pointer", a)
	}
	return d.InternalDecode(rv)
}

// InternalDecode should be used to extend only
func (d *Decoder) InternalDecode(rv reflect.Value) (err error) {
	br, ok := d.Reader.(io.ByteReader)
	if !ok {
		br = &byteReader{d.Reader}
	}
	switch rv.Kind() {
	case reflect.Bool:
		var c byte
		if c, err = br.ReadByte(); err != nil {
			return
		}
		rv.SetBool(c != 0)
	case reflect.Int8:
		var i int8
		if *(*byte)(unsafe.Pointer(&i)), err = br.ReadByte(); err != nil {
			return
		}
		rv.SetInt(int64(i))
	case reflect.Uint8:
		var u uint8
		if u, err = br.ReadByte(); err != nil {
			return
		}
		rv.SetUint(uint64(u))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = binary.ReadVarint(br); err != nil {
			return
		}
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var u uint64
		if u, err = binary.ReadUvarint(br); err != nil {
			return
		}
		rv.SetUint(u)
	case reflect.Float32:
		var v float32
		var n int
		if n, err = d.Reader.Read((*(*[4]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
		if n != 4 {
			return io.EOF
		}
		rv.SetFloat(float64(v))
	case reflect.Float64:
		var v float64
		var n int
		if n, err = d.Reader.Read((*(*[8]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
		if n != 8 {
			return io.EOF
		}
		rv.SetFloat(v)
	case reflect.Complex64:
		var v complex64
		var n int
		if n, err = d.Reader.Read((*(*[8]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
		if n != 8 {
			return io.EOF
		}
		rv.SetComplex(complex128(v))
	case reflect.Complex128:
		var v complex128
		var n int
		if n, err = d.Reader.Read((*(*[16]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
		if n != 16 {
			return io.EOF
		}
		rv.SetComplex(v)
	case reflect.String:
		var i int64
		if i, err = binary.ReadVarint(br); err != nil || i <= 0 {
			return
		}
		slc := make([]byte, i)
		var n int
		if n, err = d.Reader.Read(slc); err != nil {
			return
		}
		if n != len(slc) {
			return io.EOF
		}
		rv.SetString(*(*string)(unsafe.Pointer(&slc)))
	case reflect.Slice:
		var i int64
		if i, err = binary.ReadVarint(br); err != nil || i <= 0 {
			return
		}
		if rv.IsNil() || rv.Cap() < int(i) {
			rv.Set(reflect.MakeSlice(rv.Type(), int(i), int(i)))
		} else {
			rv.SetLen(int(i))
		}
		for x := 0; x < int(i); x++ {
			if err = d.InternalDecode(rv.Index(x)); err != nil {
				return
			}
		}
	case reflect.Array:
		for x := 0; x < rv.Len(); x++ {
			if err = d.InternalDecode(rv.Index(x)); err != nil {
				return
			}
		}
	case reflect.Map:
		var i int64
		if i, err = binary.ReadVarint(br); err != nil || i <= 0 {
			return
		}
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rv.Type()))
		}
		for x := 0; x < int(i); x++ {
			k := reflect.New(rv.Type().Key()).Elem()
			v := reflect.New(rv.Type().Elem()).Elem()
			if err = d.InternalDecode(k); err != nil {
				return
			}
			if err = d.InternalDecode(v); err != nil {
				return
			}
			rv.SetMapIndex(k, v)
		}
	case reflect.Struct:
		if fn, ok := d.Extend[rv.Type()]; ok {
			return fn(rv, d)
		}
		for x := 0; x < rv.NumField(); x++ {
			if err = d.InternalDecode(rv.Field(x)); err != nil {
				return
			}
		}
	case reflect.Ptr:
		var c byte
		if c, err = br.ReadByte(); err != nil || c <= 0 {
			return
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if err = d.InternalDecode(rv.Elem()); err != nil {
			return
		}
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		fallthrough
	default:
		return fmt.Errorf("unexpected kind: %T", rv.Kind())
	}
	return
}

type byteReader struct {
	io.Reader
}

func (r *byteReader) ReadByte() (b byte, err error) {
	var buf [1]byte
	n, err := r.Read(buf[:])
	if err != nil {
		return
	}
	if n < 1 {
		return 0, io.EOF
	}
	return buf[0], nil
}
