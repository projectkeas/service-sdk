package http

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/projectkeas/sdks-service/healthchecks"
	"github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"
)

const defaultRequestTimeout = 5 * time.Second

type HttpHealthCheck struct {
	url            string
	requestTimeout time.Duration
}

func NewHttpHealthCheck(url string) HttpHealthCheck {
	return HttpHealthCheck{
		url: url,
	}
}

func (healthCheck HttpHealthCheck) Check() healthchecks.HealthCheckResult {

	state := healthchecks.HealthCheckState_Unknown
	duration := 0 * time.Millisecond
	status := 0

	timeout := defaultRequestTimeout
	if healthCheck.requestTimeout > 0 {
		timeout = healthCheck.requestTimeout
	}

	client := http.Client{
		Timeout: timeout,
	}

	url, err := url.Parse(healthCheck.url)
	if err == nil {
		request := &http.Request{
			Method: "GET",
			URL:    url,
			Close:  true,
		}

		startTime := time.Now()
		response, err2 := client.Do(request)
		duration = time.Since(startTime)
		defer func() {
			if err2 == nil {
				response.Body.Close()
			}
		}()

		if err2 == nil {
			status = response.StatusCode
			if response.StatusCode >= http.StatusInternalServerError {
				state = healthchecks.HealthCheckState_Unhealthy
			} else if response.StatusCode == http.StatusTooManyRequests {
				state = healthchecks.HealthCheckState_Degraded
			} else {
				state = healthchecks.HealthCheckState_Healthy
			}
		} else {
			err = err2
		}
	}

	if err != nil {
		logger.Logger.Error("Error performing HTTP Health Check",
			zap.String("error", err.Error()),
			zap.String("url", healthCheck.url),
		)
		state = healthchecks.HealthCheckState_Unhealthy
	}

	return healthchecks.HealthCheckResult{
		Name:     "HTTP Health Check",
		State:    state,
		Duration: healthchecks.NewJsonTime(duration),
		Data: map[string]string{
			"URL":        healthCheck.url,
			"StatusCode": fmt.Sprintf("%d", status),
		},
	}
}
