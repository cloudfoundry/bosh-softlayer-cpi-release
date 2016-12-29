package common_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("RegistryAgentEnvService", func() {
	var (
		logger          boshlog.Logger
		agentEnvService AgentEnvService
		instanceID      string
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		instanceID = "fake-instance-id"
		agentEnvService = NewRegistryAgentEnvService("fake-registry-host", instanceID, logger)
	})

	Describe("Fetch", func() {
		Context("when instance is present in the registry", func() {
			var (
				ts           *httptest.Server
				settingsJSON string
				agentEnv     AgentEnv
			)

			BeforeEach(func() {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					Expect(r.Method).To(Equal("GET"))
					Expect(r.URL.Path).To(Equal("/instances/fake-instance-id/settings"))

					w.Write([]byte(settingsJSON))
				})
				ts = httptest.NewServer(handler)

				agentEnvService = NewRegistryAgentEnvService(ts.URL, instanceID, logger)
			})

			AfterEach(func() {
				ts.Close()
			})

			It("fetches settings from the registry", func() {
				settingsJSON = `{"settings": "{\"agent_id\":\"my-agent-id\"}"}`
				agentEnv = AgentEnv{AgentID: "fake-agent-id"}

				agentEnv, err := agentEnvService.Fetch()
				Expect(err).ToNot(HaveOccurred())
				Expect(agentEnv).To(Equal(agentEnv))
			})

			It("returns error if registry settings wrapper cannot be parsed", func() {
				settingsJSON = "invalid-json"

				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshalling registry response"))

				Expect(agentEnv).To(Equal(AgentEnv{}))
			})

			It("returns error if registry settings wrapper contains invalid json", func() {
				settingsJSON = `{"settings": "invalid-json"}`

				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshalling agent env from registry"))

				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})

		Context("when instance is not present in the registry", func() {
			var (
				ts *httptest.Server
			)

			BeforeEach(func() {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					Expect(r.Method).To(Equal("GET"))
					Expect(r.URL.Path).To(Equal("/instances/fake-instance-id/settings"))

					w.WriteHeader(http.StatusNotFound)
				})
				ts = httptest.NewServer(handler)

				agentEnvService = NewRegistryAgentEnvService(ts.URL, instanceID, logger)
			})

			AfterEach(func() {
				ts.Close()
			})

			It("returns an error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Received non-200 status code when contacting registry"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})
	})

	Describe("Update", func() {
		var (
			ts       *httptest.Server
			agentEnv AgentEnv
		)

		BeforeEach(func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.Method).To(Equal("PUT"))
				Expect(r.URL.Path).To(Equal("/instances/fake-instance-id/settings"))

				w.WriteHeader(http.StatusCreated)

			})
			ts = httptest.NewServer(handler)

			agentEnvService = NewRegistryAgentEnvService(ts.URL, instanceID, logger)
		})

		AfterEach(func() {
			ts.Close()
		})

		It("Updates settings in the registry", func() {
			agentEnv = AgentEnv{AgentID: "fake-agent-id"}
			err := agentEnvService.Update(agentEnv)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		var (
			ts *httptest.Server
		)

		BeforeEach(func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.Method).To(Equal("DELETE"))
				Expect(r.URL.Path).To(Equal("/instances/fake-instance-id/settings"))

				w.WriteHeader(http.StatusOK)
			})
			ts = httptest.NewServer(handler)

			agentEnvService = NewRegistryAgentEnvService(ts.URL, instanceID, logger)
		})

		AfterEach(func() {
			ts.Close()
		})

		It("Deletes settings in the registry", func() {
			err := agentEnvService.Delete()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
