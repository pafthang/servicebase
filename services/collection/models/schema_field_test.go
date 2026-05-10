package models_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"

	"github.com/pafthang/servicebase/tools/types"
)

func TestBaseModelFieldNames(t *testing.T) {
	result := collectionmodels.BaseModelFieldNames()
	expected := 3

	if len(result) != expected {
		t.Fatalf("Expected %d field names, got %d (%v)", expected, len(result), result)
	}
}

func TestSystemFieldNames(t *testing.T) {
	result := collectionmodels.SystemFieldNames()
	expected := 3

	if len(result) != expected {
		t.Fatalf("Expected %d field names, got %d (%v)", expected, len(result), result)
	}
}

func TestUserFieldNames(t *testing.T) {
	result := collectionmodels.UserFieldNames()
	expected := 9

	if len(result) != expected {
		t.Fatalf("Expected %d user field names, got %d (%v)", expected, len(result), result)
	}
}

func TestFieldTypes(t *testing.T) {
	result := collectionmodels.FieldTypes()
	expected := 11

	if len(result) != expected {
		t.Fatalf("Expected %d types, got %d (%v)", expected, len(result), result)
	}
}

func TestArraybleFieldTypes(t *testing.T) {
	result := collectionmodels.ArraybleFieldTypes()
	expected := 3

	if len(result) != expected {
		t.Fatalf("Expected %d arrayble types, got %d (%v)", expected, len(result), result)
	}
}

func TestSchemaFieldColDefinition(t *testing.T) {
	scenarios := []struct {
		field    collectionmodels.SchemaField
		expected string
	}{
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber, Name: "test"},
			"NUMERIC DEFAULT 0 NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool, Name: "test"},
			"BOOLEAN DEFAULT FALSE NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson, Name: "test"},
			"JSON DEFAULT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Name: "test_multiple", Options: &collectionmodels.SelectOptions{MaxSelect: 2}},
			"JSON DEFAULT '[]' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Name: "test"},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Name: "test_multiple", Options: &collectionmodels.FileOptions{MaxSelect: 2}},
			"JSON DEFAULT '[]' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Name: "test", Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"TEXT DEFAULT '' NOT NULL",
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Name: "test_multiple", Options: &collectionmodels.RelationOptions{MaxSelect: nil}},
			"JSON DEFAULT '[]' NOT NULL",
		},
	}

	for i, s := range scenarios {
		def := s.field.ColDefinition()
		if def != s.expected {
			t.Errorf("(%d) Expected definition %q, got %q", i, s.expected, def)
		}
	}
}

func TestSchemaFieldString(t *testing.T) {
	f := collectionmodels.SchemaField{
		Id:          "abc",
		Name:        "test",
		Type:        collectionmodels.FieldTypeText,
		Required:    true,
		Presentable: true,
		System:      true,
		Options: &collectionmodels.TextOptions{
			Pattern: "test",
		},
	}

	result := f.String()
	expected := `{"system":true,"id":"abc","name":"test","type":"text","required":true,"presentable":true,"unique":false,"options":{"min":null,"max":null,"pattern":"test"}}`

	if result != expected {
		t.Errorf("Expected \n%v, got \n%v", expected, result)
	}
}

func TestSchemaFieldMarshalJSON(t *testing.T) {
	scenarios := []struct {
		field    collectionmodels.SchemaField
		expected string
	}{
		// empty
		{
			collectionmodels.SchemaField{},
			`{"system":false,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		// without defined options
		{
			collectionmodels.SchemaField{
				Id:          "abc",
				Name:        "test",
				Type:        collectionmodels.FieldTypeText,
				Required:    true,
				Presentable: true,
				System:      true,
			},
			`{"system":true,"id":"abc","name":"test","type":"text","required":true,"presentable":true,"unique":false,"options":{"min":null,"max":null,"pattern":""}}`,
		},
		// with defined options
		{
			collectionmodels.SchemaField{
				Name:     "test",
				Type:     collectionmodels.FieldTypeText,
				Required: true,
				System:   true,
				Options: &collectionmodels.TextOptions{
					Pattern: "test",
				},
			},
			`{"system":true,"id":"","name":"test","type":"text","required":true,"presentable":false,"unique":false,"options":{"min":null,"max":null,"pattern":"test"}}`,
		},
	}

	for i, s := range scenarios {
		result, err := s.field.MarshalJSON()
		if err != nil {
			t.Fatalf("(%d) %v", i, err)
		}

		if string(result) != s.expected {
			t.Errorf("(%d), Expected \n%v, got \n%v", i, s.expected, string(result))
		}
	}
}

func TestSchemaFieldUnmarshalJSON(t *testing.T) {
	scenarios := []struct {
		data        []byte
		expectError bool
		expectJson  string
	}{
		{
			nil,
			true,
			`{"system":false,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			[]byte{},
			true,
			`{"system":false,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			[]byte(`{"system": true}`),
			true,
			`{"system":true,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			[]byte(`{"invalid"`),
			true,
			`{"system":false,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			[]byte(`{"type":"text","system":true}`),
			false,
			`{"system":true,"id":"","name":"","type":"text","required":false,"presentable":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}`,
		},
		{
			[]byte(`{"type":"text","options":{"pattern":"test"}}`),
			false,
			`{"system":false,"id":"","name":"","type":"text","required":false,"presentable":false,"unique":false,"options":{"min":null,"max":null,"pattern":"test"}}`,
		},
	}

	for i, s := range scenarios {
		f := collectionmodels.SchemaField{}
		err := f.UnmarshalJSON(s.data)

		hasErr := err != nil
		if hasErr != s.expectError {
			t.Errorf("(%d) Expected hasErr %v, got %v (%v)", i, s.expectError, hasErr, err)
		}

		if f.String() != s.expectJson {
			t.Errorf("(%d), Expected json \n%v, got \n%v", i, s.expectJson, f.String())
		}
	}
}

func TestSchemaFieldValidate(t *testing.T) {
	scenarios := []struct {
		name           string
		field          collectionmodels.SchemaField
		expectedErrors []string
	}{
		{
			"empty field",
			collectionmodels.SchemaField{},
			[]string{"id", "options", "name", "type"},
		},
		{
			"missing id",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "",
				Name: "test",
			},
			[]string{"id"},
		},
		{
			"invalid id length check",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234",
				Name: "test",
			},
			[]string{"id"},
		},
		{
			"valid id length check",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "12345",
				Name: "test",
			},
			[]string{},
		},
		{
			"invalid name format",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "test!@#",
			},
			[]string{"name"},
		},
		{
			"name with _via_",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "a_via_b",
			},
			[]string{"name"},
		},
		{
			"reserved name (null)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "null",
			},
			[]string{"name"},
		},
		{
			"reserved name (true)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "null",
			},
			[]string{"name"},
		},
		{
			"reserved name (false)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "false",
			},
			[]string{"name"},
		},
		{
			"reserved name (_rowid_)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "_rowid_",
			},
			[]string{"name"},
		},
		{
			"reserved name (id)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameId,
			},
			[]string{"name"},
		},
		{
			"reserved name (created)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameCreated,
			},
			[]string{"name"},
		},
		{
			"reserved name (updated)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameUpdated,
			},
			[]string{"name"},
		},
		{
			"reserved name (collectionId)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameCollectionId,
			},
			[]string{"name"},
		},
		{
			"reserved name (collectionName)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameCollectionName,
			},
			[]string{"name"},
		},
		{
			"reserved name (expand)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: collectionmodels.FieldNameExpand,
			},
			[]string{"name"},
		},
		{
			"valid name",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeText,
				Id:   "1234567890",
				Name: "test",
			},
			[]string{},
		},
		{
			"unique check for type file",
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Id:      "1234567890",
				Name:    "test",
				Options: &collectionmodels.FileOptions{MaxSelect: 1, MaxSize: 1},
			},
			[]string{"unique"},
		},
		{
			"trigger options validator (auto init)",
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeFile,
				Id:   "1234567890",
				Name: "test",
			},
			[]string{"options"},
		},
		{
			"trigger options validator (invalid option field value)",
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Id:      "1234567890",
				Name:    "test",
				Options: &collectionmodels.FileOptions{MaxSelect: 0, MaxSize: 0},
			},
			[]string{"options"},
		},
		{
			"trigger options validator (valid option field value)",
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Id:      "1234567890",
				Name:    "test",
				Options: &collectionmodels.FileOptions{MaxSelect: 1, MaxSize: 1},
			},
			[]string{},
		},
	}

	for _, s := range scenarios {
		result := s.field.Validate()

		// parse errors
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("[%s] Failed to parse errors %v", s.name, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("[%s] Expected error keys %v, got %v", s.name, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("[%s] Missing expected error key %q in %v", s.name, k, errs)
			}
		}
	}
}

func TestSchemaFieldInitOptions(t *testing.T) {
	scenarios := []struct {
		field       collectionmodels.SchemaField
		expectError bool
		expectJson  string
	}{
		{
			collectionmodels.SchemaField{},
			true,
			`{"system":false,"id":"","name":"","type":"","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			collectionmodels.SchemaField{Type: "unknown"},
			true,
			`{"system":false,"id":"","name":"","type":"unknown","required":false,"presentable":false,"unique":false,"options":null}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText},
			false,
			`{"system":false,"id":"","name":"","type":"text","required":false,"presentable":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber},
			false,
			`{"system":false,"id":"","name":"","type":"number","required":false,"presentable":false,"unique":false,"options":{"min":null,"max":null,"noDecimal":false}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool},
			false,
			`{"system":false,"id":"","name":"","type":"bool","required":false,"presentable":false,"unique":false,"options":{}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail},
			false,
			`{"system":false,"id":"","name":"","type":"email","required":false,"presentable":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl},
			false,
			`{"system":false,"id":"","name":"","type":"url","required":false,"presentable":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor},
			false,
			`{"system":false,"id":"","name":"","type":"editor","required":false,"presentable":false,"unique":false,"options":{"convertUrls":false}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate},
			false,
			`{"system":false,"id":"","name":"","type":"date","required":false,"presentable":false,"unique":false,"options":{"min":"","max":""}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect},
			false,
			`{"system":false,"id":"","name":"","type":"select","required":false,"presentable":false,"unique":false,"options":{"maxSelect":0,"values":null}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			false,
			`{"system":false,"id":"","name":"","type":"json","required":false,"presentable":false,"unique":false,"options":{"maxSize":0}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile},
			false,
			`{"system":false,"id":"","name":"","type":"file","required":false,"presentable":false,"unique":false,"options":{"mimeTypes":null,"thumbs":null,"maxSelect":0,"maxSize":0,"protected":false}}`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			false,
			`{"system":false,"id":"","name":"","type":"relation","required":false,"presentable":false,"unique":false,"options":{"collectionId":"","cascadeDelete":false,"minSelect":null,"maxSelect":null,"displayFields":null}}`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeText,
				Options: &collectionmodels.TextOptions{Pattern: "test"},
			},
			false,
			`{"system":false,"id":"","name":"","type":"text","required":false,"presentable":false,"unique":false,"options":{"min":null,"max":null,"pattern":"test"}}`,
		},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("s%d_%s", i, s.field.Type), func(t *testing.T) {
			err := s.field.InitOptions()

			hasErr := err != nil
			if hasErr != s.expectError {
				t.Fatalf("Expected %v, got %v (%v)", s.expectError, hasErr, err)
			}

			if s.field.String() != s.expectJson {
				t.Fatalf(" Expected\n%v\ngot\n%v", s.expectJson, s.field.String())
			}
		})
	}
}

func TestSchemaFieldPrepareValue(t *testing.T) {
	scenarios := []struct {
		field      collectionmodels.SchemaField
		value      any
		expectJson string
	}{
		{collectionmodels.SchemaField{Type: "unknown"}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: "unknown"}, 123, "123"},
		{collectionmodels.SchemaField{Type: "unknown"}, []int{1, 2, 1}, "[1,2,1]"},

		// text
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText}, []int{1, 2}, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText}, 123, `"123"`},

		// email
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail}, []int{1, 2}, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail}, 123, `"123"`},

		// url
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl}, []int{1, 2}, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl}, 123, `"123"`},

		// editor
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor}, []int{1, 2}, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor}, 123, `"123"`},

		// json
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, nil, "null"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "null", "null"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, 123, "123"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, -123, "-123"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "123", "123"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "-123", "-123"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, 123.456, "123.456"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, -123.456, "-123.456"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "123.456", "123.456"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "-123.456", "-123.456"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "123.456 abc", `"123.456 abc"`}, // invalid numeric string
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "-a123", `"-a123"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, true, "true"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "true", "true"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, false, "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "false", "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `test`, `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `"test"`, `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `{test":1}`, `"{test\":1}"`}, // invalid object string
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `[1 2 3]`, `"[1 2 3]"`},      // invalid array string
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, map[string]int{}, `{}`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `{}`, `{}`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, map[string]int{"test": 123}, `{"test":123}`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `{"test":123}`, `{"test":123}`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, []int{}, `[]`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `[]`, `[]`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, []int{1, 2, 1}, `[1,2,1]`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson}, `[1,2,1]`, `[1,2,1]`},

		// number
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, nil, "0"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, "", "0"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, "test", "0"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, 1, "1"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, 1.5, "1.5"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber}, "1.5", "1.5"},

		// bool
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, nil, "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, 1, "true"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, 0, "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, "", "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, "test", "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, "false", "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, "true", "true"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, false, "false"},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool}, true, "true"},

		// date
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, "test", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, 1641024040, `"2022-01-01 08:00:40.000Z"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, "2022-01-01 11:27:10.123", `"2022-01-01 11:27:10.123Z"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, "2022-01-01 11:27:10.123Z", `"2022-01-01 11:27:10.123Z"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, types.DateTime{}, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate}, time.Time{}, `""`},

		// select (single)
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect}, 123, `"123"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect}, []string{"test1", "test2"}, `"test2"`},
		{
			// no values validation/filtering
			collectionmodels.SchemaField{
				Type: collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{
					Values: []string{"test1", "test2"},
				},
			},
			"test",
			`"test"`,
		},
		// select (multiple)
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			nil,
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			"",
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			[]string{},
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			123,
			`["123"]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			"test",
			`["test"]`,
		},
		{
			// no values validation
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			[]string{"test1", "test2", "test3"},
			`["test1","test2","test3"]`,
		},
		{
			// duplicated values
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			[]string{"test1", "test2", "test1"},
			`["test1","test2"]`,
		},

		// file (single)
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile}, nil, `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile}, "", `""`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile}, 123, `"123"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile}, "test", `"test"`},
		{collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile}, []string{"test1", "test2"}, `"test2"`},
		// file (multiple)
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			nil,
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			"",
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			[]string{},
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			123,
			`["123"]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			"test",
			`["test"]`,
		},
		{
			// no values validation
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			[]string{"test1", "test2", "test3"},
			`["test1","test2","test3"]`,
		},
		{
			// duplicated values
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 2},
			},
			[]string{"test1", "test2", "test1"},
			`["test1","test2"]`,
		},

		// relation (single)
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			nil,
			`""`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			"",
			`""`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			123,
			`"123"`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			"abc",
			`"abc"`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			"1ba88b4f-e9da-42f0-9764-9a55c953e724",
			`"1ba88b4f-e9da-42f0-9764-9a55c953e724"`,
		},
		{
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			[]string{"1ba88b4f-e9da-42f0-9764-9a55c953e724", "2ba88b4f-e9da-42f0-9764-9a55c953e724"},
			`"2ba88b4f-e9da-42f0-9764-9a55c953e724"`,
		},
		// relation (multiple)
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			nil,
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			"",
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			[]string{},
			`[]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			123,
			`["123"]`,
		},
		{
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			[]string{"", "abc"},
			`["abc"]`,
		},
		{
			// no values validation
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			[]string{"1ba88b4f-e9da-42f0-9764-9a55c953e724", "2ba88b4f-e9da-42f0-9764-9a55c953e724"},
			`["1ba88b4f-e9da-42f0-9764-9a55c953e724","2ba88b4f-e9da-42f0-9764-9a55c953e724"]`,
		},
		{
			// duplicated values
			collectionmodels.SchemaField{
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(2)},
			},
			[]string{"1ba88b4f-e9da-42f0-9764-9a55c953e724", "2ba88b4f-e9da-42f0-9764-9a55c953e724", "1ba88b4f-e9da-42f0-9764-9a55c953e724"},
			`["1ba88b4f-e9da-42f0-9764-9a55c953e724","2ba88b4f-e9da-42f0-9764-9a55c953e724"]`,
		},
	}

	for i, s := range scenarios {
		result := s.field.PrepareValue(s.value)

		encoded, err := json.Marshal(result)
		if err != nil {
			t.Errorf("(%d) %v", i, err)
			continue
		}

		if string(encoded) != s.expectJson {
			t.Errorf("(%d), Expected %v, got %v", i, s.expectJson, string(encoded))
		}
	}
}

func TestSchemaFieldPrepareValueWithModifier(t *testing.T) {
	scenarios := []struct {
		name          string
		field         collectionmodels.SchemaField
		baseValue     any
		modifier      string
		modifierValue any
		expectJson    string
	}{
		// text
		{
			"text with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText},
			"base",
			"+",
			"new",
			`"base"`,
		},
		{
			"text with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText},
			"base",
			"-",
			"new",
			`"base"`,
		},
		{
			"text with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText},
			"base",
			"?",
			"new",
			`"base"`,
		},
		{
			"text cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeText},
			123,
			"?",
			"new",
			`"123"`,
		},

		// number
		{
			"number with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber},
			1,
			"+",
			4,
			`5`,
		},
		{
			"number with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber},
			1,
			"-",
			4,
			`-3`,
		},
		{
			"number with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber},
			"1",
			"?",
			4,
			`1`,
		},
		{
			"number cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeNumber},
			"test",
			"+",
			"4",
			`4`,
		},

		// bool
		{
			"bool with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool},
			true,
			"+",
			false,
			`true`,
		},
		{
			"bool with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool},
			true,
			"-",
			false,
			`true`,
		},
		{
			"bool with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool},
			true,
			"?",
			false,
			`true`,
		},
		{
			"bool cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeBool},
			"true",
			"?",
			false,
			`true`,
		},

		// email
		{
			"email with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail},
			"base",
			"+",
			"new",
			`"base"`,
		},
		{
			"email with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail},
			"base",
			"-",
			"new",
			`"base"`,
		},
		{
			"email with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail},
			"base",
			"?",
			"new",
			`"base"`,
		},
		{
			"email cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEmail},
			123,
			"?",
			"new",
			`"123"`,
		},

		// url
		{
			"url with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl},
			"base",
			"+",
			"new",
			`"base"`,
		},
		{
			"url with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl},
			"base",
			"-",
			"new",
			`"base"`,
		},
		{
			"url with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl},
			"base",
			"?",
			"new",
			`"base"`,
		},
		{
			"url cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeUrl},
			123,
			"-",
			"new",
			`"123"`,
		},

		// editor
		{
			"editor with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor},
			"base",
			"+",
			"new",
			`"base"`,
		},
		{
			"editor with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor},
			"base",
			"-",
			"new",
			`"base"`,
		},
		{
			"editor with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor},
			"base",
			"?",
			"new",
			`"base"`,
		},
		{
			"editor cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeEditor},
			123,
			"-",
			"new",
			`"123"`,
		},

		// date
		{
			"date with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate},
			"2023-01-01 00:00:00.123",
			"+",
			"2023-02-01 00:00:00.456",
			`"2023-01-01 00:00:00.123Z"`,
		},
		{
			"date with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate},
			"2023-01-01 00:00:00.123Z",
			"-",
			"2023-02-01 00:00:00.456Z",
			`"2023-01-01 00:00:00.123Z"`,
		},
		{
			"date with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate},
			"2023-01-01 00:00:00.123",
			"?",
			"2023-01-01 00:00:00.456",
			`"2023-01-01 00:00:00.123Z"`,
		},
		{
			"date cast check",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeDate},
			1672524000, // 2022-12-31 22:00:00.000Z
			"+",
			100,
			`"2022-12-31 22:00:00.000Z"`,
		},

		// json
		{
			"json with '+' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			10,
			"+",
			5,
			`10`,
		},
		{
			"json with '+' modifier (slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			[]string{"a", "b"},
			"+",
			"c",
			`["a","b"]`,
		},
		{
			"json with '-' modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			10,
			"-",
			5,
			`10`,
		},
		{
			"json with '-' modifier (slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			`["a","b"]`,
			"-",
			"c",
			`["a","b"]`,
		},
		{
			"json with unknown modifier",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeJson},
			`"base"`,
			"?",
			`"new"`,
			`"base"`,
		},

		// single select
		{
			"single select with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"",
			"+",
			"b",
			`"b"`,
		},
		{
			"single select with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"a",
			"+",
			"b",
			`"b"`,
		},
		{
			"single select with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"",
			"-",
			"a",
			`""`,
		},
		{
			"single select with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"a",
			"-",
			"",
			`"a"`,
		},
		{
			"single select with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"a",
			"-",
			"b",
			`"a"`,
		},
		{
			"single select with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"a",
			"-",
			"a",
			`""`,
		},
		{
			"single select with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"a",
			"-",
			[]string{"b", "a", "c", "123"},
			`""`,
		},
		{
			"single select with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 1}},
			"",
			"?",
			"a",
			`""`,
		},

		// multi select
		{
			"multi select with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			nil,
			"+",
			"b",
			`["b"]`,
		},
		{
			"multi select with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			[]string{"a"},
			"+",
			[]string{"b", "c"},
			`["a","b","c"]`,
		},
		{
			"multi select with '+' modifier (nonempty base; already existing value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			[]string{"a", "b"},
			"+",
			"b",
			`["a","b"]`,
		},
		{
			"multi select with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			nil,
			"-",
			[]string{"a"},
			`[]`,
		},
		{
			"multi select with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			"a",
			"-",
			"",
			`["a"]`,
		},
		{
			"multi select with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			"a",
			"-",
			"b",
			`["a"]`,
		},
		{
			"multi select with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			[]string{"a", "b", "c", "d"},
			"-",
			"c",
			`["a","b","d"]`,
		},
		{
			"multi select with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			[]string{"a", "b", "c", "d"},
			"-",
			[]string{"b", "a", "123"},
			`["c","d"]`,
		},
		{
			"multi select with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeSelect, Options: &collectionmodels.SelectOptions{MaxSelect: 10}},
			[]string{"a", "b"},
			"?",
			"a",
			`["a","b"]`,
		},

		// single relation
		{
			"single relation with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"",
			"+",
			"b",
			`"b"`,
		},
		{
			"single relation with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"a",
			"+",
			"b",
			`"b"`,
		},
		{
			"single relation with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"",
			"-",
			"a",
			`""`,
		},
		{
			"single relation with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"a",
			"-",
			"",
			`"a"`,
		},
		{
			"single relation with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"a",
			"-",
			"b",
			`"a"`,
		},
		{
			"single relation with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"a",
			"-",
			"a",
			`""`,
		},
		{
			"single relation with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"a",
			"-",
			[]string{"b", "a", "c", "123"},
			`""`,
		},
		{
			"single relation with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation, Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)}},
			"",
			"?",
			"a",
			`""`,
		},

		// multi relation
		{
			"multi relation with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			nil,
			"+",
			"b",
			`["b"]`,
		},
		{
			"multi relation with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			[]string{"a"},
			"+",
			[]string{"b", "c"},
			`["a","b","c"]`,
		},
		{
			"multi relation with '+' modifier (nonempty base; already existing value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			[]string{"a", "b"},
			"+",
			"b",
			`["a","b"]`,
		},
		{
			"multi relation with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			nil,
			"-",
			[]string{"a"},
			`[]`,
		},
		{
			"multi relation with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			"a",
			"-",
			"",
			`["a"]`,
		},
		{
			"multi relation with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			"a",
			"-",
			"b",
			`["a"]`,
		},
		{
			"multi relation with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			[]string{"a", "b", "c", "d"},
			"-",
			"c",
			`["a","b","d"]`,
		},
		{
			"multi relation with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			[]string{"a", "b", "c", "d"},
			"-",
			[]string{"b", "a", "123"},
			`["c","d"]`,
		},
		{
			"multi relation with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeRelation},
			[]string{"a", "b"},
			"?",
			"a",
			`["a","b"]`,
		},

		// single file
		{
			"single file with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"",
			"+",
			"b",
			`""`,
		},
		{
			"single file with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"a",
			"+",
			"b",
			`"a"`,
		},
		{
			"single file with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"",
			"-",
			"a",
			`""`,
		},
		{
			"single file with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"a",
			"-",
			"",
			`"a"`,
		},
		{
			"single file with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"a",
			"-",
			"b",
			`"a"`,
		},
		{
			"single file with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"a",
			"-",
			"a",
			`""`,
		},
		{
			"single file with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"a",
			"-",
			[]string{"b", "a", "c", "123"},
			`""`,
		},
		{
			"single file with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 1}},
			"",
			"?",
			"a",
			`""`,
		},

		// multi file
		{
			"multi file with '+' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			nil,
			"+",
			"b",
			`[]`,
		},
		{
			"multi file with '+' modifier (nonempty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			[]string{"a"},
			"+",
			[]string{"b", "c"},
			`["a"]`,
		},
		{
			"multi file with '+' modifier (nonempty base; already existing value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			[]string{"a", "b"},
			"+",
			"b",
			`["a","b"]`,
		},
		{
			"multi file with '-' modifier (empty base)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			nil,
			"-",
			[]string{"a"},
			`[]`,
		},
		{
			"multi file with '-' modifier (nonempty base and empty modifier value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			"a",
			"-",
			"",
			`["a"]`,
		},
		{
			"multi file with '-' modifier (nonempty base and different value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			"a",
			"-",
			"b",
			`["a"]`,
		},
		{
			"multi file with '-' modifier (nonempty base and matching value)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			[]string{"a", "b", "c", "d"},
			"-",
			"c",
			`["a","b","d"]`,
		},
		{
			"multi file with '-' modifier (nonempty base and matching value in a slice)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			[]string{"a", "b", "c", "d"},
			"-",
			[]string{"b", "a", "123"},
			`["c","d"]`,
		},
		{
			"multi file with unknown modifier (nonempty)",
			collectionmodels.SchemaField{Type: collectionmodels.FieldTypeFile, Options: &collectionmodels.FileOptions{MaxSelect: 10}},
			[]string{"a", "b"},
			"?",
			"a",
			`["a","b"]`,
		},
	}

	for _, s := range scenarios {
		result := s.field.PrepareValueWithModifier(s.baseValue, s.modifier, s.modifierValue)

		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("[%s] %v", s.name, err)
		}

		if string(encoded) != s.expectJson {
			t.Fatalf("[%s], Expected %v, got %v", s.name, s.expectJson, string(encoded))
		}
	}
}

// -------------------------------------------------------------------

type fieldOptionsScenario struct {
	name           string
	options        collectionmodels.FieldOptions
	expectedErrors []string
}

func checkFieldOptionsScenarios(t *testing.T, scenarios []fieldOptionsScenario) {
	for i, s := range scenarios {
		result := s.options.Validate()

		prefix := fmt.Sprintf("%d", i)
		if s.name != "" {
			prefix = s.name
		}

		// parse errors
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("[%s] Failed to parse errors %v", prefix, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("[%s] Expected error keys %v, got %v", prefix, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("[%s] Missing expected error key %q in %v", prefix, k, errs)
			}
		}
	}
}

func TestTextOptionsValidate(t *testing.T) {
	minus := -1
	number0 := 0
	number1 := 10
	number2 := 20
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.TextOptions{},
			[]string{},
		},
		{
			"min - failure",
			collectionmodels.TextOptions{
				Min: &minus,
			},
			[]string{"min"},
		},
		{
			"min - success",
			collectionmodels.TextOptions{
				Min: &number0,
			},
			[]string{},
		},
		{
			"max - failure without min",
			collectionmodels.TextOptions{
				Max: &minus,
			},
			[]string{"max"},
		},
		{
			"max - failure with min",
			collectionmodels.TextOptions{
				Min: &number2,
				Max: &number1,
			},
			[]string{"max"},
		},
		{
			"max - success",
			collectionmodels.TextOptions{
				Min: &number1,
				Max: &number2,
			},
			[]string{},
		},
		{
			"pattern - failure",
			collectionmodels.TextOptions{Pattern: "(test"},
			[]string{"pattern"},
		},
		{
			"pattern - success",
			collectionmodels.TextOptions{Pattern: `^\#?\w+$`},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestNumberOptionsValidate(t *testing.T) {
	int1 := 10.0
	int2 := 20.0

	decimal1 := 10.5
	decimal2 := 20.5

	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.NumberOptions{},
			[]string{},
		},
		{
			"max - without min",
			collectionmodels.NumberOptions{
				Max: &int1,
			},
			[]string{},
		},
		{
			"max - failure with min",
			collectionmodels.NumberOptions{
				Min: &int2,
				Max: &int1,
			},
			[]string{"max"},
		},
		{
			"max - success with min",
			collectionmodels.NumberOptions{
				Min: &int1,
				Max: &int2,
			},
			[]string{},
		},
		{
			"NoDecimal range failure",
			collectionmodels.NumberOptions{
				Min:       &decimal1,
				Max:       &decimal2,
				NoDecimal: true,
			},
			[]string{"min", "max"},
		},
		{
			"NoDecimal range success",
			collectionmodels.NumberOptions{
				Min:       &int1,
				Max:       &int2,
				NoDecimal: true,
			},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestBoolOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.BoolOptions{},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestEmailOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.EmailOptions{},
			[]string{},
		},
		{
			"ExceptDomains failure",
			collectionmodels.EmailOptions{
				ExceptDomains: []string{"invalid"},
			},
			[]string{"exceptDomains"},
		},
		{
			"ExceptDomains success",
			collectionmodels.EmailOptions{
				ExceptDomains: []string{"example.com", "sub.example.com"},
			},
			[]string{},
		},
		{
			"OnlyDomains check",
			collectionmodels.EmailOptions{
				OnlyDomains: []string{"invalid"},
			},
			[]string{"onlyDomains"},
		},
		{
			"OnlyDomains success",
			collectionmodels.EmailOptions{
				OnlyDomains: []string{"example.com", "sub.example.com"},
			},
			[]string{},
		},
		{
			"OnlyDomains + ExceptDomains at the same time",
			collectionmodels.EmailOptions{
				ExceptDomains: []string{"test1.com"},
				OnlyDomains:   []string{"test2.com"},
			},
			[]string{"exceptDomains", "onlyDomains"},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestUrlOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.UrlOptions{},
			[]string{},
		},
		{
			"ExceptDomains failure",
			collectionmodels.UrlOptions{
				ExceptDomains: []string{"invalid"},
			},
			[]string{"exceptDomains"},
		},
		{
			"ExceptDomains success",
			collectionmodels.UrlOptions{
				ExceptDomains: []string{"example.com", "sub.example.com"},
			},
			[]string{},
		},
		{
			"OnlyDomains check",
			collectionmodels.UrlOptions{
				OnlyDomains: []string{"invalid"},
			},
			[]string{"onlyDomains"},
		},
		{
			"OnlyDomains success",
			collectionmodels.UrlOptions{
				OnlyDomains: []string{"example.com", "sub.example.com"},
			},
			[]string{},
		},
		{
			"OnlyDomains + ExceptDomains at the same time",
			collectionmodels.UrlOptions{
				ExceptDomains: []string{"test1.com"},
				OnlyDomains:   []string{"test2.com"},
			},
			[]string{"exceptDomains", "onlyDomains"},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestEditorOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.EditorOptions{},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestDateOptionsValidate(t *testing.T) {
	date1 := types.NowDateTime()
	date2, _ := types.ParseDateTime(date1.Time().AddDate(1, 0, 0))

	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.DateOptions{},
			[]string{},
		},
		{
			"min only",
			collectionmodels.DateOptions{
				Min: date1,
			},
			[]string{},
		},
		{
			"max only",
			collectionmodels.DateOptions{
				Min: date1,
			},
			[]string{},
		},
		{
			"zero min + max",
			collectionmodels.DateOptions{
				Min: types.DateTime{},
				Max: date1,
			},
			[]string{},
		},
		{
			"min + zero max",
			collectionmodels.DateOptions{
				Min: date1,
				Max: types.DateTime{},
			},
			[]string{},
		},
		{
			"min > max",
			collectionmodels.DateOptions{
				Min: date2,
				Max: date1,
			},
			[]string{"max"},
		},
		{
			"min == max",
			collectionmodels.DateOptions{
				Min: date1,
				Max: date1,
			},
			[]string{"max"},
		},
		{
			"min < max",
			collectionmodels.DateOptions{
				Min: date1,
				Max: date2,
			},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestSelectOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.SelectOptions{},
			[]string{"values", "maxSelect"},
		},
		{
			"MaxSelect <= 0",
			collectionmodels.SelectOptions{
				Values:    []string{"test1", "test2"},
				MaxSelect: 0,
			},
			[]string{"maxSelect"},
		},
		{
			"MaxSelect > Values",
			collectionmodels.SelectOptions{
				Values:    []string{"test1", "test2"},
				MaxSelect: 3,
			},
			[]string{"maxSelect"},
		},
		{
			"MaxSelect <= Values",
			collectionmodels.SelectOptions{
				Values:    []string{"test1", "test2"},
				MaxSelect: 2,
			},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestSelectOptionsIsMultiple(t *testing.T) {
	scenarios := []struct {
		maxSelect int
		expect    bool
	}{
		{-1, false},
		{0, false},
		{1, false},
		{2, true},
	}

	for i, s := range scenarios {
		opt := collectionmodels.SelectOptions{
			MaxSelect: s.maxSelect,
		}

		if v := opt.IsMultiple(); v != s.expect {
			t.Errorf("[%d] Expected %v, got %v", i, s.expect, v)
		}
	}
}

func TestJsonOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.JsonOptions{},
			[]string{"maxSize"},
		},
		{
			"MaxSize < 0",
			collectionmodels.JsonOptions{MaxSize: -1},
			[]string{"maxSize"},
		},
		{
			"MaxSize > 0",
			collectionmodels.JsonOptions{MaxSize: 1},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestFileOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.FileOptions{},
			[]string{"maxSelect", "maxSize"},
		},
		{
			"MaxSelect <= 0 && maxSize <= 0",
			collectionmodels.FileOptions{
				MaxSize:   0,
				MaxSelect: 0,
			},
			[]string{"maxSelect", "maxSize"},
		},
		{
			"MaxSelect > 0 && maxSize > 0",
			collectionmodels.FileOptions{
				MaxSize:   2,
				MaxSelect: 1,
			},
			[]string{},
		},
		{
			"invalid thumbs format",
			collectionmodels.FileOptions{
				MaxSize:   1,
				MaxSelect: 2,
				Thumbs:    []string{"100", "200x100"},
			},
			[]string{"thumbs"},
		},
		{
			"invalid thumbs format - zero width and height",
			collectionmodels.FileOptions{
				MaxSize:   1,
				MaxSelect: 2,
				Thumbs:    []string{"0x0", "0x0t", "0x0b", "0x0f"},
			},
			[]string{"thumbs"},
		},
		{
			"valid thumbs format",
			collectionmodels.FileOptions{
				MaxSize:   1,
				MaxSelect: 2,
				Thumbs: []string{
					"100x100", "200x100", "0x100", "100x0",
					"10x10t", "10x10b", "10x10f",
				},
			},
			[]string{},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestFileOptionsIsMultiple(t *testing.T) {
	scenarios := []struct {
		maxSelect int
		expect    bool
	}{
		{-1, false},
		{0, false},
		{1, false},
		{2, true},
	}

	for i, s := range scenarios {
		opt := collectionmodels.FileOptions{
			MaxSelect: s.maxSelect,
		}

		if v := opt.IsMultiple(); v != s.expect {
			t.Errorf("[%d] Expected %v, got %v", i, s.expect, v)
		}
	}
}

func TestRelationOptionsValidate(t *testing.T) {
	scenarios := []fieldOptionsScenario{
		{
			"empty",
			collectionmodels.RelationOptions{},
			[]string{"collectionId"},
		},
		{
			"empty CollectionId",
			collectionmodels.RelationOptions{
				CollectionId: "",
				MaxSelect:    types.Pointer(1),
			},
			[]string{"collectionId"},
		},
		{
			"MinSelect < 0",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    types.Pointer(-1),
			},
			[]string{"minSelect"},
		},
		{
			"MinSelect >= 0",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    types.Pointer(0),
			},
			[]string{},
		},
		{
			"MaxSelect <= 0",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MaxSelect:    types.Pointer(0),
			},
			[]string{"maxSelect"},
		},
		{
			"MaxSelect > 0 && nonempty CollectionId",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MaxSelect:    types.Pointer(1),
			},
			[]string{},
		},
		{
			"MinSelect < MaxSelect",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    nil,
				MaxSelect:    types.Pointer(1),
			},
			[]string{},
		},
		{
			"MinSelect = MaxSelect (non-zero)",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    types.Pointer(1),
				MaxSelect:    types.Pointer(1),
			},
			[]string{},
		},
		{
			"MinSelect = MaxSelect (both zero)",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    types.Pointer(0),
				MaxSelect:    types.Pointer(0),
			},
			[]string{"maxSelect"},
		},
		{
			"MinSelect > MaxSelect",
			collectionmodels.RelationOptions{
				CollectionId: "abc",
				MinSelect:    types.Pointer(2),
				MaxSelect:    types.Pointer(1),
			},
			[]string{"maxSelect"},
		},
	}

	checkFieldOptionsScenarios(t, scenarios)
}

func TestRelationOptionsIsMultiple(t *testing.T) {
	scenarios := []struct {
		maxSelect *int
		expect    bool
	}{
		{nil, true},
		{types.Pointer(-1), false},
		{types.Pointer(0), false},
		{types.Pointer(1), false},
		{types.Pointer(2), true},
	}

	for i, s := range scenarios {
		opt := collectionmodels.RelationOptions{
			MaxSelect: s.maxSelect,
		}

		if v := opt.IsMultiple(); v != s.expect {
			t.Errorf("[%d] Expected %v, got %v", i, s.expect, v)
		}
	}
}
