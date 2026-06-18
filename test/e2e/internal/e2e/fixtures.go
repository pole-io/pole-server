//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"time"
)

type Names struct {
	Prefix    string
	Namespace string
	Gateway   string
	Order     string
	Payment   string
	Shadow    string
	User      string
	UserGroup string
	Role      string
	Policy    string
}

func NewNames(suite string) Names {
	prefix := fmt.Sprintf("e2e-%s-%d", suite, time.Now().UnixNano())
	return Names{
		Prefix:    prefix,
		Namespace: prefix + "-ns",
		Gateway:   prefix + "-gateway",
		Order:     prefix + "-order",
		Payment:   prefix + "-payment",
		Shadow:    prefix + "-shadow",
		User:      prefix + "-user",
		UserGroup: prefix + "-group",
		Role:      prefix + "-role",
		Policy:    prefix + "-policy",
	}
}

func MatchString(value string) map[string]any {
	return map[string]any{"type": "EXACT", "value": value, "value_type": "TEXT"}
}

func MatchArg(kind, key, value string) map[string]any {
	return map[string]any{"type": kind, "key": key, "value": MatchString(value)}
}

func Namespace(name string) map[string]any {
	return map[string]any{
		"name":    name,
		"comment": "e2e namespace",
		"owners":  "codex",
		"metadata": map[string]string{
			"e2e": "true",
		},
	}
}

func Service(namespace, name string) map[string]any {
	return map[string]any{
		"namespace":  namespace,
		"name":       name,
		"comment":    "e2e service",
		"department": "qa",
		"business":   "e2e",
		"metadata": map[string]string{
			"e2e": "true",
		},
	}
}

func Instance(namespace, service, host string, port int) map[string]any {
	return map[string]any{
		"namespace": namespace,
		"service":   service,
		"host":      host,
		"port":      port,
		"protocol":  "HTTP",
		"healthy":   true,
		"isolate":   false,
		"weight":    100,
		"metadata": map[string]string{
			"zone": "e2e",
		},
	}
}

func HeartbeatInstance(namespace, service, host string, port int) map[string]any {
	instance := Instance(namespace, service, host, port)
	instance["enable_health_check"] = true
	instance["health_check"] = map[string]any{
		"type": "HEARTBEAT",
		"heartbeat": map[string]any{
			"ttl": 5,
		},
	}
	return instance
}

func DiscoverRequest(namespace, service, typ string) map[string]any {
	return map[string]any{
		"type": typ,
		"service": map[string]any{
			"namespace": namespace,
			"name":      service,
		},
		"filter": map[string]any{},
	}
}

func RuleRelease(ruleID, ruleName, resource string) map[string]any {
	return map[string]any{
		"rule_id":      ruleID,
		"rule_name":    ruleName,
		"resource":     resource,
		"version":      1,
		"release_name": "e2e-normal",
		"release_type": "ReleaseTypeNormal",
		"active":       true,
	}
}

func GrayRuleRelease(ruleID, ruleName, resource string) map[string]any {
	release := RuleRelease(ruleID, ruleName, resource)
	release["version"] = 2
	release["release_name"] = "e2e-gray"
	release["release_type"] = "ReleaseTypeGray"
	release["client_labels"] = []map[string]any{{
		"key":   "e2e-gray",
		"value": MatchString("true"),
	}}
	return release
}

func User(name string) map[string]any {
	return map[string]any{
		"name":     name,
		"password": "user123",
		"comment":  "e2e user",
		"source":   "pole-io",
		"email":    name + "@example.com",
		"mobile":   "13000000000",
	}
}

func UserGroup(name string) map[string]any {
	return map[string]any{
		"name":    name,
		"comment": "e2e user group",
		"source":  "pole-io",
	}
}

func Role(name string) map[string]any {
	return map[string]any{
		"name":    name,
		"comment": "e2e role",
		"source":  "pole-io",
	}
}

func Policy(name string, principalID string, namespace string, action string) map[string]any {
	return UserNamespacePolicy(name, principalID, namespace, action)
}

func UserNamespacePolicy(name, userID, namespace, action string) map[string]any {
	return AuthPolicy(name, UserPrincipals(userID), NamespaceResources(namespace), action, FunctionsForAction(action))
}

func UserServicePolicy(name, userID, serviceID, namespace, service, action string) map[string]any {
	return AuthPolicy(name, UserPrincipals(userID), NamespaceAndServiceResources(namespace, serviceID, service), action, FunctionsForAction(action))
}

func UserNamespaceAndServicePolicy(name, userID, namespace, serviceID, service, action string) map[string]any {
	return AuthPolicy(name, UserPrincipals(userID), NamespaceAndServiceResources(namespace, serviceID, service), action, FunctionsForAction(action))
}

func GroupServicePolicy(name, groupID, serviceID, namespace, service, action string) map[string]any {
	return AuthPolicy(name, GroupPrincipals(groupID), NamespaceAndServiceResources(namespace, serviceID, service), action, FunctionsForAction(action))
}

func RoleServicePolicy(name, roleID, serviceID, namespace, service, action string, functions []string) map[string]any {
	return AuthPolicy(name, RolePrincipals(roleID), NamespaceAndServiceResources(namespace, serviceID, service), action, functions)
}

func AuthPolicy(name string, principals, resources map[string]any, action string, functions []string) map[string]any {
	if len(functions) == 0 {
		functions = FunctionsForAction(action)
	}
	return map[string]any{
		"name":       name,
		"action":     NormalizeAction(action),
		"source":     "pole-io",
		"comment":    "e2e policy",
		"principals": principals,
		"resources":  resources,
		"functions":  functions,
	}
}

func NormalizeAction(action string) string {
	switch action {
	case "DENY":
		return "DENY"
	default:
		return "ALLOW"
	}
}

func FunctionsForAction(action string) []string {
	switch action {
	case "READ", "ONLY_READ":
		return []string{"Describe*", "List*", "Get*", "DiscoverServices", "DiscoverInstances"}
	default:
		return []string{"*"}
	}
}

func UserPrincipals(userIDs ...string) map[string]any {
	return principals("users", userIDs...)
}

func GroupPrincipals(groupIDs ...string) map[string]any {
	return principals("groups", groupIDs...)
}

func RolePrincipals(roleIDs ...string) map[string]any {
	return principals("roles", roleIDs...)
}

func principals(kind string, ids ...string) map[string]any {
	items := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		items = append(items, map[string]any{"id": id})
	}
	return map[string]any{kind: items}
}

func NamespaceResources(namespace string) map[string]any {
	return map[string]any{
		"namespaces": []map[string]any{{
			"id":   namespace,
			"name": namespace,
		}},
	}
}

func ServiceResources(serviceID, namespace, service string) map[string]any {
	return map[string]any{
		"services": []map[string]any{{
			"id":        serviceID,
			"namespace": namespace,
			"name":      service,
		}},
	}
}

func NamespaceAndServiceResources(namespace, serviceID, service string) map[string]any {
	resources := NamespaceResources(namespace)
	resources["services"] = ServiceResources(serviceID, namespace, service)["services"]
	return resources
}

func AuthorizeResource(resourceType, resourceID string, principals map[string]any) map[string]any {
	return map[string]any{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"principals":    principals,
	}
}

func UserGroupWithUsers(name string, users ...map[string]any) map[string]any {
	group := UserGroup(name)
	group["relation"] = map[string]any{"users": users}
	return group
}

func RoleWithUsers(name string, users ...map[string]any) map[string]any {
	role := Role(name)
	role["users"] = users
	return role
}

func MCPServer(namespace, name, backendService string) map[string]any {
	return map[string]any{
		"name":         name,
		"namespace":    namespace,
		"description":  "e2e mcp server",
		"protocol":     "MCP_PROTOCOL_SSE",
		"backend_type": "service",
		"backend_service": map[string]any{
			"namespace": namespace,
			"name":      backendService,
		},
		"tools": []map[string]any{{
			"name":        "orders.query",
			"description": "query orders",
			"input_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		}},
	}
}

func A2AAgent(namespace, name, backendService string) map[string]any {
	return map[string]any{
		"name":        name,
		"namespace":   namespace,
		"description": "e2e a2a agent",
		"source_type": "custom",
		"backend": map[string]any{
			"type":      "service",
			"namespace": namespace,
			"service":   backendService,
		},
		"agent_card": map[string]any{
			"name":        name,
			"description": "e2e card",
			"version":     "1.0.0",
			"url":         "http://127.0.0.1/a2a",
			"skills": []map[string]any{{
				"id":          "order-query",
				"name":        "Order Query",
				"description": "query orders",
			}},
		},
		"skills": []map[string]any{{
			"id":          "order-query",
			"name":        "Order Query",
			"description": "query orders",
		}},
	}
}

func RouteRule(namespace, name, caller, callee string) map[string]any {
	return map[string]any{
		"name":        name,
		"enable":      true,
		"priority":    10,
		"description": "e2e route",
		"metadata":    map[string]string{"e2e": "true"},
		"routing_config": map[string]any{
			"@type": "type.googleapis.com/v1.RuleRoutingConfig",
			"caller": map[string]any{
				"namespace": namespace,
				"service":   caller,
			},
			"callee": map[string]any{
				"namespace": namespace,
				"service":   callee,
			},
			"rules": []map[string]any{{
				"name": "vip-route",
				"arguments": map[string]any{
					"matchMode":     "AND",
					"randomPercent": 0,
					"arguments": []map[string]any{
						MatchArg("HEADER", "x-tenant", "vip"),
					},
				},
				"destinations": []map[string]any{{
					"namespace": namespace,
					"service":   callee,
					"name":      "primary",
					"weight":    100,
					"isolate":   false,
					"labels": map[string]any{
						"version": MatchString("v1"),
					},
				}},
			}},
		},
	}
}

func RateLimitRule(namespace, name, service string) map[string]any {
	return map[string]any{
		"name":        name,
		"namespace":   namespace,
		"service":     service,
		"enable":      true,
		"priority":    10,
		"description": "e2e rate limit",
		"metadata":    map[string]string{"e2e": "true"},
		"rules": []map[string]any{{
			"method": map[string]any{
				"type":  "EXACT",
				"value": "POST /api/v1/payments",
			},
			"arguments": []map[string]any{
				MatchArg("HEADER", "x-tenant", "vip"),
			},
			"amounts": []map[string]any{{
				"validDuration": "1s",
				"maxAmount":     120,
			}},
			"resource": "QPS",
			"type":     "LOCAL",
			"action":   "REJECT",
			"failover": "FAILOVER_LOCAL",
		}},
	}
}

func CircuitBreakerRule(namespace, name, source, destination string) map[string]any {
	return map[string]any{
		"name":        name,
		"level":       "METHOD",
		"description": "e2e circuit breaker",
		"priority":    10,
		"metadata":    map[string]string{"e2e": "true"},
		"ruleMatcher": map[string]any{
			"source": map[string]any{
				"namespace": namespace,
				"service":   source,
			},
			"destination": map[string]any{
				"namespace": namespace,
				"service":   destination,
				"method":    map[string]any{"type": "EXACT", "value": "/api/v1/payments"},
			},
		},
		"block_configs": []map[string]any{{
			"block_config": map[string]any{
				"name": "payment-error-rate",
				"api": map[string]any{
					"protocol": "HTTP",
					"method":   "GET",
					"path":     MatchString("/api/v1/payments"),
				},
				"error_conditions": []map[string]any{{
					"inputType": "RET_CODE",
					"condition": map[string]any{"type": "RANGE", "value": "500-599"},
				}},
				"trigger_conditions": []map[string]any{{
					"triggerType":    "ERROR_RATE",
					"errorPercent":   50,
					"interval":       30,
					"minimumRequest": 10,
				}},
			},
			"max_ejection_percent": 100,
			"recoverCondition": map[string]any{
				"sleepWindow":        30,
				"consecutiveSuccess": 3,
			},
			"faultDetectConfig": map[string]any{"enable": false},
			"fallbackConfig":    map[string]any{"enable": true, "response": map[string]any{"code": 503, "body": "fallback"}},
		}},
	}
}

func FaultDetectRule(namespace, name, service string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": "e2e fault detect",
		"metadata":    map[string]string{"e2e": "true"},
		"targetService": map[string]any{
			"namespace": namespace,
			"service":   service,
			"api": map[string]any{
				"protocol": "HTTP",
				"method":   "GET",
				"path":     MatchString("/healthz"),
			},
		},
		"interval": 5,
		"timeout":  2,
		"port":     8080,
		"protocol": "HTTP",
		"httpConfig": map[string]any{
			"method":  "GET",
			"url":     "/healthz",
			"headers": []map[string]any{{"key": "x-e2e", "value": "true"}},
			"body":    "",
		},
	}
}

func LosslessRule(namespace, service string) map[string]any {
	return map[string]any{
		"namespace": namespace,
		"service":   service,
		"metadata":  map[string]string{"e2e": "true"},
		"lossless_online": map[string]any{
			"delay_register": map[string]any{
				"enable":                       true,
				"strategy":                     "health_check",
				"interval_second":              5,
				"health_check_protocol":        "HTTP",
				"health_check_method":          "GET",
				"health_check_path":            "/healthz",
				"health_check_interval_second": "1s",
			},
			"warmup": map[string]any{
				"enable":                        true,
				"interval_second":               30,
				"enable_overload_protection":    true,
				"overload_protection_threshold": 80,
				"curvature":                     2,
			},
		},
		"lossless_offline": map[string]any{
			"enable":          true,
			"interval_second": 10,
		},
	}
}

func LaneGroup(namespace, name, entryService, targetService string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": "e2e lane group",
		"metadata":    map[string]string{"e2e": "true"},
		"entries": []map[string]any{{
			"type": "service",
			"selector": map[string]any{
				"@type":     "type.googleapis.com/v1.ServiceSelector",
				"namespace": namespace,
				"service":   entryService,
			},
		}},
		"destinations": []map[string]any{{
			"namespace": namespace,
			"service":   targetService,
			"name":      "payment",
			"weight":    100,
			"isolate":   false,
			"labels": map[string]any{
				"lane": MatchString("blue"),
			},
		}},
		"rules": []map[string]any{{
			"name":              "blue-lane",
			"groupName":         name,
			"labelKey":          "lane",
			"defaultLabelValue": "base",
			"matchMode":         "PERMISSIVE",
			"trafficMatchRule": map[string]any{
				"matchMode": "AND",
				"arguments": []map[string]any{
					MatchArg("HEADER", "x-lane", "blue"),
				},
			},
		}},
	}
}

func TrafficSecurityRule(namespace, name, service string) map[string]any {
	return map[string]any{
		"name":           name,
		"namespace":      namespace,
		"service":        service,
		"description":    "e2e traffic security",
		"priority":       10,
		"enable":         true,
		"default_action": "TRAFFIC_SECURITY_DENY",
		"metadata":       map[string]string{"e2e": "true"},
		"policies": []map[string]any{{
			"api": map[string]any{
				"protocol": "HTTP",
				"method":   "GET",
				"path":     MatchString("/orders"),
			},
			"traffic_match_rule": map[string]any{
				"matchMode":     "AND",
				"randomPercent": 0,
				"arguments": []map[string]any{
					MatchArg("HEADER", "x-user-type", "internal"),
				},
			},
			"action": "TRAFFIC_SECURITY_ALLOW",
			"reject_effect": map[string]any{
				"status_code": 403,
				"code":        "FORBIDDEN",
				"message":     "e2e denied",
			},
		}},
	}
}

func TrafficMirrorRule(namespace, name, sourceService, targetService string) map[string]any {
	return map[string]any{
		"name":        name,
		"namespace":   namespace,
		"service":     sourceService,
		"description": "e2e mirror",
		"priority":    10,
		"enable":      true,
		"metadata":    map[string]string{"e2e": "true"},
		"rules": []map[string]any{{
			"source": map[string]any{
				"namespace": namespace,
				"service":   sourceService,
				"traffic_match_rule": map[string]any{
					"matchMode": "AND",
					"arguments": []map[string]any{
						MatchArg("HEADER", "x-traffic-mirror", "true"),
					},
				},
			},
			"destination": map[string]any{
				"namespace": namespace,
				"service":   targetService,
				"labels": map[string]any{
					"version": MatchString("shadow"),
				},
			},
			"mirror_percent": 30,
			"duration":       "300s",
			"disable":        false,
		}},
	}
}

func TrafficMockRule(namespace, name, service string) map[string]any {
	return map[string]any{
		"name":        name,
		"namespace":   namespace,
		"service":     service,
		"description": "e2e mock",
		"priority":    10,
		"enable":      true,
		"metadata":    map[string]string{"e2e": "true"},
		"rules": []map[string]any{{
			"source": map[string]any{
				"namespace": namespace,
				"service":   service,
				"api": map[string]any{
					"protocol": "HTTP",
					"method":   "GET",
					"path":     MatchString("/orders"),
				},
				"traffic_match_rule": map[string]any{
					"matchMode": "AND",
					"arguments": []map[string]any{
						MatchArg("HEADER", "x-mock", "true"),
					},
				},
			},
			"response": map[string]any{
				"status_code": 200,
				"headers":     map[string]string{"x-e2e": "mock"},
				"body":        `{"mock":true}`,
				"code":        "OK",
				"message":     "mocked",
			},
			"mock_percent": 100,
			"delay":        "1s",
			"disable":      false,
		}},
	}
}
