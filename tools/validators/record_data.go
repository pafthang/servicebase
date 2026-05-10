package validators

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"

	"github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pafthang/servicebase/daos"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/list"
	"github.com/pafthang/servicebase/tools/types"
	"github.com/pocketbase/dbx"
)

var requiredErr = validation.NewError("validation_required", "Missing required value")

// NewRecordDataValidator creates new [models.Record] data validator
// using the provided record constraints and models.
//
// Example:
//
//	validator := NewRecordDataValidator(app.Dao(), record, nil)
//	err := validator.Validate(map[string]any{"test":123})
func NewRecordDataValidator(
	dao *daos.Dao,
	record *recordmodels.Record,
	uploadedFiles map[string][]*filesystem.File,
) *RecordDataValidator {
	return &RecordDataValidator{
		dao:           dao,
		record:        record,
		uploadedFiles: uploadedFiles,
	}
}

// RecordDataValidator defines a  model.Record data validator
// using the provided record constraints and models.
type RecordDataValidator struct {
	dao           *daos.Dao
	record        *recordmodels.Record
	uploadedFiles map[string][]*filesystem.File
}

// Validate validates the provided `data` by checking it against
// the validator record constraints and models.
func (validator *RecordDataValidator) Validate(data map[string]any) error {
	keyedSchema := validator.record.Collection().Schema.AsMap()
	if len(keyedSchema) == 0 {
		return nil // no fields to check
	}

	if len(data) == 0 {
		return validation.NewError("validation_empty_data", "No data to validate")
	}

	errs := validation.Errors{}

	// check for unknown fields
	for key := range data {
		if _, ok := keyedSchema[key]; !ok {
			errs[key] = validation.NewError("validation_unknown_field", "Unknown field")
		}
	}
	if len(errs) > 0 {
		return errs
	}

	for key, field := range keyedSchema {
		// normalize value to emulate the same behavior
		// when fetching or persisting the record model
		value := field.PrepareValue(data[key])

		// check required constraint
		if field.Required && validation.Required.Validate(value) != nil {
			errs[key] = requiredErr
			continue
		}

		// validate field value by its field type
		if err := validator.checkFieldValue(field, value); err != nil {
			errs[key] = err
			continue
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func (validator *RecordDataValidator) checkFieldValue(field *models.SchemaField, value any) error {
	switch field.Type {
	case models.FieldTypeText:
		return validator.checkTextValue(field, value)
	case models.FieldTypeNumber:
		return validator.checkNumberValue(field, value)
	case models.FieldTypeBool:
		return validator.checkBoolValue(field, value)
	case models.FieldTypeEmail:
		return validator.checkEmailValue(field, value)
	case models.FieldTypeUrl:
		return validator.checkUrlValue(field, value)
	case models.FieldTypeEditor:
		return validator.checkEditorValue(field, value)
	case models.FieldTypeDate:
		return validator.checkDateValue(field, value)
	case models.FieldTypeSelect:
		return validator.checkSelectValue(field, value)
	case models.FieldTypeJson:
		return validator.checkJsonValue(field, value)
	case models.FieldTypeFile:
		return validator.checkFileValue(field, value)
	case models.FieldTypeRelation:
		return validator.checkRelationValue(field, value)
	}

	return nil
}

func (validator *RecordDataValidator) checkTextValue(field *models.SchemaField, value any) error {
	val, _ := value.(string)
	if val == "" {
		return nil // nothing to check (skip zero-defaults)
	}

	options, _ := field.Options.(*models.TextOptions)

	// note: casted to []rune to count multi-byte chars as one
	length := len([]rune(val))

	if options.Min != nil && length < *options.Min {
		return validation.NewError("validation_min_text_constraint", fmt.Sprintf("Must be at least %d character(s)", *options.Min))
	}

	if options.Max != nil && length > *options.Max {
		return validation.NewError("validation_max_text_constraint", fmt.Sprintf("Must be less than %d character(s)", *options.Max))
	}

	if options.Pattern != "" {
		match, _ := regexp.MatchString(options.Pattern, val)
		if !match {
			return validation.NewError("validation_invalid_format", "Invalid value format")
		}
	}

	return nil
}

func (validator *RecordDataValidator) checkNumberValue(field *models.SchemaField, value any) error {
	val, _ := value.(float64)
	if val == 0 {
		return nil // nothing to check (skip zero-defaults)
	}

	if math.IsInf(val, 0) || math.IsNaN(val) {
		return validation.NewError("validation_nan", "The submitted number is not properly formatted")
	}

	options, _ := field.Options.(*models.NumberOptions)

	if options.NoDecimal && val != float64(int64(val)) {
		return validation.NewError("validation_no_decimal_constraint", "Decimal numbers are not allowed")
	}

	if options.Min != nil && val < *options.Min {
		return validation.NewError("validation_min_number_constraint", fmt.Sprintf("Must be larger than %f", *options.Min))
	}

	if options.Max != nil && val > *options.Max {
		return validation.NewError("validation_max_number_constraint", fmt.Sprintf("Must be less than %f", *options.Max))
	}

	return nil
}

func (validator *RecordDataValidator) checkBoolValue(field *models.SchemaField, value any) error {
	return nil
}

func (validator *RecordDataValidator) checkEmailValue(field *models.SchemaField, value any) error {
	val, _ := value.(string)
	if val == "" {
		return nil // nothing to check
	}

	if is.EmailFormat.Validate(val) != nil {
		return validation.NewError("validation_invalid_email", "Must be a valid email")
	}

	options, _ := field.Options.(*models.EmailOptions)
	domain := val[strings.LastIndex(val, "@")+1:]

	// only domains check
	if len(options.OnlyDomains) > 0 && !list.ExistInSlice(domain, options.OnlyDomains) {
		return validation.NewError("validation_email_domain_not_allowed", "Email domain is not allowed")
	}

	// except domains check
	if len(options.ExceptDomains) > 0 && list.ExistInSlice(domain, options.ExceptDomains) {
		return validation.NewError("validation_email_domain_not_allowed", "Email domain is not allowed")
	}

	return nil
}

func (validator *RecordDataValidator) checkUrlValue(field *models.SchemaField, value any) error {
	val, _ := value.(string)
	if val == "" {
		return nil // nothing to check
	}

	if is.URL.Validate(val) != nil {
		return validation.NewError("validation_invalid_url", "Must be a valid url")
	}

	options, _ := field.Options.(*models.UrlOptions)

	// extract host/domain
	u, _ := url.Parse(val)
	host := u.Host

	// only domains check
	if len(options.OnlyDomains) > 0 && !list.ExistInSlice(host, options.OnlyDomains) {
		return validation.NewError("validation_url_domain_not_allowed", "Url domain is not allowed")
	}

	// except domains check
	if len(options.ExceptDomains) > 0 && list.ExistInSlice(host, options.ExceptDomains) {
		return validation.NewError("validation_url_domain_not_allowed", "Url domain is not allowed")
	}

	return nil
}

func (validator *RecordDataValidator) checkEditorValue(field *models.SchemaField, value any) error {
	return nil
}

func (validator *RecordDataValidator) checkDateValue(field *models.SchemaField, value any) error {
	val, _ := value.(types.DateTime)
	if val.IsZero() {
		if field.Required {
			return requiredErr
		}
		return nil // nothing to check
	}

	options, _ := field.Options.(*models.DateOptions)

	if !options.Min.IsZero() {
		if err := validation.Min(options.Min.Time()).Validate(val.Time()); err != nil {
			return err
		}
	}

	if !options.Max.IsZero() {
		if err := validation.Max(options.Max.Time()).Validate(val.Time()); err != nil {
			return err
		}
	}

	return nil
}

func (validator *RecordDataValidator) checkSelectValue(field *models.SchemaField, value any) error {
	normalizedVal := list.ToUniqueStringSlice(value)
	if len(normalizedVal) == 0 {
		if field.Required {
			return requiredErr
		}
		return nil // nothing to check
	}

	options, _ := field.Options.(*models.SelectOptions)

	// check max selected items
	if len(normalizedVal) > options.MaxSelect {
		return validation.NewError("validation_too_many_values", fmt.Sprintf("Select no more than %d", options.MaxSelect))
	}

	// check against the allowed values
	for _, val := range normalizedVal {
		if !list.ExistInSlice(val, options.Values) {
			return validation.NewError("validation_invalid_value", "Invalid value "+val)
		}
	}

	return nil
}

var emptyJsonValues = []string{
	"null", `""`, "[]", "{}",
}

func (validator *RecordDataValidator) checkJsonValue(field *models.SchemaField, value any) error {
	if is.JSON.Validate(value) != nil {
		return validation.NewError("validation_invalid_json", "Must be a valid json value")
	}

	raw, _ := types.ParseJsonRaw(value)

	options, _ := field.Options.(*models.JsonOptions)

	if len(raw) > options.MaxSize {
		return validation.NewError("validation_json_size_limit", fmt.Sprintf("The maximum allowed JSON size is %v bytes", options.MaxSize))
	}

	rawStr := strings.TrimSpace(raw.String())
	if field.Required && list.ExistInSlice(rawStr, emptyJsonValues) {
		return requiredErr
	}

	return nil
}

func (validator *RecordDataValidator) checkFileValue(field *models.SchemaField, value any) error {
	names := list.ToUniqueStringSlice(value)
	if len(names) == 0 && field.Required {
		return requiredErr
	}

	options, _ := field.Options.(*models.FileOptions)

	if len(names) > options.MaxSelect {
		return validation.NewError("validation_too_many_values", fmt.Sprintf("Select no more than %d", options.MaxSelect))
	}

	// extract the uploaded files
	files := make([]*filesystem.File, 0, len(validator.uploadedFiles[field.Name]))
	for _, file := range validator.uploadedFiles[field.Name] {
		if list.ExistInSlice(file.Name, names) {
			files = append(files, file)
		}
	}

	for _, file := range files {
		// check size
		if err := UploadedFileSize(options.MaxSize)(file); err != nil {
			return err
		}

		// check type
		if len(options.MimeTypes) > 0 {
			if err := UploadedFileMimeType(options.MimeTypes)(file); err != nil {
				return err
			}
		}
	}

	return nil
}

func (validator *RecordDataValidator) checkRelationValue(field *models.SchemaField, value any) error {
	ids := list.ToUniqueStringSlice(value)
	if len(ids) == 0 {
		if field.Required {
			return requiredErr
		}
		return nil // nothing to check
	}

	options, _ := field.Options.(*models.RelationOptions)

	if options.MinSelect != nil && len(ids) < *options.MinSelect {
		return validation.NewError("validation_not_enough_values", fmt.Sprintf("Select at least %d", *options.MinSelect))
	}

	if options.MaxSelect != nil && len(ids) > *options.MaxSelect {
		return validation.NewError("validation_too_many_values", fmt.Sprintf("Select no more than %d", *options.MaxSelect))
	}

	// check if the related records exist
	// ---
	relCollection, err := validator.dao.FindCollectionByNameOrId(options.CollectionId)
	if err != nil {
		return validation.NewError("validation_missing_rel_collection", "Relation connection is missing or cannot be accessed")
	}

	var total int
	validator.dao.RecordQuery(relCollection).
		Select("count(*)").
		AndWhere(dbx.In("id", list.ToInterfaceSlice(ids)...)).
		Row(&total)
	if total != len(ids) {
		return validation.NewError("validation_missing_rel_records", "Failed to find all relation records with the provided ids")
	}
	// ---

	return nil
}
