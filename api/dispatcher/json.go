package dispatcher

import (
	"bytes"
	"encoding/json"

	bslcaction "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	bslcapi "github.com/cloudfoundry/bosh-softlayer-cpi/api"
	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"fmt"
	"strings"
)

const (
	jsonLogTag = "json"

	jsonCloudErrorType          = "Bosh::Clouds::CloudError"
	jsonCpiErrorType            = "Bosh::Clouds::CpiError"
	jsonNotImplementedErrorType = "Bosh::Clouds::NotImplemented"
)

type Request struct {
	Method    string        `json:"method"`
	Arguments []interface{} `json:"arguments"`

	// context key is ignored
}

type Response struct {
	Result interface{}    `json:"result"`
	Error  *ResponseError `json:"error"`

	Log string `json:"log"`
}

type ResponseError struct {
	Type    string `json:"type"`
	Message string `json:"message"`

	CanRetry bool `json:"ok_to_retry"`
}

type JSON struct {
	actionFactory bslcaction.Factory
	caller        Caller
	logger        boshlog.Logger
}

func NewJSON(
	actionFactory bslcaction.Factory,
	caller Caller,
	logger boshlog.Logger,
) JSON {
	return JSON{
		actionFactory: actionFactory,
		caller:        caller,
		logger:        logger,
	}
}

func (c JSON) Dispatch(reqBytes []byte) []byte {
	var req Request

	c.logger.DebugWithDetails(jsonLogTag, "Request bytes", string(reqBytes))
	digitalDecoder := json.NewDecoder(bytes.NewReader(reqBytes))
	digitalDecoder.UseNumber()
	if err := digitalDecoder.Decode(&req); err != nil {
		return c.buildCpiError("Must provide valid JSON payload")
	}

	c.logger.DebugWithDetails(jsonLogTag, "Deserialized request", req)

	// Will remove this line of code after we make sure LocalDiskFlag is defined properly in all deployment manifest files
	c.localDiskFlagNotSet(fmt.Sprintf("%s", req))

	if req.Method == "" {
		return c.buildCpiError("Must provide method key")
	}

	if req.Arguments == nil {
		return c.buildCpiError("Must provide arguments key")
	}

	action, err := c.actionFactory.Create(req.Method)
	if err != nil {
		return c.buildNotImplementedError()
	}

	result, err := c.caller.Call(action, req.Arguments)
	if err != nil {
		return c.buildCloudError(err)
	}

	resp := Response{
		Result: result,
	}

	c.logger.DebugWithDetails(jsonLogTag, "Deserialized response", resp)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return c.buildCpiError("Failed to serialize result")
	}

	c.logger.DebugWithDetails(jsonLogTag, "Response bytes", string(respBytes))

	return respBytes
}

func (c JSON) buildCloudError(err error) []byte {
	respErr := Response{
		Error: &ResponseError{},
	}

	if typedErr, ok := err.(bslcapi.CloudError); ok {
		respErr.Error.Type = typedErr.Type()
	} else {
		respErr.Error.Type = jsonCloudErrorType
	}

	respErr.Error.Message = err.Error()

	if typedErr, ok := err.(bslcapi.RetryableError); ok {
		respErr.Error.CanRetry = typedErr.CanRetry()
	}

	respErrBytes, err := json.Marshal(respErr)
	if err != nil {
		panic(err)
	}

	c.logger.DebugWithDetails(jsonLogTag, "CloudError response bytes", string(respErrBytes))

	return respErrBytes
}

func (c JSON) buildCpiError(message string) []byte {
	respErr := Response{
		Error: &ResponseError{
			Type:    jsonCpiErrorType,
			Message: message,
		},
	}

	respErrBytes, err := json.Marshal(respErr)
	if err != nil {
		panic(err)
	}

	c.logger.DebugWithDetails(jsonLogTag, "CpiError response bytes", string(respErrBytes))

	return respErrBytes
}

func (c JSON) buildNotImplementedError() []byte {
	respErr := Response{
		Error: &ResponseError{
			Type:    jsonNotImplementedErrorType,
			Message: "Must call implemented method",
		},
	}

	respErrBytes, err := json.Marshal(respErr)
	if err != nil {
		panic(err)
	}

	c.logger.DebugWithDetails(jsonLogTag, "NotImplementedError response bytes", string(respErrBytes))

	return respErrBytes
}

// This function is used to set the default value of localDiskFlag to true if it is not specified in the input json file.
// This function will be removed if we can ensure LocalDiskFlag property is set in place in all deployment manifest
func (c JSON) localDiskFlagNotSet(reqString string) {
	if strings.Contains(strings.ToUpper(reqString), strings.ToUpper("localDiskFlag")) {
		slhelper.LocalDiskFlagNotSet = false
	} else {
		slhelper.LocalDiskFlagNotSet = true
	}
}
