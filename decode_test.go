package schema_test

import (
	"bytes"
	"database/sql"
	"github.com/fengyoulin/schema"
	"reflect"
	"testing"
	"time"
)

var (
	testBuf = []byte{1, 1, 178, 167, 190, 137, 12, 178, 167, 190, 137, 12, 0, 12, 49, 50, 51, 52, 53, 54, 12, 97, 98, 99, 100, 101, 102, 12, 33, 64, 35, 36, 37, 94, 1, 31, 133, 235, 81, 184, 30, 9, 64, 20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 8, 98, 111, 111, 108, 1}
	testDec map[reflect.Type]func(reflect.Value, *schema.Decoder) error
)

func init() {
	testDec = map[reflect.Type]func(reflect.Value, *schema.Decoder) error{
		reflect.TypeOf(time.Time{}): func(value reflect.Value, decoder *schema.Decoder) error {
			var ts int64
			if err := decoder.InternalDecode(reflect.ValueOf(&ts).Elem()); err != nil {
				return err
			}
			value.Set(reflect.ValueOf(time.Unix(ts, 0)))
			return nil
		},
		reflect.TypeOf(sql.NullTime{}): func(value reflect.Value, decoder *schema.Decoder) error {
			var tm sql.NullTime
			if err := decoder.InternalDecode(reflect.ValueOf(&tm.Valid).Elem()); err != nil {
				return err
			}
			if !tm.Valid {
				return nil
			}
			if err := decoder.InternalDecode(reflect.ValueOf(&tm.Time).Elem()); err != nil {
				return err
			}
			value.Set(reflect.ValueOf(&tm).Elem())
			return nil
		},
	}
}

func TestDecoder_Decode(t *testing.T) {
	var o testStruct
	d := &schema.Decoder{
		Reader: bytes.NewBuffer(testBuf),
		Extend: testDec,
		Types:  testTypes,
	}
	if err := d.Decode(&o); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&o, &testObj) {
		t.Errorf("%v != %v", &o, &testObj)
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {
	d := &schema.Decoder{
		Extend: testDec,
		Types:  testTypes,
	}
	for i := 0; i < b.N; i++ {
		var o testStruct
		d.Reader = bytes.NewReader(testBuf)
		if err := d.Decode(&o); err != nil {
			b.Error(err)
		}
	}
}
