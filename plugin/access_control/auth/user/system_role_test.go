package defaultuser

import (
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func TestResolveLoginRolePromotesDirectAdminRole(t *testing.T) {
	role := resolveLoginRole(
		&authtypes.User{ID: "user-1", Type: authtypes.SubAccountUserRole},
		func(principal authtypes.Principal) []*authtypes.Role {
			require.Equal(t, "user-1", principal.PrincipalID)
			return []*authtypes.Role{{ID: authtypes.SystemRoleAdminID, Type: authtypes.RoleTypeSystem}}
		},
		func(string) []string { return nil },
	)

	require.Equal(t, "admin", role)
}

func TestResolveLoginRolePromotesAdminRoleInheritedFromGroup(t *testing.T) {
	role := resolveLoginRole(
		&authtypes.User{ID: "user-1", Type: authtypes.SubAccountUserRole},
		func(principal authtypes.Principal) []*authtypes.Role {
			if principal.PrincipalType == authtypes.PrincipalGroup && principal.PrincipalID == "group-1" {
				return []*authtypes.Role{{ID: authtypes.SystemRoleAdminID, Type: authtypes.RoleTypeSystem}}
			}
			return nil
		},
		func(userID string) []string {
			require.Equal(t, "user-1", userID)
			return []string{"group-1"}
		},
	)

	require.Equal(t, "admin", role)
}

func TestResolveLoginRoleKeepsResourceRolesAsSubAccount(t *testing.T) {
	role := resolveLoginRole(
		&authtypes.User{ID: "user-1", Type: authtypes.SubAccountUserRole},
		func(authtypes.Principal) []*authtypes.Role {
			return []*authtypes.Role{{ID: authtypes.SystemRoleResourceWriterID, Type: authtypes.RoleTypeSystem}}
		},
		func(string) []string { return nil },
	)

	require.Equal(t, "sub", role)
}
