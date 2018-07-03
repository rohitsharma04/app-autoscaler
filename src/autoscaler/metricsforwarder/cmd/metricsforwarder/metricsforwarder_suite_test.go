package main_test

import (
	"path/filepath"
	"net/http"
	"database/sql"
	"golang.org/x/crypto/bcrypt"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"gopkg.in/yaml.v2"
	"io/ioutil"

	"os/exec"
	"os"
	"time"

	"testing"

	"autoscaler/metricsforwarder/config"
	"autoscaler/db"

	"autoscaler/metricsforwarder/testhelpers"

)

var (
	mfPath			string
	cfg				config.Config
	configFile      *os.File
	httpClient      *http.Client
	req				*http.Request
	username		string
	password		string 
	grpcIngressTestServer *testhelpers.TestIngressServer
)

func TestMetricsforwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metricsforwarder Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte{
	mf, err := gexec.Build("autoscaler/metricsforwarder/cmd/metricsforwarder", "-race")
	Expect(err).NotTo(HaveOccurred())

	policyDB, err := sql.Open(db.PostgresDriverName, os.Getenv("DBURL"))
	Expect(err).NotTo(HaveOccurred())

	bindingDB, err := sql.Open(db.PostgresDriverName, os.Getenv("DBURL"))
	Expect(err).NotTo(HaveOccurred())

	_, err = policyDB.Exec("DELETE from policy_json")
	Expect(err).NotTo(HaveOccurred())

	_, err = bindingDB.Exec("DELETE from credentials")
	Expect(err).NotTo(HaveOccurred())

	_, err = bindingDB.Exec("DELETE from binding")
	Expect(err).NotTo(HaveOccurred())

	_, err = bindingDB.Exec("DELETE from service_instance")
	Expect(err).NotTo(HaveOccurred())

	policy := `
		{
 			"instance_min_count": 1,
			"instance_max_count": 5,
			"scaling_rules":[
				{
					"metric_type":"custom",
					"breach_duration_secs":600,
					"threshold":30,
					"operator":"<",
					"cool_down_secs":300,
					"adjustment":"-1"
				}
			]
		}`
	query := "INSERT INTO policy_json(app_id, policy_json, guid) values($1, $2, $3)"
	_, err = policyDB.Exec(query, "an-app-id", policy, "1234")
	
	query = "INSERT INTO service_instance(service_instance_id,org_id,space_id) values($1,$2,$3)"
	_, err = bindingDB.Exec(query, "service-instance-id", "org-id", "space-id")

	query = "INSERT INTO binding(binding_id,service_instance_id,app_id,created_at) values($1,$2,$3,$4)"
	_, err = bindingDB.Exec(query, "binding-id", "service-instance-id", "an-app-id", "2011-05-18 15:36:38")

	username = "username"
	password = "password"
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 8)

	query = "INSERT INTO credentials(binding_id, username, password) values($1, $2, $3)"
	_, err = bindingDB.Exec(query, "binding-id", username, encryptedPassword)
	Expect(err).NotTo(HaveOccurred())


	err = policyDB.Close()
	Expect(err).NotTo(HaveOccurred())
	
	err = bindingDB.Close()
	Expect(err).NotTo(HaveOccurred())
	

	return []byte(mf)
}, func(pathsByte []byte){
	mfPath = string(pathsByte)

	testCertDir := "../../../../../test-certs"

	grpcIngressTestServer, _ = testhelpers.NewTestIngressServer(
		filepath.Join(testCertDir,"metricsforwarder.crt"),
		filepath.Join(testCertDir,"metricsforwarder.key"),
		filepath.Join(testCertDir,"autoscaler-ca.crt"),
	)
	
	grpcIngressTestServer.Start()

	cfg.LoggregatorConfig.CACertFile = filepath.Join(testCertDir,"autoscaler-ca.crt")
	cfg.LoggregatorConfig.ClientCertFile = filepath.Join(testCertDir,"metricsforwarder.crt")
	cfg.LoggregatorConfig.ClientKeyFile = filepath.Join(testCertDir,"metricsforwarder.key")
	
	cfg.LogLevel = "debug"
	cfg.MetronAddress= grpcIngressTestServer.GetAddr()
	cfg.ServerPort = 10000 + GinkgoParallelNode();
	cfg.Db.PolicyDb = db.DatabaseConfig{
		Url:                   os.Getenv("DBURL"),
		MaxOpenConnections:    10,
		MaxIdleConnections:    5,
		ConnectionMaxLifetime: 10 * time.Second,
	}
	configFile = writeConfig(&cfg)
	
	httpClient = &http.Client{}
})

var _ = SynchronizedAfterSuite(func() {
	grpcIngressTestServer.Stop()
	os.Remove(configFile.Name())
}, func() {
	gexec.CleanupBuildArtifacts()
})


func writeConfig(c *config.Config) *os.File {
	cfg, err := ioutil.TempFile("", "mf")
	Expect(err).NotTo(HaveOccurred())
	defer cfg.Close()

	bytes, err := yaml.Marshal(c)
	Expect(err).NotTo(HaveOccurred())

	_, err = cfg.Write(bytes)
	Expect(err).NotTo(HaveOccurred())

	return cfg
}

type MetricsForwarderRunner struct {
	configPath string
	Session    *gexec.Session
	startCheck string
}

func NewMetricsForwarderRunner() *MetricsForwarderRunner {
	return &MetricsForwarderRunner{
		configPath: configFile.Name(),
		startCheck: "metricsforwarder.started",
	}
}

func (mf *MetricsForwarderRunner) Start() {
	mfSession, err := gexec.Start(exec.Command(
		mfPath,
		"-c",
		mf.configPath,
	),
		gexec.NewPrefixedWriter("\x1b[32m[o]\x1b[32m[mc]\x1b[0m ", GinkgoWriter),
		gexec.NewPrefixedWriter("\x1b[91m[e]\x1b[32m[mc]\x1b[0m ", GinkgoWriter),
	)
	Expect(err).NotTo(HaveOccurred())

	if mf.startCheck != "" {
		Eventually(mfSession.Buffer, 2).Should(gbytes.Say(mf.startCheck))
	}

	mf.Session = mfSession
}

func (mf *MetricsForwarderRunner) Interrupt() {
	if mf.Session != nil {
		mf.Session.Interrupt().Wait(5 * time.Second)
	}
}

func (mf *MetricsForwarderRunner) KillWithFire() {
	if mf.Session != nil {
		mf.Session.Kill().Wait(5 * time.Second)
	}
}
