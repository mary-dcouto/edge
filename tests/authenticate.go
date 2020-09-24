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
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge/eid"
	"github.com/openziti/edge/internal/cert"
	"github.com/openziti/foundation/common/constants"
	"github.com/openziti/foundation/util/stringz"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	"net/http"
	"net/url"
	"sort"
)

type authenticator interface {
	Authenticate(ctx *TestContext) (*session, error)
}

type certAuthenticator struct {
	cert    *x509.Certificate
	key     crypto.PrivateKey
	certPem string
}

func (authenticator *certAuthenticator) Authenticate(ctx *TestContext) (*session, error) {
	sess := &session{
		authenticator: authenticator,
		testContext:   ctx,
	}

	sess.authenticatedRequests = authenticatedRequests{
		testContext: ctx,
		session:     sess,
	}

	transport := ctx.Transport()
	transport.TLSClientConfig.Certificates = []tls.Certificate{
		{
			Certificate: [][]byte{authenticator.cert.Raw},
			PrivateKey:  authenticator.key,
		},
	}
	client := ctx.Client(ctx.HttpClient(transport))
	client.SetHostURL("https://" + ctx.ApiHost)

	resp, err := client.R().
		SetHeader("content-type", "application/json").
		Post("/authenticate?method=cert")

	if err != nil {
		return nil, errors.Errorf("failed to authenticate via CERT: %s", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("failed to authenticate via CERT: invalid response code encountered, got %d, expected %d", resp.StatusCode(), http.StatusOK)
	}

	if err = sess.parseSessionInfoFromResponse(ctx, resp); err != nil {
		return nil, err
	}

	return sess, nil
}

func (authenticator *certAuthenticator) TLSCertificates() []tls.Certificate {
	return []tls.Certificate{
		{
			Certificate: [][]byte{authenticator.cert.Raw},
			PrivateKey:  authenticator.key,
		},
	}
}

func (authenticator *certAuthenticator) Fingerprint() string {
	return cert.NewFingerprintGenerator().FromRaw(authenticator.cert.Raw)
}

type updbAuthenticator struct {
	Username    string
	Password    string
	ConfigTypes []string
}

func (authenticator *updbAuthenticator) Authenticate(ctx *TestContext) (*session, error) {
	sess := &session{
		authenticator: authenticator,
		testContext:   ctx,
	}

	sess.authenticatedRequests = authenticatedRequests{
		testContext: ctx,
		session:     sess,
	}

	body := gabs.New()
	_, _ = body.SetP(authenticator.Username, "username")
	_, _ = body.SetP(authenticator.Password, "password")
	if len(authenticator.ConfigTypes) > 0 {
		_, _ = body.SetP(authenticator.ConfigTypes, "configTypes")
	}

	resp, err := ctx.DefaultClient().R().
		SetHeader("content-type", "application/json").
		SetBody(body.String()).
		Post("/authenticate?method=password")

	if err != nil {
		return nil, errors.Errorf("failed to authenticate via UPDB as %v: %s", authenticator, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("failed to authenticate via UPDB as %v: invalid response code encountered, got %d, expected %d", authenticator, resp.StatusCode(), http.StatusOK)
	}

	if err = sess.parseSessionInfoFromResponse(ctx, resp); err != nil {
		return nil, err
	}

	return sess, nil
}

type session struct {
	authenticator authenticator
	id            string
	token         string
	identityId    string
	configTypes   []string
	testContext   *TestContext
	authenticatedRequests
}

func (sess *session) newRequest(ctx *TestContext) *resty.Request {
	req := ctx.newRequest()
	if sess.token != "" {
		req.SetHeader(constants.ZitiSession, sess.token)
	}
	return req
}

func (sess *session) parseSessionInfoFromResponse(ctx *TestContext, response *resty.Response) error {
	respBody := response.Body()

	if len(respBody) == 0 {
		return errors.Errorf("failed to authenticate via UPDB %v: encountered zero length body", sess.authenticator)
	}

	ctx.logJson(respBody)

	bodyContainer, err := gabs.ParseJSON(respBody)

	if err != nil {
		return errors.Errorf("failed to authenticate via UPDB as %v: failed to parse response: %s", sess.authenticator, err)
	}

	if sessionId, ok := bodyContainer.Path("data.id").Data().(string); ok {
		sess.id = sessionId
	} else {
		return errors.Errorf("failed to authenticate via UPDB as %v: failed to find session id", sess.authenticator)
	}

	if sessionToken, ok := bodyContainer.Path("data.token").Data().(string); ok {
		sess.token = sessionToken
	} else {
		return errors.Errorf("failed to authenticate via UPDB as %v: failed to find session token", sess.authenticator)
	}

	if identityId, ok := bodyContainer.Path("data.identity.id").Data().(string); ok {
		sess.identityId = identityId
	} else {
		return errors.Errorf("failed to authenticate via UPDB as %v: failed to find identity id", sess.authenticator)
	}

	sess.configTypes = ctx.toStringSlice(bodyContainer.Path("data.configTypes"))
	return nil
}

func (sess *session) Exists() bool {
	resp, err := sess.newRequest(sess.testContext).Get("/current-api-session")

	if err != nil {
		return false
	}

	return resp.StatusCode() == http.StatusOK
}

func (sess *session) logout() error {
	resp, err := sess.newRequest(sess.testContext).Delete("/current-api-session")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.Errorf("could not delete current session %s for logout: got status code %d expected %d", sess.token, resp.StatusCode(), http.StatusOK)
	}

	return nil
}

type authenticatedRequests struct {
	testContext *TestContext
	session     *session
}

func (request *authenticatedRequests) newAuthenticatedRequest() *resty.Request {
	return request.session.newRequest(request.testContext)
}

func (request *authenticatedRequests) newAuthenticatedJsonRequest(body interface{}) *resty.Request {
	return request.session.newRequest(request.testContext).
		SetHeader("content-type", "application/json").
		SetBody(body)
}

func (request *authenticatedRequests) RequireCreateSdkContext() (*identity, ziti.Context) {
	identity := request.RequireNewIdentityWithOtt(false)
	config := request.testContext.EnrollIdentity(identity.Id)
	context := ziti.NewContextWithConfig(config)
	return identity, context
}

func (request *authenticatedRequests) requireCreateIdentity(name string, isAdmin bool, rolesAttributes ...string) string {
	entityData := gabs.New()
	request.testContext.setJsonValue(entityData, name, "name")
	request.testContext.setJsonValue(entityData, "User", "type")
	request.testContext.setJsonValue(entityData, isAdmin, "isAdmin")
	request.testContext.setJsonValue(entityData, rolesAttributes, "roleAttributes")

	enrollments := map[string]interface{}{}
	request.testContext.setJsonValue(entityData, enrollments, "enrollment")

	entityJson := entityData.String()
	resp := request.createEntityOfType("identities", entityJson)
	request.testContext.Req.Equal(http.StatusCreated, resp.StatusCode())
	id := request.testContext.getEntityId(resp.Body())
	return id
}

func (request *authenticatedRequests) requireCreateIdentityWithUpdbEnrollment(name string, password string, isAdmin bool, rolesAttributes ...string) (string, *updbAuthenticator) {
	userAuth := &updbAuthenticator{
		Username: name,
		Password: password,
	}

	entityData := gabs.New()
	request.testContext.setJsonValue(entityData, name, "name")
	request.testContext.setJsonValue(entityData, "User", "type")
	request.testContext.setJsonValue(entityData, isAdmin, "isAdmin")
	request.testContext.setJsonValue(entityData, rolesAttributes, "roleAttributes")

	enrollments := map[string]interface{}{
		"updb": name,
	}
	request.testContext.setJsonValue(entityData, enrollments, "enrollment")

	entityJson := entityData.String()
	resp := request.createEntityOfType("identities", entityJson)
	request.testContext.Req.Equal(http.StatusCreated, resp.StatusCode())
	id := request.testContext.getEntityId(resp.Body())
	request.testContext.completeUpdbEnrollment(id, password)
	return id, userAuth
}

func (request *authenticatedRequests) requireCreateIdentityOttEnrollment(name string, isAdmin bool, rolesAttributes ...string) (string, *certAuthenticator) {
	entityData := gabs.New()
	request.testContext.setJsonValue(entityData, name, "name")
	request.testContext.setJsonValue(entityData, "User", "type")
	request.testContext.setJsonValue(entityData, isAdmin, "isAdmin")
	request.testContext.setJsonValue(entityData, rolesAttributes, "roleAttributes")

	enrollments := map[string]interface{}{
		"ott": true,
	}
	request.testContext.setJsonValue(entityData, enrollments, "enrollment")

	entityJson := entityData.String()
	resp := request.createEntityOfType("identities", entityJson)
	request.testContext.Req.Equal(http.StatusCreated, resp.StatusCode())
	id := request.testContext.getEntityId(resp.Body())
	return id, request.testContext.completeOttEnrollment(id)
}

func (request *authenticatedRequests) requireCreateIdentityOttEnrollmentUnfinished(name string, isAdmin bool, rolesAttributes ...string) string {
	entityData := gabs.New()
	request.testContext.setJsonValue(entityData, name, "name")
	request.testContext.setJsonValue(entityData, "User", "type")
	request.testContext.setJsonValue(entityData, isAdmin, "isAdmin")
	request.testContext.setJsonValue(entityData, rolesAttributes, "roleAttributes")

	enrollments := map[string]interface{}{
		"ott": true,
	}
	request.testContext.setJsonValue(entityData, enrollments, "enrollment")

	entityJson := entityData.String()
	resp := request.createEntityOfType("identities", entityJson)
	request.testContext.Req.Equal(http.StatusCreated, resp.StatusCode())
	id := request.testContext.getEntityId(resp.Body())
	request.testContext.Req.NotEmpty(id)
	return id
}

func (request *authenticatedRequests) requireNewService(roleAttributes, configs []string) *service {
	service := request.testContext.newService(roleAttributes, configs)
	request.requireCreateEntity(service)
	return service
}

func (request *authenticatedRequests) newServiceBulk(roleAttributes, configs []string) *service {
	service := request.testContext.newService(roleAttributes, configs)
	request.requireCreateEntity(service)
	return service
}

func (request *authenticatedRequests) RequireNewServiceAccessibleToAll(terminatorStrategy string) *service {
	request.requireNewServicePolicy("Dial", s("#all"), s("#all"))
	request.requireNewServicePolicy("Bind", s("#all"), s("#all"))
	request.requireNewEdgeRouterPolicy(s("#all"), s("#all"))
	request.requireNewServiceEdgeRouterPolicy(s("#all"), s("#all"))

	service := request.testContext.newService(nil, nil)
	service.terminatorStrategy = terminatorStrategy
	request.requireCreateEntity(service)

	return service
}

func (request *authenticatedRequests) requireNewTerminator(serviceId, routerId, binding, address string) *terminator {
	terminator := request.testContext.newTerminator(serviceId, routerId, binding, address)
	request.requireCreateEntity(terminator)
	return terminator
}

func (request *authenticatedRequests) requireNewEdgeRouter(roleAttributes ...string) *edgeRouter {
	edgeRouter := newTestEdgeRouter(roleAttributes...)
	request.requireCreateEntity(edgeRouter)
	return edgeRouter
}

func (request *authenticatedRequests) requireNewTransitRouter() *transitRouter {
	transitRouter := newTestTransitRouter()
	request.requireCreateEntity(transitRouter)
	return transitRouter
}

func (request *authenticatedRequests) requireNewServicePolicy(policyType string, serviceRoles, identityRoles []string) *servicePolicy {
	policy := newServicePolicy(policyType, nil, serviceRoles, identityRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewServicePolicyWithSemantic(policyType string, semantic string, serviceRoles, identityRoles []string) *servicePolicy {
	policy := newServicePolicy(policyType, &semantic, serviceRoles, identityRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewEdgeRouterPolicy(edgeRouterRoles, identityRoles []string) *edgeRouterPolicy {
	policy := newEdgeRouterPolicy(nil, edgeRouterRoles, identityRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewEdgeRouterPolicyWithSemantic(semantic string, edgeRouterRoles, identityRoles []string) *edgeRouterPolicy {
	policy := newEdgeRouterPolicy(&semantic, edgeRouterRoles, identityRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewServiceEdgeRouterPolicy(edgeRouterRoles, serviceRoles []string) *serviceEdgeRouterPolicy {
	policy := newServiceEdgeRouterPolicy(nil, edgeRouterRoles, serviceRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewServiceEdgeRouterPolicyWithSemantic(semantic string, edgeRouterRoles, identityRoles []string) *serviceEdgeRouterPolicy {
	policy := newServiceEdgeRouterPolicy(&semantic, edgeRouterRoles, identityRoles)
	request.requireCreateEntity(policy)
	return policy
}

func (request *authenticatedRequests) requireNewIdentity(isAdmin bool, roleAttributes ...string) *identity {
	identity := newTestIdentity(isAdmin, roleAttributes...)
	request.requireCreateEntity(identity)
	return identity
}

func (request *authenticatedRequests) RequireNewIdentityWithOtt(isAdmin bool, roleAttributes ...string) *identity {
	identity := newTestIdentity(isAdmin, roleAttributes...)
	identity.enrollment = map[string]interface{}{"ott": true}
	request.requireCreateEntity(identity)
	return identity
}

func (request *authenticatedRequests) requireCreateEntity(entity entity) string {
	resp := request.createEntity(entity)
	standardJsonResponseTests(resp, http.StatusCreated, request.testContext.testing)
	id := request.testContext.getEntityId(resp.Body())
	entity.setId(id)
	return id
}

func (request *authenticatedRequests) createEntityBulk(entity entity) string {
	resp := request.createEntity(entity)
	if http.StatusCreated != resp.StatusCode() {
		panic(errors.Errorf("expected error code %v", resp.StatusCode()))
	}
	id := request.testContext.getEntityId(resp.Body())
	entity.setId(id)
	return id
}

func (request *authenticatedRequests) requireDeleteEntity(entity entity) {
	resp := request.deleteEntityOfType(entity.getEntityType(), entity.getId())
	standardJsonResponseTests(resp, http.StatusOK, request.testContext.testing)
}

func (request *authenticatedRequests) requireUpdateEntity(entity entity) {
	resp := request.updateEntity(entity)
	standardJsonResponseTests(resp, http.StatusOK, request.testContext.testing)
}

func (request *authenticatedRequests) requireList(url string) []string {
	httpStatus, body := request.query(url)
	request.testContext.logJson(body)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
	jsonBody := request.testContext.parseJson(body)
	values := request.testContext.RequirePath(jsonBody, "data")

	var result []string
	children, err := values.Children()
	request.testContext.Req.NoError(err)
	for _, child := range children {
		val, ok := child.Data().(string)
		request.testContext.Req.True(ok)
		result = append(result, val)
	}
	return result
}

func (request *authenticatedRequests) requireQuery(url string) *gabs.Container {
	httpStatus, body := request.query(url)
	request.testContext.logJson(body)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
	return request.testContext.parseJson(body)
}

func (request *authenticatedRequests) requireAddAssociation(url string, ids ...string) {
	httpStatus, _ := request.addAssociation(url, ids...)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
}

func (request *authenticatedRequests) requireRemoveAssociation(url string, ids ...string) {
	httpStatus, _ := request.removeAssociation(url, ids...)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
}

func (request *authenticatedRequests) createEntityOfType(entityType string, body interface{}) *resty.Response {
	resp, err := request.newAuthenticatedRequest().
		SetBody(body).
		Post("/" + entityType)

	request.testContext.Req.NoError(err)
	request.testContext.logJson(resp.Body())
	return resp
}

type serviceConfig struct {
	ServiceId string `json:"serviceId"`
	ConfigId  string `json:"configId"`
}

type sortableServiceConfigSlice []serviceConfig

func (s sortableServiceConfigSlice) Len() int {
	return len(s)
}

func (s sortableServiceConfigSlice) Less(i, j int) bool {
	return s[i].ServiceId < s[j].ServiceId || (s[i].ServiceId == s[j].ServiceId && s[i].ConfigId < s[j].ConfigId)
}

func (s sortableServiceConfigSlice) Swap(i, j int) {
	val := s[i]
	s[i] = s[j]
	s[j] = val
}

func (request *authenticatedRequests) requireAssignIdentityServiceConfigs(identityId string, serviceConfigs ...serviceConfig) {
	httpStatus, _ := request.updateIdentityServiceConfigs(resty.MethodPost, identityId, serviceConfigs)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
}

func (request *authenticatedRequests) requireRemoveIdentityServiceConfigs(identityId string, serviceConfigs ...serviceConfig) {
	httpStatus, _ := request.updateIdentityServiceConfigs(resty.MethodDelete, identityId, serviceConfigs)
	request.testContext.Req.Equal(http.StatusOK, httpStatus)
}

func (request *authenticatedRequests) listIdentityServiceConfigs(identityId string) []serviceConfig {
	jsonBody := request.requireQuery("identities/" + identityId + "/service-configs")
	data := request.testContext.RequirePath(jsonBody, "data")
	var children []*gabs.Container
	if data.Data() != nil {
		var err error
		children, err = data.Children()
		request.testContext.Req.NoError(err)
	}
	var result []serviceConfig
	for _, child := range children {
		service := request.testContext.requireString(child, "serviceId")
		config := request.testContext.requireString(child, "configId")
		result = append(result, serviceConfig{
			ServiceId: service,
			ConfigId:  config,
		})
	}
	sort.Sort(sortableServiceConfigSlice(result))
	return result
}

func (request *authenticatedRequests) updateIdentityServiceConfigs(method string, identityId string, serviceConfigs []serviceConfig) (int, []byte) {
	req := request.newAuthenticatedRequest()
	if len(serviceConfigs) > 0 {
		body, err := json.MarshalIndent(serviceConfigs, "", "   ")
		request.testContext.Req.NoError(err)
		if request.testContext.enabledJsonLogging {
			fmt.Println(string(body))
		}
		req.SetBody(body)
	}

	resp, err := req.Execute(method, "/identities/"+identityId+"/service-configs")

	request.testContext.Req.NoError(err)
	request.testContext.logJson(resp.Body())
	return resp.StatusCode(), resp.Body()
}

func (request *authenticatedRequests) createEntity(entity entity) *resty.Response {
	return request.createEntityOfType(entity.getEntityType(), entity.toJson(true, request.testContext))
}

func (request *authenticatedRequests) deleteEntityOfType(entityType string, id string) *resty.Response {
	resp, err := request.newAuthenticatedRequest().Delete("/" + entityType + "/" + id)

	request.testContext.Req.NoError(err)
	request.testContext.logJson(resp.Body())

	return resp
}

func (request *authenticatedRequests) updateEntity(entity entity) *resty.Response {
	return request.updateEntityOfType(entity.getId(), entity.getEntityType(), entity.toJson(false, request.testContext), false)
}

func (request *authenticatedRequests) updateEntityOfType(id string, entityType string, body string, patch bool) *resty.Response {
	if request.testContext.enabledJsonLogging {
		fmt.Printf("update body:\n%v\n", body)
	}

	urlPath := fmt.Sprintf("/%v/%v", entityType, id)
	pfxlog.Logger().Infof("url path: %v", urlPath)

	updateRequest := request.newAuthenticatedRequest().SetBody(body)

	var err error
	var resp *resty.Response

	if patch {
		resp, err = updateRequest.Patch(urlPath)
	} else {
		resp, err = updateRequest.Put(urlPath)
	}

	request.testContext.Req.NoError(err)
	request.testContext.logJson(resp.Body())
	return resp
}

func (request *authenticatedRequests) query(url string) (int, []byte) {
	resp, err := request.newAuthenticatedRequest().Get("/" + url)
	request.testContext.Req.NoError(err)
	return resp.StatusCode(), resp.Body()
}

func (request *authenticatedRequests) addAssociation(url string, ids ...string) (int, []byte) {
	return request.updateAssociation(http.MethodPut, url, ids...)
}

func (request *authenticatedRequests) removeAssociation(url string, ids ...string) (int, []byte) {
	return request.updateAssociation(http.MethodDelete, url, ids...)
}

func (request *authenticatedRequests) validateAssociations(entity entity, childType string, children ...entity) {
	var ids []string
	for _, child := range children {
		ids = append(ids, child.getId())
	}
	request.validateAssociationsAt(fmt.Sprintf("%v/%v/%v", entity.getEntityType(), entity.getId(), childType), ids...)
}

func (request *authenticatedRequests) validateAssociationContains(entity entity, childType string, children ...entity) {
	var ids []string
	for _, child := range children {
		ids = append(ids, child.getId())
	}
	request.validateAssociationsAtContains(fmt.Sprintf("%v/%v/%v", entity.getEntityType(), entity.getId(), childType), ids...)
}

func (request *authenticatedRequests) validateAssociationsAt(url string, ids ...string) {
	result := request.requireQuery(url)
	data := request.testContext.RequirePath(result, "data")
	children, err := data.Children()

	var actualIds []string
	request.testContext.Req.NoError(err)
	for _, child := range children {
		actualIds = append(actualIds, child.S("id").Data().(string))
	}

	sort.Strings(ids)
	sort.Strings(actualIds)
	request.testContext.Req.Equal(ids, actualIds)
}

func (request *authenticatedRequests) validateAssociationsAtContains(url string, ids ...string) {
	result := request.requireQuery(url)
	data := request.testContext.RequirePath(result, "data")
	children, err := data.Children()

	var actualIds []string
	request.testContext.Req.NoError(err)
	for _, child := range children {
		actualIds = append(actualIds, child.S("id").Data().(string))
	}

	for _, id := range ids {
		request.testContext.Req.True(stringz.Contains(actualIds, id), "%+v should contain %v", actualIds, id)
	}
}

func (request *authenticatedRequests) updateAssociation(method, url string, ids ...string) (int, []byte) {

	resp, err := request.newAuthenticatedRequest().
		SetBody(request.testContext.idsJson(ids...).String()).
		Execute(method, "/"+url)
	request.testContext.Req.NoError(err)
	request.testContext.logJson(resp.Body())
	return resp.StatusCode(), resp.Body()
}

func (request *authenticatedRequests) isServiceVisibleToUser(serviceId string) bool {
	query := url.QueryEscape(fmt.Sprintf(`id = "%v"`, serviceId))
	result := request.requireQuery("services?filter=" + query)
	data := request.testContext.RequirePath(result, "data")
	return nil != request.testContext.childWith(data, "id", serviceId)
}

func (request *authenticatedRequests) createUserAndLogin(isAdmin bool, roleAttributes, configTypes []string) *session {
	_, userAuth := request.requireCreateIdentityWithUpdbEnrollment(eid.New(), eid.New(), isAdmin, roleAttributes...)
	userAuth.ConfigTypes = configTypes

	session, _ := userAuth.Authenticate(request.testContext)

	return session
}

func (request *authenticatedRequests) validateEntityWithQuery(entity entity) *gabs.Container {
	query := url.QueryEscape(fmt.Sprintf(`id = "%v"`, entity.getId()))
	result := request.requireQuery(entity.getEntityType() + "?filter=" + query)
	data := request.testContext.RequirePath(result, "data")
	jsonEntity := request.testContext.RequireChildWith(data, "id", entity.getId())
	return request.testContext.validateEntity(entity, jsonEntity)
}

func (request *authenticatedRequests) validateEntityWithLookup(entity entity) *gabs.Container {
	result := request.requireQuery(entity.getEntityType() + "/" + entity.getId())
	jsonEntity := request.testContext.RequirePath(result, "data")
	return request.testContext.validateEntity(entity, jsonEntity)
}

func (request *authenticatedRequests) validateUpdate(entity entity) *gabs.Container {
	result := request.requireQuery(entity.getEntityType() + "/" + entity.getId())
	jsonConfig := request.testContext.RequirePath(result, "data")
	entity.validate(request.testContext, jsonConfig)
	return jsonConfig
}

func (request *authenticatedRequests) requireCreateNewConfig(configType string, data map[string]interface{}) *Config {
	config := request.testContext.newConfig(configType, data)
	config.Id = request.requireCreateEntity(config)
	return config
}

func (request *authenticatedRequests) requireCreateNewConfigTypeWithPrefix(prefix string) *configType {
	entity := request.testContext.newConfigType()
	entity.Name = prefix + "-" + entity.Name
	entity.Id = request.requireCreateEntity(entity)
	return entity
}

func (request *authenticatedRequests) requireCreateNewConfigType() *configType {
	entity := request.testContext.newConfigType()
	entity.Id = request.requireCreateEntity(entity)
	return entity
}

func (request *authenticatedRequests) requirePatchEntity(entity entity, fields ...string) {
	resp := request.patchEntity(entity, fields...)
	standardJsonResponseTests(resp, http.StatusOK, request.testContext.testing)
}

func (request *authenticatedRequests) patchEntity(entity entity, fields ...string) *resty.Response {
	return request.updateEntityOfType(entity.getId(), entity.getEntityType(), entity.toJson(false, request.testContext, fields...), true)
}

func (request *authenticatedRequests) getEdgeRouterJwt(edgeRouterId string) string {
	jsonBody := request.requireQuery("edge-routers/" + edgeRouterId)
	data := request.testContext.RequirePath(jsonBody, "data", "enrollmentJwt")
	return data.Data().(string)
}

func (request *authenticatedRequests) getTransitRouterJwt(transitRouterId string) string {
	jsonBody := request.requireQuery("transit-routers/" + transitRouterId)
	data := request.testContext.RequirePath(jsonBody, "data", "enrollmentJwt")
	return data.Data().(string)
}

func (request *authenticatedRequests) getIdentityJwt(identityId string) string {
	jsonBody := request.requireQuery("identities/" + identityId)
	data := request.testContext.RequirePath(jsonBody, "data", "enrollment", "ott", "jwt")
	return data.Data().(string)
}
