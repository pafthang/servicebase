package queries

import (
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/services/user/models"
	"github.com/pocketbase/dbx"

	"github.com/pafthang/servicebase/daos"
)

func FindRecordByID(dao *daos.Dao, collectionID, id string) (*recordmodels.Record, error) {
	return dao.FindRecordById(collectionID, id)
}

func FindFirstExternalAuthByRecord(
	dao *daos.Dao,
	collectionID string,
	recordID string,
	provider string,
) (*models.ExternalAuth, error) {
	_ = collectionID

	return dao.FindFirstExternalAuthByExpr(dbx.HashExp{
		"recordId": recordID,
		"provider": provider,
	})
}

func FindAllExternalAuthsByRecord(dao *daos.Dao, collectionID, recordID string) ([]*models.ExternalAuth, error) {
	record, err := FindRecordByID(dao, collectionID, recordID)
	if err != nil {
		return nil, err
	}

	return dao.FindAllExternalAuthsByRecord(record)
}

func DeleteExternalAuth(dao *daos.Dao, rel *models.ExternalAuth) error {
	return dao.DeleteExternalAuth(rel)
}
