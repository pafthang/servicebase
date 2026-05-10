package models

var (
	_ Model = (*Team)(nil)
	_ Model = (*TeamMember)(nil)
)

// Team is the typed model for the system "teams" collection.
//
// At the moment teams are still managed mainly through Record-based flows,
// but having an explicit model gives us a stable domain type to migrate to.
type Team struct {
	BaseModel

	Name string `db:"name" json:"name"`
}

// TableName returns the Team model SQL table name.
func (m *Team) TableName() string {
	return "teams"
}

// TeamMember is the typed model for the system "team_members" collection.
type TeamMember struct {
	BaseModel

	Team             string `db:"team" json:"team"`
	UserID           string `db:"userId" json:"userId"`
	UserCollectionID string `db:"userCollectionId" json:"userCollectionId"`
}

// TableName returns the TeamMember model SQL table name.
func (m *TeamMember) TableName() string {
	return "team_members"
}
