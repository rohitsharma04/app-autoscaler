package forwarder

import (
	"autoscaler/metricsforwarder/config"
	"autoscaler/models"
	"fmt"

	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"
)

type MetricForwarder struct {
	client *loggregator.IngressClient
	logger lager.Logger
}

const METRICS_FORWARDER_ORIGIN = "autoscaler_metrics_forwarder"

func NewMetricForwarder(logger lager.Logger, conf *config.Config) (MetricForwarder, error) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		conf.LoggregatorConfig.CACertFile,
		conf.LoggregatorConfig.ClientCertFile,
		conf.LoggregatorConfig.ClientKeyFile,
	)
	if err != nil {
		logger.Error("could-not-create-TLS-config", err)
		return MetricForwarder{}, err
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr(conf.MetronAddress),
		loggregator.WithTag("origin", METRICS_FORWARDER_ORIGIN),
	)

	if err != nil {
		logger.Error("could-not-create-loggregator-client", err)
		return MetricForwarder{}, err
	}

	return MetricForwarder{
		client: client,
		logger: logger,
	}, nil
}

func (mf MetricForwarder) EmitMetric(metric *models.CustomMetric) {
	mf.logger.Debug("custom-metric-emit-request-received:", lager.Data{"metric": metric})

	gauge_tags := map[string]string{
		"applicationGuid":     metric.AppGUID,
		"applicationInstance": fmt.Sprint(metric.InstanceIndex),
	}
	options := []loggregator.EmitGaugeOption{
		loggregator.WithGaugeAppInfo(metric.AppGUID, 0),
		loggregator.WithEnvelopeTags(gauge_tags),
		loggregator.WithGaugeValue(metric.Name, metric.Value, metric.Unit),
	}

	mf.client.EmitGauge(options...)
}
