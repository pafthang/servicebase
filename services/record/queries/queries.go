package queries

import (
	"github.com/pafthang/servicebase/daos"
	"github.com/pafthang/servicebase/services/record/models"
	"github.com/pocketbase/dbx"
)

func FindByID(
	dao *daos.Dao,
	collectionID string,
	recordID string,
	ruleFuncs ...func(q *dbx.SelectQuery) error,
) (*models.Record, error) {
	return dao.FindRecordById(collectionID, recordID, ruleFuncs...)
}

func Delete(dao *daos.Dao, record *models.Record) error {
	return dao.DeleteRecord(record)
}
