package publicapiserver

import (
	"autoscaler/api/config"
	"autoscaler/db"
	"net/http"

	"code.cloudfoundry.org/lager"
)

type PublicApiHandler struct {
	logger   lager.Logger
	conf     *config.Config
	policydb db.PolicyDB
}

func NewPublicApiHandler(logger lager.Logger, conf *config.Config, policydb db.PolicyDB) *PublicApiHandler {
	return &PublicApiHandler{
		logger:   logger,
		conf:     conf,
		policydb: policydb,
	}
}

func (h *PublicApiHandler) GetScalingHistories(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	w.Write([]byte("Scaling Histories"))
}
