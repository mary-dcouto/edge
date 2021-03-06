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
	"fmt"
	"github.com/openziti/edge/controller/persistence"
	"github.com/openziti/fabric/controller/models"
	"github.com/openziti/foundation/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"reflect"
)

type PostureCheck struct {
	models.BaseEntity
	Name           string
	TypeId         string
	Version        int64
	RoleAttributes []string
	SubType        PostureCheckSubType
}

type PostureCheckSubType interface {
	toBoltEntityForCreate(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error)
	toBoltEntityForUpdate(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error)
	toBoltEntityForPatch(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error)
	fillFrom(handler Handler, tx *bbolt.Tx, check *persistence.PostureCheck, subType persistence.PostureCheckSubType) error
	Evaluate(pd *PostureData) bool
}

type newPostureCheckSubType func() PostureCheckSubType

const (
	PostureCheckTypeOs      = "OS"
	PostureCheckTypeDomain  = "DOMAIN"
	PostureCheckTypeProcess = "PROCESS"
	PostureCheckTypeMAC     = "MAC"
)

var postureCheckSubTypeMap = map[string]newPostureCheckSubType{
	PostureCheckTypeOs:      newPostureCheckOperatingSystem,
	PostureCheckTypeDomain:  newPostureCheckWindowsDomains,
	PostureCheckTypeProcess: newPostureCheckProcess,
	PostureCheckTypeMAC:     newPostureCheckMacAddresses,
}

func newSubType(typeId string) PostureCheckSubType {
	if factory, ok := postureCheckSubTypeMap[typeId]; ok {
		return factory()
	}
	return nil
}

func (entity *PostureCheck) fillFrom(handler Handler, tx *bbolt.Tx, boltEntity boltz.Entity) error {
	boltPostureCheck, ok := boltEntity.(*persistence.PostureCheck)
	if !ok {
		return errors.Errorf("unexpected type %v when filling model posture check", reflect.TypeOf(boltEntity))
	}
	entity.FillCommon(boltPostureCheck)
	entity.Name = boltPostureCheck.Name
	entity.TypeId = boltPostureCheck.TypeId
	entity.Version = boltPostureCheck.Version
	entity.RoleAttributes = boltPostureCheck.RoleAttributes

	subType := newSubType(entity.TypeId)

	if subType == nil {
		return fmt.Errorf("cannot create posture check subtype [%v]", entity.TypeId)
	}

	if err := subType.fillFrom(handler, tx, boltPostureCheck, boltPostureCheck.SubType); err != nil {
		return fmt.Errorf("error filling posture check subType [%v]: %v", entity.TypeId, err)
	}

	entity.SubType = subType

	return nil
}

func (entity *PostureCheck) toBoltEntityForCreate(tx *bbolt.Tx, handler Handler) (boltz.Entity, error) {
	boltEntity := &persistence.PostureCheck{
		BaseExtEntity:  *boltz.NewExtEntity(entity.Id, entity.Tags),
		Name:           entity.Name,
		TypeId:         entity.TypeId,
		Version:        1,
		RoleAttributes: entity.RoleAttributes,
	}

	var err error
	if boltEntity.SubType, err = entity.SubType.toBoltEntityForCreate(tx, handler); err != nil {
		return nil, fmt.Errorf("error converting to bolt posture check subType [%v] for create: %v", entity.TypeId, err)
	}

	return boltEntity, nil
}

func (entity *PostureCheck) toBoltEntityForUpdate(tx *bbolt.Tx, handler Handler) (boltz.Entity, error) {
	return entity.toBoltEntityForCreate(tx, handler)
}

func (entity *PostureCheck) toBoltEntityForPatch(tx *bbolt.Tx, handler Handler) (boltz.Entity, error) {
	return entity.toBoltEntityForCreate(tx, handler)
}

func (entity *PostureCheck) Evaluate(pd *PostureData) bool {
	return entity.SubType.Evaluate(pd)
}
