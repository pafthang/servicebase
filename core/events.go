package core

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/daos"
	usermodels "github.com/pafthang/servicebase/services/user/models"

	basemodels "github.com/pafthang/servicebase/services/base/models"

	recordmodels "github.com/pafthang/servicebase/services/record/models"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
	"github.com/pafthang/servicebase/tools/auth"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/hook"
	"github.com/pafthang/servicebase/tools/mailer"
	"github.com/pafthang/servicebase/tools/search"
	"github.com/pafthang/servicebase/tools/subscriptions"
	"golang.org/x/crypto/acme/autocert"
)

var (
	_ hook.Tagger = (*BaseModelEvent)(nil)
	_ hook.Tagger = (*BaseCollectionEvent)(nil)
)

type BaseModelEvent struct {
	Model basemodels.Model
}

func (e *BaseModelEvent) Tags() []string {
	if e.Model == nil {
		return nil
	}

	if r, ok := e.Model.(*recordmodels.Record); ok && r.Collection() != nil {
		return []string{r.Collection().Id, r.Collection().Name}
	}

	return []string{e.Model.TableName()}
}

type BaseCollectionEvent struct {
	Collection *collectionmodels.Collection
}

func (e *BaseCollectionEvent) Tags() []string {
	if e.Collection == nil {
		return nil
	}

	tags := make([]string, 0, 2)

	if e.Collection.Id != "" {
		tags = append(tags, e.Collection.Id)
	}

	if e.Collection.Name != "" {
		tags = append(tags, e.Collection.Name)
	}

	return tags
}

// -------------------------------------------------------------------
// Serve events data
// -------------------------------------------------------------------

type BootstrapEvent struct {
	App App
}

type TerminateEvent struct {
	App       App
	IsRestart bool
}

type ServeEvent struct {
	App         App
	Router      *echo.Echo
	Server      *http.Server
	CertManager *autocert.Manager
}

type ApiErrorEvent struct {
	HttpContext echo.Context
	Error       error
}

// -------------------------------------------------------------------
// Model DAO events data
// -------------------------------------------------------------------

type ModelEvent struct {
	BaseModelEvent

	Dao *daos.Dao
}

// -------------------------------------------------------------------
// Mailer events data
// -------------------------------------------------------------------

type MailerRecordEvent struct {
	BaseCollectionEvent

	MailClient mailer.Mailer
	Message    *mailer.Message
	Record     *recordmodels.Record
	Meta       map[string]any
}

// -------------------------------------------------------------------
// Realtime API events data
// -------------------------------------------------------------------

type RealtimeConnectEvent struct {
	HttpContext echo.Context
	Client      subscriptions.Client
	IdleTimeout time.Duration
}

type RealtimeDisconnectEvent struct {
	HttpContext echo.Context
	Client      subscriptions.Client
}

type RealtimeMessageEvent struct {
	HttpContext echo.Context
	Client      subscriptions.Client
	Message     *subscriptions.Message
}

type RealtimeSubscribeEvent struct {
	HttpContext   echo.Context
	Client        subscriptions.Client
	Subscriptions []string
}

// -------------------------------------------------------------------
// Settings API events data
// -------------------------------------------------------------------

type SettingsListEvent struct {
	HttpContext      echo.Context
	RedactedSettings *settingsmodels.Settings
}

type SettingsUpdateEvent struct {
	HttpContext echo.Context
	OldSettings *settingsmodels.Settings
	NewSettings *settingsmodels.Settings
}

// -------------------------------------------------------------------
// Record CRUD API events data
// -------------------------------------------------------------------

type RecordsListEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Records     []*recordmodels.Record
	Result      *search.Result
}

type RecordViewEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordCreateEvent struct {
	BaseCollectionEvent

	HttpContext   echo.Context
	Record        *recordmodels.Record
	UploadedFiles map[string][]*filesystem.File
}

type RecordUpdateEvent struct {
	BaseCollectionEvent

	HttpContext   echo.Context
	Record        *recordmodels.Record
	UploadedFiles map[string][]*filesystem.File
}

type RecordDeleteEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

// -------------------------------------------------------------------
// Auth Record API events data
// -------------------------------------------------------------------

type RecordAuthEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
	Token       string
	Meta        any
}

type RecordAuthWithPasswordEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
	Identity    string
	Password    string
}

type RecordAuthWithOAuth2Event struct {
	BaseCollectionEvent

	HttpContext    echo.Context
	ProviderName   string
	ProviderClient auth.Provider
	Record         *recordmodels.Record
	OAuth2User     *auth.AuthUser
	IsNewRecord    bool
}

type RecordAuthRefreshEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordRequestPasswordResetEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordConfirmPasswordResetEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordRequestVerificationEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordConfirmVerificationEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordRequestEmailChangeEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordConfirmEmailChangeEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
}

type RecordListExternalAuthsEvent struct {
	BaseCollectionEvent

	HttpContext   echo.Context
	Record        *recordmodels.Record
	ExternalAuths []*usermodels.ExternalAuth
}

type RecordUnlinkExternalAuthEvent struct {
	BaseCollectionEvent

	HttpContext  echo.Context
	Record       *recordmodels.Record
	ExternalAuth *usermodels.ExternalAuth
}

// -------------------------------------------------------------------
// Collection API events data
// -------------------------------------------------------------------

type CollectionsListEvent struct {
	HttpContext echo.Context
	Collections []*collectionmodels.Collection
	Result      *search.Result
}

type CollectionViewEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
}

type CollectionCreateEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
}

type CollectionUpdateEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
}

type CollectionDeleteEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
}

type CollectionsImportEvent struct {
	HttpContext echo.Context
	Collections []*collectionmodels.Collection
}

// -------------------------------------------------------------------
// File API events data
// -------------------------------------------------------------------

type FileTokenEvent struct {
	BaseModelEvent

	HttpContext echo.Context
	Token       string
}

type FileDownloadEvent struct {
	BaseCollectionEvent

	HttpContext echo.Context
	Record      *recordmodels.Record
	FileField   *collectionmodels.SchemaField
	ServedPath  string
	ServedName  string
}
