package policy

import (
	"fmt"
	"maps"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func (svr *Server) ensureSystemRoles() error {
	svr.systemRolesMu.Lock()
	defer svr.systemRolesMu.Unlock()

	mainUser, err := svr.storage.GetMainUser()
	if err != nil {
		return err
	}
	if mainUser == nil || mainUser.ID == "" {
		return nil
	}

	for _, definition := range authtypes.SystemRoleDefinitions {
		saved, err := svr.storage.GetRole(definition.ID)
		if err != nil {
			return err
		}
		if saved == nil {
			if err := svr.storage.AddRole(authtypes.NewSystemRole(definition, mainUser.ID)); err != nil {
				return fmt.Errorf("create system role %s: %w", definition.Name, err)
			}
			continue
		}

		canonical := authtypes.NewSystemRole(definition, mainUser.ID)
		canonical.Users = saved.Users
		canonical.UserGroups = saved.UserGroups
		if saved.Name != canonical.Name || saved.Owner != canonical.Owner || saved.Source != canonical.Source ||
			saved.Type != canonical.Type || saved.Comment != canonical.Comment ||
			!maps.Equal(saved.Metadata, canonical.Metadata) {
			if err := svr.storage.UpdateRole(canonical); err != nil {
				return fmt.Errorf("repair system role %s: %w", definition.Name, err)
			}
		}
	}

	for _, policy := range systemRolePolicies() {
		saved, err := svr.storage.GetStrategyDetail(policy.ID)
		if err != nil {
			return err
		}
		if saved != nil {
			if !containsAllFunctions(saved.CalleeMethods, policy.CalleeMethods) {
				if err := svr.storage.UpdateStrategy(policy); err != nil {
					return fmt.Errorf("repair system role policy %s: %w", policy.Name, err)
				}
			}
			continue
		}
		tx, err := svr.storage.StartTx()
		if err != nil {
			return err
		}
		if err := svr.storage.AddStrategy(tx, policy); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("create system role policy %s: %w", policy.Name, err)
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("commit system role policy %s: %w", policy.Name, err)
		}
	}
	return nil
}

func containsAllFunctions(saved, canonical []string) bool {
	available := make(map[string]struct{}, len(saved))
	for _, function := range saved {
		available[function] = struct{}{}
	}
	for _, function := range canonical {
		if _, ok := available[function]; !ok {
			return false
		}
	}
	return true
}

func systemRolePolicies() []*authtypes.StrategyDetail {
	resources := allResourceWildcards("")
	admin := newSystemRolePolicy(
		authtypes.SystemRoleAdminPolicyID,
		"系统角色：管理员",
		apisecurity.AuthAction_ALLOW.String(),
		authtypes.SystemRoleAdminID,
		[]string{"*"},
		resources,
	)
	reader := newSystemRolePolicy(
		authtypes.SystemRoleResourceReaderPolicyID,
		"系统角色：资源全读",
		apisecurity.AuthAction_ALLOW.String(),
		authtypes.SystemRoleResourceReaderID,
		[]string{"Describe*", "Discover*", "Watch*", "Get*", "List*", "Export*"},
		resources,
	)
	writer := newSystemRolePolicy(
		authtypes.SystemRoleResourceWriterPolicyID,
		"系统角色：资源全写",
		apisecurity.AuthAction_ALLOW.String(),
		authtypes.SystemRoleResourceWriterID,
		[]string{"*"},
		resources,
	)
	guard := newSystemRolePolicy(
		authtypes.SystemRoleResourceGuardPolicyID,
		"系统角色：资源角色管理面隔离",
		apisecurity.AuthAction_DENY.String(),
		authtypes.SystemRoleResourceWriterID,
		managementFunctions(),
		resources,
	)
	guard.Principals = append(guard.Principals, authtypes.Principal{
		PrincipalID:   authtypes.SystemRoleResourceReaderID,
		PrincipalType: authtypes.PrincipalRole,
	})
	return []*authtypes.StrategyDetail{admin, reader, writer, guard}
}

func newSystemRolePolicy(id, name, action, roleID string, functions []string,
	resources []authtypes.StrategyResource) *authtypes.StrategyDetail {
	policyResources := make([]authtypes.StrategyResource, len(resources))
	copy(policyResources, resources)
	for i := range policyResources {
		policyResources[i].StrategyID = id
	}
	return &authtypes.StrategyDetail{
		ID:            id,
		Name:          name,
		Action:        action,
		Comment:       "Pole built-in role policy; immutable",
		Default:       true,
		Source:        "pole-io",
		Revision:      id + "-v1",
		Valid:         true,
		CalleeMethods: functions,
		Resources:     policyResources,
		Principals: []authtypes.Principal{{
			PrincipalID:   roleID,
			PrincipalType: authtypes.PrincipalRole,
		}},
	}
}

func allResourceWildcards(strategyID string) []authtypes.StrategyResource {
	resources := make([]authtypes.StrategyResource, 0, len(apisecurity.ResourceType_value))
	for _, value := range apisecurity.ResourceType_value {
		resources = append(resources, authtypes.StrategyResource{
			StrategyID: strategyID,
			ResType:    apisecurity.ResourceType(value),
			ResID:      "*",
		})
	}
	return resources
}

func managementFunctions() []string {
	return []string{
		string(authtypes.DescribeSystemConfiguration),
		string(authtypes.CreateUsers), string(authtypes.DeleteUsers), string(authtypes.UpdateUser),
		string(authtypes.UpdateUserPassword), string(authtypes.EnableUserToken), string(authtypes.ResetUserToken),
		string(authtypes.DescribeUsers), string(authtypes.DescribeUserToken),
		string(authtypes.CreateUserGroup), string(authtypes.UpdateUserGroups), string(authtypes.DeleteUserGroups),
		string(authtypes.DescribeUserGroups), string(authtypes.DescribeUserGroupDetail),
		string(authtypes.DescribeUserGroupToken), string(authtypes.EnableUserGroupToken), string(authtypes.ResetUserGroupToken),
		string(authtypes.CreateAuthPolicy), string(authtypes.UpdateAuthPolicies), string(authtypes.DeleteAuthPolicies),
		string(authtypes.DescribeAuthPolicies), string(authtypes.DescribeAuthPolicyDetail),
		string(authtypes.DescribePrincipalResources), string(authtypes.AuthorizeResources),
		string(authtypes.CreateAuthRoles), string(authtypes.UpdateAuthRoles), string(authtypes.DeleteAuthRoles),
		string(authtypes.DescribeAuthRoles), string(authtypes.DescribeAuthRoleDetail),
		string(authtypes.CreateLogicalServices), string(authtypes.UpdateLogicalServices),
		string(authtypes.DeleteLogicalServices), string(authtypes.DescribeLogicalServices),
		string(authtypes.BindServiceEnvironments), string(authtypes.UnbindServiceEnvironments),
		"DescribeServer*", "CloseConnections", "FreeOSMemory", "ReleaseLeaderElection", "UpdateLogOutputLevel",
	}
}
