package daos

import (
	"github.com/pafthang/servicebase/services/team/models"

	"github.com/pocketbase/dbx"
)

// TeamQuery returns a new Team select query.
func (dao *Dao) TeamQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&models.Team{})
}

// TeamMemberQuery returns a new TeamMember select query.
func (dao *Dao) TeamMemberQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&models.TeamMember{})
}

// FindTeamById finds a team by its id.
func (dao *Dao) FindTeamById(id string) (*models.Team, error) {
	model := &models.Team{}

	err := dao.TeamQuery().
		AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// FindTeamByName finds a team by its name.
func (dao *Dao) FindTeamByName(name string) (*models.Team, error) {
	model := &models.Team{}

	err := dao.TeamQuery().
		AndWhere(dbx.HashExp{"name": name}).
		Limit(1).
		One(model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// FindTeamMembersByTeamId finds all team members for the specified team id.
func (dao *Dao) FindTeamMembersByTeamId(teamId string) ([]*models.TeamMember, error) {
	modelsList := []*models.TeamMember{}

	err := dao.TeamMemberQuery().
		AndWhere(dbx.HashExp{"team": teamId}).
		OrderBy("created ASC").
		All(&modelsList)
	if err != nil {
		return nil, err
	}

	return modelsList, nil
}

// FindTeamMembersByUser finds all team memberships for the specified user.
func (dao *Dao) FindTeamMembersByUser(userId, userCollectionId string) ([]*models.TeamMember, error) {
	modelsList := []*models.TeamMember{}

	err := dao.TeamMemberQuery().
		AndWhere(dbx.HashExp{
			"userId":           userId,
			"userCollectionId": userCollectionId,
		}).
		OrderBy("created ASC").
		All(&modelsList)
	if err != nil {
		return nil, err
	}

	return modelsList, nil
}

// FindTeamMemberByTeamAndUser finds a single team member by team and user identifiers.
func (dao *Dao) FindTeamMemberByTeamAndUser(teamId, userId, userCollectionId string) (*models.TeamMember, error) {
	model := &models.TeamMember{}

	err := dao.TeamMemberQuery().
		AndWhere(dbx.HashExp{
			"team":             teamId,
			"userId":           userId,
			"userCollectionId": userCollectionId,
		}).
		Limit(1).
		One(model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// SaveTeam upserts the provided Team model.
func (dao *Dao) SaveTeam(team *models.Team) error {
	return dao.Save(team)
}

// DeleteTeam deletes the provided Team model.
func (dao *Dao) DeleteTeam(team *models.Team) error {
	return dao.Delete(team)
}

// SaveTeamMember upserts the provided TeamMember model.
func (dao *Dao) SaveTeamMember(teamMember *models.TeamMember) error {
	return dao.Save(teamMember)
}

// DeleteTeamMember deletes the provided TeamMember model.
func (dao *Dao) DeleteTeamMember(teamMember *models.TeamMember) error {
	return dao.Delete(teamMember)
}
