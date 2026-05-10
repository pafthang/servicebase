package team

import (
	"errors"

	basemodels "github.com/pafthang/servicebase/services/base/models"
	teammodels "github.com/pafthang/servicebase/services/team/models"

	recordmodels "github.com/pafthang/servicebase/services/record/models"

	teamqueries "github.com/pafthang/servicebase/services/team/queries"
)

func (s *Service) FindAdminAccessTeam() (*teammodels.Team, error) {
	return teamqueries.FindAdminAccessTeam(s.Dao())
}

func (s *Service) EnsureAdminAccessTeam() (*teammodels.Team, error) {
	adminTeam, err := s.FindAdminAccessTeam()
	if err == nil && adminTeam != nil {
		return adminTeam, nil
	}

	adminTeam = &teammodels.Team{Name: AdminTeamName}
	if err := teamqueries.SaveTeam(s.Dao(), adminTeam); err != nil {
		return nil, err
	}

	return adminTeam, nil
}

func (s *Service) HasAnyAdminTeamAccess() (bool, error) {
	adminTeam, err := s.FindAdminAccessTeam()
	if err != nil || adminTeam == nil {
		return false, err
	}

	members, err := teamqueries.FindTeamMembersByTeamID(s.Dao(), adminTeam.Id)
	if err != nil {
		return false, err
	}

	return len(members) > 0, nil
}

func (s *Service) HasAdminTeamAccess(model basemodels.Model) bool {
	record, ok := model.(*recordmodels.Record)
	if !ok {
		return false
	}

	return s.IsAdminTeamMember(record)
}

func (s *Service) IsAdminTeamMember(record *recordmodels.Record) bool {
	if record == nil || record.Collection() == nil || !record.Collection().IsUsers() {
		return false
	}

	adminTeam, err := s.FindAdminAccessTeam()
	if err != nil || adminTeam == nil {
		return false
	}

	member, err := teamqueries.FindTeamMemberByTeamAndUser(s.Dao(), adminTeam.Id, record.Id, record.Collection().Id)
	if teamqueries.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}

	return member != nil
}

func (s *Service) GrantAdminTeamAccess(record *recordmodels.Record) error {
	if record == nil || record.Collection() == nil || !record.Collection().IsUsers() {
		return errors.New("admin team access can be granted only to users records")
	}

	adminTeam, err := s.EnsureAdminAccessTeam()
	if err != nil {
		return err
	}

	if s.IsAdminTeamMember(record) {
		return nil
	}

	member := &teammodels.TeamMember{
		Team:             adminTeam.Id,
		UserID:           record.Id,
		UserCollectionID: record.Collection().Id,
	}

	return teamqueries.SaveTeamMember(s.Dao(), member)
}

func (s *Service) RevokeAdminTeamAccess(record *recordmodels.Record) error {
	if record == nil || record.Collection() == nil || !record.Collection().IsUsers() {
		return errors.New("admin team access can be revoked only from users records")
	}

	adminTeam, err := s.FindAdminAccessTeam()
	if err != nil || adminTeam == nil {
		return err
	}

	member, err := teamqueries.FindTeamMemberByTeamAndUser(s.Dao(), adminTeam.Id, record.Id, record.Collection().Id)
	if teamqueries.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	return teamqueries.DeleteTeamMember(s.Dao(), member)
}
