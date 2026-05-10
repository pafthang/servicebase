package mails

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/pafthang/servicebase/core"
	mailmodels "github.com/pafthang/servicebase/services/mails/models"
	"github.com/pafthang/servicebase/tools/types"
	"github.com/pocketbase/dbx"
)

const (
	// OAuth state token for CSRF protection
	oauthStateToken = "state-token"
)

// MailService handles Gmail integration and operations
type MailService struct {
	app          core.App
	googleConfig *oauth2.Config
	config       *MailServiceConfig
}

// NewMailService creates a new MailService instance with configuration from environment variables
func NewMailService() *MailService {
	return NewMailServiceWithApp(nil)
}

func NewMailServiceWithApp(app core.App) *MailService {
	resolved := ResolveConfig(app)
	config := &resolved

	googleConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			gmail.GmailReadonlyScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}

	setMailApp(app)

	return &MailService{
		app:          app,
		googleConfig: googleConfig,
		config:       config,
	}
}

// GetAuthUrl generates an OAuth authorization URL for Gmail authentication
func (ms *MailService) GetAuthUrl() string {
	return ms.googleConfig.AuthCodeURL(
		oauthStateToken,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("redirect_uri", ms.googleConfig.RedirectURL),
	)
}

// GetConfig returns the OAuth2 configuration
func (ms *MailService) GetConfig() *oauth2.Config {
	return ms.googleConfig
}

// tokenSource wraps the OAuth2 token source to save refreshed tokens and handle errors.
type tokenSource struct {
	app          core.App
	ctx          context.Context
	config       *oauth2.Config
	token        *oauth2.Token
	tokenId      string
	baseTokenSrc oauth2.TokenSource
}

func (ts *tokenSource) Token() (*oauth2.Token, error) {
	newToken, err := ts.baseTokenSrc.Token()
	if err != nil {
		errMsg := strings.ToLower(err.Error())

		if strings.Contains(errMsg, "invalid_grant") {
			logger.LogError(fmt.Sprintf("Invalid grant error for token %s - token may be expired or revoked", ts.tokenId), err)

			if updateErr := ts.updateToken(map[string]interface{}{
				"is_active": false,
			}); updateErr != nil {
				logger.LogError(fmt.Sprintf("Failed to mark token %s as inactive", ts.tokenId), updateErr)
			}

			return nil, fmt.Errorf("OAuth token is invalid or expired. Please re-authenticate your Gmail account: %w", err)
		}

		return nil, err
	}

	if newToken.AccessToken != ts.token.AccessToken {
		logger.Debug.Printf("Token refreshed for token ID: %s", ts.tokenId)

		expiry := types.DateTime{}
		expiry.Scan(newToken.Expiry)

		updateData := map[string]interface{}{
			"access_token": newToken.AccessToken,
			"token_type":   newToken.TokenType,
			"expiry":       expiry,
			"last_used":    types.NowDateTime(),
		}

		if newToken.RefreshToken != "" {
			updateData["refresh_token"] = newToken.RefreshToken
		}

		if updateErr := ts.updateToken(updateData); updateErr != nil {
			logger.LogError(fmt.Sprintf("Failed to save refreshed token for token ID: %s", ts.tokenId), updateErr)
		} else {
			ts.token = newToken
		}
	}

	return newToken, nil
}

func (ts *tokenSource) updateToken(data map[string]interface{}) error {
	if ts.app == nil {
		return fmt.Errorf("app is required to update OAuth token")
	}

	token := &mailmodels.Token{}

	_, err := ts.app.Dao().
		NonconcurrentDB().
		Update(token.TableName(), data, dbx.HashExp{"id": ts.tokenId}).
		Execute()

	return err
}

// FetchClient retrieves an authenticated HTTP client for a given token ID.
func (ms *MailService) FetchClient(ctx context.Context, tokenId string) (*http.Client, error) {
	if tokenId == "" {
		return nil, fmt.Errorf("token ID cannot be empty")
	}

	if ms.app == nil {
		return nil, fmt.Errorf("app is required to fetch OAuth client")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	token := &mailmodels.Token{}

	if err := ms.app.Dao().FindById(token, tokenId); err != nil {
		return nil, fmt.Errorf("failed to find token with ID %s: %w", tokenId, err)
	}

	if !token.IsActive {
		return nil, fmt.Errorf("token %s is marked as inactive - please re-authenticate", tokenId)
	}

	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry.Time(),
	}

	logger.Debug.Printf(
		"Retrieved OAuth token for token ID: %s, expires at: %v",
		tokenId,
		oauthToken.Expiry,
	)

	baseTokenSrc := ms.googleConfig.TokenSource(ctx, oauthToken)

	customTokenSrc := &tokenSource{
		app:          ms.app,
		ctx:          ctx,
		config:       ms.googleConfig,
		token:        oauthToken,
		tokenId:      tokenId,
		baseTokenSrc: baseTokenSrc,
	}

	return oauth2.NewClient(ctx, customTokenSrc), nil
}

// GetGmailService creates a Gmail API service client using the provided HTTP client.
func (ms *MailService) GetGmailService(ctx context.Context, client *http.Client) (*gmail.Service, error) {
	if client == nil {
		return nil, fmt.Errorf("HTTP client cannot be nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return service, nil
}
