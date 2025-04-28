package auth

import (
	"reflect"

	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"
)

var (
	ResourceFieldNames          = resourceFieldNames
	ResourceFieldPointerGetters = resourceFieldPointerGetters
)

var (
	resourceFieldNames = map[string]apisecurity.ResourceType{
		"namespaces":           apisecurity.ResourceType_Namespaces,
		"service":              apisecurity.ResourceType_Services,
		"config_groups":        apisecurity.ResourceType_ConfigGroups,
		"route_rules":          apisecurity.ResourceType_RouteRules,
		"ratelimit_rules":      apisecurity.ResourceType_RateLimitRules,
		"circuitbreaker_rules": apisecurity.ResourceType_CircuitBreakerRules,
		"faultdetect_rules":    apisecurity.ResourceType_FaultDetectRules,
		"lane_rules":           apisecurity.ResourceType_LaneRules,
		"users":                apisecurity.ResourceType_Users,
		"user_groups":          apisecurity.ResourceType_UserGroups,
		"roles":                apisecurity.ResourceType_Roles,
		"auth_policies":        apisecurity.ResourceType_PolicyRules,
	}

	resourceFieldPointerGetters = map[apisecurity.ResourceType]func(*apisecurity.StrategyResources) reflect.Value{
		apisecurity.ResourceType_Namespaces: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetNamespaces() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.Namespaces)
		},
		apisecurity.ResourceType_Services: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetServices() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.Services)
		},
		apisecurity.ResourceType_ConfigGroups: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetConfigGroups() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.ConfigGroups)
		},
		apisecurity.ResourceType_RouteRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetRouteRules() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.RouteRules)
		},
		apisecurity.ResourceType_RateLimitRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetRatelimitRules() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.RatelimitRules)
		},
		apisecurity.ResourceType_CircuitBreakerRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetCircuitbreakerRules() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.CircuitbreakerRules)
		},
		apisecurity.ResourceType_FaultDetectRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetFaultdetectRules() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.FaultdetectRules)
		},
		apisecurity.ResourceType_LaneRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetLaneRules() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.LaneRules)
		},
		apisecurity.ResourceType_Users: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetUsers() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.Users)
		},
		apisecurity.ResourceType_UserGroups: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetUserGroups() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.UserGroups)
		},
		apisecurity.ResourceType_Roles: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetRoles() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.Roles)
		},
		apisecurity.ResourceType_PolicyRules: func(as *apisecurity.StrategyResources) reflect.Value {
			if as.GetAuthPolicies() == nil {
				return reflect.ValueOf(&[]*apisecurity.StrategyResourceEntry{})
			}
			return reflect.ValueOf(&as.AuthPolicies)
		},
	}
)

type pbStringValue interface {
	GetValue() string
}

// collectPrincipalEntry 将 Principal 转换为对应的 []authtypes.Principal 数组
func collectPrincipalEntry(ruleID string, uType PrincipalType, res []*apisecurity.Principal) []Principal {
	principals := make([]Principal, 0, len(res)+1)
	if len(res) == 0 {
		return principals
	}

	for index := range res {
		principals = append(principals, Principal{
			StrategyID:    ruleID,
			PrincipalID:   res[index].GetId().GetValue(),
			PrincipalType: uType,
		})
	}

	return principals
}

// collectResEntry 将资源ID转换为对应的 []authtypes.StrategyResource 数组
func collectResourceEntry(ruleId string, resType apisecurity.ResourceType,
	res reflect.Value, delete bool) []StrategyResource {
	if res.Kind() != reflect.Slice || res.Len() == 0 {
		return []StrategyResource{}
	}

	resEntries := make([]StrategyResource, 0, res.Len())
	for i := 0; i < res.Len(); i++ {
		item := res.Index(i).Elem()
		resId := item.FieldByName("Id").Interface().(pbStringValue)
		resName := item.FieldByName("Name").Interface().(pbStringValue)
		// 如果是添加的动作，那么需要进行归一化处理
		if !delete {
			// 归一化处理
			if resId.GetValue() == "*" || resName.GetValue() == "*" {
				return []StrategyResource{
					{
						StrategyID: ruleId,
						ResType:    int32(resType),
						ResID:      "*",
					},
				}
			}
		}

		entry := StrategyResource{
			StrategyID: ruleId,
			ResType:    int32(resType),
			ResID:      resId.GetValue(),
		}

		resEntries = append(resEntries, entry)
	}

	return resEntries
}
