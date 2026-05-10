package daos

import (
	"errors"

	recordmodels "github.com/pafthang/servicebase/services/record/models"
	usermodels "github.com/pafthang/servicebase/services/user/models"
	"github.com/pocketbase/dbx"
)

// ExternalAuthQuery returns a new ExternalAuth select query.
func (dao *Dao) ExternalAuthQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&usermodels.ExternalAuth{})
}

// FindAllExternalAuthsByRecord returns all ExternalAuth models
// linked to the provided user record.
func (dao *Dao) FindAllExternalAuthsByRecord(userRecord *recordmodels.Record) ([]*usermodels.ExternalAuth, error) {
	auths := []*usermodels.ExternalAuth{}

	err := dao.ExternalAuthQuery().
		AndWhere(dbx.HashExp{
			"recordId": userRecord.Id,
		}).
		OrderBy("created ASC").
		All(&auths)

	if err != nil {
		return nil, err
	}

	return auths, nil
}

// FindExternalAuthByRecordAndProvider returns the first available
// ExternalAuth model for the specified user record and provider.
func (dao *Dao) FindExternalAuthByRecordAndProvider(userRecord *recordmodels.Record, provider string) (*usermodels.ExternalAuth, error) {
	model := &usermodels.ExternalAuth{}

	err := dao.ExternalAuthQuery().
		AndWhere(dbx.HashExp{
			"recordId": userRecord.Id,
			"provider": provider,
		}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}

// FindFirstExternalAuthByExpr returns the first available
// ExternalAuth model that satisfies the non-nil expression.
func (dao *Dao) FindFirstExternalAuthByExpr(expr dbx.Expression) (*usermodels.ExternalAuth, error) {
	model := &usermodels.ExternalAuth{}

	err := dao.ExternalAuthQuery().
		AndWhere(dbx.Not(dbx.HashExp{"providerId": ""})). // exclude empty providerIds
		AndWhere(expr).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}

// SaveExternalAuth upserts the provided ExternalAuth model.
func (dao *Dao) SaveExternalAuth(model *usermodels.ExternalAuth) error {
	// extra check the model data in case the provider's API response
	// has changed and no longer returns the expected fields
	if model.UserID == "" || model.Provider == "" || model.ProviderID == "" {
		return errors.New("Missing required ExternalAuth fields.")
	}

	return dao.Save(model)
}

// DeleteExternalAuth deletes the provided ExternalAuth model.
func (dao *Dao) DeleteExternalAuth(model *usermodels.ExternalAuth) error {
	return dao.Delete(model)
}
