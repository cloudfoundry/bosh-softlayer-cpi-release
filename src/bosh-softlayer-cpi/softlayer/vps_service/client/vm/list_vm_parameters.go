package vm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewListVMParams creates a new ListVMParams object
// with the default values initialized.
func NewListVMParams() *ListVMParams {

	return &ListVMParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewListVMParamsWithTimeout creates a new ListVMParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewListVMParamsWithTimeout(timeout time.Duration) *ListVMParams {

	return &ListVMParams{

		timeout: timeout,
	}
}

// NewListVMParamsWithContext creates a new ListVMParams object
// with the default values initialized, and the ability to set a context for a request
func NewListVMParamsWithContext(ctx context.Context) *ListVMParams {

	return &ListVMParams{

		Context: ctx,
	}
}

/*ListVMParams contains all the parameters to send to the API endpoint
for the list Vm operation typically these are written to a http.Request
*/
type ListVMParams struct {
	timeout time.Duration
	Context context.Context
}

// WithTimeout adds the timeout to the list Vm params
func (o *ListVMParams) WithTimeout(timeout time.Duration) *ListVMParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list Vm params
func (o *ListVMParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list Vm params
func (o *ListVMParams) WithContext(ctx context.Context) *ListVMParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list Vm params
func (o *ListVMParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WriteToRequest writes these params to a swagger request
func (o *ListVMParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
