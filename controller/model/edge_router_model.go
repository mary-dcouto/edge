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

package model

import (
	"github.com/openziti/edge/controller/persistence"
	"github.com/openziti/fabric/controller/db"
	"github.com/openziti/fabric/controller/models"
	"github.com/openziti/foundation/common"
	"github.com/openziti/foundation/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"reflect"
)

type EdgeRouter struct {
	models.BaseEntity
	Name                string
	RoleAttributes      []string
	IsVerified          bool
	Fingerprint         *string
	CertPem             *string
	Hostname            *string
	EdgeRouterProtocols map[string]string
	VersionInfo         *common.VersionInfo
}

func (entity *EdgeRouter) toBoltEntityForCreate(*bbolt.Tx, Handler) (boltz.Entity, error) {
	boltEntity := &persistence.EdgeRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
		},
		RoleAttributes: entity.RoleAttributes,
		IsVerified:     false,
	}

	return boltEntity, nil
}

func (entity *EdgeRouter) toBoltEntityForUpdate(_ *bbolt.Tx, _ Handler) (boltz.Entity, error) {
	return &persistence.EdgeRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
			Fingerprint:   entity.Fingerprint,
		},
		RoleAttributes:      entity.RoleAttributes,
		IsVerified:          entity.IsVerified,
		CertPem:             entity.CertPem,
		Hostname:            entity.Hostname,
		EdgeRouterProtocols: entity.EdgeRouterProtocols,
	}, nil
}

func (entity *EdgeRouter) toBoltEntityForPatch(tx *bbolt.Tx, handler Handler) (boltz.Entity, error) {
	return entity.toBoltEntityForUpdate(tx, handler)
}

func (entity *EdgeRouter) fillFrom(_ Handler, _ *bbolt.Tx, boltEntity boltz.Entity) error {
	boltEdgeRouter, ok := boltEntity.(*persistence.EdgeRouter)
	if !ok {
		return errors.Errorf("unexpected type %v when filling model edge router", reflect.TypeOf(boltEntity))
	}

	entity.FillCommon(boltEdgeRouter)
	entity.Name = boltEdgeRouter.Name
	entity.RoleAttributes = boltEdgeRouter.RoleAttributes
	entity.IsVerified = boltEdgeRouter.IsVerified
	entity.Fingerprint = boltEdgeRouter.Fingerprint
	entity.CertPem = boltEdgeRouter.CertPem
	entity.Hostname = boltEdgeRouter.Hostname
	entity.EdgeRouterProtocols = boltEdgeRouter.EdgeRouterProtocols

	return nil
}
