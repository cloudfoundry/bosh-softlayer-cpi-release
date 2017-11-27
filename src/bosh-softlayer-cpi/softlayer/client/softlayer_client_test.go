package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"io"
	"log"

	boslc "bosh-softlayer-cpi/softlayer/client"
)

var _ = Describe("SoftlayerClient", func() {
	It("NewSoftlayerClientSession", func() {
		// Initialize session of softlayer client
		var errOut, errOutLog bytes.Buffer
		multiWriter := io.MultiWriter(&errOut, &errOutLog)
		outLogger := log.New(multiWriter, "fake-uuid", log.LstdFlags)
		username := "SL_USERNAME"
		apiKey := "SL_API_KEY"
		timeout := 300
		retries := 1
		retryTimout := 60

		sess := boslc.NewSoftlayerClientSession(boslc.SoftlayerAPIEndpointPublicDefault, username, apiKey, false, timeout, retries, retryTimout, outLogger)
		Expect(sess.Endpoint).To(Equal(boslc.SoftlayerAPIEndpointPublicDefault))
		Expect(sess.APIKey).To(Equal(apiKey))
		Expect(sess.UserName).To(Equal(username))
		Expect(sess.Debug).To(Equal(false))

		defer func() {
			// Inspect no panic occur
			r := recover()
			Expect(r).To(BeNil())
		}()
	})
})
