package schema_test

import (
	"bytes"
	"database/sql"
	"github.com/fengyoulin/schema"
	"reflect"
	"testing"
	"time"
)

type testType struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
	UUID      string
	OpenID    string
	UnionID   sql.NullString
	Level     float64
	Extra     []byte
}

var (
	testObj testType
	testEnc map[reflect.Type]func(reflect.Value, *schema.Encoder) error
)

func init() {
	testObj = testType{
		ID:        1,
		CreatedAt: time.Unix(1620560345, 0),
		UpdatedAt: time.Unix(1620560345, 0),
		DeletedAt: sql.NullTime{},
		UUID:      "123456",
		OpenID:    "abcdef",
		UnionID:   sql.NullString{Valid: true, String: "!@#$%^"},
		Level:     3.14,
		Extra:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}
	testEnc = map[reflect.Type]func(reflect.Value, *schema.Encoder) error{
		reflect.TypeOf(time.Time{}): func(value reflect.Value, encoder *schema.Encoder) error {
			var ts int64
			if value.CanAddr() {
				ts = value.Addr().Interface().(*time.Time).Unix()
			} else {
				ts = value.Interface().(time.Time).Unix()
			}
			return encoder.InternalEncode(reflect.ValueOf(&ts).Elem())
		},
		reflect.TypeOf(sql.NullTime{}): func(value reflect.Value, encoder *schema.Encoder) error {
			var tm sql.NullTime
			if value.CanAddr() {
				tm = *value.Addr().Interface().(*sql.NullTime)
			} else {
				tm = value.Interface().(sql.NullTime)
			}
			if err := encoder.InternalEncode(reflect.ValueOf(&tm.Valid).Elem()); err != nil {
				return err
			}
			if !tm.Valid {
				return nil
			}
			return encoder.InternalEncode(reflect.ValueOf(&tm.Time).Elem())
		},
	}
}

func TestEncoder_Encode(t *testing.T) {
	b := &bytes.Buffer{}
	e := &schema.Encoder{
		Writer: b,
		Extend: testEnc,
	}
	if err := e.Encode(&testObj); err != nil {
		t.Error(err)
	}
	if d := b.Bytes(); !reflect.DeepEqual(&d, &testBuf) {
		t.Errorf("%v != %v", &d, &testBuf)
	}
}

func BenchmarkEncoder_Encode(b *testing.B) {
	w := &bytes.Buffer{}
	e := &schema.Encoder{
		Writer: w,
		Extend: testEnc,
	}
	for i := 0; i < b.N; i++ {
		w.Reset()
		if err := e.Encode(&testObj); err != nil {
			b.Error(err)
		}
	}
}
