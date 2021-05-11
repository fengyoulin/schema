package schema_test

import (
	"database/sql"
	"encoding/json"
	"github.com/fengyoulin/schema"
	"reflect"
	"testing"
	"time"
)

const testDef = `
{"name":"Record","fields":[
{"name":"ID","type":"uint","tags":{"json":"id,omitempty"}},
{"name":"CreatedAt","type":"Time","tags":{"json":"created_at,omitempty"}},
{"name":"UpdatedAt","type":"Time","tags":{"json":"updated_at,omitempty"}},
{"name":"DeletedAt","type":"NullTime","tags":{"json":"deleted_at,omitempty"}},
{"name":"UUID","type":"string","tags":{"json":"uuid,omitempty"}},
{"name":"OpenID","type":"string","tags":{"json":"open_id,omitempty"}},
{"name":"UnionID","type":"NullString","tags":{"json":"union_id,omitempty"}},
{"name":"Level","type":"float64","tags":{"json":"level,omitempty"}},
{"name":"Extra","type":"[]byte","tags":{"json":"extra,omitempty"}}
]}
`

var (
	testTypes *schema.Types
	testType  reflect.Type
)

func init() {
	testTypes = schema.New()
}

func TestTypes_AddType(t *testing.T) {
	tp := reflect.TypeOf(time.Time{})
	if err := testTypes.AddType(tp); err != nil {
		t.Error(err)
	}
	tp = reflect.TypeOf(sql.NullTime{})
	if err := testTypes.AddType(tp); err != nil {
		t.Error(err)
	}
	tp = reflect.TypeOf(sql.NullString{})
	if err := testTypes.AddType(tp); err != nil {
		t.Error(err)
	}
}

func TestTypes_CreateSchema(t *testing.T) {
	var s schema.Schema
	if err := json.Unmarshal([]byte(testDef), &s); err != nil {
		t.Error(err)
	}
	if tp, err := testTypes.CreateSchema(s); err != nil {
		t.Error(err)
	} else {
		testType = tp
	}
}

func TestTypes_CreateType(t *testing.T) {
	if _, err := testTypes.CreateType("map[uint]Record"); err != nil {
		t.Error(err)
	}
}

func TestTypes_TypeByName(t *testing.T) {
	tp, ok := testTypes.TypeByName("Record")
	if !ok {
		t.Error("not found")
	}
	if tp != testType {
		t.Errorf("%v != %v", tp, testType)
	}
}

func TestTypes_NameByType(t *testing.T) {
	nm, ok := testTypes.NameByType(testType)
	if !ok {
		t.Error("not found")
	}
	if nm != "Record" {
		t.Errorf("unexpected name: %s", nm)
	}
}
