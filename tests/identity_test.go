// +build apitests

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

package tests

import (
	"github.com/Jeffail/gabs"
	"github.com/openziti/edge/eid"
	"github.com/openziti/foundation/util/stringz"
	"net/http"
	"net/url"
	"sort"
	"testing"
)

func Test_Identity(t *testing.T) {
	ctx := NewTestContext(t)
	defer ctx.Teardown()
	ctx.StartServer()
	ctx.RequireAdminLogin()

	t.Run("role attributes should be created", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		identity := newTestIdentity(false, role1, role2)
		identity.Id = ctx.AdminSession.requireCreateEntity(identity)
		ctx.AdminSession.validateEntityWithQuery(identity)
		ctx.AdminSession.validateEntityWithLookup(identity)
	})

	t.Run("role attributes should be updated", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		identity := newTestIdentity(false, role1, role2)
		identity.Id = ctx.AdminSession.requireCreateEntity(identity)

		role3 := eid.New()
		identity.roleAttributes = []string{role2, role3}
		ctx.AdminSession.requireUpdateEntity(identity)
		ctx.AdminSession.validateEntityWithLookup(identity)
	})

	t.Run("role attributes should be queryable", func(t *testing.T) {
		ctx.testContextChanged(t)
		prefix := "rol3-attribut3-qu3ry-t3st-"
		role1 := prefix + "sales"
		role2 := prefix + "support"
		role3 := prefix + "engineering"
		role4 := prefix + "field-ops"
		role5 := prefix + "executive"

		ctx.AdminSession.requireNewIdentity(false, role1, role2)
		ctx.AdminSession.requireNewIdentity(false, role2, role3)
		ctx.AdminSession.requireNewIdentity(false, role3, role4)
		identity := ctx.AdminSession.requireNewIdentity(false, role5)
		ctx.AdminSession.requireNewIdentity(false)

		list := ctx.AdminSession.requireList("identity-role-attributes")
		ctx.Req.True(len(list) >= 5)
		ctx.Req.True(stringz.ContainsAll(list, role1, role2, role3, role4, role5))

		filter := url.QueryEscape(`id contains "e" and id contains "` + prefix + `" sort by id`)
		list = ctx.AdminSession.requireList("identity-role-attributes?filter=" + filter)
		ctx.Req.Equal(4, len(list))

		expected := []string{role1, role3, role4, role5}
		sort.Strings(expected)
		ctx.Req.Equal(expected, list)

		identity.roleAttributes = nil
		ctx.AdminSession.requireUpdateEntity(identity)
		list = ctx.AdminSession.requireList("identity-role-attributes")
		ctx.Req.True(len(list) >= 4)
		ctx.Req.True(stringz.ContainsAll(list, role1, role2, role3, role4))
		ctx.Req.False(stringz.Contains(list, role5))
	})

	t.Run("update (PUT) an identity", func(t *testing.T) {
		ctx.testContextChanged(t)
		enrolledId, _ := ctx.AdminSession.requireCreateIdentityOttEnrollment(eid.New(), false)
		enrolledIdentity := ctx.AdminSession.requireQuery("identities/" + enrolledId)

		unenrolledId := ctx.AdminSession.requireCreateIdentityOttEnrollmentUnfinished(eid.New(), false)
		unenrolledIdentity := ctx.AdminSession.requireQuery("identities/" + unenrolledId)

		t.Run("should not alter authenticators", func(t *testing.T) {
			ctx.testContextChanged(t)
			updateContent := gabs.New()
			_, _ = updateContent.SetP(eid.New(), "name")
			_, _ = updateContent.SetP("Device", "type")
			_, _ = updateContent.SetP(map[string]interface{}{}, "tags")
			_, _ = updateContent.SetP(false, "isAdmin")

			resp := ctx.AdminSession.updateEntityOfType(enrolledId, "identities", updateContent.String(), false)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			updatedIdentity := ctx.AdminSession.requireQuery("identities/" + enrolledId)

			data := enrolledIdentity.Path("data.authenticators").Data()
			expectedAuths := data.(map[string]interface{})
			ctx.Req.NotEmpty(expectedAuths)

			updatedAuths := updatedIdentity.Path("data.authenticators").Data().(map[string]interface{})
			ctx.Req.NotEmpty(updatedAuths)
			ctx.Req.Equal(expectedAuths, updatedAuths)
		})

		t.Run("should not alter enrollments", func(t *testing.T) {
			ctx.testContextChanged(t)
			updateContent := gabs.New()
			_, _ = updateContent.SetP(eid.New(), "name")
			_, _ = updateContent.SetP("Device", "type")
			_, _ = updateContent.SetP(map[string]interface{}{}, "tags")
			_, _ = updateContent.SetP(false, "isAdmin")

			resp := ctx.AdminSession.updateEntityOfType(unenrolledId, "identities", updateContent.String(), false)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			updatedIdentity := ctx.AdminSession.requireQuery("identities/" + unenrolledId)

			expectedEnrollments := unenrolledIdentity.Path("data.enrollment").Data().(map[string]interface{})
			ctx.Req.NotEmpty(expectedEnrollments)

			updatedEnrollments := updatedIdentity.Path("data.enrollment").Data().(map[string]interface{})
			ctx.Req.NotEmpty(updatedEnrollments)

			ctx.Req.Equal(expectedEnrollments, updatedEnrollments)
		})

		t.Run("should not allow isDefaultAdmin to be altered", func(t *testing.T) {
			ctx.testContextChanged(t)
			identityId := ctx.AdminSession.requireCreateIdentity(eid.New(), true)

			updateContent := gabs.New()
			_, _ = updateContent.SetP(eid.New(), "name")
			_, _ = updateContent.SetP("Device", "type")
			_, _ = updateContent.SetP(map[string]interface{}{}, "tags")
			_, _ = updateContent.SetP(true, "isAdmin")
			_, _ = updateContent.SetP(true, "isDefaultAdmin")

			resp := ctx.AdminSession.updateEntityOfType(identityId, "identities", updateContent.String(), false)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			updatedIdentity := ctx.AdminSession.requireQuery("identities/" + unenrolledId)

			ctx.Req.Equal(true, updatedIdentity.ExistsP("data.isDefaultAdmin"))
			isDefaultAdmin := updatedIdentity.Path("data.isDefaultAdmin").Data().(bool)
			ctx.Req.Equal(false, isDefaultAdmin)
		})

		t.Run("can update", func(t *testing.T) {
			ctx.testContextChanged(t)
			identityId := ctx.AdminSession.requireCreateIdentity(eid.New(), true)

			newName := eid.New()
			updateContent := gabs.New()
			_, _ = updateContent.SetP(newName, "name")
			_, _ = updateContent.SetP("Device", "type")
			_, _ = updateContent.SetP(map[string]interface{}{}, "tags")
			_, _ = updateContent.SetP(false, "isAdmin")

			resp := ctx.AdminSession.updateEntityOfType(identityId, "identities", updateContent.String(), false)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			updatedIdentity := ctx.AdminSession.requireQuery("identities/" + identityId)

			t.Run("name", func(t *testing.T) {
				ctx.testContextChanged(t)
				ctx.Req.Equal(true, updatedIdentity.ExistsP("data.name"))
				updatedName := updatedIdentity.Path("data.name").Data().(string)
				ctx.Req.Equal(newName, updatedName)
			})

			t.Run("isAdmin", func(t *testing.T) {
				ctx.testContextChanged(t)
				ctx.Req.Equal(true, updatedIdentity.ExistsP("data.isAdmin"))
				newIsAdmin := updatedIdentity.Path("data.isAdmin").Data().(bool)
				ctx.Req.Equal(false, newIsAdmin)
			})
		})

	})
}
