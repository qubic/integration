// Code generated by go-swagger; DO NOT EDIT.

package tx_status_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewTxStatusServiceGetBlockHeightParams creates a new TxStatusServiceGetBlockHeightParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewTxStatusServiceGetBlockHeightParams() *TxStatusServiceGetBlockHeightParams {
	return &TxStatusServiceGetBlockHeightParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewTxStatusServiceGetBlockHeightParamsWithTimeout creates a new TxStatusServiceGetBlockHeightParams object
// with the ability to set a timeout on a request.
func NewTxStatusServiceGetBlockHeightParamsWithTimeout(timeout time.Duration) *TxStatusServiceGetBlockHeightParams {
	return &TxStatusServiceGetBlockHeightParams{
		timeout: timeout,
	}
}

// NewTxStatusServiceGetBlockHeightParamsWithContext creates a new TxStatusServiceGetBlockHeightParams object
// with the ability to set a context for a request.
func NewTxStatusServiceGetBlockHeightParamsWithContext(ctx context.Context) *TxStatusServiceGetBlockHeightParams {
	return &TxStatusServiceGetBlockHeightParams{
		Context: ctx,
	}
}

// NewTxStatusServiceGetBlockHeightParamsWithHTTPClient creates a new TxStatusServiceGetBlockHeightParams object
// with the ability to set a custom HTTPClient for a request.
func NewTxStatusServiceGetBlockHeightParamsWithHTTPClient(client *http.Client) *TxStatusServiceGetBlockHeightParams {
	return &TxStatusServiceGetBlockHeightParams{
		HTTPClient: client,
	}
}

/*
TxStatusServiceGetBlockHeightParams contains all the parameters to send to the API endpoint

	for the tx status service get block height operation.

	Typically these are written to a http.Request.
*/
type TxStatusServiceGetBlockHeightParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the tx status service get block height params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *TxStatusServiceGetBlockHeightParams) WithDefaults() *TxStatusServiceGetBlockHeightParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the tx status service get block height params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *TxStatusServiceGetBlockHeightParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) WithTimeout(timeout time.Duration) *TxStatusServiceGetBlockHeightParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) WithContext(ctx context.Context) *TxStatusServiceGetBlockHeightParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) WithHTTPClient(client *http.Client) *TxStatusServiceGetBlockHeightParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the tx status service get block height params
func (o *TxStatusServiceGetBlockHeightParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *TxStatusServiceGetBlockHeightParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}