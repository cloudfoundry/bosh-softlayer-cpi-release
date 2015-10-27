package vm_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cppforlife/bosh-warden-cpi/vm"
)

var _ = Describe("RegistryAgentEnvService", func() {
	var (
		logger               boshlog.Logger
		agentEnvService      AgentEnvService
		registryServer       *registryServer
		expectedAgentEnv     AgentEnv
		expectedAgentEnvJSON []byte
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		registryOptions := RegistryOptions{
			Host:     "127.0.0.1",
			Port:     6307,
			Username: "fake-username",
			Password: "fake-password",
		}

		if registryServer == nil {
			registryServer = NewRegistryServer(registryOptions)

			readyCh := make(chan struct{})
			stopCh := make(chan error)

			go func() {
				err := registryServer.Start(readyCh)
				if err != nil {
					stopCh <- err
				}
			}()

			select {
			case <-readyCh:
				// ok, continue
			case err := <-stopCh:
				panic(fmt.Sprintf("Error occurred waiting for registry server to start: %s", err))
			}
		}

		instanceID := "fake-instance-id"
		agentEnvService = NewRegistryAgentEnvService(registryOptions, instanceID, logger)

		expectedAgentEnv = AgentEnv{AgentID: "fake-agent-id"}
		var err error
		expectedAgentEnvJSON, err = json.Marshal(expectedAgentEnv)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Leaking registry server on purpose
		registryServer.Stop()
	})

	Describe("Fetch", func() {
		Context("when settings for the instance exist in the registry", func() {
			BeforeEach(func() {
				registryServer.InstanceSettings = expectedAgentEnvJSON
			})

			It("fetches settings from the registry", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).ToNot(HaveOccurred())
				Expect(agentEnv).To(Equal(expectedAgentEnv))
			})
		})

		Context("when settings for instance do not exist", func() {
			It("returns an error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})
	})

	Describe("Update", func() {
		It("updates settings in the registry", func() {
			Expect(registryServer.InstanceSettings).To(BeNil())
			err := agentEnvService.Update(expectedAgentEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(registryServer.InstanceSettings).To(Equal(expectedAgentEnvJSON))
		})
	})
})

type registryServer struct {
	InstanceSettings []byte
	options          RegistryOptions
	listener         net.Listener
}

func NewRegistryServer(options RegistryOptions) *registryServer {
	return &registryServer{options: options}
}

func (s *registryServer) Start(readyCh chan struct{}) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.options.Host, s.options.Port))
	if err != nil {
		return err
	}

	s.listener = listener

	readyCh <- struct{}{}

	mux := http.NewServeMux()
	mux.HandleFunc("/instances/fake-instance-id/settings", s.instanceHandler)

	server := &http.Server{Handler: mux}

	return server.Serve(s.listener)
}

func (s *registryServer) Stop() error {
	// if client keeps connection alive, server will still be running
	s.InstanceSettings = nil

	return nil
}

type registryResp struct {
	Settings string `json:"settings"`
}

func (s *registryServer) instanceHandler(w http.ResponseWriter, req *http.Request) {
	if !s.isAuthorized(req) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if req.Method == "GET" {
		resp := registryResp{Settings: string(s.InstanceSettings)}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal instance settings", 500)
			return
		}

		if s.InstanceSettings != nil {
			w.Write(respBytes)
			return
		}

		http.NotFound(w, req)
		return
	}

	if req.Method == "PUT" {
		reqBody, _ := ioutil.ReadAll(req.Body)
		s.InstanceSettings = reqBody
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *registryServer) isAuthorized(req *http.Request) bool {
	auth := s.options.Username + ":" + s.options.Password
	expectedAuthorizationHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	return expectedAuthorizationHeader == req.Header.Get("Authorization")
}
