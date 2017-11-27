package integration

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshot", func() {

	It("executes the disk snapshot lifecycle", func() {
		var diskCID, snapshotCID string
		By("creating a disk")
		request := fmt.Sprintf(`{
		  "method": "create_disk",
		  "arguments": [20480, {"datacenter": "lon02", "snapshot_space": 5, "iops": 500}, 0]
		}`)
		diskCID = assertSucceedsWithResult(request).(string)
		diskID, err := strconv.Atoi(diskCID)
		Expect(err).NotTo(HaveOccurred())

		By("snapshotting the disk")
		request = fmt.Sprintf(`{
		  "method": "snapshot_disk",
		  "arguments": [%d, {}]
		}`, diskID)
		snapshotCID = assertSucceedsWithResult(request).(string)
		snapshotID, err := strconv.Atoi(snapshotCID)
		Expect(err).NotTo(HaveOccurred())

		By("deleting the snapshot")
		request = fmt.Sprintf(`{
			  "method": "delete_snapshot",
			  "arguments": [%d]
			}`, snapshotID)
		assertSucceeds(request)

		By("creating a snapshot with metadata")
		request = fmt.Sprintf(`{
			  "method": "snapshot_disk",
			  "arguments": [%d, {"deployment": "integration_test", "job": "creating_snapshot", "index": "job_index"}]
			}`, diskID)
		snapshotCID = assertSucceedsWithResult(request).(string)
		snapshotID, err = strconv.Atoi(snapshotCID)
		Expect(err).NotTo(HaveOccurred())

		By("deleting the snapshot")
		request = fmt.Sprintf(`{
			  "method": "delete_snapshot",
			  "arguments": [%d]
			}`, snapshotID)
		assertSucceeds(request)

		By("deleting the disk")
		request = fmt.Sprintf(`{
			  "method": "delete_disk",
			  "arguments": [%d]
			}`, diskID)
		assertSucceeds(request)
	})
})
