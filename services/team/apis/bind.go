package apis

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	collectioncache "github.com/pafthang/servicebase/services/cache"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	teamservice "github.com/pafthang/servicebase/services/team"
	teammodels "github.com/pafthang/servicebase/services/team/models"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type MiddlewareByStringFactory func(string) echo.MiddlewareFunc
type ErrorFunc func(message string, rawErr any) error

type BindDeps struct {
	ActivityLogger             MiddlewareFactory
	LoadFixedCollectionContext func(app core.App, collectionName string, collectionType string) echo.MiddlewareFunc
	RequireAdminTeamAccess     func() echo.MiddlewareFunc
	RequireRecordAuth          MiddlewareByStringFactory
	NewUnauthorizedError       ErrorFunc
	NewNotFoundError           ErrorFunc
	NewBadRequestError         ErrorFunc
	ContextAuthRecordKey       string
	ListHandler                echo.HandlerFunc
	CreateHandler              echo.HandlerFunc
	ViewHandler                echo.HandlerFunc
	UpdateHandler              echo.HandlerFunc
	DeleteHandler              echo.HandlerFunc
}

type teamMemberUpsertBody struct {
	UserID string `json:"userId"`
}

type teamMemberDetails struct {
	Membership *teammodels.TeamMember `json:"membership"`
	User       *recordmodels.Record   `json:"user,omitempty"`
}

// Bind registers dedicated team endpoints inside the team service module.
func Bind(app core.App, rg *echo.Group, deps BindDeps) {
	group := rg.Group(
		"/teams",
		deps.ActivityLogger(app),
		deps.LoadFixedCollectionContext(app, teamservice.TeamsCollectionName, collectionmodels.CollectionTypeBase),
	)

	group.GET("", deps.ListHandler, deps.RequireAdminTeamAccess())
	group.POST("", deps.CreateHandler, deps.RequireAdminTeamAccess())
	group.GET("/me", listCurrentUserTeams(app, deps), deps.RequireRecordAuth("users"))
	group.GET("/:id", deps.ViewHandler, deps.RequireAdminTeamAccess())
	group.PATCH("/:id", deps.UpdateHandler, deps.RequireAdminTeamAccess())
	group.DELETE("/:id", deps.DeleteHandler, deps.RequireAdminTeamAccess())
	group.GET("/:id/members", listTeamMembers(app, deps), deps.RequireAdminTeamAccess())
	group.POST("/:id/members", addTeamMember(app, deps), deps.RequireAdminTeamAccess())
	group.DELETE("/:id/members/:userId", removeTeamMember(app, deps), deps.RequireAdminTeamAccess())
}

func listCurrentUserTeams(app core.App, deps BindDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		record, _ := c.Get(deps.ContextAuthRecordKey).(*recordmodels.Record)
		if record == nil {
			return deps.NewUnauthorizedError("The request requires valid record authorization token to be set.", nil)
		}

		usersCollection, err := collectioncache.FindByNameOrId(app, "users")
		if err != nil || usersCollection == nil {
			return deps.NewNotFoundError("", err)
		}

		memberships, err := app.Dao().FindTeamMembersByUser(record.Id, usersCollection.Id)
		if err != nil {
			return deps.NewBadRequestError("Failed to fetch current user team memberships.", err)
		}

		teams := make([]*teammodels.Team, 0, len(memberships))
		for _, membership := range memberships {
			team, err := app.Dao().FindTeamById(membership.Team)
			if err != nil || team == nil {
				continue
			}

			teams = append(teams, team)
		}

		return c.JSON(http.StatusOK, teams)
	}
}

func listTeamMembers(app core.App, deps BindDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		teamID := c.PathParam("id")
		if teamID == "" {
			return deps.NewNotFoundError("", nil)
		}

		if _, err := app.Dao().FindTeamById(teamID); err != nil {
			return deps.NewNotFoundError("", err)
		}

		members, err := app.Dao().FindTeamMembersByTeamId(teamID)
		if err != nil {
			return deps.NewBadRequestError("Failed to fetch team members.", err)
		}

		result := make([]*teamMemberDetails, 0, len(members))
		for _, membership := range members {
			item := &teamMemberDetails{Membership: membership}

			user, err := app.Dao().FindRecordById(membership.UserCollectionID, membership.UserID)
			if err == nil && user != nil {
				item.User = user
			}

			result = append(result, item)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func addTeamMember(app core.App, deps BindDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		teamID := c.PathParam("id")
		if teamID == "" {
			return deps.NewNotFoundError("", nil)
		}

		if _, err := app.Dao().FindTeamById(teamID); err != nil {
			return deps.NewNotFoundError("", err)
		}

		body := &teamMemberUpsertBody{}
		if err := c.Bind(body); err != nil {
			return deps.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
		}

		if body.UserID == "" {
			return deps.NewBadRequestError("Missing required userId.", nil)
		}

		usersCollection, err := collectioncache.FindByNameOrId(app, "users")
		if err != nil || usersCollection == nil {
			return deps.NewNotFoundError("", err)
		}

		user, err := app.Dao().FindRecordById(usersCollection.Id, body.UserID)
		if err != nil || user == nil {
			return deps.NewNotFoundError("User not found.", err)
		}

		member := &teammodels.TeamMember{
			Team:             teamID,
			UserID:           user.Id,
			UserCollectionID: usersCollection.Id,
		}
		if err := app.Dao().SaveTeamMember(member); err != nil {
			return deps.NewBadRequestError("Failed to add team member.", err)
		}

		return c.JSON(http.StatusOK, &teamMemberDetails{
			Membership: member,
			User:       user,
		})
	}
}

func removeTeamMember(app core.App, deps BindDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		teamID := c.PathParam("id")
		userID := c.PathParam("userId")
		if teamID == "" || userID == "" {
			return deps.NewNotFoundError("", nil)
		}

		usersCollection, err := collectioncache.FindByNameOrId(app, "users")
		if err != nil || usersCollection == nil {
			return deps.NewNotFoundError("", err)
		}

		member, err := app.Dao().FindTeamMemberByTeamAndUser(teamID, userID, usersCollection.Id)
		if err != nil || member == nil {
			return deps.NewNotFoundError("", err)
		}

		if err := app.Dao().DeleteTeamMember(member); err != nil {
			return deps.NewBadRequestError("Failed to remove team member.", err)
		}

		return c.NoContent(http.StatusNoContent)
	}
}
