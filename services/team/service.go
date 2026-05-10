package team

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

const (
	TeamsCollectionName       = "teams"
	TeamMembersCollectionName = "team_members"
	AdminTeamName             = "admin"
)

var Descriptor = servicebase.Descriptor{
	Name:    "team",
	Purpose: "Manages admin team lookup and membership grants for access control.",
	Dependencies: []string{
		"core.App",
	},
	Operations: []string{
		"FindAdminAccessTeam",
		"EnsureAdminAccessTeam",
		"HasAnyAdminTeamAccess",
		"HasAdminTeamAccess",
		"IsAdminTeamMember",
		"GrantAdminTeamAccess",
		"RevokeAdminTeamAccess",
	},
}

type Service struct {
	servicebase.Service
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}
