package schema

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// Encoder in binary mode
type Encoder struct {
	io.Writer
	Extend map[reflect.Type]func(reflect.Value, *Encoder) error
	Types  *Types
}

// Encode the data
func (e *Encoder) Encode(a interface{}) (err error) {
	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not pointer", a)
	}
	return e.InternalEncode(rv)
}

// InternalEncode should be used to extend only
func (e *Encoder) InternalEncode(rv reflect.Value) (err error) {
	switch rv.Kind() {
	case reflect.Bool:
		var b [1]byte
		*(*bool)(unsafe.Pointer(&b[0])) = rv.Bool()
		if _, err = e.Writer.Write(b[:]); err != nil {
			return
		}
	case reflect.Int8:
		var b [1]byte
		*(*int8)(unsafe.Pointer(&b[0])) = int8(rv.Int())
		if _, err = e.Writer.Write(b[:]); err != nil {
			return
		}
	case reflect.Uint8:
		var b [1]byte
		b[0] = uint8(rv.Uint())
		if _, err = e.Writer.Write(b[:]); err != nil {
			return
		}
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		var buf [32]byte
		l := binary.PutVarint(buf[:], rv.Int())
		if _, err = e.Writer.Write(buf[:l]); err != nil {
			return
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var buf [32]byte
		l := binary.PutUvarint(buf[:], rv.Uint())
		if _, err = e.Writer.Write(buf[:l]); err != nil {
			return
		}
	case reflect.Float32:
		v := float32(rv.Float())
		if _, err = e.Writer.Write((*(*[4]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
	case reflect.Float64:
		v := rv.Float()
		if _, err = e.Writer.Write((*(*[8]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
	case reflect.Complex64:
		v := complex64(rv.Complex())
		if _, err = e.Writer.Write((*(*[8]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
	case reflect.Complex128:
		v := rv.Complex()
		if _, err = e.Writer.Write((*(*[16]byte)(unsafe.Pointer(&v)))[:]); err != nil {
			return
		}
	case reflect.String:
		var slc []byte
		str := rv.String()
		*(*string)(unsafe.Pointer(&slc)) = str
		(*reflect.SliceHeader)(unsafe.Pointer(&slc)).Cap = len(str)
		var buf [32]byte
		n := binary.PutVarint(buf[:], int64(len(slc)))
		if _, err = e.Writer.Write(buf[:n]); err != nil {
			return
		}
		if _, err = e.Writer.Write(slc); err != nil {
			return
		}
	case reflect.Slice:
		l := int64(rv.Len())
		var buf [32]byte
		n := binary.PutVarint(buf[:], l)
		if _, err = e.Writer.Write(buf[:n]); err != nil {
			return
		}
		switch rv.Type().Elem().Kind() {
		case reflect.Int8, reflect.Uint8, reflect.Bool:
			var b []byte
			*(*reflect.SliceHeader)(unsafe.Pointer(&b)) = reflect.SliceHeader{
				Data: rv.Pointer(),
				Len:  rv.Len(),
				Cap:  rv.Cap(),
			}
			_, err = e.Writer.Write(b)
			return
		}
		for x := 0; x < int(l); x++ {
			if err = e.InternalEncode(rv.Index(x)); err != nil {
				return
			}
		}
	case reflect.Array:
		for x := 0; x < rv.Len(); x++ {
			if err = e.InternalEncode(rv.Index(x)); err != nil {
				return
			}
		}
	case reflect.Map:
		l := int64(rv.Len())
		var buf [32]byte
		n := binary.PutVarint(buf[:], l)
		if _, err = e.Writer.Write(buf[:n]); err != nil {
			return
		}
		it := rv.MapRange()
		for it.Next() {
			if err = e.InternalEncode(it.Key()); err != nil {
				return
			}
			if err = e.InternalEncode(it.Value()); err != nil {
				return
			}
		}
	case reflect.Struct:
		if fn, ok := e.Extend[rv.Type()]; ok {
			return fn(rv, e)
		}
		n := rv.NumField()
		for x := 0; x < n; x++ {
			if err = e.InternalEncode(rv.Field(x)); err != nil {
				return
			}
		}
	case reflect.Ptr:
		var b [1]byte
		*(*bool)(unsafe.Pointer(&b[0])) = !rv.IsNil()
		if _, err = e.Writer.Write(b[:]); err != nil {
			return
		}
		if b[0] > 0 {
			if err = e.InternalEncode(rv.Elem()); err != nil {
				return
			}
		}
	case reflect.Interface:
		if rv.IsNil() {
			if _, err = e.Writer.Write([]byte{0}); err != nil {
				return
			}
		} else {
			tp := rv.Elem().Type()
			if e.Types == nil {
				return fmt.Errorf("unknown type: %s", tp.String())
			}
			nm, ok := e.Types.NameByType(tp)
			if !ok {
				return fmt.Errorf("unknown type: %s", tp.String())
			}
			if err = e.InternalEncode(reflect.ValueOf(&nm).Elem()); err != nil {
				return
			}
			return e.InternalEncode(rv.Elem())
		}
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		fallthrough
	default:
		return fmt.Errorf("unexpected kind: %v", rv.Kind())
	}
	return
}
