package daos_test

import (
	"testing"

	"github.com/pafthang/servicebase/services/team/models"
	"github.com/pafthang/servicebase/tests"
)

func TestTeamQuery(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	expected := "SELECT {{teams}}.* FROM `teams`"

	sql := app.Dao().TeamQuery().Build().SQL()
	if sql != expected {
		t.Fatalf("Expected sql %s, got %s", expected, sql)
	}
}

func TestTeamMemberQuery(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	expected := "SELECT {{team_members}}.* FROM `team_members`"

	sql := app.Dao().TeamMemberQuery().Build().SQL()
	if sql != expected {
		t.Fatalf("Expected sql %s, got %s", expected, sql)
	}
}

func TestSaveFindAndDeleteTeam(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	team := &models.Team{Name: "ops"}
	if err := app.Dao().SaveTeam(team); err != nil {
		t.Fatal(err)
	}

	foundById, err := app.Dao().FindTeamById(team.Id)
	if err != nil {
		t.Fatal(err)
	}
	if foundById.Name != team.Name {
		t.Fatalf("Expected team name %q, got %q", team.Name, foundById.Name)
	}

	foundByName, err := app.Dao().FindTeamByName(team.Name)
	if err != nil {
		t.Fatal(err)
	}
	if foundByName.Id != team.Id {
		t.Fatalf("Expected team id %q, got %q", team.Id, foundByName.Id)
	}

	if err := app.Dao().DeleteTeam(team); err != nil {
		t.Fatal(err)
	}
}

func TestSaveFindAndDeleteTeamMember(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	team := &models.Team{Name: "ops"}
	if err := app.Dao().SaveTeam(team); err != nil {
		t.Fatal(err)
	}

	member := &models.TeamMember{
		Team:             team.Id,
		UserID:           "user1",
		UserCollectionID: "_pb_users_auth_",
	}
	if err := app.Dao().SaveTeamMember(member); err != nil {
		t.Fatal(err)
	}

	found, err := app.Dao().FindTeamMemberByTeamAndUser(team.Id, member.UserID, member.UserCollectionID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Id != member.Id {
		t.Fatalf("Expected team member id %q, got %q", member.Id, found.Id)
	}

	members, err := app.Dao().FindTeamMembersByTeamId(team.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 1 {
		t.Fatalf("Expected 1 team member, got %d", len(members))
	}

	if err := app.Dao().DeleteTeamMember(member); err != nil {
		t.Fatal(err)
	}
}
