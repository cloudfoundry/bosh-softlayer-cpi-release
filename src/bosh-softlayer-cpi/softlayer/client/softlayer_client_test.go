package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boslc "bosh-softlayer-cpi/softlayer/client"
	"bytes"
	"io"
)

var _ = Describe("SoftlayerClient", func() {
	It("NewSoftlayerClientSession", func() {
		// Initialize session of softlayer client
		var errOut, errOutLog bytes.Buffer
		multiWriter := io.MultiWriter(&errOut, &errOutLog)
		username := "SL_USERNAME"
		apiKey := "SL_API_KEY"
		timeout := 300
		retries := 1
		retryTimout := 60

		sess := boslc.NewSoftlayerClientSession(boslc.SoftlayerAPIEndpointPublicDefault, username, apiKey, false, timeout, retries, retryTimout, multiWriter)
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
