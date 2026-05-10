package record

import (
	recordforms "github.com/pafthang/servicebase/services/record/forms"
	"github.com/pafthang/servicebase/services/record/models"
	recordqueries "github.com/pafthang/servicebase/services/record/queries"
	"github.com/pocketbase/dbx"
)

func (s *Service) FindByID(collectionID, recordID string, ruleFuncs ...func(q *dbx.SelectQuery) error) (*models.Record, error) {
	return recordqueries.FindByID(s.Dao(), collectionID, recordID, ruleFuncs...)
}

func (s *Service) NewUpsertForm(record *models.Record) *recordforms.RecordUpsert {
	return recordforms.NewRecordUpsert(s.App(), record)
}

func (s *Service) Delete(record *models.Record) error {
	return recordqueries.Delete(s.Dao(), record)
}
