package models

var _ Model = (*ExternalAuth)(nil)

type ExternalAuth struct {
	BaseModel

	UserCollectionID string `db:"collectionId" json:"-"`
	UserID           string `db:"recordId" json:"userId"`
	Provider         string `db:"provider" json:"provider"`
	ProviderID       string `db:"providerId" json:"providerId"`
}

func (m *ExternalAuth) TableName() string {
	return "_externalAuths"
}
