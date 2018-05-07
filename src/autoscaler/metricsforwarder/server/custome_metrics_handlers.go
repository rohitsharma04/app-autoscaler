package server

import (
	"autoscaler/metricsforwarder/forwarder"
	"autoscaler/models"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cfhttp/handlers"
)

type CustomMetricsHandler struct {
	metricForwarder forwarder.MetricForwarder
}

func NewCustomMetricsHandler(metricForwarder forwarder.MetricForwarder) *CustomMetricsHandler {
	return &CustomMetricsHandler{
		metricForwarder: metricForwarder,
	}
}

func (mh *CustomMetricsHandler) PublishMetrics(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Error posting custom metrics"})
		return
	}
	var metric models.CustomMetric
	err = json.Unmarshal(body, &metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	mh.metricForwarder.EmitMetric(metric)
	w.WriteHeader(http.StatusCreated)
}
