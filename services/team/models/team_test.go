package models_test

import (
	"testing"

	"github.com/pafthang/servicebase/services/team/models"
)

func TestTeamTableName(t *testing.T) {
	t.Parallel()

	m := models.Team{}
	if m.TableName() != "teams" {
		t.Fatalf("Unexpected table name, got %q", m.TableName())
	}
}

func TestTeamMemberTableName(t *testing.T) {
	t.Parallel()

	m := models.TeamMember{}
	if m.TableName() != "team_members" {
		t.Fatalf("Unexpected table name, got %q", m.TableName())
	}
}
