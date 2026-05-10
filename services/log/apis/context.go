package apis

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	logmodels "github.com/pafthang/servicebase/services/log/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pocketbase/dbx"
)

const (
	contextAuthRecordKey         = "authRecord"
	contextProjectLoggingProject = "projectLoggingProject"
)

func currentProjectUserID(c echo.Context) (string, bool) {
	record, _ := c.Get(contextAuthRecordKey).(*recordmodels.Record)
	if record == nil || record.Id == "" {
		return "", false
	}
	return record.Id, true
}

func currentProjectLoggingProject(c echo.Context) (*logmodels.LoggingProject, bool) {
	project, _ := c.Get(contextProjectLoggingProject).(*logmodels.LoggingProject)
	return project, project != nil && project.Id != ""
}

func requireProjectLoggingAPIKey(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenValue := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
			if tokenValue == "" {
				return httpError(401, "API key required.", nil)
			}

			devToken, err := findActiveDevTokenByPlaintext(app, tokenValue)
			if err != nil {
				return httpError(401, "Invalid API key.", err)
			}

			project, err := findActiveLoggingProjectByDevToken(app, devToken.Id)
			if err != nil {
				return httpError(403, "No active logging project found for API key.", err)
			}

			c.Set(contextProjectLoggingProject, project)
			return next(c)
		}
	}
}

func generateSecureProjectToken(prefix string) (string, error) {
	randomBytes := make([]byte, 48)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(randomBytes)
	if prefix != "" {
		return prefix + "_" + token, nil
	}
	return token, nil
}

func findActiveDevTokenByPlaintext(app core.App, plaintext string) (*basemodels.DevToken, error) {
	var items []*basemodels.DevToken
	if err := app.Dao().ModelQuery(&basemodels.DevToken{}).Where(dbx.HashExp{"is_active": true}).All(&items); err != nil {
		return nil, err
	}

	encryptionKey := os.Getenv(app.EncryptionEnv())
	for _, item := range items {
		if item.Token == plaintext {
			return item, nil
		}
		if encryptionKey == "" {
			continue
		}
		decrypted, err := security.Decrypt(item.Token, encryptionKey)
		if err == nil && string(decrypted) == plaintext {
			return item, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}

func findActiveLoggingProjectByDevToken(app core.App, devTokenID string) (*logmodels.LoggingProject, error) {
	item := &logmodels.LoggingProject{}
	err := app.Dao().ModelQuery(item).Where(dbx.HashExp{"dev_token": devTokenID, "active": true}).Limit(1).One(item)
	return item, err
}
