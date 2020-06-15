// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package authenticator

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/edge/rest_model"
)

// CreateAuthenticatorReader is a Reader for the CreateAuthenticator structure.
type CreateAuthenticatorReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateAuthenticatorReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewCreateAuthenticatorOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreateAuthenticatorBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewCreateAuthenticatorUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewCreateAuthenticatorOK creates a CreateAuthenticatorOK with default headers values
func NewCreateAuthenticatorOK() *CreateAuthenticatorOK {
	return &CreateAuthenticatorOK{}
}

/*CreateAuthenticatorOK handles this case with default header values.

The create was successful
*/
type CreateAuthenticatorOK struct {
	Payload *rest_model.AuthenticatorCreate
}

func (o *CreateAuthenticatorOK) Error() string {
	return fmt.Sprintf("[POST /authenticators][%d] createAuthenticatorOK  %+v", 200, o.Payload)
}

func (o *CreateAuthenticatorOK) GetPayload() *rest_model.AuthenticatorCreate {
	return o.Payload
}

func (o *CreateAuthenticatorOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.AuthenticatorCreate)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateAuthenticatorBadRequest creates a CreateAuthenticatorBadRequest with default headers values
func NewCreateAuthenticatorBadRequest() *CreateAuthenticatorBadRequest {
	return &CreateAuthenticatorBadRequest{}
}

/*CreateAuthenticatorBadRequest handles this case with default header values.

The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information
*/
type CreateAuthenticatorBadRequest struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *CreateAuthenticatorBadRequest) Error() string {
	return fmt.Sprintf("[POST /authenticators][%d] createAuthenticatorBadRequest  %+v", 400, o.Payload)
}

func (o *CreateAuthenticatorBadRequest) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *CreateAuthenticatorBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateAuthenticatorUnauthorized creates a CreateAuthenticatorUnauthorized with default headers values
func NewCreateAuthenticatorUnauthorized() *CreateAuthenticatorUnauthorized {
	return &CreateAuthenticatorUnauthorized{}
}

/*CreateAuthenticatorUnauthorized handles this case with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type CreateAuthenticatorUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *CreateAuthenticatorUnauthorized) Error() string {
	return fmt.Sprintf("[POST /authenticators][%d] createAuthenticatorUnauthorized  %+v", 401, o.Payload)
}

func (o *CreateAuthenticatorUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *CreateAuthenticatorUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}