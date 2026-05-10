package apis

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	realtimeservice "github.com/pafthang/servicebase/services/realtime"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/resolvers"
	"github.com/pafthang/servicebase/tools/rest"
	"github.com/pafthang/servicebase/tools/routine"
	"github.com/pafthang/servicebase/tools/search"
	"github.com/pafthang/servicebase/tools/subscriptions"
	"github.com/pocketbase/dbx"
	"github.com/spf13/cast"
)

// bindRealtimeApi registers the realtime api endpoints.
// Bind registers realtime-owned API endpoints and model event hooks.
func Bind(app core.App, rg *echo.Group, service *realtimeservice.Service, activityMiddleware ...echo.MiddlewareFunc) {
	if app == nil || rg == nil {
		return
	}
	if service == nil {
		service = realtimeservice.New(app)
	}

	api := realtimeApi{app: app, service: service}

	subGroup := rg.Group("/realtime")
	subGroup.GET("", api.connect)
	subGroup.POST("", api.setSubscriptions, activityMiddleware...)

	api.bindEvents()
}

type realtimeApi struct {
	app     core.App
	service *realtimeservice.Service
}

func (api *realtimeApi) connect(c echo.Context) error {
	cancelCtx, cancelRequest := context.WithCancel(c.Request().Context())
	defer cancelRequest()
	c.SetRequest(c.Request().Clone(cancelCtx))

	// register new subscription client
	client := subscriptions.NewDefaultClient()
	api.app.SubscriptionsBroker().Register(client)
	defer func() {
		disconnectEvent := &core.RealtimeDisconnectEvent{
			HttpContext: c,
			Client:      client,
		}

		if err := api.app.OnRealtimeDisconnectRequest().Trigger(disconnectEvent); err != nil {
			api.app.Logger().Debug(
				"OnRealtimeDisconnectRequest error",
				slog.String("clientId", client.Id()),
				slog.String("error", err.Error()),
			)
		}

		api.app.SubscriptionsBroker().Unregister(client.Id())
	}()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-store")
	// https://github.com/pocketbase/pocketbase/discussions/480#discussioncomment-3657640
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering
	c.Response().Header().Set("X-Accel-Buffering", "no")

	connectEvent := &core.RealtimeConnectEvent{
		HttpContext: c,
		Client:      client,
		IdleTimeout: 5 * time.Minute,
	}

	if err := api.app.OnRealtimeConnectRequest().Trigger(connectEvent); err != nil {
		return err
	}

	api.app.Logger().Debug("Realtime connection established.", slog.String("clientId", client.Id()))

	// signalize established connection (aka. fire "connect" message)
	connectMsgEvent := &core.RealtimeMessageEvent{
		HttpContext: c,
		Client:      client,
		Message: &subscriptions.Message{
			Name: "PB_CONNECT",
			Data: []byte(`{"clientId":"` + client.Id() + `"}`),
		},
	}
	connectMsgErr := api.app.OnRealtimeBeforeMessageSend().Trigger(connectMsgEvent, func(e *core.RealtimeMessageEvent) error {
		w := e.HttpContext.Response()
		w.Write([]byte("id:" + client.Id() + "\n"))
		w.Write([]byte("event:" + e.Message.Name + "\n"))
		w.Write([]byte("data:"))
		w.Write(e.Message.Data)
		w.Write([]byte("\n\n"))
		w.Flush()
		return api.app.OnRealtimeAfterMessageSend().Trigger(e)
	})
	if connectMsgErr != nil {
		api.app.Logger().Debug(
			"Realtime connection closed (failed to deliver PB_CONNECT)",
			slog.String("clientId", client.Id()),
			slog.String("error", connectMsgErr.Error()),
		)
		return nil
	}

	// start an idle timer to keep track of inactive/forgotten connections
	idleTimeout := connectEvent.IdleTimeout
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case <-idleTimer.C:
			cancelRequest()
		case msg, ok := <-client.Channel():
			if !ok {
				// channel is closed
				api.app.Logger().Debug(
					"Realtime connection closed (closed channel)",
					slog.String("clientId", client.Id()),
				)
				return nil
			}

			msgEvent := &core.RealtimeMessageEvent{
				HttpContext: c,
				Client:      client,
				Message:     &msg,
			}
			msgErr := api.app.OnRealtimeBeforeMessageSend().Trigger(msgEvent, func(e *core.RealtimeMessageEvent) error {
				w := e.HttpContext.Response()
				w.Write([]byte("id:" + e.Client.Id() + "\n"))
				w.Write([]byte("event:" + e.Message.Name + "\n"))
				w.Write([]byte("data:"))
				w.Write(e.Message.Data)
				w.Write([]byte("\n\n"))
				w.Flush()
				return api.app.OnRealtimeAfterMessageSend().Trigger(msgEvent)
			})
			if msgErr != nil {
				api.app.Logger().Debug(
					"Realtime connection closed (failed to deliver message)",
					slog.String("clientId", client.Id()),
					slog.String("error", msgErr.Error()),
				)
				return nil
			}

			idleTimer.Stop()
			idleTimer.Reset(idleTimeout)
		case <-c.Request().Context().Done():
			// connection is closed
			api.app.Logger().Debug(
				"Realtime connection closed (cancelled request)",
				slog.String("clientId", client.Id()),
			)
			return nil
		}
	}
}

// note: in case of reconnect, clients will have to resubmit all subscriptions again
func (api *realtimeApi) setSubscriptions(c echo.Context) error {
	form := api.service.NewSubscribeForm()

	// read request data
	if err := c.Bind(form); err != nil {
		return httpError(http.StatusBadRequest, "", err)
	}

	// validate request data
	if err := form.Validate(); err != nil {
		return httpError(http.StatusBadRequest, "", err)
	}

	// find subscription client
	client, err := api.app.SubscriptionsBroker().ClientById(form.ClientId)
	if err != nil {
		return httpError(http.StatusNotFound, "Missing or invalid client id.", err)
	}

	// check if the previous request was authorized
	oldAuthId := realtimeservice.ExtractAuthID(client)
	newAuthId := realtimeservice.ExtractAuthID(c)
	if oldAuthId != "" && oldAuthId != newAuthId {
		return httpError(http.StatusForbidden, "The current and the previous request authorization don't match.", nil)
	}

	event := &core.RealtimeSubscribeEvent{
		HttpContext:   c,
		Client:        client,
		Subscriptions: form.Subscriptions,
	}

	return api.app.OnRealtimeBeforeSubscribeRequest().Trigger(event, func(e *core.RealtimeSubscribeEvent) error {
		// update auth state
		e.Client.Set(ContextAdminTeamAccessKey, e.HttpContext.Get(ContextAdminTeamAccessKey))
		e.Client.Set(ContextAuthRecordKey, e.HttpContext.Get(ContextAuthRecordKey))

		// unsubscribe from any previous existing subscriptions
		e.Client.Unsubscribe()

		// subscribe to the new subscriptions
		e.Client.Subscribe(e.Subscriptions...)

		api.app.Logger().Debug(
			"Realtime subscriptions updated.",
			slog.String("clientId", e.Client.Id()),
			slog.Any("subscriptions", e.Subscriptions),
		)

		return api.app.OnRealtimeAfterSubscribeRequest().Trigger(event, func(e *core.RealtimeSubscribeEvent) error {
			if e.HttpContext.Response().Committed {
				return nil
			}

			return e.HttpContext.NoContent(http.StatusNoContent)
		})
	})
}

// updateClientsAuthModel updates the existing clients auth model with the new one (matched by ID).
func (api *realtimeApi) updateClientsAuthModel(contextKey string, newModel basemodels.Model) error {
	api.service.UpdateClientsAuthModel(api.app.SubscriptionsBroker().Clients(), contextKey, newModel)
	return nil
}

// unregisterClientsByAuthModel unregister all clients that has the provided auth model.
func (api *realtimeApi) unregisterClientsByAuthModel(contextKey string, model basemodels.Model) error {
	api.service.UnregisterClientsByAuthModel(api.app.SubscriptionsBroker().Clients(), contextKey, model)
	return nil
}

func (api *realtimeApi) bindEvents() {
	// update the clients that has admin or users record association
	api.app.OnModelAfterUpdate().PreAdd(func(e *core.ModelEvent) error {
		if record := api.service.ResolveRecord(e.Model); record != nil && record.Collection().IsUsers() {
			return api.updateClientsAuthModel(ContextAuthRecordKey, record)
		}

		return nil
	})

	// remove the client(s) associated to the deleted admin or users record
	api.app.OnModelAfterDelete().PreAdd(func(e *core.ModelEvent) error {
		if collection := api.service.ResolveRecordCollection(e.Model); collection != nil && collection.IsUsers() {
			return api.unregisterClientsByAuthModel(ContextAuthRecordKey, e.Model)
		}

		return nil
	})

	api.app.OnModelAfterCreate().PreAdd(func(e *core.ModelEvent) error {
		if record := api.service.ResolveRecord(e.Model); record != nil {
			if err := api.broadcastRecord("create", record, false); err != nil {
				api.app.Logger().Debug(
					"Failed to broadcast record create",
					slog.String("id", record.Id),
					slog.String("collectionName", record.Collection().Name),
					slog.String("error", err.Error()),
				)
			}
		}
		return nil
	})

	api.app.OnModelAfterUpdate().PreAdd(func(e *core.ModelEvent) error {
		if record := api.service.ResolveRecord(e.Model); record != nil {
			if err := api.broadcastRecord("update", record, false); err != nil {
				api.app.Logger().Debug(
					"Failed to broadcast record update",
					slog.String("id", record.Id),
					slog.String("collectionName", record.Collection().Name),
					slog.String("error", err.Error()),
				)
			}
		}
		return nil
	})

	api.app.OnModelBeforeDelete().Add(func(e *core.ModelEvent) error {
		if record := api.service.ResolveRecord(e.Model); record != nil {
			if err := api.broadcastRecord("delete", record, true); err != nil {
				api.app.Logger().Debug(
					"Failed to dry cache record delete",
					slog.String("id", record.Id),
					slog.String("collectionName", record.Collection().Name),
					slog.String("error", err.Error()),
				)
			}
		}
		return nil
	})

	api.app.OnModelAfterDelete().Add(func(e *core.ModelEvent) error {
		if record := api.service.ResolveRecord(e.Model); record != nil {
			if err := api.broadcastDryCachedRecord("delete", record); err != nil {
				api.app.Logger().Debug(
					"Failed to broadcast record delete",
					slog.String("id", record.Id),
					slog.String("collectionName", record.Collection().Name),
					slog.String("error", err.Error()),
				)
			}
		}
		return nil
	})
}

// resolveRecord converts *if possible* the provided model interface to a Record.
// This is usually helpful if the provided model is a custom Record model struct.
// recordData represents the broadcasted record subscrition message data.
type recordData struct {
	Record any    `json:"record"` /* map or recordmodels.Record */
	Action string `json:"action"`
}

func (api *realtimeApi) broadcastRecord(action string, record *recordmodels.Record, dryCache bool) error {
	collection := record.Collection()
	if collection == nil {
		return errors.New("[broadcastRecord] Record collection not set")
	}

	clients := api.app.SubscriptionsBroker().Clients()
	if len(clients) == 0 {
		return nil // no subscribers
	}

	subscriptionRuleMap := map[string]*string{
		(collection.Name + "/" + record.Id + "?"): collection.ViewRule,
		(collection.Id + "/" + record.Id + "?"):   collection.ViewRule,
		(collection.Name + "/*?"):                 collection.ListRule,
		(collection.Id + "/*?"):                   collection.ListRule,
		(collection.Id + "?"):                     collection.ListRule,
	}

	dryCacheKey := action + "/" + record.Id

	for _, client := range clients {
		client := client

		// note: not executed concurrently to avoid races and to ensure
		// that the access checks are applied for the current record db state
		for prefix, rule := range subscriptionRuleMap {
			subs := client.Subscriptions(prefix)
			if len(subs) == 0 {
				continue
			}

			for sub, options := range subs {
				// create a clean record copy without expand and unknown fields
				// because we don't know yet which exact fields the client subscription has permissions to access
				cleanRecord := record.CleanCopy()

				// mock request data
				requestInfo := &recordmodels.RequestInfo{
					Context:         recordmodels.RequestInfoContextRealtime,
					Method:          "GET",
					Query:           options.Query,
					Headers:         options.Headers,
					AdminTeamAccess: cast.ToBool(client.Get(ContextAdminTeamAccessKey)),
				}
				requestInfo.AuthRecord, _ = client.Get(ContextAuthRecordKey).(*recordmodels.Record)

				if !api.canAccessRecord(cleanRecord, requestInfo, rule) {
					continue
				}

				rawExpand := cast.ToString(options.Query[expandQueryParam])
				if rawExpand != "" {
					expandErrs := api.app.Dao().ExpandRecord(
						cleanRecord,
						strings.Split(rawExpand, ","),
						nil,
					)
					if len(expandErrs) > 0 {
						api.app.Logger().Debug(
							"[broadcastRecord] expand errors",
							slog.String("id", cleanRecord.Id),
							slog.String("collectionName", cleanRecord.Collection().Name),
							slog.String("sub", sub),
							slog.String("expand", rawExpand),
							slog.Any("errors", expandErrs),
						)
					}
				}
				// ignore the users record email visibility checks
				// for auth owner, admin or manager
				if requestInfo.AdminTeamAccess {
					cleanRecord.IgnoreEmailVisibility(true)
				} else if collection.IsUsers() {
					authId := realtimeservice.ExtractAuthID(client)
					if authId == cleanRecord.Id {
						if api.canAccessRecord(cleanRecord, requestInfo, collection.UserOptions().ManageRule) {
							cleanRecord.IgnoreEmailVisibility(true)
						}
					}
				}

				data := &recordData{
					Action: action,
					Record: cleanRecord,
				}

				// check fields
				rawFields := cast.ToString(options.Query[fieldsQueryParam])
				if rawFields != "" {
					decoded, err := rest.PickFields(cleanRecord, rawFields)
					if err == nil {
						data.Record = decoded
					} else {
						api.app.Logger().Debug(
							"[broadcastRecord] pick fields error",
							slog.String("id", cleanRecord.Id),
							slog.String("collectionName", cleanRecord.Collection().Name),
							slog.String("sub", sub),
							slog.String("fields", rawFields),
							slog.String("error", err.Error()),
						)
					}
				}

				dataBytes, err := json.Marshal(data)
				if err != nil {
					api.app.Logger().Debug(
						"[broadcastRecord] data marshal error",
						slog.String("id", cleanRecord.Id),
						slog.String("collectionName", cleanRecord.Collection().Name),
						slog.String("error", err.Error()),
					)
					continue
				}

				msg := subscriptions.Message{
					Name: sub,
					Data: dataBytes,
				}

				if dryCache {
					messages, ok := client.Get(dryCacheKey).([]subscriptions.Message)
					if !ok {
						messages = []subscriptions.Message{msg}
					} else {
						messages = append(messages, msg)
					}
					client.Set(dryCacheKey, messages)
				} else {
					routine.FireAndForget(func() {
						client.Send(msg)
					})
				}
			}
		}
	}

	return nil
}

// broadcastDryCachedRecord broadcasts all cached record related messages.
func (api *realtimeApi) broadcastDryCachedRecord(action string, record *recordmodels.Record) error {
	key := action + "/" + record.Id

	clients := api.app.SubscriptionsBroker().Clients()

	for _, client := range clients {
		messages, ok := client.Get(key).([]subscriptions.Message)
		if !ok {
			continue
		}

		client.Unset(key)

		client := client

		routine.FireAndForget(func() {
			for _, msg := range messages {
				client.Send(msg)
			}
		})
	}

	return nil
}

// canAccessRecord checks if the subscription client has access to the specified record model.
func (api *realtimeApi) canAccessRecord(
	record *recordmodels.Record,
	requestInfo *recordmodels.RequestInfo,
	accessRule *string,
) bool {
	// check the access rule
	// ---
	if ok, _ := api.app.Dao().CanAccessRecord(record, requestInfo, accessRule); !ok {
		return false
	}

	// check the subscription client-side filter (if any)
	// ---
	filter := cast.ToString(requestInfo.Query[search.FilterQueryParam])
	if filter == "" {
		return true // no further checks needed
	}

	if err := checkForAdminOnlyRuleFields(requestInfo); err != nil {
		return false
	}

	ruleFunc := func(q *dbx.SelectQuery) error {
		resolver := resolvers.NewRecordFieldResolver(api.app.Dao(), record.Collection(), requestInfo, false)

		expr, err := search.FilterData(filter).BuildExpr(resolver)
		if err != nil {
			return err
		}
		q.AndWhere(expr)

		resolver.UpdateQuery(q)

		return nil
	}

	_, err := api.app.Dao().FindRecordById(record.Collection().Id, record.Id, ruleFunc)

	return err == nil
}
