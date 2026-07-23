package paramcheck

import (
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

func TestHasBuiltInRolePrincipal(t *testing.T) {
	require.False(t, hasBuiltInRolePrincipal(nil))
	require.False(t, hasBuiltInRolePrincipal(&apisecurity.Principals{
		Roles: []*apisecurity.Principal{{Id: "custom-role"}},
	}))
	require.True(t, hasBuiltInRolePrincipal(&apisecurity.Principals{
		Roles: []*apisecurity.Principal{{Id: authtypes.SystemRoleAdminID}},
	}))
}
