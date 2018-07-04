package server_test

import (
	"autoscaler/metricsforwarder/config"
	"autoscaler/metricsforwarder/fakes"
	. "autoscaler/metricsforwarder/server"
	"autoscaler/models"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	serverProcess ifrit.Process
	serverUrl     string
	policyDB      *fakes.FakePolicyDB
)

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {

	port := 2222 + GinkgoParallelNode()
	testCertDir := "../../../../test-certs"
	loggregatorConfig := config.LoggregatorConfig{
		CACertFile:     filepath.Join(testCertDir, "loggregator-ca.crt"),
		ClientCertFile: filepath.Join(testCertDir, "metron.crt"),
		ClientKeyFile:  filepath.Join(testCertDir, "metron.key"),
	}
	conf := &config.Config{
		ServerPort:        port,
		LoggregatorConfig: loggregatorConfig,
	}
	policyDB = &fakes.FakePolicyDB{}
	httpServer, err := NewServer(lager.NewLogger("test"), conf, policyDB)
	Expect(err).NotTo(HaveOccurred())
	serverUrl = fmt.Sprintf("http://127.0.0.1:%d", conf.ServerPort)
	serverProcess = ginkgomon.Invoke(httpServer)
})

var _ = SynchronizedAfterSuite(func() {
	ginkgomon.Interrupt(serverProcess)
}, func() {
})

var _ = Describe("Server", func() {
	var (
		resp *http.Response
		req  *http.Request
		body []byte
		err  error
	)

	Context("when a request to forward custom metrics comes", func() {
		BeforeEach(func() {
			policyDB.ValidateCustomMetricsCredsReturns(true)
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "dummy-guid"})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req, err = http.NewRequest("POST", serverUrl+"/v1/metrics", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Basic M2YxZWY2MTJiMThlYTM5YmJlODRjZjUxMzY4MWYwYjc6YWYyNjk1Y2RmZDE0MzA4NThhMWY3MzJhYTI5NTQ2ZTk=")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns with a 201", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes without Authorization header", func() {
		BeforeEach(func() {
			policyDB.ValidateCustomMetricsCredsReturns(true)
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "dummy-guid"})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req, err = http.NewRequest("POST", serverUrl+"/v1/metrics", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns with a 401", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes without 'Basic'", func() {
		BeforeEach(func() {
			policyDB.ValidateCustomMetricsCredsReturns(true)
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "dummy-guid"})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req, err = http.NewRequest("POST", serverUrl+"/v1/metrics", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "M2YxZWY2MTJiMThlYTM5YmJlODRjZjUxMzY4MWYwYjc6YWYyNjk1Y2RmZDE0MzA4NThhMWY3MzJhYTI5NTQ2ZTk=")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns with a 401", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes with  wrong user credentials", func() {
		BeforeEach(func() {
			policyDB.ValidateCustomMetricsCredsReturns(false)
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "dummy-guid"})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req, err = http.NewRequest("POST", serverUrl+"/v1/metrics", bytes.NewReader(body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Basic M2YxZWY2MTJiMThlYTM5YmJlODRjZjUxMzY4MWYwYjc6YWYyNjk1Y2RmZDE0MzA4NThhMWY3MzJhYTI5NTQ2ZTk=")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns with a 401", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

})
