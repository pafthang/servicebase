package realtime

import (
	"strings"

	basemodels "github.com/pafthang/servicebase/services/base/models"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	realtimeforms "github.com/pafthang/servicebase/services/realtime/forms"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/subscriptions"
)

func (s *Service) NewSubscribeForm() *realtimeforms.RealtimeSubscribe {
	return realtimeforms.NewRealtimeSubscribe()
}

func (s *Service) ResolveRecord(model basemodels.Model) (record *recordmodels.Record) {
	record, _ = model.(*recordmodels.Record)

	if record == nil && !strings.HasPrefix(model.TableName(), "_") {
		record, _ = s.Dao().FindRecordById(model.TableName(), model.GetId())
	}

	return record
}

func (s *Service) ResolveRecordCollection(model basemodels.Model) (collection *collectionmodels.Collection) {
	if record, ok := model.(*recordmodels.Record); ok {
		collection = record.Collection()
	} else if !strings.HasPrefix(model.TableName(), "_") {
		collection, _ = s.Dao().FindCollectionByNameOrId(model.TableName())
	}

	return collection
}

func (s *Service) UpdateClientsAuthModel(clients map[string]subscriptions.Client, contextKey string, newModel basemodels.Model) {
	for _, client := range clients {
		clientModel, _ := client.Get(contextKey).(basemodels.Model)
		if clientModel != nil &&
			clientModel.TableName() == newModel.TableName() &&
			clientModel.GetId() == newModel.GetId() {
			client.Set(contextKey, newModel)
		}
	}
}

func (s *Service) UnregisterClientsByAuthModel(clients map[string]subscriptions.Client, contextKey string, model basemodels.Model) {
	for _, client := range clients {
		clientModel, _ := client.Get(contextKey).(basemodels.Model)
		if clientModel != nil &&
			clientModel.TableName() == model.TableName() &&
			clientModel.GetId() == model.GetId() {
			client.Unset(contextKey)
		}
	}
}

func ExtractAuthID(val Getter) string {
	record, _ := val.Get("authRecord").(*recordmodels.Record)
	if record != nil {
		return record.Id
	}

	return ""
}
