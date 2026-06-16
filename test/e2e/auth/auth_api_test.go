//go:build e2e
// +build e2e

package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/pole-io/pole-server/test/e2e/internal/e2e"
)

func TestAuthSwitchMatrixForConsoleAndClientAPI(t *testing.T) {
	cases := []struct {
		name        string
		portBase    int
		consoleOpen bool
		clientOpen  bool
	}{
		{name: "console-off-client-off", portBase: 30100, consoleOpen: false, clientOpen: false},
		{name: "console-on-client-off", portBase: 30120, consoleOpen: true, clientOpen: false},
		{name: "console-off-client-on", portBase: 30140, consoleOpen: false, clientOpen: true},
		{name: "console-on-client-on", portBase: 30160, consoleOpen: true, clientOpen: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := e2e.Start(t, e2e.Options{Suite: tc.name, PortBase: tc.portBase, ConsoleOpen: tc.consoleOpen, ClientOpen: tc.clientOpen})
			api := e2e.NewHTTPClient(t, env)

			consoleResp := api.ConsoleQuery(http.MethodGet, "/core/v1/namespaces", e2e.Query(0, 10, nil), nil)
			if tc.consoleOpen {
				e2e.RequireDenied(t, consoleResp, "anonymous console request when consoleOpen=true")
			} else {
				e2e.RequireSuccess(t, consoleResp, "anonymous console request when consoleOpen=false")
			}

			clientResp := api.Client(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest("default", "not-exist", "SERVICES"))
			if tc.clientOpen {
				e2e.RequireDenied(t, clientResp, "anonymous client request when clientOpen=true")
			} else {
				e2e.RequireSuccess(t, clientResp, "anonymous client request when clientOpen=false")
			}

			api.LoginAsMainUser()
			names := e2e.NewNames("auth-switch")
			e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{e2e.Namespace(names.Namespace)}), "authenticated console create namespace")
			e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, names.Payment)}), "authenticated console create service")

			anonymousInstance := e2e.HeartbeatInstance(names.Namespace, names.Payment, "10.3.0.10", 19180)
			registerResp := api.Client(http.MethodPost, "/naming/v1/RegisterInstance", anonymousInstance)
			if tc.clientOpen {
				e2e.RequireDenied(t, registerResp, "anonymous register instance when clientOpen=true")
				heartbeatTarget := e2e.HeartbeatInstance(names.Namespace, names.Payment, "10.3.0.11", 19181)
				e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/instances", []map[string]any{heartbeatTarget}), "create heartbeat target for anonymous denied check")
				heartbeatTargets := api.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{
					"namespace": names.Namespace,
					"service":   names.Payment,
				}), nil)
				e2e.RequireSuccess(t, heartbeatTargets, "list heartbeat target for anonymous denied check")
				heartbeatItem, ok := e2e.FindByField(heartbeatTargets.Slice(t), "host", "10.3.0.11")
				if !ok {
					t.Fatalf("heartbeat target host 10.3.0.11 not found")
				}
				heartbeatTarget["id"] = e2e.StringField(t, heartbeatItem, "id")
				e2e.RequireDenied(t, api.Client(http.MethodPost, "/naming/v1/Heartbeat", heartbeatTarget), "anonymous heartbeat when clientOpen=true")
				return
			}
			e2e.RequireSuccess(t, registerResp, "anonymous register instance when clientOpen=false")
			instances := api.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{
				"namespace": names.Namespace,
				"service":   names.Payment,
			}), nil)
			e2e.RequireSuccess(t, instances, "list anonymous registered instance")
			instanceItem, ok := e2e.FindByField(instances.Slice(t), "host", "10.3.0.10")
			if !ok {
				t.Fatalf("anonymous registered instance host 10.3.0.10 not found")
			}
			anonymousInstance["id"] = e2e.StringField(t, instanceItem, "id")
			e2e.RequireSuccess(t, api.Client(http.MethodPost, "/naming/v1/Heartbeat", anonymousInstance), "anonymous heartbeat when clientOpen=false")
		})
	}
}

func TestAuthPolicyUserGroupRoleAndTokenInterfaces(t *testing.T) {
	env := e2e.Start(t, e2e.Options{Suite: "auth-api", PortBase: 30200, ConsoleOpen: true, ClientOpen: true})
	api := e2e.NewHTTPClient(t, env)
	api.LoginAsMainUser()
	names := e2e.NewNames("auth")

	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{e2e.Namespace(names.Namespace)}), "create auth namespace")
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, names.Payment)}), "create auth service")
	paymentServiceID := requireServiceID(t, api, names.Namespace, names.Payment)

	user := e2e.User(names.User)
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/users", []map[string]any{user}), "create auth user")
	users := api.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": names.User}), nil)
	e2e.RequireSuccess(t, users, "list auth users")
	userItem, ok := e2e.FindByName(users.Slice(t), names.User)
	if !ok {
		t.Fatalf("auth user %s not found", names.User)
	}
	userID := e2e.StringField(t, userItem, "id")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/user/token", map[string]string{"id": userID}, nil), "get auth user token")
	e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/user/token/enable", map[string]any{"id": userID, "token_enable": false}), "disable auth user token")
	e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/user/token/refresh", map[string]any{"id": userID}), "refresh auth user token")

	group := e2e.UserGroup(names.UserGroup)
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/usergroups", []map[string]any{group}), "create auth user group")
	groups := api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroups", e2e.Query(0, 10, map[string]string{"name": names.UserGroup}), nil)
	e2e.RequireSuccess(t, groups, "list auth user groups")
	groupItem, ok := e2e.FindByName(groups.Slice(t), names.UserGroup)
	if !ok {
		t.Fatalf("auth user group %s not found", names.UserGroup)
	}
	groupID := e2e.StringField(t, groupItem, "id")
	group["id"] = groupID
	group["comment"] = "e2e group updated"
	e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/usergroups", []map[string]any{group}), "update auth user group")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroup/detail", map[string]string{"id": groupID}, nil), "get auth user group detail")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroup/token", map[string]string{"id": groupID}, nil), "get auth user group token")

	role := e2e.Role(names.Role)
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/roles", []map[string]any{role}), "create auth role")
	roles := api.ConsoleQuery(http.MethodGet, "/auth/v1/roles", e2e.Query(0, 10, map[string]string{"name": names.Role}), nil)
	e2e.RequireSuccess(t, roles, "list auth roles")
	roleItem, ok := e2e.FindByName(roles.Slice(t), names.Role)
	if !ok {
		t.Fatalf("auth role %s not found", names.Role)
	}
	roleID := e2e.StringField(t, roleItem, "id")
	role["id"] = roleID
	role["comment"] = "e2e role updated"
	e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/roles", []map[string]any{role}), "update auth role")

	policy := e2e.UserServicePolicy(names.Policy, userID, paymentServiceID, names.Namespace, names.Payment, "READ_WRITE")
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{policy}), "create auth policy")
	policies := api.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy}), nil)
	e2e.RequireSuccess(t, policies, "list auth policies")
	policyItem, ok := e2e.FindByName(policies.Slice(t), names.Policy)
	if !ok {
		t.Fatalf("auth policy %s not found", names.Policy)
	}
	policyID := e2e.StringField(t, policyItem, "id")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/policies/detail", map[string]string{"id": policyID}, nil), "get auth policy detail")
	policy["id"] = policyID
	policy["comment"] = "e2e policy updated"
	e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/policies", []map[string]any{policy}), "update auth policy")
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/resources/authorize", []map[string]any{
		e2e.AuthorizeResource("Services", paymentServiceID, e2e.UserPrincipals(userID)),
	}), "authorize service resource")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/principal/resources", map[string]string{"principal_id": userID, "principal_type": "PrincipalUser"}, nil), "query principal resources")
	e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/resources/principals", map[string]string{"res_type": "Services", "res_id": paymentServiceID}, nil), "query resource principals")
}

func TestAuthPolicyEnforcesConsoleAndClientRequests(t *testing.T) {
	env := e2e.Start(t, e2e.Options{Suite: "auth-enforce", PortBase: 30240, ConsoleOpen: true, ClientOpen: true})
	admin := e2e.NewHTTPClient(t, env)
	admin.LoginAsMainUser()
	names := e2e.NewNames("auth-enforce")
	otherNamespace := names.Prefix + "-other-ns"

	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{
		e2e.Namespace(names.Namespace),
		e2e.Namespace(otherNamespace),
	}), "create policy namespaces")
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/naming/v1/services", []map[string]any{
		e2e.Service(names.Namespace, names.Order),
		e2e.Service(names.Namespace, names.Payment),
		e2e.Service(names.Namespace, names.Shadow),
	}), "create policy services")
	paymentServiceID := requireServiceID(t, admin, names.Namespace, names.Payment)
	shadowServiceID := requireServiceID(t, admin, names.Namespace, names.Shadow)
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/naming/v1/instances", []map[string]any{
		e2e.Instance(names.Namespace, names.Payment, "10.2.0.10", 19080),
	}), "create authorized instance")

	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(names.User)}), "create policy user")
	users := admin.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": names.User}), nil)
	e2e.RequireSuccess(t, users, "list policy user")
	userItem, ok := e2e.FindByName(users.Slice(t), names.User)
	if !ok {
		t.Fatalf("policy user %s not found", names.User)
	}
	userID := e2e.StringField(t, userItem, "id")

	readPolicy := e2e.UserNamespaceAndServicePolicy(names.Policy, userID, names.Namespace, paymentServiceID, names.Payment, "READ")
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{readPolicy}), "create readonly policy")
	readPolicies := admin.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy}), nil)
	e2e.RequireSuccess(t, readPolicies, "list readonly policy")
	readPolicyItem, ok := e2e.FindByName(readPolicies.Slice(t), names.Policy)
	if !ok {
		t.Fatalf("readonly policy %s not found", names.Policy)
	}
	readPolicyID := e2e.StringField(t, readPolicyItem, "id")
	readPolicy["id"] = readPolicyID

	userConsole := e2e.NewHTTPClient(t, env)
	userConsole.Login(names.User, "user123")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := userConsole.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "name": names.Payment}), nil)
		return resp.Code == e2e.ExecuteSuccess
	}, "readonly user can list authorized service")
	deniedCreate := userConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, names.Prefix+"-forbidden")})
	e2e.RequireDenied(t, deniedCreate, "readonly user cannot create service")

	writeUserName := names.User + "-writer"
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(writeUserName)}), "create write policy user")
	writeUserID := requireUserID(t, admin, writeUserName)
	writePolicy := e2e.UserNamespacePolicy(names.Policy+"-writer", writeUserID, names.Namespace, "READ_WRITE")
	writePolicy["functions"] = []string{"CreateServices", "DescribeServices"}
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{writePolicy}), "create write namespace policy")
	writeConsole := e2e.NewHTTPClient(t, env)
	writeConsole.Login(writeUserName, "user123")
	authorizedService := names.Prefix + "-writer-service"
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := writeConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, authorizedService)})
		return resp.Code == e2e.ExecuteSuccess
	}, "write user can create service in authorized namespace")
	deniedOtherNamespace := writeConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(otherNamespace, names.Prefix+"-writer-denied")})
	e2e.RequireDenied(t, deniedOtherNamespace, "write user cannot create service in unauthorized namespace")

	governanceUserName := names.User + "-governance"
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(governanceUserName)}), "create governance policy user")
	governanceUserID := requireUserID(t, admin, governanceUserName)
	governancePolicy := e2e.AuthPolicy(
		names.Policy+"-governance",
		e2e.UserPrincipals(governanceUserID),
		map[string]any{},
		"READ_WRITE",
		[]string{"CreateRouteRules", "CreateTrafficSecurityRules"},
	)
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{governancePolicy}), "create governance type policy")
	governanceConsole := e2e.NewHTTPClient(t, env)
	governanceConsole.Login(governanceUserName, "user123")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := governanceConsole.Console(http.MethodPost, "/naming/v1/traffic/security", []map[string]any{
			e2e.TrafficSecurityRule(names.Namespace, names.Prefix+"-security-allowed", names.Payment),
		})
		return resp.Code == e2e.ExecuteSuccess
	}, "governance user can create authorized traffic security rule")
	deniedMirror := governanceConsole.Console(http.MethodPost, "/naming/v1/traffic/mirrors", []map[string]any{
		e2e.TrafficMirrorRule(names.Namespace, names.Prefix+"-mirror-denied", names.Payment, names.Shadow),
	})
	e2e.RequireDenied(t, deniedMirror, "governance user cannot create unauthorized traffic mirror rule")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := governanceConsole.Console(http.MethodPost, "/naming/v1/routings", []map[string]any{
			e2e.RouteRule(names.Namespace, names.Prefix+"-route-allowed", names.Order, names.Payment),
		})
		return resp.Code == e2e.ExecuteSuccess
	}, "governance user can create authorized route rule")
	deniedRateLimit := governanceConsole.Console(http.MethodPost, "/naming/v1/ratelimits", []map[string]any{
		e2e.RateLimitRule(names.Namespace, names.Prefix+"-ratelimit-denied", names.Payment),
	})
	e2e.RequireDenied(t, deniedRateLimit, "governance user cannot create unauthorized rate limit rule")

	tokenResp := admin.ConsoleQuery(http.MethodGet, "/auth/v1/user/token", map[string]string{"id": userID}, nil)
	e2e.RequireSuccess(t, tokenResp, "get policy user token")
	userToken := e2e.ExtractToken(t, tokenResp)
	headers := map[string]string{"Authorization": userToken}
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Payment, "INSTANCE"), headers)
		return resp.Code == e2e.ExecuteSuccess || resp.Code == 200001
	}, "authorized token can discover authorized service")
	unauthorized := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Shadow, "INSTANCE"), headers)
	e2e.RequireDenied(t, unauthorized, "authorized token cannot discover unauthorized service")
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/user/token/enable", map[string]any{"id": userID, "token_enable": false}), "disable readonly user token")
	e2e.RequireDenied(t, admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Payment, "INSTANCE"), headers), "disabled token cannot discover authorized service")
	refreshedTokenResp := admin.Console(http.MethodPut, "/auth/v1/user/token/refresh", map[string]any{"id": userID})
	e2e.RequireSuccess(t, refreshedTokenResp, "refresh readonly user token")
	refreshedToken := e2e.ExtractToken(t, refreshedTokenResp)
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/user/token/enable", map[string]any{"id": userID, "token_enable": true}), "enable refreshed readonly user token")
	refreshedHeaders := map[string]string{"Authorization": refreshedToken}
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Payment, "INSTANCE"), refreshedHeaders)
		return resp.Code == e2e.ExecuteSuccess || resp.Code == 200001
	}, "refreshed token can discover authorized service")
	e2e.RequireDenied(t, admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Payment, "INSTANCE"), headers), "old token cannot discover after refresh")

	clientUserName := names.User + "-client"
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(clientUserName)}), "create client policy user")
	clientUsers := admin.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": clientUserName}), nil)
	e2e.RequireSuccess(t, clientUsers, "list client policy user")
	clientUserItem, ok := e2e.FindByName(clientUsers.Slice(t), clientUserName)
	if !ok {
		t.Fatalf("client policy user %s not found", clientUserName)
	}
	clientUserID := e2e.StringField(t, clientUserItem, "id")
	clientPolicy := e2e.UserServicePolicy(names.Policy+"-client", clientUserID, paymentServiceID, names.Namespace, names.Payment, "READ_WRITE")
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{clientPolicy}), "create client write policy")
	clientPolicies := admin.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy + "-client"}), nil)
	e2e.RequireSuccess(t, clientPolicies, "list client write policy")
	clientPolicyItem, ok := e2e.FindByName(clientPolicies.Slice(t), names.Policy+"-client")
	if !ok {
		t.Fatalf("client write policy %s not found", names.Policy+"-client")
	}
	clientPolicyID := e2e.StringField(t, clientPolicyItem, "id")
	clientPolicy["id"] = clientPolicyID
	clientTokenResp := admin.ConsoleQuery(http.MethodGet, "/auth/v1/user/token", map[string]string{"id": clientUserID}, nil)
	e2e.RequireSuccess(t, clientTokenResp, "get client policy user token")
	clientHeaders := map[string]string{"Authorization": e2e.ExtractToken(t, clientTokenResp)}

	registerInstance := e2e.HeartbeatInstance(names.Namespace, names.Payment, "10.2.0.11", 19081)
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/RegisterInstance", registerInstance, clientHeaders)
		return resp.Code == e2e.ExecuteSuccess
	}, "authorized token can register instance on authorized service")
	forbiddenRegister := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/RegisterInstance", e2e.HeartbeatInstance(names.Namespace, names.Shadow, "10.2.0.12", 19082), clientHeaders)
	e2e.RequireDenied(t, forbiddenRegister, "authorized token cannot register instance on unauthorized service")

	registeredInstances := admin.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{
		"namespace": names.Namespace,
		"service":   names.Payment,
	}), nil)
	e2e.RequireSuccess(t, registeredInstances, "list registered instances")
	registeredItem, ok := e2e.FindByField(registeredInstances.Slice(t), "host", "10.2.0.11")
	if !ok {
		t.Fatalf("registered instance host 10.2.0.11 not found")
	}
	registerInstance["id"] = e2e.StringField(t, registeredItem, "id")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Heartbeat", registerInstance, clientHeaders)
		return resp.Code == e2e.ExecuteSuccess
	}, "authorized token can heartbeat authorized instance")

	shadowInstance := e2e.HeartbeatInstance(names.Namespace, names.Shadow, "10.2.0.13", 19083)
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/naming/v1/instances", []map[string]any{shadowInstance}), "create unauthorized heartbeat target")
	shadowInstances := admin.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{
		"namespace": names.Namespace,
		"service":   names.Shadow,
	}), nil)
	e2e.RequireSuccess(t, shadowInstances, "list unauthorized heartbeat targets")
	shadowItem, ok := e2e.FindByField(shadowInstances.Slice(t), "host", "10.2.0.13")
	if !ok {
		t.Fatalf("shadow instance host 10.2.0.13 not found")
	}
	shadowInstance["id"] = e2e.StringField(t, shadowItem, "id")
	forbiddenHeartbeat := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Heartbeat", shadowInstance, clientHeaders)
	e2e.RequireDenied(t, forbiddenHeartbeat, "authorized token cannot heartbeat unauthorized instance")

	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies/delete", []map[string]any{{"id": clientPolicyID}}), "delete client write policy")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/RegisterInstance", e2e.HeartbeatInstance(names.Namespace, names.Payment, "10.2.0.14", 19084), clientHeaders)
		return resp.Code != e2e.ExecuteSuccess
	}, "client write permission is revoked after policy delete")

	groupUserName := names.User + "-group-member"
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(groupUserName)}), "create group member user")
	groupUserID := requireUserID(t, admin, groupUserName)
	groupName := names.UserGroup + "-enforce"
	group := e2e.UserGroupWithUsers(groupName, map[string]any{"id": groupUserID})
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/usergroups", []map[string]any{group}), "create user group with member")
	groupID := requireUserGroupID(t, admin, groupName)
	groupPolicy := e2e.GroupServicePolicy(names.Policy+"-group", groupID, paymentServiceID, names.Namespace, names.Payment, "READ")
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{groupPolicy}), "create group read policy")
	groupMemberConsole := e2e.NewHTTPClient(t, env)
	groupMemberConsole.Login(groupUserName, "user123")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := groupMemberConsole.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "name": names.Payment}), nil)
		return resp.Code == e2e.ExecuteSuccess
	}, "user inherits group read permission")
	group["id"] = groupID
	group["relation"] = map[string]any{"users": []map[string]any{}}
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/usergroups", []map[string]any{group}), "remove user from group")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := groupMemberConsole.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "name": names.Payment}), nil)
		return resp.Code != e2e.ExecuteSuccess
	}, "group inherited permission is revoked after membership removal")

	roleUserName := names.User + "-role-member"
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/users", []map[string]any{e2e.User(roleUserName)}), "create role member user")
	roleUserID := requireUserID(t, admin, roleUserName)
	roleName := names.Role + "-enforce"
	role := e2e.RoleWithUsers(roleName, map[string]any{"id": roleUserID})
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/roles", []map[string]any{role}), "create role with member")
	roleID := requireRoleID(t, admin, roleName)
	rolePolicy := e2e.RoleServicePolicy(names.Policy+"-role", roleID, paymentServiceID, names.Namespace, names.Payment, "READ_WRITE", []string{"DescribeServices"})
	e2e.RequireSuccess(t, admin.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{rolePolicy}), "create role service policy")
	rolePolicies := admin.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy + "-role"}), nil)
	e2e.RequireSuccess(t, rolePolicies, "list role service policy")
	rolePolicyItem, ok := e2e.FindByName(rolePolicies.Slice(t), names.Policy+"-role")
	if !ok {
		t.Fatalf("role policy %s not found", names.Policy+"-role")
	}
	rolePolicyID := e2e.StringField(t, rolePolicyItem, "id")
	rolePolicy["id"] = rolePolicyID
	roleConsole := e2e.NewHTTPClient(t, env)
	roleConsole.Login(roleUserName, "user123")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := roleConsole.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "name": names.Payment}), nil)
		return resp.Code == e2e.ExecuteSuccess
	}, "role member can list service with DescribeServices")
	e2e.RequireDenied(t, roleConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, names.Prefix+"-role-denied")}), "role member cannot create service before function update")
	rolePolicy["functions"] = []string{"DescribeServices", "CreateServices"}
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/policies", []map[string]any{rolePolicy}), "add CreateServices to role policy")
	roleServiceName := names.Prefix + "-role-created"
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := roleConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, roleServiceName)})
		return resp.Code == e2e.ExecuteSuccess
	}, "role member can create service after function update")
	rolePolicy["functions"] = []string{"DescribeServices"}
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/policies", []map[string]any{rolePolicy}), "remove CreateServices from role policy")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := roleConsole.Console(http.MethodPost, "/naming/v1/services", []map[string]any{e2e.Service(names.Namespace, names.Prefix+"-role-revoked")})
		return resp.Code != e2e.ExecuteSuccess
	}, "role member cannot create service after function removal")

	readPolicy["resources"] = e2e.NamespaceAndServiceResources(names.Namespace, shadowServiceID, names.Shadow)
	e2e.RequireSuccess(t, admin.Console(http.MethodPut, "/auth/v1/policies", []map[string]any{readPolicy}), "move readonly policy to shadow service")
	e2e.Eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		resp := admin.ClientWithHeaders(http.MethodPost, "/naming/v1/Discover", e2e.DiscoverRequest(names.Namespace, names.Payment, "INSTANCE"), refreshedHeaders)
		return resp.Code != e2e.ExecuteSuccess && resp.Code != 200001
	}, "client discover permission is revoked after policy update")
}

func requireServiceID(t *testing.T, api *e2e.HTTPClient, namespace, service string) string {
	t.Helper()
	resp := api.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{
		"namespace": namespace,
		"name":      service,
	}), nil)
	e2e.RequireSuccess(t, resp, "list service "+namespace+"/"+service)
	item, ok := e2e.FindByName(resp.Slice(t), service)
	if !ok {
		t.Fatalf("service %s/%s not found", namespace, service)
	}
	return e2e.StringField(t, item, "id")
}

func requireUserID(t *testing.T, api *e2e.HTTPClient, name string) string {
	t.Helper()
	resp := api.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": name}), nil)
	e2e.RequireSuccess(t, resp, "list user "+name)
	item, ok := e2e.FindByName(resp.Slice(t), name)
	if !ok {
		t.Fatalf("user %s not found", name)
	}
	return e2e.StringField(t, item, "id")
}

func requireUserGroupID(t *testing.T, api *e2e.HTTPClient, name string) string {
	t.Helper()
	resp := api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroups", e2e.Query(0, 10, map[string]string{"name": name}), nil)
	e2e.RequireSuccess(t, resp, "list user group "+name)
	item, ok := e2e.FindByName(resp.Slice(t), name)
	if !ok {
		t.Fatalf("user group %s not found", name)
	}
	return e2e.StringField(t, item, "id")
}

func requireRoleID(t *testing.T, api *e2e.HTTPClient, name string) string {
	t.Helper()
	resp := api.ConsoleQuery(http.MethodGet, "/auth/v1/roles", e2e.Query(0, 10, map[string]string{"name": name}), nil)
	e2e.RequireSuccess(t, resp, "list role "+name)
	item, ok := e2e.FindByName(resp.Slice(t), name)
	if !ok {
		t.Fatalf("role %s not found", name)
	}
	return e2e.StringField(t, item, "id")
}
