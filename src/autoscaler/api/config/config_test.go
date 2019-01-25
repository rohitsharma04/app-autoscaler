package config_test

import (
	. "autoscaler/api/config"
	"autoscaler/cf"
	"autoscaler/db"
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("Config", func() {

	var (
		conf        *Config
		err         error
		configBytes []byte
	)

	Describe("Load Config", func() {
		JustBeforeEach(func() {
			conf, err = LoadConfig(bytes.NewReader(configBytes))
		})

		Context("with invalid yaml", func() {
			BeforeEach(func() {
				configBytes = []byte(`
 broker_server:
  port: 8080,
public_api_server:
  port: 8081
logging:
  level: debug
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/se.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/se.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
metrics_collector:
  metrics_collector_url: http://localhost:8084
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/mc.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/mc.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
event_generator:
  event_generator_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/eg.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/eg.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
use_buildin_mode: false
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials
`)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(MatchRegexp("yaml: .*")))
			})
		})
		Context("with valid yaml", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_server:
  port: 8080
public_api_server:
  port: 8081
logging:
  level: debug
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/se.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/se.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
metrics_collector:
  metrics_collector_url: http://localhost:8084
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/mc.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/mc.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
event_generator:
  event_generator_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/eg.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/eg.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
use_buildin_mode: false
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials`)
			})

			It("It returns the config", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(conf.Logging.Level).To(Equal("debug"))
				Expect(conf.BrokerServer.Port).To(Equal(8080))
				Expect(conf.DB.BindingDB).To(Equal(
					db.DatabaseConfig{
						URL:                   "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable",
						MaxOpenConnections:    10,
						MaxIdleConnections:    5,
						ConnectionMaxLifetime: 60 * time.Second,
					}))
				Expect(conf.BrokerUsername).To(Equal("brokeruser"))
				Expect(conf.BrokerPassword).To(Equal("supersecretpassword"))
				Expect(conf.CatalogPath).To(Equal("../exampleconfig/catalog-example.json"))
				Expect(conf.CatalogSchemaPath).To(Equal("../schemas/catalog.schema.json"))
			})
		})
		Context("with partial config", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
metrics_collector:
  metrics_collector_url: http://localhost:8084
event_generator:
  event_generator_url: http://localhost:8083
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials
`)
			})
			It("It returns the default values", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(conf.Logging.Level).To(Equal("info"))
				Expect(conf.BrokerServer.Port).To(Equal(8080))
				Expect(conf.DB.BindingDB).To(Equal(
					db.DatabaseConfig{
						URL:                   "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable",
						MaxOpenConnections:    0,
						MaxIdleConnections:    0,
						ConnectionMaxLifetime: 0 * time.Second,
					}))
			})
		})

		Context("when it gives a non integer port", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_server:
  port: port
`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into int")))
			})
		})

		Context("when it gives a non integer max_open_connections of policydb", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_server:
  port: 8080
public_api_server:
  port: 8081
logging:
  level: debug
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: NOT-INTEGER-VALUE
    max_idle_connections: 5
    connection_max_lifetime: 60s
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/se.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/se.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
metrics_collector:
  metrics_collector_url: http://localhost:8084
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/mc.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/mc.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
event_generator:
  event_generator_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/eg.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/eg.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
use_buildin_mode: false
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials`)
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into int")))
			})
		})
		Context("when it gives a non integer max_idle_connections of policydb", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_server:
  port: 8080
public_api_server:
  port: 8081
logging:
  level: debug
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: NOT-INTEGER-VALUE
    connection_max_lifetime: 60s
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/se.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/se.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
metrics_collector:
  metrics_collector_url: http://localhost:8084
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/mc.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/mc.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
event_generator:
  event_generator_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/eg.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/eg.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
use_buildin_mode: false
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials`)
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into int")))
			})
		})
		Context("when it gives a non integer connection_max_lifetime of policydb", func() {
			BeforeEach(func() {
				configBytes = []byte(`
broker_server:
  port: 8080
public_api_server:
  port: 8081
logging:
  level: debug
broker_username: brokeruser
broker_password: supersecretpassword
db:
  binding_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: NOT-TIME-DURATION
  policy_db:
    url: postgres://postgres:postgres@localhost/autoscaler?sslmode=disable
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
catalog_schema_path: '../schemas/catalog.schema.json'
catalog_path: '../exampleconfig/catalog-example.json'
scaling_engine:
  scaling_engine_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/se.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/se.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
metrics_collector:
  metrics_collector_url: http://localhost:8084
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/mc.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/mc.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
event_generator:
  event_generator_url: http://localhost:8083
  tls:
    key_file: /var/vcap/jobs/autoscaler/config/certs/eg.key
    cert_file: /var/vcap/jobs/autoscaler/config/certs/eg.crt
    ca_file: /var/vcap/jobs/autoscaler/config/certs/autoscaler-ca.crt
use_buildin_mode: false
cf:
  api: https://api.example.com
  client_id: client-id
  secret: client-secret
  skip_ssl_validation: false
  grant_type: client_credentials`)
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into time.Duration")))
			})
		})

	})

	Describe("Validate", func() {
		BeforeEach(func() {
			conf = &Config{}
			conf.DB.BindingDB = db.DatabaseConfig{
				URL:                   "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable",
				MaxOpenConnections:    10,
				MaxIdleConnections:    5,
				ConnectionMaxLifetime: 60 * time.Second,
			}
			conf.DB.PolicyDB = db.DatabaseConfig{
				URL:                   "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable",
				MaxOpenConnections:    10,
				MaxIdleConnections:    5,
				ConnectionMaxLifetime: 60 * time.Second,
			}
			conf.BrokerUsername = "brokeruser"
			conf.BrokerPassword = "supersecretpassword"
			conf.CatalogSchemaPath = "../schemas/catalog.schema.json"
			conf.CatalogPath = "../exampleconfig/catalog-example.json"

			conf.MetricsCollector.MetricsCollectorUrl = "http://localhost:8083"
			conf.ScalingEngine.ScalingEngineUrl = "http://localhost:8084"
			conf.EventGenerator.EventGeneratorUrl = "http://localhost:8085"

			conf.CF.API = "http://api.bosh-lite.com"
			conf.CF.ClientID = "client-id"
			conf.CF.Secret = "secret"
			conf.CF.GrantType = cf.GrantTypeClientCredentials

		})
		JustBeforeEach(func() {
			err = conf.Validate()
		})

		Context("When all the configs are valid", func() {
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when bindingdb url is not set", func() {
			BeforeEach(func() {
				conf.DB.BindingDB.URL = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: BindingDB URL is empty")))
			})
		})

		Context("when broker username is not set", func() {
			BeforeEach(func() {
				conf.BrokerUsername = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: BrokerUsername is empty")))
			})
		})

		Context("when broker password is not set", func() {
			BeforeEach(func() {
				conf.BrokerPassword = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: BrokerPassword is empty")))
			})
		})

		Context("when metricscollector url is not set", func() {
			BeforeEach(func() {
				conf.MetricsCollector.MetricsCollectorUrl = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: metrics_collector.metrics_collector_url is empty")))
			})
		})

		Context("when eventgenerator url is not set", func() {
			BeforeEach(func() {
				conf.EventGenerator.EventGeneratorUrl = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: event_generator.event_generator_url is empty")))
			})
		})

		Context("when scalingengine url is not set", func() {
			BeforeEach(func() {
				conf.ScalingEngine.ScalingEngineUrl = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: scaling_engine.scaling_engine_url is empty")))
			})
		})

		Context("when catalog schema path is not set", func() {
			BeforeEach(func() {
				conf.CatalogSchemaPath = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: CatalogSchemaPath is empty")))
			})
		})

		Context("when catalog path is not set", func() {
			BeforeEach(func() {
				conf.CatalogPath = ""
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: CatalogPath is empty")))
			})
		})

		Context("when catalog is not valid json", func() {
			BeforeEach(func() {
				conf.CatalogPath = "../exampleconfig/catalog-invalid-json-example.json"
			})
			It("should err", func() {
				Expect(err).To(MatchError("invalid character '[' after object key"))
			})
		})

		Context("when catalog is missing required fields", func() {
			BeforeEach(func() {
				conf.CatalogPath = "../exampleconfig/catalog-missing-example.json"
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("{\"name is required\"}")))
			})
		})

		Context("when catalog has invalid type fields", func() {
			BeforeEach(func() {
				conf.CatalogPath = "../exampleconfig/catalog-invalid-example.json"
			})
			It("should err", func() {
				Expect(err).To(MatchError(MatchRegexp("{\"Invalid type. Expected: boolean, given: integer\"}")))
			})
		})
	})
})
