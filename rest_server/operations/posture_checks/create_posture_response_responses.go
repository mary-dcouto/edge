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

package posture_checks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/edge/rest_model"
)

// CreatePostureResponseOKCode is the HTTP code returned for type CreatePostureResponseOK
const CreatePostureResponseOKCode int = 200

/*CreatePostureResponseOK Base empty response

swagger:response createPostureResponseOK
*/
type CreatePostureResponseOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model.Empty `json:"body,omitempty"`
}

// NewCreatePostureResponseOK creates CreatePostureResponseOK with default headers values
func NewCreatePostureResponseOK() *CreatePostureResponseOK {

	return &CreatePostureResponseOK{}
}

// WithPayload adds the payload to the create posture response o k response
func (o *CreatePostureResponseOK) WithPayload(payload *rest_model.Empty) *CreatePostureResponseOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create posture response o k response
func (o *CreatePostureResponseOK) SetPayload(payload *rest_model.Empty) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreatePostureResponseOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreatePostureResponseBadRequestCode is the HTTP code returned for type CreatePostureResponseBadRequest
const CreatePostureResponseBadRequestCode int = 400

/*CreatePostureResponseBadRequest The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information

swagger:response createPostureResponseBadRequest
*/
type CreatePostureResponseBadRequest struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewCreatePostureResponseBadRequest creates CreatePostureResponseBadRequest with default headers values
func NewCreatePostureResponseBadRequest() *CreatePostureResponseBadRequest {

	return &CreatePostureResponseBadRequest{}
}

// WithPayload adds the payload to the create posture response bad request response
func (o *CreatePostureResponseBadRequest) WithPayload(payload *rest_model.APIErrorEnvelope) *CreatePostureResponseBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create posture response bad request response
func (o *CreatePostureResponseBadRequest) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreatePostureResponseBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreatePostureResponseUnauthorizedCode is the HTTP code returned for type CreatePostureResponseUnauthorized
const CreatePostureResponseUnauthorizedCode int = 401

/*CreatePostureResponseUnauthorized The currently supplied session does not have the correct access rights to request this resource

swagger:response createPostureResponseUnauthorized
*/
type CreatePostureResponseUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewCreatePostureResponseUnauthorized creates CreatePostureResponseUnauthorized with default headers values
func NewCreatePostureResponseUnauthorized() *CreatePostureResponseUnauthorized {

	return &CreatePostureResponseUnauthorized{}
}

// WithPayload adds the payload to the create posture response unauthorized response
func (o *CreatePostureResponseUnauthorized) WithPayload(payload *rest_model.APIErrorEnvelope) *CreatePostureResponseUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create posture response unauthorized response
func (o *CreatePostureResponseUnauthorized) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreatePostureResponseUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
