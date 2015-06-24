/*
Integration test for tags
1. spin up vm with softlayer_go
2. set tags with ./out/cpi after bin/build
	2a. manually build before or build/clean during test? https://github.com/cloudfoundry/cli/blob/master/bin/test
	2b. runner.run() argument format
3. verify tags set with softlayer_go

Problems:
1. Ginkgo timed out waiting for all parallel nodes to report back!
	1a. Do I need to get godeps myself? Perhaps unrelated.
	1b. Possibility of getting all required dependencies in a shell script
	Solution: one of the test threads is failing, make sure all test dependencies are satisfied
2. {"result":null,"error":{"type":"Bosh::Clouds::CloudError","message":"Extracting method arguments from payload: Unmarshalling action argument: strconv.ParseInt: parsing \"4310eade-7084-4167-a326-3fa0aed81ed1.softlayer.com\": invalid syntax","ok_to_retry":false},"log":""}
3. Should I set username and api key in the test file
4. get the machine id?

Notes:
look for cli integration in cloudfoundry/cf-acceptance-tests and pivotal-cf-experimental/GATS
integration-test-support in ruby, look for golang equivalent in cloudfoundry-incubator/cli or cf-testhelpers?
*/

// am I using package correctly?
package virtual_guest_tags_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	// why this import format?
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// import from cf-test-helpers
	runner "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	// datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

var _ = Describe("SoftLayer Tags - to be updated", func() {
	var (
		err error

		accountService      softlayer.SoftLayer_Account_Service
		virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service
	)

	// determine if I need stuff in cloudfoundry/cli/cf/configuration/core_config/ for auth
	// I don't think so
	BeforeEach(func() {
		// fmt.Printf("<=================LINE 28=================>")
		// set env var for username
		os.Setenv("SL_USERNAME", "maxim@us.ibm.com")
		// set env var for api key
		os.Setenv("SL_API_KEY", "4afb3febf56ebbfb607468abbd862936e85aa9397458722dd3758c73c3598879")

		accountService, err = testhelpers.CreateAccountService()
		Expect(err).ToNot(HaveOccurred())

		virtualGuestService, err = testhelpers.CreateVirtualGuestService()
		Expect(err).ToNot(HaveOccurred())

		testhelpers.TIMEOUT = 35 * time.Minute
		testhelpers.POLLING_INTERVAL = 10 * time.Second

		// fmt.Printf("<=================LINE 35=================>")
	})

	// TODO: figure out which commands are required for my purposes
	// How to get ssh key?
	Context("SoftLayer_VirtualGuestService#setUserMetadata and SoftLayer_VirtualGuestService#configureMetadataDisk", func() {
		It("creates ssh key, VirtualGuest, waits for it to be RUNNING, set user data, configures disk, verifies user data, and delete VG", func() {
			/*
			sshKeyPath := os.Getenv("SOFTLAYER_GO_TEST_SSH_KEY_PATH2")
			Expect(sshKeyPath).ToNot(Equal(""), "SOFTLAYER_GO_TEST_SSH_KEY_PATH2 env variable is not set")

			err = testhelpers.FindAndDeleteTestSshKeys()
			Expect(err).ToNot(HaveOccurred())

			createdSshKey := testhelpers.CreateTestSshKey(sshKeyPath)
			testhelpers.WaitForCreatedSshKeyToBePresent(createdSshKey.Id)

			virtualGuest := testhelpers.CreateVirtualGuestAndMarkItTest([]datatypes.SoftLayer_Security_Ssh_Key{createdSshKey})

			testhelpers.WaitForVirtualGuestToBeRunning(virtualGuest.Id)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)

			startTime := time.Now()
			userMetadata := "softlayer-go test fake metadata"
			transaction := testhelpers.SetUserMetadataAndConfigureDisk(virtualGuest.Id, userMetadata)
			averageTransactionDuration, err := time.ParseDuration(transaction.TransactionStatus.AverageDuration + "m")
			Ω(err).ShouldNot(HaveOccurred())

			testhelpers.WaitForVirtualGuest(virtualGuest.Id, "RUNNING", averageTransactionDuration)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)
			fmt.Printf("====> Set Metadata and configured disk on instance: %d in %d time\n", virtualGuest.Id, time.Since(startTime))

			sshKeyFilePath := os.Getenv("SOFTLAYER_GO_TEST_SSH_KEY_PATH2")
			Expect(sshKeyFilePath).ToNot(Equal(""), "SOFTLAYER_GO_TEST_SSH_KEY_PATH2 env variable is not set")

			testhelpers.TestUserMetadata(userMetadata, sshKeyFilePath)

			startTime = time.Now()
			userMetadata = "softlayer-go test MODIFIED fake metadata"
			testhelpers.SetUserMetadataAndConfigureDisk(virtualGuest.Id, userMetadata)

			testhelpers.WaitForVirtualGuest(virtualGuest.Id, "RUNNING", averageTransactionDuration)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)
			fmt.Printf("====> Set Metadata and configured disk on instance: %d in %d time\n", virtualGuest.Id, time.Since(startTime))

			testhelpers.TestUserMetadata(userMetadata, sshKeyFilePath)

			doesn't get here - not sure if this is the right function yet
			*/

			// Find way to get working directly, unless I don't have to run in command line
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(dir)
			runner.Run("/Users/jlo/workspace/golang/src/github.com/maximilien/bosh-softlayer-cpi/out/cpi", )

			// TODO: figure out cleanup steps
			// is this sufficient to delete the softlayer VM?
			/*
			testhelpers.DeleteVirtualGuest(virtualGuest.Id)
			testhelpers.DeleteSshKey(createdSshKey.Id)
			*/
		})
	})
})
