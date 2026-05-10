package daos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"

	recordmodels "github.com/pafthang/servicebase/services/record/models"

	"github.com/pafthang/servicebase/tools/inflector"
	"github.com/pafthang/servicebase/tools/list"
	"github.com/pafthang/servicebase/tools/resolvers"
	"github.com/pafthang/servicebase/tools/search"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pafthang/servicebase/tools/types"
	"github.com/pocketbase/dbx"
	"github.com/spf13/cast"
)

// RecordQuery returns a new Record select query from a collection model, id or name.
//
// In case a collection id or name is provided and that collection doesn't
// actually exists, the generated query will be created with a cancelled context
// and will fail once an executor (Row(), One(), All(), etc.) is called.
func (dao *Dao) RecordQuery(collectionModelOrIdentifier any) *dbx.SelectQuery {
	var tableName string
	var collection *collectionmodels.Collection
	var collectionErr error
	switch c := collectionModelOrIdentifier.(type) {
	case *collectionmodels.Collection:
		collection = c
		tableName = collection.Name
	case collectionmodels.Collection:
		collection = &c
		tableName = collection.Name
	case string:
		collection, collectionErr = dao.FindCollectionByNameOrId(c)
		if collection != nil {
			tableName = collection.Name
		}
	default:
		collectionErr = errors.New("unsupported collection identifier, must be collection model, id or name")
	}

	// update with some fake table name for easier debugging
	if tableName == "" {
		tableName = "@@__invalidCollectionModelOrIdentifier"
	}

	selectCols := fmt.Sprintf("%s.*", dao.DB().QuoteSimpleColumnName(tableName))

	query := dao.DB().Select(selectCols).From(tableName)

	// in case of an error attach a new context and cancel it immediately with the error
	if collectionErr != nil {
		// @todo consider changing to WithCancelCause when upgrading
		// the min Go requirement to 1.20, so that we can pass the error
		ctx, cancelFunc := context.WithCancel(context.Background())
		query.WithContext(ctx)
		cancelFunc()
	}

	return query.WithBuildHook(func(q *dbx.Query) {
		q.WithExecHook(execLockRetry(dao.ModelQueryTimeout, dao.MaxLockRetries)).
			WithOneHook(func(q *dbx.Query, a any, op func(b any) error) error {
				switch v := a.(type) {
				case *recordmodels.Record:
					if v == nil {
						return op(a)
					}

					row := dbx.NullStringMap{}
					if err := op(&row); err != nil {
						return err
					}

					record := recordmodels.NewRecordFromNullStringMap(collection, row)

					*v = *record

					return nil
				default:
					return op(a)
				}
			}).
			WithAllHook(func(q *dbx.Query, sliceA any, op func(sliceB any) error) error {
				switch v := sliceA.(type) {
				case *[]*recordmodels.Record:
					if v == nil {
						return op(sliceA)
					}

					rows := []dbx.NullStringMap{}
					if err := op(&rows); err != nil {
						return err
					}

					records := recordmodels.NewRecordsFromNullStringMaps(collection, rows)

					*v = records

					return nil
				case *[]recordmodels.Record:
					if v == nil {
						return op(sliceA)
					}

					rows := []dbx.NullStringMap{}
					if err := op(&rows); err != nil {
						return err
					}

					records := recordmodels.NewRecordsFromNullStringMaps(collection, rows)

					nonPointers := make([]recordmodels.Record, len(records))
					for i, r := range records {
						nonPointers[i] = *r
					}

					*v = nonPointers

					return nil
				default:
					return op(sliceA)
				}
			})
	})
}

// FindRecordById finds the Record model by its id.
func (dao *Dao) FindRecordById(
	collectionNameOrId string,
	recordId string,
	optFilters ...func(q *dbx.SelectQuery) error,
) (*recordmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	query := dao.RecordQuery(collection).
		AndWhere(dbx.HashExp{collection.Name + ".id": recordId})

	for _, filter := range optFilters {
		if filter == nil {
			continue
		}
		if err := filter(query); err != nil {
			return nil, err
		}
	}

	record := &recordmodels.Record{}

	if err := query.Limit(1).One(record); err != nil {
		return nil, err
	}

	return record, nil
}

// FindRecordsByIds finds all Record models by the provided ids.
// If no records are found, returns an empty slice.
func (dao *Dao) FindRecordsByIds(
	collectionNameOrId string,
	recordIds []string,
	optFilters ...func(q *dbx.SelectQuery) error,
) ([]*recordmodels.Record, error) {
	if len(recordIds) == 0 {
		return make([]*recordmodels.Record, 0), nil
	}

	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	query := dao.RecordQuery(collection).
		AndWhere(dbx.In(
			collection.Name+".id",
			list.ToInterfaceSlice(recordIds)...,
		))

	for _, filter := range optFilters {
		if filter == nil {
			continue
		}
		if err := filter(query); err != nil {
			return nil, err
		}
	}

	records := make([]*recordmodels.Record, 0, len(recordIds))

	if err := query.All(&records); err != nil {
		return nil, err
	}

	return records, nil
}

// FindRecordsByExpr finds all records by the specified db expression.
//
// Returns all collection records if no expressions are provided.
//
// Returns an empty slice if no records are found.
//
// Example:
//
//	expr1 := dbx.HashExp{"email": "test@example.com"}
//	expr2 := dbx.NewExp("LOWER(username) = {:username}", dbx.Params{"username": "test"})
//	dao.FindRecordsByExpr("example", expr1, expr2)
func (dao *Dao) FindRecordsByExpr(collectionNameOrId string, exprs ...dbx.Expression) ([]*recordmodels.Record, error) {
	query := dao.RecordQuery(collectionNameOrId)

	// add only the non-nil expressions
	for _, expr := range exprs {
		if expr != nil {
			query.AndWhere(expr)
		}
	}

	var records []*recordmodels.Record

	if err := query.All(&records); err != nil {
		return nil, err
	}

	return records, nil
}

// FindFirstRecordByData returns the first found record matching
// the provided key-value pair.
func (dao *Dao) FindFirstRecordByData(
	collectionNameOrId string,
	key string,
	value any,
) (*recordmodels.Record, error) {
	record := &recordmodels.Record{}

	err := dao.RecordQuery(collectionNameOrId).
		AndWhere(dbx.HashExp{inflector.Columnify(key): value}).
		Limit(1).
		One(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// FindRecordsByFilter returns limit number of records matching the
// provided string filter.
//
// NB! Use the last "params" argument to bind untrusted user variables!
//
// The sort argument is optional and can be empty string OR the same format
// used in the web APIs, eg. "-created,title".
//
// If the limit argument is <= 0, no limit is applied to the query and
// all matching records are returned.
//
// Example:
//
//	dao.FindRecordsByFilter(
//		"posts",
//		"title ~ {:title} && visible = {:visible}",
//		"-created",
//		10,
//		0,
//		dbx.Params{"title": "lorem ipsum", "visible": true}
//	)
func (dao *Dao) FindRecordsByFilter(
	collectionNameOrId string,
	filter string,
	sort string,
	limit int,
	offset int,
	params ...dbx.Params,
) ([]*recordmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	q := dao.RecordQuery(collection)

	// build a fields resolver and attach the generated conditions to the query
	// ---
	resolver := resolvers.NewRecordFieldResolver(
		dao,
		collection, // the base collection
		nil,        // no request data
		true,       // allow searching hidden/protected fields like "email"
	)

	expr, err := search.FilterData(filter).BuildExpr(resolver, params...)
	if err != nil || expr == nil {
		return nil, errors.New("invalid or empty filter expression")
	}
	q.AndWhere(expr)

	if sort != "" {
		for _, sortField := range search.ParseSortFromString(sort) {
			expr, err := sortField.BuildExpr(resolver)
			if err != nil {
				return nil, err
			}
			if expr != "" {
				q.AndOrderBy(expr)
			}
		}
	}

	resolver.UpdateQuery(q) // attaches any adhoc joins and aliases
	// ---

	if offset > 0 {
		q.Offset(int64(offset))
	}

	if limit > 0 {
		q.Limit(int64(limit))
	}

	records := []*recordmodels.Record{}

	if err := q.All(&records); err != nil {
		return nil, err
	}

	return records, nil
}

// FindFirstRecordByFilter returns the first available record matching the provided filter.
//
// NB! Use the last params argument to bind untrusted user variables!
//
// Example:
//
//	dao.FindFirstRecordByFilter("posts", "slug={:slug} && status='public'", dbx.Params{"slug": "test"})
func (dao *Dao) FindFirstRecordByFilter(
	collectionNameOrId string,
	filter string,
	params ...dbx.Params,
) (*recordmodels.Record, error) {
	result, err := dao.FindRecordsByFilter(collectionNameOrId, filter, "", 1, 0, params...)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result[0], nil
}

// IsRecordValueUnique checks if the provided key-value pair is a unique Record value.
//
// For correctness, if the collection is "auth" and the key is "username",
// the unique check will be case insensitive.
//
// NB! Array values (eg. from multiple select fields) are matched
// as a serialized json strings (eg. `["a","b"]`), so the value uniqueness
// depends on the elements order. Or in other words the following values
// are considered different: `[]string{"a","b"}` and `[]string{"b","a"}`
func (dao *Dao) IsRecordValueUnique(
	collectionNameOrId string,
	key string,
	value any,
	excludeIds ...string,
) bool {
	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return false
	}

	var expr dbx.Expression
	if collection.IsUsers() && key == collectionmodels.FieldNameUsername {
		expr = dbx.NewExp("LOWER([["+collectionmodels.FieldNameUsername+"]])={:username}", dbx.Params{
			"username": strings.ToLower(cast.ToString(value)),
		})
	} else {
		var normalizedVal any
		switch val := value.(type) {
		case []string:
			normalizedVal = append(types.JsonArray[string]{}, val...)
		case []any:
			normalizedVal = append(types.JsonArray[any]{}, val...)
		default:
			normalizedVal = val
		}

		expr = dbx.HashExp{inflector.Columnify(key): normalizedVal}
	}

	query := dao.RecordQuery(collection).
		Select("count(*)").
		AndWhere(expr).
		Limit(1)

	if uniqueExcludeIds := list.NonzeroUniques(excludeIds); len(uniqueExcludeIds) > 0 {
		query.AndWhere(dbx.NotIn(collection.Name+".id", list.ToInterfaceSlice(uniqueExcludeIds)...))
	}

	var exists bool

	return query.Row(&exists) == nil && !exists
}

// FindUserRecordByToken finds the user record associated with the provided JWT.
//
// Returns an error if the JWT is invalid, expired or not associated to a user record from a users collection.
func (dao *Dao) FindUserRecordByToken(token string, baseTokenKey string) (*recordmodels.Record, error) {
	unverifiedClaims, err := security.ParseUnverifiedJWT(token)
	if err != nil {
		return nil, err
	}

	// check required claims
	id, _ := unverifiedClaims["id"].(string)
	collectionId, _ := unverifiedClaims["collectionId"].(string)
	if id == "" || collectionId == "" {
		return nil, errors.New("missing or invalid token claims")
	}

	record, err := dao.FindRecordById(collectionId, id)
	if err != nil {
		return nil, err
	}

	if !record.Collection().IsUsers() {
		return nil, errors.New("the token is not associated to a user record from a users collection")
	}

	verificationKey := record.TokenKey() + baseTokenKey

	// verify token signature
	if _, err := security.ParseJWT(token, verificationKey); err != nil {
		return nil, err
	}

	return record, nil
}

// FindUserRecordByEmail finds the user record associated with the provided email.
//
// Returns an error if it is not a users collection or the user record is not found.
func (dao *Dao) FindUserRecordByEmail(collectionNameOrId string, email string) (*recordmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users collection %q (%w)", collectionNameOrId, err)
	}
	if !collection.IsUsers() {
		return nil, fmt.Errorf("%q is not a users collection", collectionNameOrId)
	}

	record := &recordmodels.Record{}

	err = dao.RecordQuery(collection).
		AndWhere(dbx.HashExp{collectionmodels.FieldNameEmail: email}).
		Limit(1).
		One(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// FindUserRecordByUsername finds the user record associated with the provided username (case insensitive).
//
// Returns an error if it is not a users collection or the user record is not found.
func (dao *Dao) FindUserRecordByUsername(collectionNameOrId string, username string) (*recordmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users collection %q (%w)", collectionNameOrId, err)
	}
	if !collection.IsUsers() {
		return nil, fmt.Errorf("%q is not a users collection", collectionNameOrId)
	}

	record := &recordmodels.Record{}

	err = dao.RecordQuery(collection).
		AndWhere(dbx.NewExp("LOWER([["+collectionmodels.FieldNameUsername+"]])={:username}", dbx.Params{
			"username": strings.ToLower(username),
		})).
		Limit(1).
		One(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// SuggestUniqueUserRecordUsername checks if the provided username is unique
// and return a new "unique" username with appended random numeric part
// (eg. "existingName" -> "existingName583").
//
// The same username will be returned if the provided string is already unique.
func (dao *Dao) SuggestUniqueUserRecordUsername(
	collectionNameOrId string,
	baseUsername string,
	excludeIds ...string,
) string {
	username := baseUsername

	for i := 0; i < 10; i++ { // max 10 attempts
		isUnique := dao.IsRecordValueUnique(
			collectionNameOrId,
			collectionmodels.FieldNameUsername,
			username,
			excludeIds...,
		)
		if isUnique {
			break // already unique
		}
		username = baseUsername + security.RandomStringWithAlphabet(3+i, "123456789")
	}

	return username
}

// CanAccessRecord checks if a record is allowed to be accessed by the
// specified requestInfo and accessRule.
//
// Rule and db checks are ignored in case requestInfo has admin access.
//
// The returned error indicate that something unexpected happened during
// the check (eg. invalid rule or db error).
//
// The method always return false on invalid access rule or db error.
//
// Example:
//
//	requestInfo := apis.RequestInfo(c /* echo.Context */)
//	record, _ := dao.FindRecordById("example", "RECORD_ID")
//	rule := types.Pointer("@request.auth.id != '' || status = 'public'")
//	// ... or use one of the record collection's rule, eg. record.Collection().ViewRule
//
//	if ok, _ := dao.CanAccessRecord(record, requestInfo, rule); ok { ... }
func (dao *Dao) CanAccessRecord(record *recordmodels.Record, requestInfo *recordmodels.RequestInfo, accessRule *string) (bool, error) {
	if requestInfo != nil && requestInfo.AdminTeamAccess {
		// records with admin team access can access everything
		return true, nil
	}

	if accessRule == nil {
		// only admins can access this record
		return false, nil
	}

	if *accessRule == "" {
		// empty public rule, aka. everyone can access
		return true, nil
	}

	var exists bool

	query := dao.RecordQuery(record.Collection()).
		Select("(1)").
		AndWhere(dbx.HashExp{record.Collection().Name + ".id": record.Id})

	// parse and apply the access rule filter
	resolver := resolvers.NewRecordFieldResolver(dao, record.Collection(), requestInfo, true)
	expr, err := search.FilterData(*accessRule).BuildExpr(resolver)
	if err != nil {
		return false, err
	}
	resolver.UpdateQuery(query)
	query.AndWhere(expr)

	if err := query.Limit(1).Row(&exists); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return exists, nil
}

// SaveRecord persists the provided Record model in the database.
//
// If record.IsNew() is true, the method will perform a create, otherwise an update.
// To explicitly mark a record for update you can use record.MarkAsNotNew().
func (dao *Dao) SaveRecord(record *recordmodels.Record) error {
	var needToDeleteExternalAuths bool

	if record.Collection().IsUsers() {
		if record.Username() == "" {
			return errors.New("unable to save user record without username")
		}

		// Cross-check that the user record id is unique across all users collections.
		// This is to make sure that the filter `@request.auth.id` always returns a unique id.
		authCollections, err := dao.FindCollectionsByType(collectionmodels.CollectionTypeUsers)
		if err != nil {
			return fmt.Errorf("unable to fetch users collections for the cross-id unique check: %w", err)
		}
		for _, collection := range authCollections {
			if record.Collection().Id == collection.Id {
				continue // skip current collection (sqlite will do the check for us)
			}
			isUnique := dao.IsRecordValueUnique(collection.Id, collectionmodels.FieldNameId, record.Id)
			if !isUnique {
				return errors.New("the user record ID must be unique across all users collections")
			}
		}

		if !record.IsNew() && record.Verified() {
			lastSavedRecord, err := dao.FindRecordById(record.Collection().Id, record.Id)
			if err != nil {
				return err
			}

			// in case upgrading from "unverified" -> "verified" mark all pre-existing OAuth2 links
			// for deletion since there is no reliable way to verify that they weren't created by an attacker
			if !lastSavedRecord.Verified() && record.Verified() {
				needToDeleteExternalAuths = true
			}
		}
	}

	if !needToDeleteExternalAuths {
		return dao.Save(record)
	}

	return dao.RunInTransaction(func(txDao *Dao) error {
		externalAuths, err := txDao.FindAllExternalAuthsByRecord(record)
		if err != nil {
			return err
		}
		if len(externalAuths) > 0 {
			// delete all pre-existing external auths
			for _, ea := range externalAuths {
				if err := txDao.DeleteExternalAuth(ea); err != nil {
					return err
				}
			}

			// force refresh tokens reset (if not already)
			record.RefreshTokenKey()
		}

		return txDao.Save(record)
	})
}

// DeleteRecord deletes the provided Record model.
//
// This method will also cascade the delete operation to all linked
// relational records (delete or unset, depending on the rel settings).
//
// The delete operation may fail if the record is part of a required
// reference in another record (aka. cannot be deleted or unset).
func (dao *Dao) DeleteRecord(record *recordmodels.Record) error {
	// fetch rel references (if any)
	//
	// note: the select is outside of the transaction to minimize
	// SQLITE_BUSY errors when mixing read&write in a single transaction
	refs, err := dao.FindCollectionReferences(record.Collection())
	if err != nil {
		return err
	}

	return dao.RunInTransaction(func(txDao *Dao) error {
		// manually trigger delete on any linked external auth to ensure
		// that the `OnModel*` hooks are triggered
		if record.Collection().IsUsers() {
			// note: the select is outside of the transaction to minimize
			// SQLITE_BUSY errors when mixing read&write in a single transaction
			externalAuths, err := dao.FindAllExternalAuthsByRecord(record)
			if err != nil {
				return err
			}
			for _, auth := range externalAuths {
				if err := txDao.DeleteExternalAuth(auth); err != nil {
					return err
				}
			}
		}

		// delete the record before the relation references to ensure that there
		// will be no "A<->B" relations to prevent deadlock when calling DeleteRecord recursively
		if err := txDao.Delete(record); err != nil {
			return err
		}

		return txDao.cascadeRecordDelete(record, refs)
	})
}

// cascadeRecordDelete triggers cascade deletion for the provided references.
//
// NB! This method is expected to be called inside a transaction.
func (dao *Dao) cascadeRecordDelete(mainRecord *recordmodels.Record, refs map[*collectionmodels.Collection][]*collectionmodels.SchemaField) error {
	// @todo consider changing refs to a slice
	//
	// Sort the refs keys to ensure that the cascade events firing order is always the same.
	// This is not necessary for the operation to function correctly but it helps having deterministic output during testing.
	sortedRefKeys := make([]*collectionmodels.Collection, 0, len(refs))
	for k := range refs {
		sortedRefKeys = append(sortedRefKeys, k)
	}
	sort.Slice(sortedRefKeys, func(i, j int) bool {
		return sortedRefKeys[i].Name < sortedRefKeys[j].Name
	})

	for _, refCollection := range sortedRefKeys {
		fields, ok := refs[refCollection]

		if refCollection.IsView() || !ok {
			continue // skip missing or view collections
		}

		for _, field := range fields {
			recordTableName := inflector.Columnify(refCollection.Name)
			prefixedFieldName := recordTableName + "." + inflector.Columnify(field.Name)

			query := dao.RecordQuery(refCollection)

			if opt, ok := field.Options.(collectionmodels.MultiValuer); !ok || !opt.IsMultiple() {
				query.AndWhere(dbx.HashExp{prefixedFieldName: mainRecord.Id})
			} else {
				query.AndWhere(dbx.Exists(dbx.NewExp(fmt.Sprintf(
					`SELECT 1 FROM json_each(CASE WHEN json_valid([[%s]]) THEN [[%s]] ELSE json_array([[%s]]) END) {{__je__}} WHERE [[__je__.value]]={:jevalue}`,
					prefixedFieldName, prefixedFieldName, prefixedFieldName,
				), dbx.Params{
					"jevalue": mainRecord.Id,
				})))
			}

			if refCollection.Id == mainRecord.Collection().Id {
				query.AndWhere(dbx.Not(dbx.HashExp{recordTableName + ".id": mainRecord.Id}))
			}

			// trigger cascade for each batchSize rel items until there is none
			batchSize := 4000
			rows := make([]dbx.NullStringMap, 0, batchSize)
			for {
				if err := query.Limit(int64(batchSize)).All(&rows); err != nil {
					return err
				}

				total := len(rows)
				if total == 0 {
					break
				}

				refRecords := recordmodels.NewRecordsFromNullStringMaps(refCollection, rows)

				err := dao.deleteRefRecords(mainRecord, refRecords, field)
				if err != nil {
					return err
				}

				if total < batchSize {
					break // no more items
				}

				rows = rows[:0] // keep allocated memory
			}
		}
	}

	return nil
}

// deleteRefRecords checks if related records has to be deleted (if `CascadeDelete` is set)
// OR
// just unset the record id from any relation field values (if they are not required).
//
// NB! This method is expected to be called inside a transaction.
func (dao *Dao) deleteRefRecords(mainRecord *recordmodels.Record, refRecords []*recordmodels.Record, field *collectionmodels.SchemaField) error {
	options, _ := field.Options.(*collectionmodels.RelationOptions)
	if options == nil {
		return errors.New("relation field options are not initialized")
	}

	for _, refRecord := range refRecords {
		ids := refRecord.GetStringSlice(field.Name)

		// unset the record id
		for i := len(ids) - 1; i >= 0; i-- {
			if ids[i] == mainRecord.Id {
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
		}

		// cascade delete the reference
		// (only if there are no other active references in case of multiple select)
		if options.CascadeDelete && len(ids) == 0 {
			if err := dao.DeleteRecord(refRecord); err != nil {
				return err
			}
			// no further actions are needed (the reference is deleted)
			continue
		}

		if field.Required && len(ids) == 0 {
			return fmt.Errorf("the record cannot be deleted because it is part of a required reference in record %s (%s collection)", refRecord.Id, refRecord.Collection().Name)
		}

		// save the reference changes
		refRecord.Set(field.Name, field.PrepareValue(ids))
		if err := dao.SaveRecord(refRecord); err != nil {
			return err
		}
	}

	return nil
}
