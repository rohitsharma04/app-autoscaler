package routes

import (
	"github.com/gorilla/mux"

	"net/http"
)

const (
	MetricHistoriesPath         = "/v1/apps/{appid}/metric_histories/{metrictype}"
	GetMetricHistoriesRouteName = "GetMetricHistories"

	AggregatedMetricHistoriesPath         = "/v1/apps/{appid}/aggregated_metric_histories/{metrictype}"
	GetAggregatedMetricHistoriesRouteName = "GetAggregatedMetricHistories"

	ScalePath      = "/v1/apps/{appid}/scale"
	ScaleRouteName = "Scale"

	ScalingHistoriesPath         = "/v1/apps/{appid}/scaling_histories"
	GetScalingHistoriesRouteName = "GetScalingHistories"

	ActiveSchedulePath            = "/v1/apps/{appid}/active_schedules/{scheduleid}"
	SetActiveScheduleRouteName    = "SetActiveSchedule"
	DeleteActiveScheduleRouteName = "DeleteActiveSchedule"

	ActiveSchedulesPath         = "/v1/apps/{appid}/active_schedules"
	GetActiveSchedulesRouteName = "GetActiveSchedules"

	SyncActiveSchedulesPath      = "/v1/syncSchedules"
	SyncActiveSchedulesRouteName = "SyncActiveSchedules"

	ApiPolicyPath            = "/v1/apps/{appId}/policy"
	ApiGetPolicyRouteName    = "GetPolicy"
	ApiPutPolicyRouteName    = "PutPolicy"
	ApiDeletePolicyRouteName = "DeletePolicy"

	ApiGetScalingHistoryPath   = "/v1/apps/{appId}/scaling_histories"
	ApiScalingHistoryRouteName = "GetScalingHistory"

	ApiGetMetricsHistoryPath      = "/v1/apps/{appId}/metric_histories/{metricType}"
	ApiGetMetricsHistoryRouteName = "GetMetricsHistory"

	ApiCatalogPath      = "/sb/v2/catalog"
	ApiCatalogRouteName = "GetCatalog"

	ApiInstancePath            = "/sb/v2/service_instances/{instanceId}"
	ApiCreateInstanceRouteName = "CreateInstance"
	ApiDeleteInstanceRouteName = "DeleteInstance"

	ApiBindingPath            = "/sb/v2/service_instances/{instanceId}/service_bindings/{bindingId}"
	ApiCreateBindingRouteName = "CreateBinding"
	ApiDeleteBindingRouteName = "DeleteBinding"
)

type AutoScalerRoute struct {
	metricsCollectorRoutes *mux.Router
	eventGeneratorRoutes   *mux.Router
	scalingEngineRoutes    *mux.Router
	apiRoutes              *mux.Router
}

var autoScalerRouteInstance = newRouters()

func newRouters() *AutoScalerRoute {
	instance := &AutoScalerRoute{
		metricsCollectorRoutes: mux.NewRouter(),
		eventGeneratorRoutes:   mux.NewRouter(),
		scalingEngineRoutes:    mux.NewRouter(),
		apiRoutes:              mux.NewRouter(),
	}

	instance.metricsCollectorRoutes.Path(MetricHistoriesPath).Methods(http.MethodGet).Name(GetMetricHistoriesRouteName)

	instance.eventGeneratorRoutes.Path(AggregatedMetricHistoriesPath).Methods(http.MethodGet).Name(GetAggregatedMetricHistoriesRouteName)

	instance.scalingEngineRoutes.Path(ScalePath).Methods(http.MethodPost).Name(ScaleRouteName)
	instance.scalingEngineRoutes.Path(ScalingHistoriesPath).Methods(http.MethodGet).Name(GetScalingHistoriesRouteName)
	instance.scalingEngineRoutes.Path(ActiveSchedulePath).Methods(http.MethodPut).Name(SetActiveScheduleRouteName)
	instance.scalingEngineRoutes.Path(ActiveSchedulePath).Methods(http.MethodDelete).Name(DeleteActiveScheduleRouteName)
	instance.scalingEngineRoutes.Path(ActiveSchedulesPath).Methods(http.MethodGet).Name(GetActiveSchedulesRouteName)
	instance.scalingEngineRoutes.Path(SyncActiveSchedulesPath).Methods(http.MethodPut).Name(SyncActiveSchedulesRouteName)

	instance.apiRoutes.Path(ApiPolicyPath).Methods(http.MethodGet).Name(ApiGetPolicyRouteName)
	instance.apiRoutes.Path(ApiPolicyPath).Methods(http.MethodPut).Name(ApiPutPolicyRouteName)
	instance.apiRoutes.Path(ApiPolicyPath).Methods(http.MethodDelete).Name(ApiDeletePolicyRouteName)

	instance.apiRoutes.Path(ApiCatalogPath).Methods(http.MethodGet).Name(ApiCatalogRouteName)

	instance.apiRoutes.Path(ApiInstancePath).Methods(http.MethodPut).Name(ApiCreateInstanceRouteName)
	instance.apiRoutes.Path(ApiInstancePath).Methods(http.MethodDelete).Name(ApiDeleteInstanceRouteName)

	instance.apiRoutes.Path(ApiBindingPath).Methods(http.MethodPut).Name(ApiCreateBindingRouteName)
	instance.apiRoutes.Path(ApiBindingPath).Methods(http.MethodDelete).Name(ApiDeleteBindingRouteName)

	instance.apiRoutes.Path(ApiGetScalingHistoryPath).Methods(http.MethodGet).Name(ApiScalingHistoryRouteName)
	instance.apiRoutes.Path(ApiGetMetricsHistoryPath).Methods(http.MethodGet).Name(ApiGetMetricsHistoryRouteName)

	return instance

}
func MetricsCollectorRoutes() *mux.Router {
	return autoScalerRouteInstance.metricsCollectorRoutes
}

func EventGeneratorRoutes() *mux.Router {
	return autoScalerRouteInstance.eventGeneratorRoutes
}

func ScalingEngineRoutes() *mux.Router {
	return autoScalerRouteInstance.scalingEngineRoutes
}

func ApiRoutes() *mux.Router {
	return autoScalerRouteInstance.apiRoutes
}
