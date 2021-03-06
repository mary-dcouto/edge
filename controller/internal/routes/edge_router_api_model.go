/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package routes

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge/controller/env"
	"github.com/openziti/edge/controller/model"
	"github.com/openziti/edge/controller/response"
	"github.com/openziti/edge/rest_model"
	"github.com/openziti/fabric/controller/models"
	"github.com/openziti/foundation/util/stringz"
	"strings"
)

const (
	EntityNameEdgeRouter = "edge-routers"
)

var EdgeRouterLinkFactory = NewEdgeRouterLinkFactory()

type EdgeRouterLinkFactoryImpl struct {
	BasicLinkFactory
}

func NewEdgeRouterLinkFactory() *EdgeRouterLinkFactoryImpl {
	return &EdgeRouterLinkFactoryImpl{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameEdgeRouter),
	}
}

func (factory *EdgeRouterLinkFactoryImpl) Links(entity models.Entity) rest_model.Links {
	links := factory.BasicLinkFactory.Links(entity)
	links[EntityNameEdgeRouterPolicy] = factory.NewNestedLink(entity, EntityNameEdgeRouterPolicy)
	return links
}

func MapCreateEdgeRouterToModel(router *rest_model.EdgeRouterCreate) *model.EdgeRouter {
	ret := &model.EdgeRouter{
		BaseEntity: models.BaseEntity{
			Tags: router.Tags,
		},
		Name:           stringz.OrEmpty(router.Name),
		RoleAttributes: router.RoleAttributes,
	}

	return ret
}

func MapUpdateEdgeRouterToModel(id string, router *rest_model.EdgeRouterUpdate) *model.EdgeRouter {
	ret := &model.EdgeRouter{
		BaseEntity: models.BaseEntity{
			Tags: router.Tags,
			Id:   id,
		},
		Name:           stringz.OrEmpty(router.Name),
		RoleAttributes: router.RoleAttributes,
	}

	return ret
}

func MapPatchEdgeRouterToModel(id string, router *rest_model.EdgeRouterPatch) *model.EdgeRouter {
	ret := &model.EdgeRouter{
		BaseEntity: models.BaseEntity{
			Tags: router.Tags,
			Id:   id,
		},
		Name:           router.Name,
		RoleAttributes: router.RoleAttributes,
	}

	return ret
}

func MapEdgeRouterToRestEntity(ae *env.AppEnv, _ *response.RequestContext, e models.Entity) (interface{}, error) {
	router, ok := e.(*model.EdgeRouter)

	if !ok {
		err := fmt.Errorf("entity is not a EdgeRouter \"%s\"", e.GetId())
		log := pfxlog.Logger()
		log.Error(err)
		return nil, err
	}

	restModel, err := MapEdgeRouterToRestModel(ae, router)

	if err != nil {
		err := fmt.Errorf("could not convert to API entity \"%s\": %s", e.GetId(), err)
		log := pfxlog.Logger()
		log.Error(err)
		return nil, err
	}
	return restModel, nil
}

func MapEdgeRouterToRestModel(ae *env.AppEnv, router *model.EdgeRouter) (*rest_model.EdgeRouterDetail, error) {
	hostname := ""
	protocols := map[string]string{}
	os := ""
	arch := ""
	buildDate := ""
	revision := ""
	version := ""

	onlineEdgeRouter := ae.Broker.GetOnlineEdgeRouter(router.Id)

	isOnline := onlineEdgeRouter != nil

	if isOnline {
		hostname = *onlineEdgeRouter.Hostname
		protocols = onlineEdgeRouter.EdgeRouterProtocols

		if onlineEdgeRouter.VersionInfo != nil {
			os = onlineEdgeRouter.VersionInfo.OS
			arch = onlineEdgeRouter.VersionInfo.Arch
			buildDate = onlineEdgeRouter.VersionInfo.BuildDate
			revision = onlineEdgeRouter.VersionInfo.Revision
			version = onlineEdgeRouter.VersionInfo.Version
		}
	}

	ret := &rest_model.EdgeRouterDetail{
		BaseEntity:          BaseEntityToRestModel(router, EdgeRouterLinkFactory),
		Name:                &router.Name,
		RoleAttributes:      router.RoleAttributes,
		EnrollmentToken:     nil,
		EnrollmentCreatedAt: nil,
		EnrollmentExpiresAt: nil,
		EnrollmentJwt:       nil,
		IsOnline:            &isOnline,
		IsVerified:          &router.IsVerified,
		Fingerprint:         stringz.OrEmpty(router.Fingerprint),
		Hostname:            &hostname,
		SupportedProtocols:  protocols,
		VersionInfo: &rest_model.VersionInfo{
			Os:       &os,
			Arch:     &arch,
			Revision: &revision,
			BuilDate: &buildDate,
			Version:  &version,
		},
	}

	if !router.IsVerified {
		var enrollments []*model.Enrollment

		err := ae.GetHandlers().EdgeRouter.CollectEnrollments(router.Id, func(entity *model.Enrollment) error {
			enrollments = append(enrollments, entity)
			return nil
		})

		if err != nil {
			return nil, err
		}

		if len(enrollments) != 1 {
			return nil, fmt.Errorf("expected enrollment not found for unverified edge router %s", router.Id)
		}
		enrollment := enrollments[0]

		createdAt := strfmt.DateTime(enrollment.CreatedAt)
		expiresAt := strfmt.DateTime(*enrollment.ExpiresAt)

		ret.EnrollmentExpiresAt = &expiresAt
		ret.EnrollmentCreatedAt = &createdAt
		ret.EnrollmentJwt = &enrollment.Jwt
		ret.EnrollmentToken = &enrollment.Token
	}

	return ret, nil
}

func GetNamedEdgeRouterRoles(edgeRouterHandler *model.EdgeRouterHandler, roles []string) rest_model.NamedRoles {
	result := rest_model.NamedRoles{}
	for _, role := range roles {
		if strings.HasPrefix(role, "@") {

			edgeRouter, err := edgeRouterHandler.Read(role[1:])
			if err != nil {
				pfxlog.Logger().Errorf("error converting edgeRouter role [%s] to a named role: %v", role, err)
				continue
			}

			result = append(result, &rest_model.NamedRole{
				Role: role,
				Name: "@" + edgeRouter.Name,
			})
		} else {
			result = append(result, &rest_model.NamedRole{
				Role: role,
				Name: role,
			})
		}

	}
	return result
}
