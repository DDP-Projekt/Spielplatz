package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/DDP-Projekt/Spielplatz/server/kddp"
	"github.com/gin-gonic/gin"
)

func healthcheckHandler(c *gin.Context) {
	logger := getLogger(c)

	logger.Debug("starting healthcheck")
	healthcheckResult := performHealthcheck(logger)
	logger.Debug("healthcheck done")

	json_content, err := json.MarshalIndent(healthcheckResult, "", "\t")
	if err != nil {
		logger.Warn("Error marshalling healthcheck result", "err", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	status_code := http.StatusOK
	if !healthcheckResult.Healthy {
		status_code = http.StatusServiceUnavailable
	}
	c.Data(status_code, "application/json", json_content)
}

type KddpHealthcheckResult struct {
	Healthy    bool    `json:"healthy"`
	Error      *string `json:"error,omitempty"`
	Version    *string `json:"version,omitempty"`
	Exitstatus *int    `json:"exit-status,omitempty"`
}

type HealthcheckResult struct {
	Healthy    bool                  `json:"healthy"`
	KddpStatus KddpHealthcheckResult `json:"kddp-status"`
}

func performHealthcheck(logger *slog.Logger) (result HealthcheckResult) {
	result.Healthy = true

	if kddpResult, err := kddp.GetKDDPVersion(); err != nil {
		logger.Error("could not read kddp version", "err", err.Error())
		errstr := err.Error()

		result.Healthy = false
		result.KddpStatus = KddpHealthcheckResult{
			Healthy:    false,
			Error:      &errstr,
			Exitstatus: &kddpResult.ReturnCode,
		}
	} else {
		result.KddpStatus = KddpHealthcheckResult{
			Healthy:    true,
			Version:    &kddpResult.Stdout,
			Exitstatus: &kddpResult.ReturnCode,
		}
	}

	return result
}
