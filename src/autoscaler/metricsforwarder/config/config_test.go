package config_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"autoscaler/db"
	. "autoscaler/metricsforwarder/config"
)

var _ = Describe("Config", func() {

	var (
		conf        *Config
		err         error
		configBytes []byte
	)

	Describe("LoadConfig", func() {

		JustBeforeEach(func() {
			conf, err = LoadConfig(bytes.NewReader(configBytes))
		})

		Context("with invalid yaml", func() {
			BeforeEach(func() {
				configBytes = []byte(`
  server_port: 8081
log_level: info
metron_address: 127.0.0.1:3457
loggregator_ca_cert_path: "../testcerts/ca.crt"
`)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(MatchRegexp("yaml: .*")))
			})
		})

		Context("with valid yaml", func() {
			BeforeEach(func() {
				configBytes = []byte(`
server_port: 8081
log_level: debug
metron_address: 127.0.0.1:3457
loggregator_ca_cert_path: "../testcerts/ca.crt"
loggregator_cert_path: "../testcerts/client.crt"
loggregator_key_path: "../testcerts/client.key"
db:
  policy_db:
    url: "postgres://pqgotest:password@localhost/pqgotest"
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
`)
			})

			It("returns the config", func() {
				Expect(conf.ServerPort).To(Equal(8081))
				Expect(conf.LogLevel).To(Equal("debug"))
				Expect(conf.MetronAddress).To(Equal("127.0.0.1:3457"))
				Expect(conf.Db.PolicyDb).To(Equal(
					db.DatabaseConfig{
						Url:                   "postgres://pqgotest:password@localhost/pqgotest",
						MaxOpenConnections:    10,
						MaxIdleConnections:    5,
						ConnectionMaxLifetime: 60 * time.Second,
					}))
			})
		})

		Context("with partial config", func() {
			BeforeEach(func() {
				configBytes = []byte(`
loggregator_ca_cert_path: "../testcerts/ca.crt"
loggregator_cert_path: "../testcerts/client.crt"
loggregator_key_path: "../testcerts/client.key"
db:
  policy_db:
    url: "postgres://pqgotest:password@localhost/pqgotest"
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 60s
`)
			})

			It("returns default values", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(conf.ServerPort).To(Equal(6110))
				Expect(conf.LogLevel).To(Equal("info"))
				Expect(conf.MetronAddress).To(Equal("127.0.0.1:3458"))
			})
		})

		Context("when it gives a non integer port", func() {
			BeforeEach(func() {
				configBytes = []byte(`
server_port: port
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
loggregator_ca_cert_path: "../testcerts/ca.crt"
loggregator_cert_path: "../testcerts/client.crt"
loggregator_key_path: "../testcerts/client.key"
db:
  policy_db:
    url: postgres://pqgotest:password@localhost/pqgotest
    max_open_connections: NOT-INTEGER-VALUE
    max_idle_connections: 5
    connection_max_lifetime: 60s
`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into int")))
			})
		})

		Context("when it gives a non integer max_idle_connections of policydb", func() {
			BeforeEach(func() {
				configBytes = []byte(`
loggregator_ca_cert_path: "../testcerts/ca.crt"
loggregator_cert_path: "../testcerts/client.crt"
loggregator_key_path: "../testcerts/client.key"
db:
  policy_db:
    url: postgres://pqgotest:password@localhost/pqgotest
    max_open_connections: 10
    max_idle_connections: NOT-INTEGER-VALUE
    connection_max_lifetime: 60s
`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal.*into int")))
			})
		})

		Context("when connection_max_lifetime of policydb is not a time duration", func() {
			BeforeEach(func() {
				configBytes = []byte(`
loggregator_ca_cert_path: "../testcerts/ca.crt"
loggregator_cert_path: "../testcerts/client.crt"
loggregator_key_path: "../testcerts/client.key"
db:
  policy_db:
    url: postgres://pqgotest:password@localhost/pqgotest
    max_open_connections: 10
    max_idle_connections: 5
    connection_max_lifetime: 6K
`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal .* into time.Duration")))
			})
		})

	})

	Describe("Validate", func() {
		BeforeEach(func() {
			conf = &Config{}
			conf.ServerPort = 8081
			conf.LogLevel = "debug"
			conf.MetronAddress = "127.0.0.1:3458"
			conf.LoggregatorConfig.CACertFile = "../testcerts/ca.crt"
			conf.LoggregatorConfig.ClientCertFile = "../testcerts/client.crt"
			conf.LoggregatorConfig.ClientKeyFile = "../testcerts/client.crt"
			conf.Db.PolicyDb = db.DatabaseConfig{
				Url:                   "postgres://pqgotest:password@localhost/pqgotest",
				MaxOpenConnections:    10,
				MaxIdleConnections:    5,
				ConnectionMaxLifetime: 60 * time.Second,
			}

		})

		JustBeforeEach(func() {
			err = conf.Validate()
		})

		Context("when all the configs are valid", func() {
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when policy db url is not set", func() {
			BeforeEach(func() {
				conf.Db.PolicyDb.Url = ""
			})

			It("should error", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: Policy DB url is empty")))
			})
		})

		Context("when policy db url is not set", func() {
			BeforeEach(func() {
				conf.Db.PolicyDb.Url = ""
			})

			It("should error", func() {
				Expect(err).To(MatchError(MatchRegexp("Configuration error: Policy DB url is empty")))
			})
		})

		// Context("when loggregator CA certificate file not set", func() {
		// 	BeforeEach(func() {
		// 		conf.LoggregatorConfig.CACertFile = ""
		// 	})

		// 	It("should error", func() {
		// 		Expect(err).To(MatchError(MatchRegexp("Configuration error: Loggregator CA Certificate file not provided")))
		// 	})
		// })

		// Context("when loggregator Metron Client certificate file not set", func() {
		// 	BeforeEach(func() {
		// 		conf.LoggregatorConfig.CACertFile = ""
		// 	})

		// 	It("should error", func() {
		// 		Expect(err).To(MatchError(MatchRegexp("Configuration error: Loggregator Metron Client Certificate file not provided")))
		// 	})
		// })

		// Context("when loggregator Metron Client Key  file not set", func() {
		// 	BeforeEach(func() {
		// 		conf.LoggregatorConfig.CACertFile = ""
		// 	})

		// 	It("should error", func() {
		// 		Expect(err).To(MatchError(MatchRegexp("Configuration error: Loggregator Metron Client Key file not provided")))
		// 	})
		// })

	})
})
