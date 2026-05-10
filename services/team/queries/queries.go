package queries

import (
	"database/sql"
	"errors"

	"github.com/pafthang/servicebase/daos"
	"github.com/pafthang/servicebase/services/team/models"
)

func FindAdminAccessTeam(dao *daos.Dao) (*models.Team, error) {
	team, err := dao.FindTeamByName("admin")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return team, err
}

func FindTeamMembersByTeamID(dao *daos.Dao, teamID string) ([]*models.TeamMember, error) {
	return dao.FindTeamMembersByTeamId(teamID)
}

func FindTeamMemberByTeamAndUser(dao *daos.Dao, teamID, userID, userCollectionID string) (*models.TeamMember, error) {
	return dao.FindTeamMemberByTeamAndUser(teamID, userID, userCollectionID)
}

func SaveTeam(dao *daos.Dao, team *models.Team) error {
	return dao.SaveTeam(team)
}

func SaveTeamMember(dao *daos.Dao, member *models.TeamMember) error {
	return dao.SaveTeamMember(member)
}

func DeleteTeamMember(dao *daos.Dao, member *models.TeamMember) error {
	return dao.DeleteTeamMember(member)
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
