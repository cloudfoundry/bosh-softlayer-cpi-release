package integration

import (
	//. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func assertSucceeds(request string) {
	response, err := execCPI(request)
	Expect(err).ToNot(HaveOccurred())
	Expect(response.Error).To(BeNil())
}

func assertFails(request string) error {
	response, err := execCPI(request)
	Expect(err).To(HaveOccurred())
	Expect(response.Error).ToNot(BeNil())
	return response.Error
}

func assertSucceedsWithResult(request string) interface{} {
	response, err := execCPI(request)
	Expect(err).ToNot(HaveOccurred())
	Expect(response.Error).To(BeNil())
	Expect(response.Result).ToNot(BeNil())
	return response.Result
}

func assertSucceedsWithResultOrCatchCapacityError(request string) interface{} {
	response, err := execCPI(request)
	Expect(err).ToNot(HaveOccurred())
	if response.Error != nil {
		Expect(response.Error.Error()).To(ContainSubstring("There is insufficient capacity to complete the request."))
		Expect(response.Error).To(BeNil())
		return response.Result
	} else {
		Expect(response.Result).ToNot(BeNil())
		return response.Result
	}
}

func toStringArray(raw []interface{}) []string {
	strings := make([]string, len(raw), len(raw))
	for i := range raw {
		strings[i] = raw[i].(string)
	}
	return strings
}
