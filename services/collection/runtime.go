package collection

import (
	collectionforms "github.com/pafthang/servicebase/services/collection/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	collectionqueries "github.com/pafthang/servicebase/services/collection/queries"
)

func (s *Service) List(query string) (*collectionqueries.ListResult, []*collectionmodels.Collection, error) {
	return collectionqueries.List(s.Dao(), query)
}

func (s *Service) FindByNameOrID(nameOrID string) (*collectionmodels.Collection, error) {
	return collectionqueries.FindByNameOrID(s.Dao(), nameOrID)
}

func (s *Service) NewUpsertForm(collection *collectionmodels.Collection) *collectionforms.Upsert {
	return collectionforms.NewUpsert(s.App(), collection)
}

func (s *Service) SubmitUpsert(
	form *collectionforms.Upsert,
	interceptors ...collectionforms.UpsertInterceptorFunc,
) error {
	return form.Submit(interceptors...)
}

func (s *Service) Delete(collection *collectionmodels.Collection) error {
	return collectionqueries.Delete(s.Dao(), collection)
}

func (s *Service) NewImportForm() *collectionforms.Import {
	return collectionforms.NewImport(s.App())
}

func (s *Service) SubmitImport(
	form *collectionforms.Import,
	interceptors ...collectionforms.ImportInterceptorFunc,
) error {
	return form.Submit(interceptors...)
}
