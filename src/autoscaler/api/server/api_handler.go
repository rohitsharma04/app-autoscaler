package server

import (
	"autoscaler/api/config"
	"autoscaler/db"
	"autoscaler/models"
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cfhttp/handlers"
	"code.cloudfoundry.org/lager"
)

type ApiHandler struct {
	logger    lager.Logger
	conf      *config.Config
	bindingdb db.BindingDB
}

func NewApiHandler(logger lager.Logger, conf *config.Config, bindingdb db.BindingDB) *ApiHandler {
	return &ApiHandler{
		logger:    logger,
		conf:      conf,
		bindingdb: bindingdb,
	}
}

func (h *ApiHandler) GetBrokerCatalog(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	w.Write([]byte(h.conf.Catalog))
}

func (h *ApiHandler) CreateServiceInstance(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	instanceId := vars["instanceId"]

	body := &models.InstanceCreationRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Failed to read request body"})
		return
	}

	err = h.bindingdb.CreateServiceInstance(instanceId, body.OrgGuid, body.SpaceGuid)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Error creating service instance"})
		return
	}

	if h.conf.DashboardRedirectURI == "" {
		w.WriteHeader(http.StatusCreated)
		w.Write(nil)
	} else {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("{\"dashboard_url\":\"%s\"}", GetDashboardURL(h.conf, instanceId))))
	}
}

func (h *ApiHandler) DeleteServiceInstance(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	instanceId := vars["instanceId"]
	err := h.bindingdb.DeleteServiceInstance(instanceId)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Error deleting service instance"})
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}

func (h *ApiHandler) BindServiceInstance(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	instanceId := vars["instanceId"]
	bindingId := vars["bindingId"]

	body := &models.BindingRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Failed to read request body"})
		return
	}

	err = h.bindingdb.CreateServiceBinding(bindingId, instanceId, body.AppId)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Error creating service binding"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(nil)
}

func (h *ApiHandler) UnbindServiceInstance(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	bindingId := vars["bindingId"]

	err := h.bindingdb.DeleteServiceBinding(bindingId)
	if err != nil {
		handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "Interal-Server-Error",
			Message: "Error creating service binding"})
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}
