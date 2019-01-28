package publicapiserver

import (
	"autoscaler/api/config"
	"autoscaler/cf"
	"autoscaler/models"
	"encoding/json"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cfhttp/handlers"
	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
)

type OAuthMiddleware struct {
	logger          lager.Logger
	cf              cf.CFConfig
	httpClient      *http.Client
	cfTokenEndpoint string
}

func NewOauthMiddleware(logger lager.Logger, conf *config.Config) *OAuthMiddleware {
	return &OAuthMiddleware{
		logger:     logger,
		cf:         conf.CF,
		httpClient: &http.Client{},
	}
}

func (oam *OAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		appId := vars["appId"]
		userToken := r.Header.Get("Authorization")

		if appId == "" {
			handlers.WriteJSONResponse(w, http.StatusBadRequest, models.ErrorResponse{
				Code:    "Bad Request",
				Message: "Malformed or missing appId",
			})
			return
		}

		if userToken == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userId, err := oam.getUserId(userToken)
		if err != nil {
			handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
				Code:    "Interal-Server-Error",
				Message: "Failed to get user ID"})
			return
		}

		scopes, err := oam.getUserScope(userToken, userId)
		if err != nil {
			handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
				Code:    "Interal-Server-Error",
				Message: "Failed to get user scope"})
			return
		}
		for _, scope := range scopes {
			if scope == "cloud_controller.admin" {
				oam.logger.Info("user is cc admin", lager.Data{"userId": userId})
				next.ServeHTTP(w, r)
				return
			}
		}

		isSpaceDev, err := oam.isUserSpaceDeveloper(userToken, userId, appId)
		if err != nil {
			handlers.WriteJSONResponse(w, http.StatusInternalServerError, models.ErrorResponse{
				Code:    "Interal-Server-Error",
				Message: "Failed to check space developer permission"})
			return
		}
		if !isSpaceDev {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}

func (oam *OAuthMiddleware) isUserSpaceDeveloper(userToken string, userId string, appId string) (bool, error) {
	userSpaceDeveloperPermissionEndpoint := oam.cf.API + "/v2/users/" + userId + "/spaces?q=app_guid:" + appId + "&q=developer_guid:" + userId

	req, err := http.NewRequest("GET", userSpaceDeveloperPermissionEndpoint, nil)
	if err != nil {
		oam.logger.Error("Failed to create check space dev permission request", err, lager.Data{"userSpaceDeveloperPermissionEndpoint": userSpaceDeveloperPermissionEndpoint})
		return false, err
	}
	req.Header.Set("Authorization", userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := oam.httpClient.Do(req)
	if err != nil {
		oam.logger.Error("failed to get user space dev permission, request failed", err, lager.Data{"userSpaceDeveloperPermissionEndpoint": userSpaceDeveloperPermissionEndpoint})
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		oam.logger.Error("Failed to get user space dev permission", err, lager.Data{"userSpaceDeveloperPermissionEndpoint": userSpaceDeveloperPermissionEndpoint, "statusCode": resp.StatusCode})
		return false, err
	}

	spaces := struct {
		Total int `json:"total_results"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&spaces)
	if err != nil {
		oam.logger.Error("Failed to parse user space dev permission response body", err, lager.Data{"userSpaceDeveloperPermissionEndpoint": userSpaceDeveloperPermissionEndpoint})
	}
	return spaces.Total > 0, nil
}

func (oam *OAuthMiddleware) getUserScope(userToken string, userId string) ([]string, error) {
	userScopeEndpoint := oam.getCFTokenEndpoint() + "/check_token?token=" + strings.Split(userToken, " ")[1]
	req, err := http.NewRequest("POST", userScopeEndpoint, nil)
	if err != nil {
		oam.logger.Error("failed to create getuserscope request", err, lager.Data{"userScopeEndpoint": userScopeEndpoint})
		return nil, err
	}
	req.SetBasicAuth(oam.cf.ClientID, oam.cf.Secret)

	resp, err := oam.httpClient.Do(req)
	if err != nil {
		oam.logger.Error("failed to get user scope, request failed", err, lager.Data{"userScopeEndpoint": userScopeEndpoint})
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		oam.logger.Error("Failed to get user scope", err, lager.Data{"userScopeEndpoint": userScopeEndpoint, "statusCode": resp.StatusCode})
		return nil, err
	}
	userScope := struct {
		Scope []string `json:"scope"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&userScope)
	if err != nil {
		oam.logger.Error("Failed to parse user scope response body", err, lager.Data{"userScopeEndpoint": userScopeEndpoint})
	}
	return userScope.Scope, nil
}

func (oam *OAuthMiddleware) getUserId(userToken string) (string, error) {
	userInfoEndpoint := oam.getCFTokenEndpoint() + "/userinfo"

	req, err := http.NewRequest("GET", userInfoEndpoint, nil)
	if err != nil {
		oam.logger.Error("failed to get user info, create request failed", err, lager.Data{"userInfoEndpoint": userInfoEndpoint})
		return "", err
	}
	req.Header.Set("Authorization", userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := oam.httpClient.Do(req)
	if err != nil {
		oam.logger.Error("failed to get user info, request failed", err, lager.Data{"userInfoEndpoint": userInfoEndpoint})
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		oam.logger.Error("Failed to get user info", err, lager.Data{"userInfoEndpoint": userInfoEndpoint, "statusCode": resp.StatusCode})
		return "", err
	}

	userInfo := struct {
		UserId string `json:"user_id"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		oam.logger.Error("Failed to parse user info response body", err, lager.Data{"userInfoEndpoint": userInfoEndpoint})
		return "", err
	}

	return userInfo.UserId, nil
}

func (oam *OAuthMiddleware) getCFTokenEndpoint() string {
	if oam.cfTokenEndpoint == "" {
		infoEndpoint := oam.cf.API + "/v2/info"

		resp, err := oam.httpClient.Get(infoEndpoint)
		if err != nil {
			oam.logger.Error("failed to get cf info, request failed", err, lager.Data{"infoEndpoint": infoEndpoint})
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			oam.logger.Error("Failed to get cf info", err, lager.Data{"infoEndpoint": infoEndpoint, "statusCode": resp.StatusCode})
		}

		info := struct {
			Endpoint string `json:"token_endpoint"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&info)
		if err != nil {
			oam.logger.Error("Failed to parse cf info response body", err, lager.Data{"infoEndpoint": infoEndpoint})
		}
		oam.cfTokenEndpoint = info.Endpoint
	}
	return oam.cfTokenEndpoint
}
