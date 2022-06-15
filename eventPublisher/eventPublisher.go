package eventPublisher

import (
	"context"
	"encoding/json"

	"github.com/projectkeas/sdks-service/configuration"
	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	log "github.com/projectkeas/sdks-service/logger"
)

const (
	SERVICE_NAME string = "EventPublisher"
)

type EventPublisherService interface {
	Publish(event cloudevents.Event) bool
}

type eventPublisherExecutionService struct {
	client *cloudevents.Client
	config *configuration.ConfigurationRoot
}

func New(config *configuration.ConfigurationRoot) EventPublisherService {
	service := eventPublisherExecutionService{}
	config.RegisterChangeNotificationHandler(func(c configuration.ConfigurationRoot) {
		service.config = &c
		service.client = nil
	})

	return &service
}

func (ep *eventPublisherExecutionService) Publish(event cloudevents.Event) bool {
	err := event.Validate()
	if err != nil {
		log.Logger.Error("Unable to validate outbound CloudEvent", zap.Error(err))
		return false
	}

	ctx := cloudevents.ContextWithTarget(context.Background(), ep.config.GetStringValueOrDefault("ingestion.uri", "http://keas-ingestion.keas.svc.cluster.local/ingest"))
	if ep.client == nil {
		sender, err := cloudevents.NewHTTP(cloudevents.WithHeader("Authorization", "ApiKey "+ep.config.GetStringValueOrDefault("ingestion.auth.token", "")))
		if err != nil {
			log.Logger.Error("Unable to create new HTTP sender", zap.Error(err))
			return false
		}

		client, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			log.Logger.Error("Unable to create new HTTP Client", zap.Error(err))
			return false
		}

		ep.client = &client
	}

	result := (*ep.client).Send(ctx, event)
	ack := cloudevents.IsACK(result)
	if cloudevents.IsUndelivered(result) {
		log.Logger.Error("Unable to send HTTP request", zap.Error(result), zap.Any("event", map[string]interface{}{
			"uuid":         event.ID(),
			"acknowledged": cloudevents.IsACK(result),
		}))
	} else {
		if !ack {
			var httpResult *cehttp.Result
			cloudevents.ResultAs(result, &httpResult)

			if httpResult.StatusCode == 400 {
				res := map[string]interface{}{}
				err := json.Unmarshal([]byte(httpResult.Args[1].(string)), &res)
				if err == nil && res["reason"] == "ingestion-service-rejected" {
					ack = true
				}
			}
		}

		log.Logger.Debug("Sent event", zap.Any("event", map[string]interface{}{
			"uuid":         event.ID(),
			"acknowledged": ack,
		}))
	}

	return ack
}
