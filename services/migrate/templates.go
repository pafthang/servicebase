package migrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
)

const TemplateLangGo = "go"

var emptyTemplateErr = errors.New("empty template")

func (s *Service) goBlankTemplate() (string, error) {
	const template = `
package %s

import (
	"github.com/pocketbase/dbx"
	m "github.com/pafthang/servicebase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		// add up queries...

		return nil
	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
`

	return fmt.Sprintf(template, filepath.Base(s.config.Dir)), nil
}

func (s *Service) goSnapshotTemplate(collections []*collectionmodels.Collection) (string, error) {
	jsonData, err := marhshalWithoutEscape(collections, "\t\t", "\t")
	if err != nil {
		return "", fmt.Errorf("failed to serialize collections list: %w", err)
	}

	const template = `package %s

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pafthang/servicebase/daos"
	m "github.com/pafthang/servicebase/migrations"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := ` + "`%s`" + `

		collections := []*collectionmodels.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return daos.New(db).ImportCollections(collections, true, nil)
	}, func(db dbx.Builder) error {
		return nil
	})
}
`
	return fmt.Sprintf(
		template,
		filepath.Base(s.config.Dir),
		escapeBacktick(string(jsonData)),
	), nil
}

func (s *Service) goCreateTemplate(collection *collectionmodels.Collection) (string, error) {
	jsonData, err := marhshalWithoutEscape(collection, "\t\t", "\t")
	if err != nil {
		return "", fmt.Errorf("failed to serialize collections list: %w", err)
	}

	const template = `package %s

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pafthang/servicebase/daos"
	m "github.com/pafthang/servicebase/migrations"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := ` + "`%s`" + `

		collection := &collectionmodels.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId(%q)
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
`

	return fmt.Sprintf(
		template,
		filepath.Base(s.config.Dir),
		escapeBacktick(string(jsonData)),
		collection.Id,
	), nil
}

func (s *Service) goDeleteTemplate(collection *collectionmodels.Collection) (string, error) {
	jsonData, err := marhshalWithoutEscape(collection, "\t\t", "\t")
	if err != nil {
		return "", fmt.Errorf("failed to serialize collections list: %w", err)
	}

	const template = `package %s

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pafthang/servicebase/daos"
	m "github.com/pafthang/servicebase/migrations"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId(%q)
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	}, func(db dbx.Builder) error {
		jsonData := ` + "`%s`" + `

		collection := &collectionmodels.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	})
}
`

	return fmt.Sprintf(
		template,
		filepath.Base(s.config.Dir),
		collection.Id,
		escapeBacktick(string(jsonData)),
	), nil
}

func (s *Service) goDiffTemplate(new *collectionmodels.Collection, old *collectionmodels.Collection) (string, error) {
	if new == nil && old == nil {
		return "", errors.New("the diff template require at least one of the collection to be non-nil")
	}

	if new == nil {
		return s.goDeleteTemplate(old)
	}

	if old == nil {
		return s.goCreateTemplate(new)
	}

	upParts := []string{}
	downParts := []string{}
	varName := "collection"
	if old.Name != new.Name {
		upParts = append(upParts, fmt.Sprintf("%s.Name = %q\n", varName, new.Name))
		downParts = append(downParts, fmt.Sprintf("%s.Name = %q\n", varName, old.Name))
	}

	if old.Type != new.Type {
		upParts = append(upParts, fmt.Sprintf("%s.Type = %q\n", varName, new.Type))
		downParts = append(downParts, fmt.Sprintf("%s.Type = %q\n", varName, old.Type))
	}

	if old.System != new.System {
		upParts = append(upParts, fmt.Sprintf("%s.System = %t\n", varName, new.System))
		downParts = append(downParts, fmt.Sprintf("%s.System = %t\n", varName, old.System))
	}

	formatRule := func(prop string, rule *string) string {
		if rule == nil {
			return fmt.Sprintf("%s.%s = nil\n", varName, prop)
		}

		return fmt.Sprintf("%s.%s = types.Pointer(%s)\n", varName, prop, strconv.Quote(*rule))
	}

	if old.ListRule != new.ListRule {
		oldRule := formatRule("ListRule", old.ListRule)
		newRule := formatRule("ListRule", new.ListRule)

		if oldRule != newRule {
			upParts = append(upParts, newRule)
			downParts = append(downParts, oldRule)
		}
	}

	if old.ViewRule != new.ViewRule {
		oldRule := formatRule("ViewRule", old.ViewRule)
		newRule := formatRule("ViewRule", new.ViewRule)

		if oldRule != newRule {
			upParts = append(upParts, newRule)
			downParts = append(downParts, oldRule)
		}
	}

	if old.CreateRule != new.CreateRule {
		oldRule := formatRule("CreateRule", old.CreateRule)
		newRule := formatRule("CreateRule", new.CreateRule)

		if oldRule != newRule {
			upParts = append(upParts, newRule)
			downParts = append(downParts, oldRule)
		}
	}

	if old.UpdateRule != new.UpdateRule {
		oldRule := formatRule("UpdateRule", old.UpdateRule)
		newRule := formatRule("UpdateRule", new.UpdateRule)

		if oldRule != newRule {
			upParts = append(upParts, newRule)
			downParts = append(downParts, oldRule)
		}
	}

	if old.DeleteRule != new.DeleteRule {
		oldRule := formatRule("DeleteRule", old.DeleteRule)
		newRule := formatRule("DeleteRule", new.DeleteRule)

		if oldRule != newRule {
			upParts = append(upParts, newRule)
			downParts = append(downParts, oldRule)
		}
	}

	rawNewOptions, err := marhshalWithoutEscape(new.Options, "\t\t", "\t")
	if err != nil {
		return "", err
	}
	rawOldOptions, err := marhshalWithoutEscape(old.Options, "\t\t", "\t")
	if err != nil {
		return "", err
	}
	if !bytes.Equal(rawNewOptions, rawOldOptions) {
		upParts = append(upParts, "options := map[string]any{}")
		upParts = append(upParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), &options)", escapeBacktick(string(rawNewOptions)))))
		upParts = append(upParts, fmt.Sprintf("%s.SetOptions(options)\n", varName))
		downParts = append(downParts, "options := map[string]any{}")
		downParts = append(downParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), &options)", escapeBacktick(string(rawOldOptions)))))
		downParts = append(downParts, fmt.Sprintf("%s.SetOptions(options)\n", varName))
	}

	rawNewIndexes, err := marhshalWithoutEscape(new.Indexes, "\t\t", "\t")
	if err != nil {
		return "", err
	}
	rawOldIndexes, err := marhshalWithoutEscape(old.Indexes, "\t\t", "\t")
	if err != nil {
		return "", err
	}
	if !bytes.Equal(rawNewIndexes, rawOldIndexes) {
		upParts = append(upParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), &%s.Indexes)", escapeBacktick(string(rawNewIndexes)), varName))+"\n")
		downParts = append(downParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), &%s.Indexes)", escapeBacktick(string(rawOldIndexes)), varName))+"\n")
	}

	for _, oldField := range old.Schema.Fields() {
		if new.Schema.GetFieldById(oldField.Id) != nil {
			continue
		}

		rawOldField, err := marhshalWithoutEscape(oldField, "\t\t", "\t")
		if err != nil {
			return "", err
		}

		fieldVar := fmt.Sprintf("del_%s", oldField.Name)

		upParts = append(upParts, "// remove")
		upParts = append(upParts, fmt.Sprintf("%s.Schema.RemoveField(%q)\n", varName, oldField.Id))

		downParts = append(downParts, "// add")
		downParts = append(downParts, fmt.Sprintf("%s := &collectionmodels.SchemaField{}", fieldVar))
		downParts = append(downParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), %s)", escapeBacktick(string(rawOldField)), fieldVar)))
		downParts = append(downParts, fmt.Sprintf("%s.Schema.AddField(%s)\n", varName, fieldVar))
	}

	for _, newField := range new.Schema.Fields() {
		if old.Schema.GetFieldById(newField.Id) != nil {
			continue
		}

		rawNewField, err := marhshalWithoutEscape(newField, "\t\t", "\t")
		if err != nil {
			return "", err
		}

		fieldVar := fmt.Sprintf("new_%s", newField.Name)

		upParts = append(upParts, "// add")
		upParts = append(upParts, fmt.Sprintf("%s := &collectionmodels.SchemaField{}", fieldVar))
		upParts = append(upParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), %s)", escapeBacktick(string(rawNewField)), fieldVar)))
		upParts = append(upParts, fmt.Sprintf("%s.Schema.AddField(%s)\n", varName, fieldVar))

		downParts = append(downParts, "// remove")
		downParts = append(downParts, fmt.Sprintf("%s.Schema.RemoveField(%q)\n", varName, newField.Id))
	}

	for _, newField := range new.Schema.Fields() {
		oldField := old.Schema.GetFieldById(newField.Id)
		if oldField == nil {
			continue
		}

		rawNewField, err := marhshalWithoutEscape(newField, "\t\t", "\t")
		if err != nil {
			return "", err
		}

		rawOldField, err := marhshalWithoutEscape(oldField, "\t\t", "\t")
		if err != nil {
			return "", err
		}

		if bytes.Equal(rawNewField, rawOldField) {
			continue
		}

		fieldVar := fmt.Sprintf("edit_%s", newField.Name)

		upParts = append(upParts, "// update")
		upParts = append(upParts, fmt.Sprintf("%s := &collectionmodels.SchemaField{}", fieldVar))
		upParts = append(upParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), %s)", escapeBacktick(string(rawNewField)), fieldVar)))
		upParts = append(upParts, fmt.Sprintf("%s.Schema.AddField(%s)\n", varName, fieldVar))

		downParts = append(downParts, "// update")
		downParts = append(downParts, fmt.Sprintf("%s := &collectionmodels.SchemaField{}", fieldVar))
		downParts = append(downParts, goErrIf(fmt.Sprintf("json.Unmarshal([]byte(`%s`), %s)", escapeBacktick(string(rawOldField)), fieldVar)))
		downParts = append(downParts, fmt.Sprintf("%s.Schema.AddField(%s)\n", varName, fieldVar))
	}

	if len(upParts) == 0 && len(downParts) == 0 {
		return "", emptyTemplateErr
	}

	up := strings.Join(upParts, "\n\t\t")
	down := strings.Join(downParts, "\n\t\t")
	combined := up + down

	var imports string
	if strings.Contains(combined, "json.Unmarshal(") || strings.Contains(combined, "json.Marshal(") {
		imports += "\n\t\"encoding/json\"\n"
	}

	imports += "\n\t\"github.com/pocketbase/dbx\""
	imports += "\n\t\"github.com/pafthang/servicebase/daos\""
	imports += "\n\tm \"github.com/pafthang/servicebase/migrations\""

	if strings.Contains(combined, "collectionmodels.SchemaField{") {
		imports += "\n\tcollectionmodels \"github.com/pafthang/servicebase/services/collection/models\""
	}

	if strings.Contains(combined, "types.Pointer(") {
		imports += "\n\t\"github.com/pafthang/servicebase/tools/types\""
	}

	const template = `package %s

import (%s
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId(%q)
		if err != nil {
			return err
		}

		%s

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId(%q)
		if err != nil {
			return err
		}

		%s

		return dao.SaveCollection(collection)
	})
}
`

	return fmt.Sprintf(
		template,
		filepath.Base(s.config.Dir),
		imports,
		old.Id, strings.TrimSpace(up),
		new.Id, strings.TrimSpace(down),
	), nil
}

func marhshalWithoutEscape(v any, prefix string, indent string) ([]byte, error) {
	raw, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, err
	}

	return []byte(strings.ReplaceAll(string(raw), "\\u003c", "<")), nil
}

func goErrIf(expr string) string {
	return fmt.Sprintf("if err := %s; err != nil {\n\t\t\treturn err\n\t\t}", expr)
}

func escapeBacktick(str string) string {
	return strings.ReplaceAll(str, "`", "` + \"`\" + `")
}
