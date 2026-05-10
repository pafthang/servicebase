package models_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pocketbase/dbx"

	"github.com/pafthang/servicebase/tools/list"
	"github.com/pafthang/servicebase/tools/types"
)

func TestNewRecord(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Name: "test_collection",
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "test",
				Type: collectionmodels.FieldTypeText,
			},
		),
	}

	m := recordmodels.NewRecord(collection)

	if m.Collection().Name != collection.Name {
		t.Fatalf("Expected collection with name %q, got %q", collection.Id, m.Collection().Id)
	}

	if len(m.SchemaData()) != 0 {
		t.Fatalf("Expected empty schema data, got %v", m.SchemaData())
	}
}

func TestNewRecordFromNullStringMap(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Name: "test",
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeBool,
			},
			&collectionmodels.SchemaField{
				Name: "field4",
				Type: collectionmodels.FieldTypeNumber,
			},
			&collectionmodels.SchemaField{
				Name: "field5",
				Type: collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{
					Values:    []string{"test1", "test2"},
					MaxSelect: 1,
				},
			},
			&collectionmodels.SchemaField{
				Name: "field6",
				Type: collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{
					MaxSelect: 2,
					MaxSize:   1,
				},
			},
		),
	}

	data := dbx.NullStringMap{
		"id": sql.NullString{
			String: "test_id",
			Valid:  true,
		},
		"created": sql.NullString{
			String: "2022-01-01 10:00:00.123Z",
			Valid:  true,
		},
		"updated": sql.NullString{
			String: "2022-01-01 10:00:00.456Z",
			Valid:  true,
		},
		// users collection specific fields
		"username": sql.NullString{
			String: "test_username",
			Valid:  true,
		},
		"email": sql.NullString{
			String: "test_email",
			Valid:  true,
		},
		"emailVisibility": sql.NullString{
			String: "true",
			Valid:  true,
		},
		"verified": sql.NullString{
			String: "",
			Valid:  false,
		},
		"tokenKey": sql.NullString{
			String: "test_tokenKey",
			Valid:  true,
		},
		"passwordHash": sql.NullString{
			String: "test_passwordHash",
			Valid:  true,
		},
		"lastResetSentAt": sql.NullString{
			String: "2022-01-02 10:00:00.123Z",
			Valid:  true,
		},
		"lastVerificationSentAt": sql.NullString{
			String: "2022-02-03 10:00:00.456Z",
			Valid:  true,
		},
		// custom schema fields
		"field1": sql.NullString{
			String: "test",
			Valid:  true,
		},
		"field2": sql.NullString{
			String: "test",
			Valid:  false, // test invalid db serialization
		},
		"field3": sql.NullString{
			String: "true",
			Valid:  true,
		},
		"field4": sql.NullString{
			String: "123.123",
			Valid:  true,
		},
		"field5": sql.NullString{
			String: `["test1","test2"]`, // will select only the last elem
			Valid:  true,
		},
		"field6": sql.NullString{
			String: "test", // will be converted to slice
			Valid:  true,
		},
		"unknown": sql.NullString{
			String: "test",
			Valid:  true,
		},
	}

	scenarios := []struct {
		collectionType string
		expectedJson   string
	}{
		{
			collectionmodels.CollectionTypeBase,
			`{"collectionId":"","collectionName":"test","created":"2022-01-01 10:00:00.123Z","field1":"test","field2":"","field3":true,"field4":123.123,"field5":"test2","field6":["test"],"id":"test_id","updated":"2022-01-01 10:00:00.456Z"}`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			`{"collectionId":"","collectionName":"test","created":"2022-01-01 10:00:00.123Z","email":"test_email","emailVisibility":true,"field1":"test","field2":"","field3":true,"field4":123.123,"field5":"test2","field6":["test"],"id":"test_id","updated":"2022-01-01 10:00:00.456Z","username":"test_username","verified":false}`,
		},
	}

	for i, s := range scenarios {
		collection.Type = s.collectionType
		m := recordmodels.NewRecordFromNullStringMap(collection, data)
		m.IgnoreEmailVisibility(true)

		encoded, err := m.MarshalJSON()
		if err != nil {
			t.Errorf("(%d) Unexpected error: %v", i, err)
			continue
		}

		if string(encoded) != s.expectedJson {
			t.Errorf("(%d) Expected \n%v \ngot \n%v", i, s.expectedJson, string(encoded))
		}

		// additional data checks
		if collection.IsUsers() {
			if v := m.GetString(collectionmodels.FieldNamePasswordHash); v != "test_passwordHash" {
				t.Errorf("(%d) Expected %q, got %q", i, "test_passwordHash", v)
			}
			if v := m.GetString(collectionmodels.FieldNameTokenKey); v != "test_tokenKey" {
				t.Errorf("(%d) Expected %q, got %q", i, "test_tokenKey", v)
			}
			if v := m.GetString(collectionmodels.FieldNameLastResetSentAt); v != "2022-01-02 10:00:00.123Z" {
				t.Errorf("(%d) Expected %q, got %q", i, "2022-01-02 10:00:00.123Z", v)
			}
			if v := m.GetString(collectionmodels.FieldNameLastVerificationSentAt); v != "2022-02-03 10:00:00.456Z" {
				t.Errorf("(%d) Expected %q, got %q", i, "2022-01-02 10:00:00.123Z", v)
			}
		}
	}
}

func TestNewRecordsFromNullStringMaps(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Name: "test",
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeUrl,
			},
		),
	}

	data := []dbx.NullStringMap{
		{
			"id": sql.NullString{
				String: "test_id1",
				Valid:  true,
			},
			"created": sql.NullString{
				String: "2022-01-01 10:00:00.123Z",
				Valid:  true,
			},
			"updated": sql.NullString{
				String: "2022-01-01 10:00:00.456Z",
				Valid:  true,
			},
			// partial auth fields
			"email": sql.NullString{
				String: "test_email",
				Valid:  true,
			},
			"tokenKey": sql.NullString{
				String: "test_tokenKey",
				Valid:  true,
			},
			"emailVisibility": sql.NullString{
				String: "true",
				Valid:  true,
			},
			// custom schema fields
			"field1": sql.NullString{
				String: "test",
				Valid:  true,
			},
			"field2": sql.NullString{
				String: "123.123",
				Valid:  true,
			},
			"field3": sql.NullString{
				String: "test",
				Valid:  false, // should force resolving to empty string
			},
			"unknown": sql.NullString{
				String: "test",
				Valid:  true,
			},
		},
		{
			"field3": sql.NullString{
				String: "test",
				Valid:  true,
			},
			"email": sql.NullString{
				String: "test_email",
				Valid:  true,
			},
			"emailVisibility": sql.NullString{
				String: "false",
				Valid:  true,
			},
		},
	}

	scenarios := []struct {
		collectionType string
		expectedJson   string
	}{
		{
			collectionmodels.CollectionTypeBase,
			`[{"collectionId":"","collectionName":"test","created":"2022-01-01 10:00:00.123Z","field1":"test","field2":123.123,"field3":"","id":"test_id1","updated":"2022-01-01 10:00:00.456Z"},{"collectionId":"","collectionName":"test","created":"","field1":"","field2":0,"field3":"test","id":"","updated":""}]`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			`[{"collectionId":"","collectionName":"test","created":"2022-01-01 10:00:00.123Z","email":"test_email","emailVisibility":true,"field1":"test","field2":123.123,"field3":"","id":"test_id1","updated":"2022-01-01 10:00:00.456Z","username":"","verified":false},{"collectionId":"","collectionName":"test","created":"","emailVisibility":false,"field1":"","field2":0,"field3":"test","id":"","updated":"","username":"","verified":false}]`,
		},
	}

	for i, s := range scenarios {
		collection.Type = s.collectionType
		result := recordmodels.NewRecordsFromNullStringMaps(collection, data)

		encoded, err := json.Marshal(result)
		if err != nil {
			t.Errorf("(%d) Unexpected error: %v", i, err)
			continue
		}

		if string(encoded) != s.expectedJson {
			t.Errorf("(%d) Expected \n%v \ngot \n%v", i, s.expectedJson, string(encoded))
		}
	}
}

func TestRecordTableName(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}
	collection.Name = "test"
	collection.RefreshId()

	m := recordmodels.NewRecord(collection)

	if m.TableName() != collection.Name {
		t.Fatalf("Expected table %q, got %q", collection.Name, m.TableName())
	}
}

func TestRecordCollection(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}
	collection.RefreshId()

	m := recordmodels.NewRecord(collection)

	if m.Collection().Id != collection.Id {
		t.Fatalf("Expected collection with id %v, got %v", collection.Id, m.Collection().Id)
	}
}

func TestRecordOriginalCopy(t *testing.T) {
	t.Parallel()

	m := recordmodels.NewRecord(&collectionmodels.Collection{})
	m.Load(map[string]any{"f": "123"})

	// change the field
	m.Set("f", "456")

	if v := m.GetString("f"); v != "456" {
		t.Fatalf("Expected f to be %q, got %q", "456", v)
	}

	if v := m.OriginalCopy().GetString("f"); v != "123" {
		t.Fatalf("Expected the initial/original f to be %q, got %q", "123", v)
	}

	// loading new data shouldn't affect the original state
	m.Load(map[string]any{"f": "789"})

	if v := m.GetString("f"); v != "789" {
		t.Fatalf("Expected f to be %q, got %q", "789", v)
	}

	if v := m.OriginalCopy().GetString("f"); v != "123" {
		t.Fatalf("Expected the initial/original f still to be %q, got %q", "123", v)
	}
}

func TestRecordCleanCopy(t *testing.T) {
	t.Parallel()

	m := recordmodels.NewRecord(&collectionmodels.Collection{
		Name: "cname",
		Type: collectionmodels.CollectionTypeUsers,
	})
	m.Load(map[string]any{
		"id":       "id1",
		"created":  "2023-01-01 00:00:00.000Z",
		"updated":  "2023-01-02 00:00:00.000Z",
		"username": "test",
		"verified": true,
		"email":    "test@example.com",
		"unknown":  "456",
	})

	// make a change to ensure that the latest data is targeted
	m.Set("id", "id2")

	// allow the special flags and options to check whether they will be ignored
	m.SetExpand(map[string]any{"test": 123})
	m.IgnoreEmailVisibility(true)
	m.WithUnknownData(true)

	copy := m.CleanCopy()
	copyExport, _ := copy.MarshalJSON()

	expectedExport := []byte(`{"collectionId":"","collectionName":"cname","created":"2023-01-01 00:00:00.000Z","emailVisibility":false,"id":"id2","updated":"2023-01-02 00:00:00.000Z","username":"test","verified":true}`)
	if !bytes.Equal(copyExport, expectedExport) {
		t.Fatalf("Expected clean export \n%s, \ngot \n%s", expectedExport, copyExport)
	}
}

func TestRecordSetAndGetExpand(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}
	m := recordmodels.NewRecord(collection)

	data := map[string]any{"test": 123}

	m.SetExpand(data)

	// change the original data to check if it was shallow copied
	data["test"] = 456

	expand := m.Expand()
	if v, ok := expand["test"]; !ok || v != 123 {
		t.Fatalf("Expected expand.test to be %v, got %v", 123, v)
	}
}

func TestRecordMergeExpand(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}
	m := recordmodels.NewRecord(collection)
	m.Id = "m"

	// a
	a := recordmodels.NewRecord(collection)
	a.Id = "a"
	a1 := recordmodels.NewRecord(collection)
	a1.Id = "a1"
	a2 := recordmodels.NewRecord(collection)
	a2.Id = "a2"
	a3 := recordmodels.NewRecord(collection)
	a3.Id = "a3"
	a31 := recordmodels.NewRecord(collection)
	a31.Id = "a31"
	a32 := recordmodels.NewRecord(collection)
	a32.Id = "a32"
	a.SetExpand(map[string]any{
		"a1":  a1,
		"a23": []*recordmodels.Record{a2, a3},
	})
	a3.SetExpand(map[string]any{
		"a31": a31,
		"a32": []*recordmodels.Record{a32},
	})

	// b
	b := recordmodels.NewRecord(collection)
	b.Id = "b"
	b1 := recordmodels.NewRecord(collection)
	b1.Id = "b1"
	b.SetExpand(map[string]any{
		"b1": b1,
	})

	// c
	c := recordmodels.NewRecord(collection)
	c.Id = "c"

	// load initial expand
	m.SetExpand(map[string]any{
		"a": a,
		"b": b,
		"c": []*recordmodels.Record{c},
	})

	// a (new)
	aNew := recordmodels.NewRecord(collection)
	aNew.Id = a.Id
	a3New := recordmodels.NewRecord(collection)
	a3New.Id = a3.Id
	a32New := recordmodels.NewRecord(collection)
	a32New.Id = "a32New"
	a33New := recordmodels.NewRecord(collection)
	a33New.Id = "a33New"
	a3New.SetExpand(map[string]any{
		"a32":    []*recordmodels.Record{a32New},
		"a33New": a33New,
	})
	aNew.SetExpand(map[string]any{
		"a23": []*recordmodels.Record{a2, a3New},
	})

	// b (new)
	bNew := recordmodels.NewRecord(collection)
	bNew.Id = "bNew"
	dNew := recordmodels.NewRecord(collection)
	dNew.Id = "dNew"

	// merge expands
	m.MergeExpand(map[string]any{
		"a":    aNew,
		"b":    []*recordmodels.Record{bNew},
		"dNew": dNew,
	})

	result := m.Expand()

	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	rawStr := string(raw)

	expected := `{"a":{"collectionId":"","collectionName":"","created":"","expand":{"a1":{"collectionId":"","collectionName":"","created":"","id":"a1","updated":""},"a23":[{"collectionId":"","collectionName":"","created":"","id":"a2","updated":""},{"collectionId":"","collectionName":"","created":"","expand":{"a31":{"collectionId":"","collectionName":"","created":"","id":"a31","updated":""},"a32":[{"collectionId":"","collectionName":"","created":"","id":"a32","updated":""},{"collectionId":"","collectionName":"","created":"","id":"a32New","updated":""}],"a33New":{"collectionId":"","collectionName":"","created":"","id":"a33New","updated":""}},"id":"a3","updated":""}]},"id":"a","updated":""},"b":[{"collectionId":"","collectionName":"","created":"","expand":{"b1":{"collectionId":"","collectionName":"","created":"","id":"b1","updated":""}},"id":"b","updated":""},{"collectionId":"","collectionName":"","created":"","id":"bNew","updated":""}],"c":[{"collectionId":"","collectionName":"","created":"","id":"c","updated":""}],"dNew":{"collectionId":"","collectionName":"","created":"","id":"dNew","updated":""}}`

	if expected != rawStr {
		t.Fatalf("Expected \n%v, \ngot \n%v", expected, rawStr)
	}
}

func TestRecordMergeExpandNilCheck(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}

	scenarios := []struct {
		name     string
		expand   map[string]any
		expected string
	}{
		{
			"nil expand",
			nil,
			`{"collectionId":"","collectionName":"","created":"","id":"","updated":""}`,
		},
		{
			"empty expand",
			map[string]any{},
			`{"collectionId":"","collectionName":"","created":"","id":"","updated":""}`,
		},
		{
			"non-empty expand",
			map[string]any{"test": recordmodels.NewRecord(collection)},
			`{"collectionId":"","collectionName":"","created":"","expand":{"test":{"collectionId":"","collectionName":"","created":"","id":"","updated":""}},"id":"","updated":""}`,
		},
	}

	for _, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.MergeExpand(s.expand)

		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		rawStr := string(raw)

		if rawStr != s.expected {
			t.Fatalf("[%s] Expected \n%v, \ngot \n%v", s.name, s.expected, rawStr)
		}
	}
}

func TestRecordExpandedRel(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}

	main := recordmodels.NewRecord(collection)

	single := recordmodels.NewRecord(collection)
	single.Id = "single"

	multiple1 := recordmodels.NewRecord(collection)
	multiple1.Id = "multiple1"

	multiple2 := recordmodels.NewRecord(collection)
	multiple2.Id = "multiple2"

	main.SetExpand(map[string]any{
		"single":   single,
		"multiple": []*recordmodels.Record{multiple1, multiple2},
	})

	if v := main.ExpandedOne("missing"); v != nil {
		t.Fatalf("Expected nil, got %v", v)
	}

	if v := main.ExpandedOne("single"); v == nil || v.Id != "single" {
		t.Fatalf("Expected record with id %q, got %v", "single", v)
	}

	if v := main.ExpandedOne("multiple"); v == nil || v.Id != "multiple1" {
		t.Fatalf("Expected record with id %q, got %v", "multiple1", v)
	}
}

func TestRecordExpandedAll(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}

	main := recordmodels.NewRecord(collection)

	single := recordmodels.NewRecord(collection)
	single.Id = "single"

	multiple1 := recordmodels.NewRecord(collection)
	multiple1.Id = "multiple1"

	multiple2 := recordmodels.NewRecord(collection)
	multiple2.Id = "multiple2"

	main.SetExpand(map[string]any{
		"single":   single,
		"multiple": []*recordmodels.Record{multiple1, multiple2},
	})

	if v := main.ExpandedAll("missing"); v != nil {
		t.Fatalf("Expected nil, got %v", v)
	}

	if v := main.ExpandedAll("single"); len(v) != 1 || v[0].Id != "single" {
		t.Fatalf("Expected [single] slice, got %v", v)
	}

	if v := main.ExpandedAll("multiple"); len(v) != 2 || v[0].Id != "multiple1" || v[1].Id != "multiple2" {
		t.Fatalf("Expected [multiple1, multiple2] slice, got %v", v)
	}
}

func TestRecordSchemaData(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Type: collectionmodels.CollectionTypeUsers,
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
		),
	}

	m := recordmodels.NewRecord(collection)
	m.Set("email", "test@example.com")
	m.Set("field1", 123)
	m.Set("field2", 456)
	m.Set("unknown", 789)

	encoded, err := json.Marshal(m.SchemaData())
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"field1":"123","field2":456}`

	if v := string(encoded); v != expected {
		t.Fatalf("Expected \n%v \ngot \n%v", v, expected)
	}
}

func TestRecordUnknownData(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
		),
	}

	data := map[string]any{
		"id":                     "test_id",
		"created":                "2022-01-01 00:00:00.000",
		"updated":                "2022-01-01 00:00:00.000",
		"collectionId":           "test_collectionId",
		"collectionName":         "test_collectionName",
		"expand":                 "test_expand",
		"field1":                 "test_field1",
		"field2":                 "test_field1",
		"unknown1":               "test_unknown1",
		"unknown2":               "test_unknown2",
		"passwordHash":           "test_passwordHash",
		"username":               "test_username",
		"emailVisibility":        true,
		"email":                  "test_email",
		"verified":               true,
		"tokenKey":               "test_tokenKey",
		"lastResetSentAt":        "2022-01-01 00:00:00.000",
		"lastVerificationSentAt": "2022-01-01 00:00:00.000",
	}

	scenarios := []struct {
		collectionType string
		expectedKeys   []string
	}{
		{
			collectionmodels.CollectionTypeBase,
			[]string{
				"unknown1",
				"unknown2",
				"passwordHash",
				"username",
				"emailVisibility",
				"email",
				"verified",
				"tokenKey",
				"lastResetSentAt",
				"lastVerificationSentAt",
			},
		},
		{
			collectionmodels.CollectionTypeUsers,
			[]string{"unknown1", "unknown2"},
		},
	}

	for i, s := range scenarios {
		collection.Type = s.collectionType
		m := recordmodels.NewRecord(collection)
		m.Load(data)

		result := m.UnknownData()

		if len(result) != len(s.expectedKeys) {
			t.Errorf("(%d) Expected data \n%v \ngot \n%v", i, s.expectedKeys, result)
			continue
		}

		for _, key := range s.expectedKeys {
			if _, ok := result[key]; !ok {
				t.Errorf("(%d) Missing expected key %q in \n%v", i, key, result)
			}
		}
	}
}

func TestRecordSetAndGet(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
			// fields that are not explicitly set to check
			// the default retrieval value (single and multiple)
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeBool,
			},
			&collectionmodels.SchemaField{
				Name:    "field4",
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 2},
			},
			&collectionmodels.SchemaField{
				Name:    "field5",
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
		),
	}

	m := recordmodels.NewRecord(collection)
	m.Set("id", "test_id")
	m.Set("created", "2022-09-15 00:00:00.123Z")
	m.Set("updated", "invalid")
	m.Set("field1", 123)                         // should be casted to string
	m.Set("field2", "invlaid")                   // should be casted to zero-number
	m.Set("unknown", 456)                        // undefined fields are allowed but not exported by default
	m.Set("expand", map[string]any{"test": 123}) // should store the value in m.expand

	if v := m.Get("id"); v != "test_id" {
		t.Fatalf("Expected id %q, got %q", "test_id", v)
	}

	if v := m.GetString("created"); v != "2022-09-15 00:00:00.123Z" {
		t.Fatalf("Expected created %q, got %q", "2022-09-15 00:00:00.123Z", v)
	}

	if v := m.GetString("updated"); v != "" {
		t.Fatalf("Expected updated to be empty, got %q", v)
	}

	if v, ok := m.Get("field1").(string); !ok || v != "123" {
		t.Fatalf("Expected field1 %#v, got %#v", "123", m.Get("field1"))
	}

	if v, ok := m.Get("field2").(float64); !ok || v != 0.0 {
		t.Fatalf("Expected field2 %#v, got %#v", 0.0, m.Get("field2"))
	}

	if v, ok := m.Get("field3").(bool); !ok || v != false {
		t.Fatalf("Expected field3 %#v, got %#v", false, m.Get("field3"))
	}

	if v, ok := m.Get("field4").([]string); !ok || len(v) != 0 {
		t.Fatalf("Expected field4 %#v, got %#v", "[]", m.Get("field4"))
	}

	if v, ok := m.Get("field5").(string); !ok || len(v) != 0 {
		t.Fatalf("Expected field5 %#v, got %#v", "", m.Get("field5"))
	}

	if v := m.Get("unknown"); v != 456 {
		t.Fatalf("Expected unknown %v, got %v", 456, v)
	}

	if m.Expand()["test"] != 123 {
		t.Fatalf("Expected expand to be %v, got %v", map[string]any{"test": 123}, m.Expand())
	}
}

func TestRecordGetBool(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		value    any
		expected bool
	}{
		{nil, false},
		{"", false},
		{0, false},
		{1, true},
		{[]string{"true"}, false},
		{time.Now(), false},
		{"test", false},
		{"false", false},
		{"true", true},
		{false, false},
		{true, true},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetBool("test")
		if result != s.expected {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetString(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		value    any
		expected string
	}{
		{nil, ""},
		{"", ""},
		{0, "0"},
		{1.4, "1.4"},
		{[]string{"true"}, ""},
		{map[string]int{"test": 1}, ""},
		{[]byte("abc"), "abc"},
		{"test", "test"},
		{false, "false"},
		{true, "true"},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetString("test")
		if result != s.expected {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetInt(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		value    any
		expected int
	}{
		{nil, 0},
		{"", 0},
		{[]string{"true"}, 0},
		{map[string]int{"test": 1}, 0},
		{time.Now(), 0},
		{"test", 0},
		{123, 123},
		{2.4, 2},
		{"123", 123},
		{"123.5", 0},
		{false, 0},
		{true, 1},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetInt("test")
		if result != s.expected {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetFloat(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		value    any
		expected float64
	}{
		{nil, 0},
		{"", 0},
		{[]string{"true"}, 0},
		{map[string]int{"test": 1}, 0},
		{time.Now(), 0},
		{"test", 0},
		{123, 123},
		{2.4, 2.4},
		{"123", 123},
		{"123.5", 123.5},
		{false, 0},
		{true, 1},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetFloat("test")
		if result != s.expected {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetTime(t *testing.T) {
	t.Parallel()

	nowTime := time.Now()
	testTime, _ := time.Parse(types.DefaultDateLayout, "2022-01-01 08:00:40.000Z")

	scenarios := []struct {
		value    any
		expected time.Time
	}{
		{nil, time.Time{}},
		{"", time.Time{}},
		{false, time.Time{}},
		{true, time.Time{}},
		{"test", time.Time{}},
		{[]string{"true"}, time.Time{}},
		{map[string]int{"test": 1}, time.Time{}},
		{1641024040, testTime},
		{"2022-01-01 08:00:40.000", testTime},
		{nowTime, nowTime},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetTime("test")
		if !result.Equal(s.expected) {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetDateTime(t *testing.T) {
	t.Parallel()

	nowTime := time.Now()
	testTime, _ := time.Parse(types.DefaultDateLayout, "2022-01-01 08:00:40.000Z")

	scenarios := []struct {
		value    any
		expected time.Time
	}{
		{nil, time.Time{}},
		{"", time.Time{}},
		{false, time.Time{}},
		{true, time.Time{}},
		{"test", time.Time{}},
		{[]string{"true"}, time.Time{}},
		{map[string]int{"test": 1}, time.Time{}},
		{1641024040, testTime},
		{"2022-01-01 08:00:40.000", testTime},
		{nowTime, nowTime},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetDateTime("test")
		if !result.Time().Equal(s.expected) {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, result)
		}
	}
}

func TestRecordGetStringSlice(t *testing.T) {
	t.Parallel()

	nowTime := time.Now()

	scenarios := []struct {
		value    any
		expected []string
	}{
		{nil, []string{}},
		{"", []string{}},
		{false, []string{"false"}},
		{true, []string{"true"}},
		{nowTime, []string{}},
		{123, []string{"123"}},
		{"test", []string{"test"}},
		{map[string]int{"test": 1}, []string{}},
		{`["test1", "test2"]`, []string{"test1", "test2"}},
		{[]int{123, 123, 456}, []string{"123", "456"}},
		{[]string{"test", "test", "123"}, []string{"test", "123"}},
	}

	collection := &collectionmodels.Collection{}

	for i, s := range scenarios {
		m := recordmodels.NewRecord(collection)
		m.Set("test", s.value)

		result := m.GetStringSlice("test")

		if len(result) != len(s.expected) {
			t.Errorf("(%d) Expected %d elements, got %d: %v", i, len(s.expected), len(result), result)
			continue
		}

		for _, v := range result {
			if !list.ExistInSlice(v, s.expected) {
				t.Errorf("(%d) Cannot find %v in %v", i, v, s.expected)
			}
		}
	}
}

func TestRecordUnmarshalJSONField(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(&collectionmodels.SchemaField{
			Name: "field",
			Type: collectionmodels.FieldTypeJson,
		}),
	}
	m := recordmodels.NewRecord(collection)

	var testPointer *string
	var testStr string
	var testInt int
	var testBool bool
	var testSlice []int
	var testMap map[string]any

	scenarios := []struct {
		value        any
		destination  any
		expectError  bool
		expectedJson string
	}{
		{nil, testStr, true, `""`},
		{"", testStr, false, `""`},
		{1, testInt, false, `1`},
		{true, testBool, false, `true`},
		{[]int{1, 2, 3}, testSlice, false, `[1,2,3]`},
		{map[string]any{"test": 123}, testMap, false, `{"test":123}`},
		// json encoded values
		{`null`, testPointer, false, `null`},
		{`true`, testBool, false, `true`},
		{`456`, testInt, false, `456`},
		{`"test"`, testStr, false, `"test"`},
		{`[4,5,6]`, testSlice, false, `[4,5,6]`},
		{`{"test":456}`, testMap, false, `{"test":456}`},
	}

	for i, s := range scenarios {
		m.Set("field", s.value)

		err := m.UnmarshalJSONField("field", &s.destination)
		hasErr := err != nil

		if hasErr != s.expectError {
			t.Errorf("(%d) Expected hasErr %v, got %v", i, s.expectError, hasErr)
			continue
		}

		raw, _ := json.Marshal(s.destination)
		if v := string(raw); v != s.expectedJson {
			t.Errorf("(%d) Expected %q, got %q", i, s.expectedJson, v)
		}
	}
}

func TestRecordBaseFilesPath(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{}
	collection.RefreshId()
	collection.Name = "test"

	m := recordmodels.NewRecord(collection)
	m.RefreshId()

	expected := collection.BaseFilesPath() + "/" + m.Id
	result := m.BaseFilesPath()

	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestRecordFindFileFieldByFile(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{
					MaxSelect: 1,
					MaxSize:   1,
				},
			},
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{
					MaxSelect: 2,
					MaxSize:   1,
				},
			},
		),
	}

	m := recordmodels.NewRecord(collection)
	m.Set("field1", "test")
	m.Set("field2", "test.png")
	m.Set("field3", []string{"test1.png", "test2.png"})

	scenarios := []struct {
		filename    string
		expectField string
	}{
		{"", ""},
		{"test", ""},
		{"test2", ""},
		{"test.png", "field2"},
		{"test2.png", "field3"},
	}

	for i, s := range scenarios {
		result := m.FindFileFieldByFile(s.filename)

		var fieldName string
		if result != nil {
			fieldName = result.Name
		}

		if s.expectField != fieldName {
			t.Errorf("(%d) Expected field %v, got %v", i, s.expectField, result)
			continue
		}
	}
}

func TestRecordLoadAndData(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
		),
	}

	data := map[string]any{
		"id":      "test_id",
		"created": "2022-01-01 10:00:00.123Z",
		"updated": "2022-01-01 10:00:00.456Z",
		"field1":  "test_field",
		"field2":  "123", // should be casted to float
		"unknown": "test_unknown",
		// users collection specific casting test
		"passwordHash":           "test_passwordHash",
		"emailVisibility":        "12345", // should be casted to bool only for users collections
		"username":               123,     // should be casted to string only for users collections
		"email":                  "test_email",
		"verified":               true,
		"tokenKey":               "test_tokenKey",
		"lastResetSentAt":        "2022-01-01 11:00:00.000", // should be casted to DateTime only for users collections
		"lastVerificationSentAt": "2022-01-01 12:00:00.000", // should be casted to DateTime only for users collections
	}

	scenarios := []struct {
		collectionType string
	}{
		{collectionmodels.CollectionTypeBase},
		{collectionmodels.CollectionTypeUsers},
	}

	for i, s := range scenarios {
		collection.Type = s.collectionType
		m := recordmodels.NewRecord(collection)

		m.Load(data)

		expectations := map[string]any{}
		for k, v := range data {
			expectations[k] = v
		}

		expectations["created"], _ = types.ParseDateTime("2022-01-01 10:00:00.123Z")
		expectations["updated"], _ = types.ParseDateTime("2022-01-01 10:00:00.456Z")
		expectations["field2"] = 123.0

		// extra casting test
		if collection.IsUsers() {
			lastResetSentAt, _ := types.ParseDateTime(expectations["lastResetSentAt"])
			lastVerificationSentAt, _ := types.ParseDateTime(expectations["lastVerificationSentAt"])
			expectations["emailVisibility"] = false
			expectations["username"] = "123"
			expectations["verified"] = true
			expectations["lastResetSentAt"] = lastResetSentAt
			expectations["lastVerificationSentAt"] = lastVerificationSentAt
		}

		for k, v := range expectations {
			if m.Get(k) != v {
				t.Errorf("(%d) Expected field %s to be %v, got %v", i, k, v, m.Get(k))
			}
		}
	}
}

func TestRecordColumnValueMap(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{
					MaxSelect: 1,
					MaxSize:   1,
				},
			},
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{
					MaxSelect: 2,
					Values:    []string{"test1", "test2", "test3"},
				},
			},
			&collectionmodels.SchemaField{
				Name: "field4",
				Type: collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{
					MaxSelect: types.Pointer(2),
				},
			},
		),
	}

	scenarios := []struct {
		collectionType string
		expectedJson   string
	}{
		{
			collectionmodels.CollectionTypeBase,
			`{"created":"2022-01-01 10:00:30.123Z","field1":"test","field2":"test.png","field3":["test1","test2"],"field4":["test11","test12"],"id":"test_id","updated":""}`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			`{"created":"2022-01-01 10:00:30.123Z","email":"test_email","emailVisibility":true,"field1":"test","field2":"test.png","field3":["test1","test2"],"field4":["test11","test12"],"id":"test_id","lastLoginAlertSentAt":"","lastResetSentAt":"2022-01-02 10:00:30.123Z","lastVerificationSentAt":"","passwordHash":"test_passwordHash","tokenKey":"test_tokenKey","updated":"","username":"test_username","verified":false}`,
		},
	}

	created, _ := types.ParseDateTime("2022-01-01 10:00:30.123Z")
	lastResetSentAt, _ := types.ParseDateTime("2022-01-02 10:00:30.123Z")
	data := map[string]any{
		"id":              "test_id",
		"created":         created,
		"field1":          "test",
		"field2":          "test.png",
		"field3":          []string{"test1", "test2"},
		"field4":          []string{"test11", "test12", "test11"}, // strip duplicate,
		"unknown":         "test_unknown",
		"passwordHash":    "test_passwordHash",
		"username":        "test_username",
		"emailVisibility": true,
		"email":           "test_email",
		"verified":        "invalid", // should be casted
		"tokenKey":        "test_tokenKey",
		"lastResetSentAt": lastResetSentAt,
	}

	m := recordmodels.NewRecord(collection)

	for i, s := range scenarios {
		collection.Type = s.collectionType

		m.Load(data)

		result := m.ColumnValueMap()

		encoded, err := json.Marshal(result)
		if err != nil {
			t.Errorf("(%d) Unexpected error %v", i, err)
			continue
		}

		if str := string(encoded); str != s.expectedJson {
			t.Errorf("(%d) Expected \n%v \ngot \n%v", i, s.expectedJson, str)
		}
	}
}

func TestRecordPublicExportAndMarshalJSON(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Name: "c_name",
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{
					MaxSelect: 1,
					MaxSize:   1,
				},
			},
			&collectionmodels.SchemaField{
				Name: "field3",
				Type: collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{
					MaxSelect: 2,
					Values:    []string{"test1", "test2", "test3"},
				},
			},
		),
	}
	collection.Id = "c_id"

	scenarios := []struct {
		collectionType string
		exportHidden   bool
		exportUnknown  bool
		expectedJson   string
	}{
		// base
		{
			collectionmodels.CollectionTypeBase,
			false,
			false,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","updated":""}`,
		},
		{
			collectionmodels.CollectionTypeBase,
			true,
			false,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","updated":""}`,
		},
		{
			collectionmodels.CollectionTypeBase,
			false,
			true,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","email":"test_email","emailVisibility":"test_invalid","expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","lastResetSentAt":"2022-01-02 10:00:30.123Z","lastVerificationSentAt":"test_lastVerificationSentAt","passwordHash":"test_passwordHash","tokenKey":"test_tokenKey","unknown":"test_unknown","updated":"","username":123,"verified":true}`,
		},
		{
			collectionmodels.CollectionTypeBase,
			true,
			true,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","email":"test_email","emailVisibility":"test_invalid","expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","lastResetSentAt":"2022-01-02 10:00:30.123Z","lastVerificationSentAt":"test_lastVerificationSentAt","passwordHash":"test_passwordHash","tokenKey":"test_tokenKey","unknown":"test_unknown","updated":"","username":123,"verified":true}`,
		},

		// auth
		{
			collectionmodels.CollectionTypeUsers,
			false,
			false,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","emailVisibility":false,"expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","updated":"","username":"123","verified":true}`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			true,
			false,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","email":"test_email","emailVisibility":false,"expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","updated":"","username":"123","verified":true}`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			false,
			true,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","emailVisibility":false,"expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","unknown":"test_unknown","updated":"","username":"123","verified":true}`,
		},
		{
			collectionmodels.CollectionTypeUsers,
			true,
			true,
			`{"collectionId":"c_id","collectionName":"c_name","created":"2022-01-01 10:00:30.123Z","email":"test_email","emailVisibility":false,"expand":{"test":123},"field1":"test","field2":"test.png","field3":["test1","test2"],"id":"test_id","unknown":"test_unknown","updated":"","username":"123","verified":true}`,
		},
	}

	created, _ := types.ParseDateTime("2022-01-01 10:00:30.123Z")
	lastResetSentAt, _ := types.ParseDateTime("2022-01-02 10:00:30.123Z")

	data := map[string]any{
		"id":                     "test_id",
		"created":                created,
		"field1":                 "test",
		"field2":                 "test.png",
		"field3":                 []string{"test1", "test2"},
		"expand":                 map[string]any{"test": 123},
		"collectionId":           "m_id",   // should be always ignored
		"collectionName":         "m_name", // should be always ignored
		"unknown":                "test_unknown",
		"passwordHash":           "test_passwordHash",
		"username":               123,            // for users collections should be casted to string
		"emailVisibility":        "test_invalid", // for users collections should be casted to bool
		"email":                  "test_email",
		"verified":               true,
		"tokenKey":               "test_tokenKey",
		"lastResetSentAt":        lastResetSentAt,
		"lastVerificationSentAt": "test_lastVerificationSentAt",
	}

	m := recordmodels.NewRecord(collection)

	for i, s := range scenarios {
		collection.Type = s.collectionType

		m.Load(data)
		m.IgnoreEmailVisibility(s.exportHidden)
		m.WithUnknownData(s.exportUnknown)

		exportResult, err := json.Marshal(m.PublicExport())
		if err != nil {
			t.Errorf("(%d) Unexpected error %v", i, err)
			continue
		}
		exportResultStr := string(exportResult)

		// MarshalJSON and PublicExport should return the same
		marshalResult, err := m.MarshalJSON()
		if err != nil {
			t.Errorf("(%d) Unexpected error %v", i, err)
			continue
		}
		marshalResultStr := string(marshalResult)

		if exportResultStr != marshalResultStr {
			t.Errorf("(%d) Expected the PublicExport to be the same as MarshalJSON, but got \n%v \nvs \n%v", i, exportResultStr, marshalResultStr)
		}

		if exportResultStr != s.expectedJson {
			t.Errorf("(%d) Expected json \n%v \ngot \n%v", i, s.expectedJson, exportResultStr)
		}
	}
}

func TestRecordUnmarshalJSON(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "field1",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "field2",
				Type: collectionmodels.FieldTypeNumber,
			},
		),
	}

	data := map[string]any{
		"id":      "test_id",
		"created": "2022-01-01 10:00:00.123Z",
		"updated": "2022-01-01 10:00:00.456Z",
		"field1":  "test_field",
		"field2":  "123", // should be casted to float
		"unknown": "test_unknown",
		// users collection specific casting test
		"passwordHash":           "test_passwordHash",
		"emailVisibility":        "12345", // should be casted to bool only for users collections
		"username":               123.123, // should be casted to string only for users collections
		"email":                  "test_email",
		"verified":               true,
		"tokenKey":               "test_tokenKey",
		"lastResetSentAt":        "2022-01-01 11:00:00.000", // should be casted to DateTime only for users collections
		"lastVerificationSentAt": "2022-01-01 12:00:00.000", // should be casted to DateTime only for users collections
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Unexpected data marshal error %v", err)
	}

	scenarios := []struct {
		collectionType string
	}{
		{collectionmodels.CollectionTypeBase},
		{collectionmodels.CollectionTypeUsers},
	}

	// with invalid data
	m0 := recordmodels.NewRecord(collection)
	if err := m0.UnmarshalJSON([]byte("test")); err == nil {
		t.Fatal("Expected error, got nil")
	}

	// with valid data (it should be pretty much the same as load)
	for i, s := range scenarios {
		collection.Type = s.collectionType
		m := recordmodels.NewRecord(collection)

		err := m.UnmarshalJSON(dataRaw)
		if err != nil {
			t.Errorf("(%d) Unexpected error %v", i, err)
			continue
		}

		expectations := map[string]any{}
		for k, v := range data {
			expectations[k] = v
		}

		expectations["created"], _ = types.ParseDateTime("2022-01-01 10:00:00.123Z")
		expectations["updated"], _ = types.ParseDateTime("2022-01-01 10:00:00.456Z")
		expectations["field2"] = 123.0

		// extra casting test
		if collection.IsUsers() {
			lastResetSentAt, _ := types.ParseDateTime(expectations["lastResetSentAt"])
			lastVerificationSentAt, _ := types.ParseDateTime(expectations["lastVerificationSentAt"])
			expectations["emailVisibility"] = false
			expectations["username"] = "123.123"
			expectations["verified"] = true
			expectations["lastResetSentAt"] = lastResetSentAt
			expectations["lastVerificationSentAt"] = lastVerificationSentAt
		}

		for k, v := range expectations {
			if m.Get(k) != v {
				t.Errorf("(%d) Expected field %s to be %v, got %v", i, k, v, m.Get(k))
			}
		}
	}
}

func TestRecordReplaceModifers(t *testing.T) {
	t.Parallel()

	collection := &collectionmodels.Collection{
		Schema: collectionmodels.NewSchema(
			&collectionmodels.SchemaField{
				Name: "text",
				Type: collectionmodels.FieldTypeText,
			},
			&collectionmodels.SchemaField{
				Name: "number",
				Type: collectionmodels.FieldTypeNumber,
			},
			&collectionmodels.SchemaField{
				Name:    "rel_one",
				Type:    collectionmodels.FieldTypeRelation,
				Options: &collectionmodels.RelationOptions{MaxSelect: types.Pointer(1)},
			},
			&collectionmodels.SchemaField{
				Name: "rel_many",
				Type: collectionmodels.FieldTypeRelation,
			},
			&collectionmodels.SchemaField{
				Name:    "select_one",
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 1},
			},
			&collectionmodels.SchemaField{
				Name:    "select_many",
				Type:    collectionmodels.FieldTypeSelect,
				Options: &collectionmodels.SelectOptions{MaxSelect: 10},
			},
			&collectionmodels.SchemaField{
				Name:    "file_one",
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 1},
			},
			&collectionmodels.SchemaField{
				Name:    "file_one_index",
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 1},
			},
			&collectionmodels.SchemaField{
				Name:    "file_one_name",
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 1},
			},
			&collectionmodels.SchemaField{
				Name:    "file_many",
				Type:    collectionmodels.FieldTypeFile,
				Options: &collectionmodels.FileOptions{MaxSelect: 10},
			},
		),
	}

	record := recordmodels.NewRecord(collection)

	record.Load(map[string]any{
		"text":           "test",
		"number":         10,
		"rel_one":        "a",
		"rel_many":       []string{"a", "b"},
		"select_one":     "a",
		"select_many":    []string{"a", "b", "c"},
		"file_one":       "a",
		"file_one_index": "b",
		"file_one_name":  "c",
		"file_many":      []string{"a", "b", "c", "d", "e", "f"},
	})

	result := record.ReplaceModifers(map[string]any{
		"text-":            "m-",
		"text+":            "m+",
		"number-":          3,
		"number+":          5,
		"rel_one-":         "a",
		"rel_one+":         "b",
		"rel_many-":        []string{"a"},
		"rel_many+":        []string{"c", "d", "e"},
		"select_one-":      "a",
		"select_one+":      "c",
		"select_many-":     []string{"b", "c"},
		"select_many+":     []string{"d", "e"},
		"file_one+":        "skip", // should be ignored
		"file_one-":        "a",
		"file_one_index.0": "",
		"file_one_name.c":  "",
		"file_many+":       []string{"e", "f"}, // should be ignored
		"file_many-":       []string{"c", "d"},
		"file_many.f":      nil,
		"file_many.0":      nil,
	})

	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"file_many":["b","e"],"file_one":"","file_one_index":"","file_one_name":"","number":12,"rel_many":["b","c","d","e"],"rel_one":"b","select_many":["a","d","e"],"select_one":"c","text":"test"}`

	if v := string(raw); v != expected {
		t.Fatalf("Expected \n%s, \ngot \n%s", expected, v)
	}
}

// -------------------------------------------------------------------
// Auth helpers:
// -------------------------------------------------------------------

func TestRecordUsername(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue := "test 1232 !@#%" // formatting isn't checked

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetUsername(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.Username(); v != "" {
				t.Fatalf("(%d) Expected empty string, got %q", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameUsername); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameUsername, v)
			}
		} else {
			if err := m.SetUsername(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.Username(); v != testValue {
				t.Fatalf("(%d) Expected %q, got %q", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameUsername); v != testValue {
				t.Fatalf("(%d) Expected data field value %q, got %q", i, testValue, v)
			}
		}
	}
}

func TestRecordEmail(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue := "test 1232 !@#%" // formatting isn't checked

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetEmail(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.Email(); v != "" {
				t.Fatalf("(%d) Expected empty string, got %q", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameEmail); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameEmail, v)
			}
		} else {
			if err := m.SetEmail(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.Email(); v != testValue {
				t.Fatalf("(%d) Expected %q, got %q", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameEmail); v != testValue {
				t.Fatalf("(%d) Expected data field value %q, got %q", i, testValue, v)
			}
		}
	}
}

func TestRecordEmailVisibility(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		value          bool
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true, true},
		{collectionmodels.CollectionTypeBase, true, true},
		{collectionmodels.CollectionTypeUsers, false, false},
		{collectionmodels.CollectionTypeUsers, true, false},
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetEmailVisibility(s.value); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.EmailVisibility(); v != false {
				t.Fatalf("(%d) Expected empty string, got %v", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameEmailVisibility); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameEmailVisibility, v)
			}
		} else {
			if err := m.SetEmailVisibility(s.value); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.EmailVisibility(); v != s.value {
				t.Fatalf("(%d) Expected %v, got %v", i, s.value, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameEmailVisibility); v != s.value {
				t.Fatalf("(%d) Expected data field value %v, got %v", i, s.value, v)
			}
		}
	}
}

func TestRecordEmailVerified(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		value          bool
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true, true},
		{collectionmodels.CollectionTypeBase, true, true},
		{collectionmodels.CollectionTypeUsers, false, false},
		{collectionmodels.CollectionTypeUsers, true, false},
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetVerified(s.value); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.Verified(); v != false {
				t.Fatalf("(%d) Expected empty string, got %v", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameVerified); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameVerified, v)
			}
		} else {
			if err := m.SetVerified(s.value); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.Verified(); v != s.value {
				t.Fatalf("(%d) Expected %v, got %v", i, s.value, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameVerified); v != s.value {
				t.Fatalf("(%d) Expected data field value %v, got %v", i, s.value, v)
			}
		}
	}
}

func TestRecordTokenKey(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue := "test 1232 !@#%" // formatting isn't checked

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetTokenKey(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.TokenKey(); v != "" {
				t.Fatalf("(%d) Expected empty string, got %q", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameTokenKey); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameTokenKey, v)
			}
		} else {
			if err := m.SetTokenKey(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.TokenKey(); v != testValue {
				t.Fatalf("(%d) Expected %q, got %q", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameTokenKey); v != testValue {
				t.Fatalf("(%d) Expected data field value %q, got %q", i, testValue, v)
			}
		}
	}
}

func TestRecordRefreshTokenKey(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.RefreshTokenKey(); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.TokenKey(); v != "" {
				t.Fatalf("(%d) Expected empty string, got %q", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameTokenKey); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameTokenKey, v)
			}
		} else {
			if err := m.RefreshTokenKey(); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.TokenKey(); len(v) != 50 {
				t.Fatalf("(%d) Expected 50 chars, got %d", i, len(v))
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameTokenKey); v != m.TokenKey() {
				t.Fatalf("(%d) Expected data field value %q, got %q", i, m.TokenKey(), v)
			}
		}
	}
}

func TestRecordLastPasswordLoginAlertSentAt(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue, err := types.ParseDateTime("2022-01-01 00:00:00.123Z")
	if err != nil {
		t.Fatal(err)
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetLastLoginAlertSentAt(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.LastLoginAlertSentAt(); !v.IsZero() {
				t.Fatalf("(%d) Expected empty value, got %v", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastLoginAlertSentAt); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameLastLoginAlertSentAt, v)
			}
		} else {
			if err := m.SetLastLoginAlertSentAt(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.LastLoginAlertSentAt(); v != testValue {
				t.Fatalf("(%d) Expected %v, got %v", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastLoginAlertSentAt); v != testValue {
				t.Fatalf("(%d) Expected data field value %v, got %v", i, testValue, v)
			}
		}
	}
}

func TestRecordLastResetSentAt(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue, err := types.ParseDateTime("2022-01-01 00:00:00.123Z")
	if err != nil {
		t.Fatal(err)
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetLastResetSentAt(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.LastResetSentAt(); !v.IsZero() {
				t.Fatalf("(%d) Expected empty value, got %v", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastResetSentAt); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameLastResetSentAt, v)
			}
		} else {
			if err := m.SetLastResetSentAt(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.LastResetSentAt(); v != testValue {
				t.Fatalf("(%d) Expected %v, got %v", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastResetSentAt); v != testValue {
				t.Fatalf("(%d) Expected data field value %v, got %v", i, testValue, v)
			}
		}
	}
}

func TestRecordLastVerificationSentAt(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, true},
		{collectionmodels.CollectionTypeUsers, false},
	}

	testValue, err := types.ParseDateTime("2022-01-01 00:00:00.123Z")
	if err != nil {
		t.Fatal(err)
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetLastVerificationSentAt(testValue); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.LastVerificationSentAt(); !v.IsZero() {
				t.Fatalf("(%d) Expected empty value, got %v", i, v)
			}
			// verify that nothing is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastVerificationSentAt); v != nil {
				t.Fatalf("(%d) Didn't expect data field %q: %v", i, collectionmodels.FieldNameLastVerificationSentAt, v)
			}
		} else {
			if err := m.SetLastVerificationSentAt(testValue); err != nil {
				t.Fatalf("(%d) Expected nil, got error %v", i, err)
			}
			if v := m.LastVerificationSentAt(); v != testValue {
				t.Fatalf("(%d) Expected %v, got %v", i, testValue, v)
			}
			// verify that the field is stored in the record data slice
			if v := m.Get(collectionmodels.FieldNameLastVerificationSentAt); v != testValue {
				t.Fatalf("(%d) Expected data field value %v, got %v", i, testValue, v)
			}
		}
	}
}

func TestRecordPasswordHash(t *testing.T) {
	t.Parallel()

	m := recordmodels.NewRecord(&collectionmodels.Collection{})

	if v := m.PasswordHash(); v != "" {
		t.Errorf("Expected PasswordHash() to be empty, got %v", v)
	}

	m.Set(collectionmodels.FieldNamePasswordHash, "test")

	if v := m.PasswordHash(); v != "test" {
		t.Errorf("Expected PasswordHash() to be 'test', got %v", v)
	}
}

func TestRecordValidatePassword(t *testing.T) {
	t.Parallel()

	// 123456
	hash := "$2a$10$YKU8mPP8sTE3xZrpuM.xQuq27KJ7aIJB2oUeKPsDDqZshbl5g5cDK"

	scenarios := []struct {
		collectionType string
		password       string
		hash           string
		expected       bool
	}{
		{collectionmodels.CollectionTypeBase, "123456", hash, false},
		{collectionmodels.CollectionTypeUsers, "", "", false},
		{collectionmodels.CollectionTypeUsers, "", hash, false},
		{collectionmodels.CollectionTypeUsers, "123456", hash, true},
		{collectionmodels.CollectionTypeUsers, "654321", hash, false},
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)
		m.Set(collectionmodels.FieldNamePasswordHash, hash)

		if v := m.ValidatePassword(s.password); v != s.expected {
			t.Errorf("(%d) Expected %v, got %v", i, s.expected, v)
		}
	}
}

func TestRecordSetPassword(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		collectionType string
		password       string
		expectError    bool
	}{
		{collectionmodels.CollectionTypeBase, "", true},
		{collectionmodels.CollectionTypeBase, "123456", true},
		{collectionmodels.CollectionTypeUsers, "", true},
		{collectionmodels.CollectionTypeUsers, "123456", false},
	}

	for i, s := range scenarios {
		collection := &collectionmodels.Collection{Type: s.collectionType}
		m := recordmodels.NewRecord(collection)

		if s.expectError {
			if err := m.SetPassword(s.password); err == nil {
				t.Errorf("(%d) Expected error, got nil", i)
			}
			if v := m.GetString(collectionmodels.FieldNamePasswordHash); v != "" {
				t.Errorf("(%d) Expected empty hash, got %q", i, v)
			}
		} else {
			if err := m.SetPassword(s.password); err != nil {
				t.Errorf("(%d) Expected nil, got err", i)
			}
			if v := m.GetString(collectionmodels.FieldNamePasswordHash); v == "" {
				t.Errorf("(%d) Expected non empty hash", i)
			}
			if !m.ValidatePassword(s.password) {
				t.Errorf("(%d) Expected true, got false", i)
			}
		}
	}
}
