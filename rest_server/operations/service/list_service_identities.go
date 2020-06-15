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

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// ListServiceIdentitiesHandlerFunc turns a function with the right signature into a list service identities handler
type ListServiceIdentitiesHandlerFunc func(ListServiceIdentitiesParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn ListServiceIdentitiesHandlerFunc) Handle(params ListServiceIdentitiesParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// ListServiceIdentitiesHandler interface for that can handle valid list service identities params
type ListServiceIdentitiesHandler interface {
	Handle(ListServiceIdentitiesParams, interface{}) middleware.Responder
}

// NewListServiceIdentities creates a new http.Handler for the list service identities operation
func NewListServiceIdentities(ctx *middleware.Context, handler ListServiceIdentitiesHandler) *ListServiceIdentities {
	return &ListServiceIdentities{Context: ctx, Handler: handler}
}

/*ListServiceIdentities swagger:route GET /services/{id}/identities Service listServiceIdentities

List identities with access

Retrieves a list of identities that have access to this service. Supports filtering, sorting, and pagination. Requires admin access.


*/
type ListServiceIdentities struct {
	Context *middleware.Context
	Handler ListServiceIdentitiesHandler
}

func (o *ListServiceIdentities) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewListServiceIdentitiesParams()

	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
	}
	var principal interface{}
	if uprinc != nil {
		principal = uprinc
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}