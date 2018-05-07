package forwarder

import (
	"autoscaler/metricsforwarder/config"
	"autoscaler/models"
	"fmt"
	"log"

	"code.cloudfoundry.org/go-loggregator"
)

type MetricForwarder struct {
	client *loggregator.IngressClient
}

func NewMetricForwarder(conf *config.Config) (MetricForwarder, error) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		conf.LoggregatorConfig.CACertFile,
		conf.LoggregatorConfig.ClientCertFile,
		conf.LoggregatorConfig.ClientKeyFile,
	)
	if err != nil {
		log.Fatal("Could not create TLS config", err)
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr(conf.MetronAddress),
		loggregator.WithTag("origin", "autoscaler_metrics_forwarder"),
	)

	if err != nil {
		log.Fatal("Could not create client", err)
		return MetricForwarder{}, err
	}

	return MetricForwarder{
		client: client,
	}, nil
}

func (mf MetricForwarder) EmitMetric(metric models.CustomMetric) {
	fmt.Println("Metric Emit request received:", metric)

	gauge_tags := map[string]string{
		"applicationGuid":     metric.AppGUID,
		"applicationInstance": fmt.Sprint(metric.InstanceID),
	}
	options := []loggregator.EmitGaugeOption{
		loggregator.WithGaugeAppInfo(metric.AppGUID, 0),
		loggregator.WithEnvelopeTags(gauge_tags),
		loggregator.WithGaugeValue(metric.Name, metric.Value, metric.Unit),
	}

	mf.client.EmitGauge(options...)
}
