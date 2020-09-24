// +build apitests

package tests

import (
	"fmt"
	"github.com/openziti/edge/controller/model"
	"github.com/openziti/edge/eid"
	"net/http"
	"sort"
	"testing"
)

func Test_CA(t *testing.T) {
	ctx := NewTestContext(t)
	defer ctx.Teardown()
	ctx.StartServer()
	ctx.RequireAdminLogin()

	t.Run("identity attributes should be created", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		ca := newTestCa(role1, role2)
		ca.id = ctx.AdminSession.requireCreateEntity(ca)
		ctx.AdminSession.validateEntityWithQuery(ca)
		ctx.AdminSession.validateEntityWithLookup(ca)
	})

	t.Run("identity attributes should be updated", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		ca := newTestCa(role1, role2)
		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		role3 := eid.New()
		ca.identityRoles = []string{role2, role3}
		ctx.AdminSession.requireUpdateEntity(ca)
		ctx.AdminSession.validateEntityWithLookup(ca)
	})

	t.Run("identity name format should default if not specified", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		ca := newTestCa(role1, role2)

		ca.identityNameFormat = ""
		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		//set to default for verification
		ca.identityNameFormat = model.DefaultCaIdentityNameFormat

		ctx.AdminSession.validateEntityWithQuery(ca)
		ctx.AdminSession.validateEntityWithLookup(ca)
	})

	t.Run("identities from auto enrollment inherit CA identity roles", func(t *testing.T) {
		ctx.testContextChanged(t)
		role1 := eid.New()
		role2 := eid.New()
		ca := newTestCa(role1, role2)
		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		caValues := ctx.AdminSession.requireQuery("cas/" + ca.id)
		verificationToken := caValues.Path("data.verificationToken").Data().(string)

		ctx.Req.NotEmpty(verificationToken)

		validationAuth := ca.CreateSignedCert(verificationToken)

		clientAuthenticator := ca.CreateSignedCert(eid.New())

		resp, err := ctx.AdminSession.newAuthenticatedRequest().
			SetHeader("content-type", "text/plain").
			SetBody(validationAuth.certPem).
			Post("cas/" + ca.id + "/verify")

		ctx.Req.NoError(err)
		ctx.logJson(resp.Body())
		ctx.Req.Equal(http.StatusOK, resp.StatusCode())

		ctx.completeCaAutoEnrollment(clientAuthenticator)

		enrolledSession, err := clientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)

		identity := ctx.AdminSession.requireQuery("identities/" + enrolledSession.identityId)
		sort.Strings(ca.identityRoles)
		ctx.pathEqualsStringSlice(identity, ca.identityRoles, path("data", "roleAttributes"))
	})

	t.Run("identities from auto enrollment use identity name format for naming", func(t *testing.T) {
		ctx.testContextChanged(t)
		ca := newTestCa()

		expectedName := "singular.name.not.great"
		ca.identityNameFormat = expectedName
		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		caValues := ctx.AdminSession.requireQuery("cas/" + ca.id)
		verificationToken := caValues.Path("data.verificationToken").Data().(string)

		ctx.Req.NotEmpty(verificationToken)

		validationAuth := ca.CreateSignedCert(verificationToken)

		clientAuthenticator := ca.CreateSignedCert(eid.New())

		resp, err := ctx.AdminSession.newAuthenticatedRequest().
			SetHeader("content-type", "text/plain").
			SetBody(validationAuth.certPem).
			Post("cas/" + ca.id + "/verify")

		ctx.Req.NoError(err)
		ctx.logJson(resp.Body())
		ctx.Req.Equal(http.StatusOK, resp.StatusCode())

		ctx.completeCaAutoEnrollment(clientAuthenticator)

		enrolledSession, err := clientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)

		identity := ctx.AdminSession.requireQuery("identities/" + enrolledSession.identityId)
		ctx.Req.Equal(expectedName, identity.Path("data.name").Data().(string))
	})

	t.Run("identities from auto enrollment identity name collisions add numbers to the end", func(t *testing.T) {
		ctx.testContextChanged(t)

		firstExpectedName := "some.static.name.no.replacements"
		secondExpectedName := "some.static.name.no.replacements000001"

		//create CA
		ca := newTestCa()
		ca.identityNameFormat = firstExpectedName
		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		caValues := ctx.AdminSession.requireQuery("cas/" + ca.id)
		verificationToken := caValues.Path("data.verificationToken").Data().(string)

		ctx.Req.NotEmpty(verificationToken)

		//validate CA
		validationAuth := ca.CreateSignedCert(verificationToken)
		resp, err := ctx.AdminSession.newAuthenticatedRequest().
			SetHeader("content-type", "text/plain").
			SetBody(validationAuth.certPem).
			Post("cas/" + ca.id + "/verify")

		ctx.Req.NoError(err)
		ctx.logJson(resp.Body())
		ctx.Req.Equal(http.StatusOK, resp.StatusCode())

		//first firstIdentity, no issues
		firstClientAuthenticator := ca.CreateSignedCert(eid.New())
		ctx.completeCaAutoEnrollment(firstClientAuthenticator)

		firstEnrolledSession, err := firstClientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)

		firstIdentity := ctx.AdminSession.requireQuery("identities/" + firstEnrolledSession.identityId)
		ctx.Req.Equal(firstExpectedName, firstIdentity.Path("data.name").Data().(string))

		//second firstIdentity that collides, becomes
		secondClientAuthenticator := ca.CreateSignedCert(eid.New())
		ctx.completeCaAutoEnrollment(secondClientAuthenticator)

		secondEnrolledSession, err := secondClientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)

		secondIdentity := ctx.AdminSession.requireQuery("identities/" + secondEnrolledSession.identityId)
		ctx.Req.Equal(secondExpectedName, secondIdentity.Path("data.name").Data().(string))
	})

	t.Run("identities from auto enrollment use identity name format for naming with replacements", func(t *testing.T) {
		ctx.testContextChanged(t)
		ca := newTestCa()
		ca.identityNameFormat = "[caName] - [caId] - [commonName] - [requestedName] - [identityId]"

		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		caValues := ctx.AdminSession.requireQuery("cas/" + ca.id)
		verificationToken := caValues.Path("data.verificationToken").Data().(string)

		ctx.Req.NotEmpty(verificationToken)

		validationAuth := ca.CreateSignedCert(verificationToken)

		clientAuthenticator := ca.CreateSignedCert(eid.New())

		resp, err := ctx.AdminSession.newAuthenticatedRequest().
			SetHeader("content-type", "text/plain").
			SetBody(validationAuth.certPem).
			Post("cas/" + ca.id + "/verify")

		ctx.Req.NoError(err)
		ctx.logJson(resp.Body())
		ctx.Req.Equal(http.StatusOK, resp.StatusCode())

		requestedName := "bobby"
		ctx.completeCaAutoEnrollmentWithName(clientAuthenticator, requestedName)

		enrolledSession, err := clientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)

		identity := ctx.AdminSession.requireQuery("identities/" + enrolledSession.identityId)
		expectedName := fmt.Sprintf("%s - %s - %s - %s - %s", ca.name, ca.id, clientAuthenticator.cert.Subject.CommonName, requestedName, enrolledSession.identityId)

		ctx.Req.Equal(expectedName, identity.Path("data.name").Data().(string))
	})

	t.Run("CAs with auth enabled can authenticate", func(t *testing.T) {
		ctx.testContextChanged(t)
		ca := newTestCa()

		ca.id = ctx.AdminSession.requireCreateEntity(ca)

		caValues := ctx.AdminSession.requireQuery("cas/" + ca.id)
		verificationToken := caValues.Path("data.verificationToken").Data().(string)
		ctx.Req.NotEmpty(verificationToken)

		validationAuth := ca.CreateSignedCert(verificationToken)
		clientAuthenticator := ca.CreateSignedCert(eid.New())

		resp, err := ctx.AdminSession.newAuthenticatedRequest().
			SetHeader("content-type", "text/plain").
			SetBody(validationAuth.certPem).
			Post("cas/" + ca.id + "/verify")

		ctx.Req.NoError(err)
		ctx.logJson(resp.Body())
		ctx.Req.Equal(http.StatusOK, resp.StatusCode())

		ctx.completeCaAutoEnrollment(clientAuthenticator)

		enrolledSession, err := clientAuthenticator.Authenticate(ctx)

		ctx.Req.NoError(err)
		ctx.Req.NotEmpty(enrolledSession)

		t.Run("CAs with auth disabled can no longer authenticate", func(t *testing.T) {
			ctx.testContextChanged(t)
			ca.isAuthEnabled = false
			resp := ctx.AdminSession.patchEntity(ca, "isAuthEnabled")
			ctx.Req.NotEmpty(resp)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			enrolledSession, err := clientAuthenticator.Authenticate(ctx)

			ctx.Req.Error(err)
			ctx.Req.Empty(enrolledSession)
		})

		t.Run("CAs with auth re-enabled can authenticate", func(t *testing.T) {
			ctx.testContextChanged(t)
			ca.isAuthEnabled = true
			resp := ctx.AdminSession.patchEntity(ca, "isAuthEnabled")
			ctx.Req.NotEmpty(resp)
			ctx.Req.Equal(http.StatusOK, resp.StatusCode())

			enrolledSession, err := clientAuthenticator.Authenticate(ctx)

			ctx.Req.NoError(err)
			ctx.Req.NotEmpty(enrolledSession)
		})

		t.Run("deleting a CA no longer allows authentication", func(t *testing.T) {
			t.Skip("skipping till edge#311")
			ctx.testContextChanged(t)

			beforeDeleteSession, err := clientAuthenticator.Authenticate(ctx)

			ctx.AdminSession.requireDeleteEntity(ca)

			afterDeleteSession, err := clientAuthenticator.Authenticate(ctx)

			ctx.Req.Error(err)
			ctx.Req.Empty(afterDeleteSession)

			t.Run("sessions authenticated through a deleted CA are also deleted", func(t *testing.T) {
				ctx.Req.False(beforeDeleteSession.Exists())
			})
		})
	})
}
