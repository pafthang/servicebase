package team_test

import (
	"testing"

	teamservice "github.com/pafthang/servicebase/services/team"
	teammodels "github.com/pafthang/servicebase/services/team/models"
	"github.com/pafthang/servicebase/tests"
)

func TestHasAnyAdminTeamAccess(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	service := teamservice.New(app)

	hasAccess, err := service.HasAnyAdminTeamAccess()
	if err != nil {
		t.Fatal(err)
	}
	if hasAccess {
		t.Fatal("expected no admin access without admin team members")
	}

	record, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.GrantAdminTeamAccess(record); err != nil {
		t.Fatal(err)
	}

	hasAccess, err = service.HasAnyAdminTeamAccess()
	if err != nil {
		t.Fatal(err)
	}
	if !hasAccess {
		t.Fatal("expected admin access after adding admin team member")
	}
}

func TestFindAndEnsureAdminAccessTeam(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	service := teamservice.New(app)

	adminTeam, err := service.FindAdminAccessTeam()
	if err != nil {
		t.Fatal(err)
	}
	if adminTeam == nil || adminTeam.Name != teamservice.AdminTeamName {
		t.Fatalf("expected existing %q team, got %#v", teamservice.AdminTeamName, adminTeam)
	}

	ensuredTeam, err := service.EnsureAdminAccessTeam()
	if err != nil {
		t.Fatal(err)
	}
	if ensuredTeam.Id != adminTeam.Id {
		t.Fatalf("expected EnsureAdminTeam to reuse %q team, got %q", adminTeam.Id, ensuredTeam.Id)
	}
}

func TestHasAdminTeamAccess(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	service := teamservice.New(app)

	record, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if service.HasAdminTeamAccess(record) {
		t.Fatal("expected record without team membership to not have admin access")
	}

	if err := service.GrantAdminTeamAccess(record); err != nil {
		t.Fatal(err)
	}
	if !service.HasAdminTeamAccess(record) {
		t.Fatal("expected admin team member record to have admin access")
	}

	if service.HasAdminTeamAccess(&teammodels.Team{}) {
		t.Fatal("expected unrelated model type to not have admin access")
	}
}

func TestGrantAndRevokeAdminTeamAccess(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	service := teamservice.New(app)

	record, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.GrantAdminTeamAccess(record); err != nil {
		t.Fatal(err)
	}

	adminTeam, err := service.FindAdminAccessTeam()
	if err != nil {
		t.Fatal(err)
	}

	member, err := app.Dao().FindTeamMemberByTeamAndUser(adminTeam.Id, record.Id, record.Collection().Id)
	if err != nil {
		t.Fatal(err)
	}
	if member == nil {
		t.Fatal("expected team membership to be created")
	}

	if err := service.RevokeAdminTeamAccess(record); err != nil {
		t.Fatal(err)
	}

	member, err = app.Dao().FindTeamMemberByTeamAndUser(adminTeam.Id, record.Id, record.Collection().Id)
	if member != nil {
		t.Fatalf("expected team membership to be removed, got %#v", member)
	}
}
