// Code generated by go-swagger; DO NOT EDIT.

package tx_status_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/qubic/integration/go-clients/tx-status/models"
)

// TxStatusServiceGetTransactionStatusReader is a Reader for the TxStatusServiceGetTransactionStatus structure.
type TxStatusServiceGetTransactionStatusReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *TxStatusServiceGetTransactionStatusReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewTxStatusServiceGetTransactionStatusOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewTxStatusServiceGetTransactionStatusDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewTxStatusServiceGetTransactionStatusOK creates a TxStatusServiceGetTransactionStatusOK with default headers values
func NewTxStatusServiceGetTransactionStatusOK() *TxStatusServiceGetTransactionStatusOK {
	return &TxStatusServiceGetTransactionStatusOK{}
}

/*
TxStatusServiceGetTransactionStatusOK describes a response with status code 200, with default header values.

A successful response.
*/
type TxStatusServiceGetTransactionStatusOK struct {
	Payload *models.PbGetTransactionStatusResponse
}

// IsSuccess returns true when this tx status service get transaction status o k response has a 2xx status code
func (o *TxStatusServiceGetTransactionStatusOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this tx status service get transaction status o k response has a 3xx status code
func (o *TxStatusServiceGetTransactionStatusOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this tx status service get transaction status o k response has a 4xx status code
func (o *TxStatusServiceGetTransactionStatusOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this tx status service get transaction status o k response has a 5xx status code
func (o *TxStatusServiceGetTransactionStatusOK) IsServerError() bool {
	return false
}

// IsCode returns true when this tx status service get transaction status o k response a status code equal to that given
func (o *TxStatusServiceGetTransactionStatusOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the tx status service get transaction status o k response
func (o *TxStatusServiceGetTransactionStatusOK) Code() int {
	return 200
}

func (o *TxStatusServiceGetTransactionStatusOK) Error() string {
	return fmt.Sprintf("[GET /tx-status/{txId}][%d] txStatusServiceGetTransactionStatusOK  %+v", 200, o.Payload)
}

func (o *TxStatusServiceGetTransactionStatusOK) String() string {
	return fmt.Sprintf("[GET /tx-status/{txId}][%d] txStatusServiceGetTransactionStatusOK  %+v", 200, o.Payload)
}

func (o *TxStatusServiceGetTransactionStatusOK) GetPayload() *models.PbGetTransactionStatusResponse {
	return o.Payload
}

func (o *TxStatusServiceGetTransactionStatusOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.PbGetTransactionStatusResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewTxStatusServiceGetTransactionStatusDefault creates a TxStatusServiceGetTransactionStatusDefault with default headers values
func NewTxStatusServiceGetTransactionStatusDefault(code int) *TxStatusServiceGetTransactionStatusDefault {
	return &TxStatusServiceGetTransactionStatusDefault{
		_statusCode: code,
	}
}

/*
TxStatusServiceGetTransactionStatusDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type TxStatusServiceGetTransactionStatusDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this tx status service get transaction status default response has a 2xx status code
func (o *TxStatusServiceGetTransactionStatusDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this tx status service get transaction status default response has a 3xx status code
func (o *TxStatusServiceGetTransactionStatusDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this tx status service get transaction status default response has a 4xx status code
func (o *TxStatusServiceGetTransactionStatusDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this tx status service get transaction status default response has a 5xx status code
func (o *TxStatusServiceGetTransactionStatusDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this tx status service get transaction status default response a status code equal to that given
func (o *TxStatusServiceGetTransactionStatusDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the tx status service get transaction status default response
func (o *TxStatusServiceGetTransactionStatusDefault) Code() int {
	return o._statusCode
}

func (o *TxStatusServiceGetTransactionStatusDefault) Error() string {
	return fmt.Sprintf("[GET /tx-status/{txId}][%d] TxStatusService_GetTransactionStatus default  %+v", o._statusCode, o.Payload)
}

func (o *TxStatusServiceGetTransactionStatusDefault) String() string {
	return fmt.Sprintf("[GET /tx-status/{txId}][%d] TxStatusService_GetTransactionStatus default  %+v", o._statusCode, o.Payload)
}

func (o *TxStatusServiceGetTransactionStatusDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *TxStatusServiceGetTransactionStatusDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}