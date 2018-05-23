package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boslc "bosh-softlayer-cpi/softlayer/client"
	"time"
)

var _ = Describe("SwiftClient", func() {
	It("NewSwiftClient", func() {
		// Initialize session of swift client
		swiftEndpoint := "fake-swift-endpoint"
		swiftUsername := "fake-swift-username"
		swiftPassword := "fake-swfit-password"
		timeout := 300
		retries := 1

		conn := boslc.NewSwiftClient(swiftEndpoint, swiftUsername, swiftPassword, timeout, retries)
		Expect(conn.AuthUrl).To(Equal(swiftEndpoint))
		Expect(conn.UserName).To(Equal(swiftUsername))
		Expect(conn.ApiKey).To(Equal(swiftPassword))
		Expect(conn.Timeout).To(Equal(time.Duration(timeout) * time.Second))
		Expect(conn.Retries).To(Equal(retries))

		defer func() {
			// Inspect no panic occur
			r := recover()
			Expect(r).To(BeNil())
		}()
	})
})
