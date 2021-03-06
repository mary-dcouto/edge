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
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/edge/rest_model"
)

// CreateAuthenticatorOKCode is the HTTP code returned for type CreateAuthenticatorOK
const CreateAuthenticatorOKCode int = 200

/*CreateAuthenticatorOK The create was successful

swagger:response createAuthenticatorOK
*/
type CreateAuthenticatorOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model.AuthenticatorCreate `json:"body,omitempty"`
}

// NewCreateAuthenticatorOK creates CreateAuthenticatorOK with default headers values
func NewCreateAuthenticatorOK() *CreateAuthenticatorOK {

	return &CreateAuthenticatorOK{}
}

// WithPayload adds the payload to the create authenticator o k response
func (o *CreateAuthenticatorOK) WithPayload(payload *rest_model.AuthenticatorCreate) *CreateAuthenticatorOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create authenticator o k response
func (o *CreateAuthenticatorOK) SetPayload(payload *rest_model.AuthenticatorCreate) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAuthenticatorOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAuthenticatorBadRequestCode is the HTTP code returned for type CreateAuthenticatorBadRequest
const CreateAuthenticatorBadRequestCode int = 400

/*CreateAuthenticatorBadRequest The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information

swagger:response createAuthenticatorBadRequest
*/
type CreateAuthenticatorBadRequest struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewCreateAuthenticatorBadRequest creates CreateAuthenticatorBadRequest with default headers values
func NewCreateAuthenticatorBadRequest() *CreateAuthenticatorBadRequest {

	return &CreateAuthenticatorBadRequest{}
}

// WithPayload adds the payload to the create authenticator bad request response
func (o *CreateAuthenticatorBadRequest) WithPayload(payload *rest_model.APIErrorEnvelope) *CreateAuthenticatorBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create authenticator bad request response
func (o *CreateAuthenticatorBadRequest) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAuthenticatorBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAuthenticatorUnauthorizedCode is the HTTP code returned for type CreateAuthenticatorUnauthorized
const CreateAuthenticatorUnauthorizedCode int = 401

/*CreateAuthenticatorUnauthorized The currently supplied session does not have the correct access rights to request this resource

swagger:response createAuthenticatorUnauthorized
*/
type CreateAuthenticatorUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewCreateAuthenticatorUnauthorized creates CreateAuthenticatorUnauthorized with default headers values
func NewCreateAuthenticatorUnauthorized() *CreateAuthenticatorUnauthorized {

	return &CreateAuthenticatorUnauthorized{}
}

// WithPayload adds the payload to the create authenticator unauthorized response
func (o *CreateAuthenticatorUnauthorized) WithPayload(payload *rest_model.APIErrorEnvelope) *CreateAuthenticatorUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create authenticator unauthorized response
func (o *CreateAuthenticatorUnauthorized) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAuthenticatorUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
