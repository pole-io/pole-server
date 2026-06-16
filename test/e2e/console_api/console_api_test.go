//go:build e2e
// +build e2e

package consoleapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pole-io/pole-server/test/e2e/internal/e2e"
)

func TestConsoleAPIResourcesAndGovernanceRules(t *testing.T) {
	env := e2e.Start(t, e2e.Options{Suite: "console-api", PortBase: 30000, ConsoleOpen: true, ClientOpen: false})
	api := e2e.NewHTTPClient(t, env)
	api.LoginAsMainUser()
	names := e2e.NewNames("console")

	createBaseResources(t, api, names)

	t.Run("namespace create list update delete", func(t *testing.T) {
		namespace := names.Prefix + "-tmp-ns"
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{e2e.Namespace(namespace)}), "create temporary namespace")
		updated := e2e.Namespace(namespace)
		updated["comment"] = "e2e namespace updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/core/v1/namespaces", []map[string]any{updated}), "update namespace")
		resp := api.ConsoleQuery(http.MethodGet, "/core/v1/namespaces", e2e.Query(0, 10, map[string]string{"name": namespace}), nil)
		e2e.RequireSuccess(t, resp, "list namespace")
		if _, ok := e2e.FindByName(resp.Slice(t), namespace); !ok {
			t.Fatalf("namespace %s not found after update", namespace)
		}
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces/delete", []map[string]any{{"name": namespace}}), "delete namespace")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/core/v1/namespaces", e2e.Query(0, 10, map[string]string{"name": namespace}), nil), namespace, "list namespace after delete")
	})

	t.Run("mcp server create list update tools delete", func(t *testing.T) {
		name := names.Prefix + "-mcp"
		payload := e2e.MCPServer(names.Namespace, name, names.Gateway)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/ai/mcp/v1/servers", []map[string]any{payload}), "create mcp server")
		list := api.ConsoleQuery(http.MethodGet, "/ai/mcp/v1/servers", e2e.Query(0, 10, map[string]string{"name": name, "namespace": names.Namespace}), nil)
		e2e.RequireSuccess(t, list, "list mcp server")
		item, ok := e2e.FindByName(list.Slice(t), name)
		if !ok {
			t.Fatalf("mcp server %s not found", name)
		}
		id := e2e.StringField(t, item, "id")
		payload["id"] = id
		payload["description"] = "e2e mcp server updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/ai/mcp/v1/servers", []map[string]any{payload}), "update mcp server")
		tools := api.ConsoleQuery(http.MethodGet, "/ai/mcp/v1/server/tools", e2e.Query(0, 10, map[string]string{"server_id": id}), nil)
		e2e.RequireSuccess(t, tools, "list mcp tools")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/ai/mcp/v1/servers/delete", map[string]any{"server_ids": []string{id}}), "delete mcp server")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/ai/mcp/v1/servers", e2e.Query(0, 10, map[string]string{"name": name, "namespace": names.Namespace}), nil), name, "list mcp server after delete")
		requireResponseNotContaining(t, api.ConsoleQuery(http.MethodGet, "/ai/mcp/v1/server/tools", e2e.Query(0, 10, map[string]string{"server_id": id}), nil), name, "list mcp tools after delete")
	})

	t.Run("a2a agent create list update card skills delete", func(t *testing.T) {
		name := names.Prefix + "-agent"
		payload := e2e.A2AAgent(names.Namespace, name, names.Gateway)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/ai/a2a/v1/agents", []map[string]any{payload}), "create a2a agent")
		list := api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agents", e2e.Query(0, 10, map[string]string{"name": name, "namespace": names.Namespace}), nil)
		e2e.RequireSuccess(t, list, "list a2a agent")
		item, ok := e2e.FindByName(list.Slice(t), name)
		if !ok {
			t.Fatalf("a2a agent %s not found", name)
		}
		id := e2e.StringField(t, item, "id")
		payload["id"] = id
		payload["description"] = "e2e a2a agent updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/ai/a2a/v1/agents", []map[string]any{payload}), "update a2a agent")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agents/"+id+"/card", nil, nil), "get a2a card")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agent/skills", e2e.Query(0, 10, map[string]string{"agent_id": id}), nil), "list a2a skills")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/ai/a2a/v1/agents/delete", map[string]any{"agent_ids": []string{id}}), "delete a2a agent")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agents", e2e.Query(0, 10, map[string]string{"name": name, "namespace": names.Namespace}), nil), name, "list a2a agent after delete")
		requireResponseNotContaining(t, api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agents/"+id+"/card", nil, nil), name, "get a2a card after delete")
		requireResponseNotContaining(t, api.ConsoleQuery(http.MethodGet, "/ai/a2a/v1/agent/skills", e2e.Query(0, 10, map[string]string{"agent_id": id}), nil), "order-query", "list a2a skills after delete")
	})

	t.Run("service alias instance lifecycle", func(t *testing.T) {
		serviceName := names.Prefix + "-temporary-service"
		service := e2e.Service(names.Namespace, serviceName)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services", []map[string]any{service}), "create temporary service")
		service["comment"] = "e2e service updated"
		service["business"] = "updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/naming/v1/services", []map[string]any{service}), "update temporary service")
		serviceList := api.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"name": serviceName, "namespace": names.Namespace}), nil)
		e2e.RequireSuccess(t, serviceList, "list temporary service")
		if _, ok := e2e.FindByName(serviceList.Slice(t), serviceName); !ok {
			t.Fatalf("service %s not found after update", serviceName)
		}

		aliasName := names.Prefix + "-alias"
		alias := map[string]any{
			"alias":           aliasName,
			"alias_namespace": names.Namespace,
			"service":         names.Payment,
			"namespace":       names.Namespace,
			"comment":         "e2e alias",
		}
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/service/alias", alias), "create service alias")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/naming/v1/service/aliases", e2e.Query(0, 10, map[string]string{"alias": aliasName, "alias_namespace": names.Namespace}), nil), "list service aliases")
		alias["comment"] = "e2e alias updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/naming/v1/service/alias", alias), "update service alias")

		instance := e2e.Instance(names.Namespace, names.Payment, "10.0.0.10", 18080)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/instances", []map[string]any{instance}), "create instance")
		instances := api.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "service": names.Payment}), nil)
		e2e.RequireSuccess(t, instances, "list instances")
		instance["weight"] = 80
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/naming/v1/instances", []map[string]any{instance}), "update instance")
		e2e.RequireSuccess(t, api.Console(http.MethodGet, "/naming/v1/service/subscribers", nil), "list service subscribers")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/instances/delete", []map[string]any{instance}), "delete instance")
		e2e.RequireNotFoundByField(t, api.ConsoleQuery(http.MethodGet, "/naming/v1/instances", e2e.Query(0, 10, map[string]string{"namespace": names.Namespace, "service": names.Payment}), nil), "host", "10.0.0.10", "list instances after delete")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/service/aliases/delete", []map[string]any{alias}), "delete service alias")
		e2e.RequireNotFoundByField(t, api.ConsoleQuery(http.MethodGet, "/naming/v1/service/aliases", e2e.Query(0, 10, map[string]string{"alias": aliasName, "alias_namespace": names.Namespace}), nil), "alias", aliasName, "list service alias after delete")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services/delete", []map[string]any{service}), "delete temporary service")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 10, map[string]string{"name": serviceName, "namespace": names.Namespace}), nil), serviceName, "list temporary service after delete")
	})

	t.Run("governance rules current and release lifecycle", func(t *testing.T) {
		cases := governanceCases(names)
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path, []map[string]any{tc.payload}), "create "+tc.name)
				list := api.ConsoleQuery(http.MethodGet, tc.path, e2e.Query(0, 10, tc.query), nil)
				e2e.RequireSuccess(t, list, "list "+tc.name)
				item, ok := e2e.FindByName(list.Slice(t), tc.ruleName)
				if !ok {
					t.Fatalf("rule %s not found in %s list", tc.ruleName, tc.name)
				}
				id := e2e.StringField(t, item, "id")
				tc.payload["id"] = id
				tc.payload["description"] = "e2e updated " + tc.name
				e2e.RequireSuccess(t, api.Console(http.MethodPut, tc.path, []map[string]any{tc.payload}), "update "+tc.name)
				e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, tc.path+"/detail", map[string]string{"id": id}, nil), "detail "+tc.name)
				normalRelease := e2e.RuleRelease(id, tc.ruleName, tc.releaseResource)
				grayRelease := e2e.GrayRuleRelease(id, tc.ruleName, tc.releaseResource)
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/releases", []map[string]any{normalRelease}), "release "+tc.name)
				e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, tc.path+"/releases", e2e.Query(0, 10, map[string]string{"id": id}), nil), "list releases "+tc.name)
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/releases", []map[string]any{grayRelease}), "gray release "+tc.name)
				e2e.RequireSuccess(t, api.Console(http.MethodPut, tc.path+"/releases/stopbeta", []map[string]any{grayRelease}), "stop gray release "+tc.name)
				if tc.supportsRollback {
					e2e.RequireSuccess(t, api.Console(http.MethodPut, tc.path+"/releases/rollback", []map[string]any{normalRelease}), "rollback release "+tc.name)
				}
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/releases/delete", []map[string]any{normalRelease}), "delete release "+tc.name)
				e2e.RequireSuccess(t, api.Console(http.MethodPost, tc.path+"/delete", []map[string]any{{"id": id}}), "delete "+tc.name)
				e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, tc.path, e2e.Query(0, 10, tc.query), nil), tc.ruleName, "list "+tc.name+" after delete")
			})
		}
	})

	t.Run("auth user group role policy lifecycle", func(t *testing.T) {
		user := e2e.User(names.User)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/users", []map[string]any{user}), "create user")
		users := api.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": names.User}), nil)
		e2e.RequireSuccess(t, users, "list users")
		userItem, ok := e2e.FindByName(users.Slice(t), names.User)
		if !ok {
			t.Fatalf("user %s not found", names.User)
		}
		userID := e2e.StringField(t, userItem, "id")
		user["id"] = userID
		user["comment"] = "e2e user updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/users", []map[string]any{user}), "update user")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/user/token", map[string]string{"id": userID}, nil), "get user token")
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/user/token/enable", map[string]any{"id": userID, "token_enable": true}), "enable user token")
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/user/token/refresh", map[string]any{"id": userID}), "refresh user token")

		group := e2e.UserGroup(names.UserGroup)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/usergroups", []map[string]any{group}), "create user group")
		groups := api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroups", e2e.Query(0, 10, map[string]string{"name": names.UserGroup}), nil)
		e2e.RequireSuccess(t, groups, "list user groups")
		groupItem, ok := e2e.FindByName(groups.Slice(t), names.UserGroup)
		if !ok {
			t.Fatalf("user group %s not found", names.UserGroup)
		}
		groupID := e2e.StringField(t, groupItem, "id")
		group["id"] = groupID
		group["comment"] = "e2e user group updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/usergroups", []map[string]any{group}), "update user group")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroup/detail", map[string]string{"id": groupID}, nil), "get user group detail")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroup/token", map[string]string{"id": groupID}, nil), "get user group token")
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/usergroup/token/enable", map[string]any{"id": groupID, "token_enable": true}), "enable user group token")
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/usergroup/token/refresh", map[string]any{"id": groupID}), "refresh user group token")

		role := e2e.Role(names.Role)
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/roles", []map[string]any{role}), "create role")
		roles := api.ConsoleQuery(http.MethodGet, "/auth/v1/roles", e2e.Query(0, 10, map[string]string{"name": names.Role}), nil)
		e2e.RequireSuccess(t, roles, "list roles")
		roleItem, ok := e2e.FindByName(roles.Slice(t), names.Role)
		if !ok {
			t.Fatalf("role %s not found", names.Role)
		}
		roleID := e2e.StringField(t, roleItem, "id")
		role["id"] = roleID
		role["comment"] = "e2e role updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/roles", []map[string]any{role}), "update role")

		policy := e2e.Policy(names.Policy, userID, names.Namespace, "READ_WRITE")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/policies", []map[string]any{policy}), "create policy")
		policies := api.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy}), nil)
		e2e.RequireSuccess(t, policies, "list policies")
		policyItem, ok := e2e.FindByName(policies.Slice(t), names.Policy)
		if !ok {
			t.Fatalf("policy %s not found", names.Policy)
		}
		policyID := e2e.StringField(t, policyItem, "id")
		policy["id"] = policyID
		policy["comment"] = "e2e policy updated"
		e2e.RequireSuccess(t, api.Console(http.MethodPut, "/auth/v1/policies", []map[string]any{policy}), "update policy")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/resources/authorize", []map[string]any{
			e2e.AuthorizeResource("Namespaces", names.Namespace, e2e.UserPrincipals(userID)),
		}), "authorize resources")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/principal/resources", map[string]string{"principal_id": userID, "principal_type": "PrincipalUser"}, nil), "query principal resources")
		e2e.RequireSuccess(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/resources/principals", map[string]string{"res_type": "Namespaces", "res_id": names.Namespace}, nil), "query resource principals")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/policies/delete", []map[string]any{{"id": policyID}}), "delete policy")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/policies", e2e.Query(0, 10, map[string]string{"name": names.Policy}), nil), names.Policy, "list policy after delete")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/roles/delete", []map[string]any{{"id": roleID}}), "delete role")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/roles", e2e.Query(0, 10, map[string]string{"name": names.Role}), nil), names.Role, "list role after delete")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/usergroups/delete", []map[string]any{{"id": groupID}}), "delete user group")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/usergroups", e2e.Query(0, 10, map[string]string{"name": names.UserGroup}), nil), names.UserGroup, "list user group after delete")
		e2e.RequireSuccess(t, api.Console(http.MethodPost, "/auth/v1/users/delete", []map[string]any{{"id": userID}}), "delete user")
		e2e.RequireNotFoundByName(t, api.ConsoleQuery(http.MethodGet, "/auth/v1/users", e2e.Query(0, 10, map[string]string{"name": names.User}), nil), names.User, "list user after delete")
	})
}

type governanceCase struct {
	name             string
	path             string
	ruleName         string
	releaseResource  string
	query            map[string]string
	payload          map[string]any
	supportsRollback bool
}

func governanceCases(names e2e.Names) []governanceCase {
	return []governanceCase{
		{name: "route", path: "/naming/v1/routings", ruleName: names.Prefix + "-route", releaseResource: "RouteRules", query: map[string]string{"name": names.Prefix + "-route"}, payload: e2e.RouteRule(names.Namespace, names.Prefix+"-route", names.Order, names.Payment), supportsRollback: true},
		{name: "ratelimit", path: "/naming/v1/ratelimits", ruleName: names.Prefix + "-ratelimit", releaseResource: "RateLimitRules", query: map[string]string{"name": names.Prefix + "-ratelimit", "namespace": names.Namespace, "service": names.Payment}, payload: e2e.RateLimitRule(names.Namespace, names.Prefix+"-ratelimit", names.Payment), supportsRollback: true},
		{name: "circuitbreaker", path: "/naming/v1/circuitbreakers", ruleName: names.Prefix + "-circuitbreaker", releaseResource: "CircuitBreakerRules", query: map[string]string{"name": names.Prefix + "-circuitbreaker"}, payload: e2e.CircuitBreakerRule(names.Namespace, names.Prefix+"-circuitbreaker", names.Order, names.Payment), supportsRollback: true},
		{name: "faultdetect", path: "/naming/v1/faultdetectors", ruleName: names.Prefix + "-faultdetect", releaseResource: "FaultDetectRules", query: map[string]string{"name": names.Prefix + "-faultdetect"}, payload: e2e.FaultDetectRule(names.Namespace, names.Prefix+"-faultdetect", names.Payment), supportsRollback: true},
		{name: "lossless", path: "/naming/v1/lossless", ruleName: names.Payment, releaseResource: "LosslessRules", query: map[string]string{"namespace": names.Namespace, "service": names.Payment}, payload: e2e.LosslessRule(names.Namespace, names.Payment)},
		{name: "lane", path: "/naming/v1/lane/groups", ruleName: names.Prefix + "-lane", releaseResource: "LaneRules", query: map[string]string{"name": names.Prefix + "-lane"}, payload: e2e.LaneGroup(names.Namespace, names.Prefix+"-lane", names.Gateway, names.Payment), supportsRollback: true},
		{name: "traffic-security", path: "/naming/v1/traffic/security", ruleName: names.Prefix + "-security", releaseResource: "TrafficSecurityRules", query: map[string]string{"name": names.Prefix + "-security", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficSecurityRule(names.Namespace, names.Prefix+"-security", names.Gateway)},
		{name: "traffic-mirror", path: "/naming/v1/traffic/mirrors", ruleName: names.Prefix + "-mirror", releaseResource: "TrafficMirrorRules", query: map[string]string{"name": names.Prefix + "-mirror", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficMirrorRule(names.Namespace, names.Prefix+"-mirror", names.Gateway, names.Shadow)},
		{name: "traffic-mock", path: "/naming/v1/traffic/mocks", ruleName: names.Prefix + "-mock", releaseResource: "TrafficMockRules", query: map[string]string{"name": names.Prefix + "-mock", "namespace": names.Namespace, "service": names.Gateway}, payload: e2e.TrafficMockRule(names.Namespace, names.Prefix+"-mock", names.Gateway)},
	}
}

func requireResponseNotContaining(t *testing.T, resp e2e.APIResponse, text, action string) {
	t.Helper()
	if resp.Code == e2e.ExecuteSuccess && strings.Contains(string(resp.Data), text) {
		t.Fatalf("%s: response still contains %q: %s", action, text, string(resp.Data))
	}
}

func createBaseResources(t *testing.T, api *e2e.HTTPClient, names e2e.Names) {
	t.Helper()
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/core/v1/namespaces", []map[string]any{e2e.Namespace(names.Namespace)}), "create namespace")
	services := []map[string]any{
		e2e.Service(names.Namespace, names.Gateway),
		e2e.Service(names.Namespace, names.Order),
		e2e.Service(names.Namespace, names.Payment),
		e2e.Service(names.Namespace, names.Shadow),
	}
	e2e.RequireSuccess(t, api.Console(http.MethodPost, "/naming/v1/services", services), "create services")
	list := api.ConsoleQuery(http.MethodGet, "/naming/v1/services", e2e.Query(0, 20, map[string]string{"namespace": names.Namespace}), nil)
	e2e.RequireSuccess(t, list, "list services")
}
