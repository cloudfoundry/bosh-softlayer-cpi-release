package action_test

import (
	. "bosh-softlayer-cpi/action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		legalIDBytes []byte
		illegaBytes  []byte

		stemcellCID StemcellCID
		vmCID       VMCID
		diskCID     DiskCID
		snapshotCID SnapshotCID
	)

	BeforeEach(func() {
		legalIDBytes = []byte("12345678")
		illegaBytes = []byte("12vi345678")
	})

	Describe("StemcellCID", func() {
		It("UnmarshalJSON if UnmarshalJSON successfully", func() {
			err := stemcellCID.UnmarshalJSON(legalIDBytes)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if strconv.Atoi occur invalid syntax", func() {
			err := stemcellCID.UnmarshalJSON(illegaBytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("MarshalJSON successful", func() {
			stemcellCID = 12345678
			_, err := stemcellCID.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("VMCID", func() {
		It("does not return error if UnmarshalJSON successfully", func() {

			err := vmCID.UnmarshalJSON(legalIDBytes)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if strconv.Atoi occur invalid syntax", func() {
			err := vmCID.UnmarshalJSON(illegaBytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("MarshalJSON successful", func() {
			vmCID = 12345678
			_, err := vmCID.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("DiskCID", func() {
		It("does not return error if UnmarshalJSON successfully", func() {
			err := diskCID.UnmarshalJSON(legalIDBytes)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if strconv.Atoi occur invalid syntax", func() {
			err := diskCID.UnmarshalJSON(illegaBytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("MarshalJSON successful", func() {
			diskCID = 12345678
			_, err := diskCID.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("snapshotCID", func() {
		It("does not return error if UnmarshalJSON successfully", func() {
			err := snapshotCID.UnmarshalJSON(legalIDBytes)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if strconv.Atoi occur invalid syntax", func() {
			err := snapshotCID.UnmarshalJSON(illegaBytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("MarshalJSON successful", func() {
			snapshotCID = 12345678
			_, err := snapshotCID.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
