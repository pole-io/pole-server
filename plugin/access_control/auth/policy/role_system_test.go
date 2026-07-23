package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

type roleSystemTestStore struct {
	storeapi.Store
	getMainUser func() (*authtypes.User, error)
	getRole     func(string) (*authtypes.Role, error)
	addRole     func(*authtypes.Role) error
	updateRole  func(*authtypes.Role) error
	getStrategy func(string) (*authtypes.StrategyDetail, error)
	startTx     func() (storeapi.Tx, error)
	addStrategy func(storeapi.Tx, *authtypes.StrategyDetail) error
}

func (s roleSystemTestStore) GetMainUser() (*authtypes.User, error) { return s.getMainUser() }

func (s roleSystemTestStore) GetRole(id string) (*authtypes.Role, error) {
	return s.getRole(id)
}

func (s roleSystemTestStore) UpdateRole(role *authtypes.Role) error {
	return s.updateRole(role)
}

func (s roleSystemTestStore) AddRole(role *authtypes.Role) error { return s.addRole(role) }

func (s roleSystemTestStore) GetStrategyDetail(id string) (*authtypes.StrategyDetail, error) {
	return s.getStrategy(id)
}

func (s roleSystemTestStore) StartTx() (storeapi.Tx, error) { return s.startTx() }

func (s roleSystemTestStore) AddStrategy(tx storeapi.Tx, strategy *authtypes.StrategyDetail) error {
	return s.addStrategy(tx, strategy)
}

type roleSystemTestTx struct{}

func (roleSystemTestTx) Commit() error              { return nil }
func (roleSystemTestTx) Rollback() error            { return nil }
func (roleSystemTestTx) GetDelegateTx() interface{} { return nil }
func (roleSystemTestTx) CreateReadView() error      { return nil }

func TestCreateCustomRoleSucceeds(t *testing.T) {
	storage := roleSystemTestStore{
		addRole: func(role *authtypes.Role) error {
			require.NotEmpty(t, role.ID)
			require.Equal(t, "custom", role.Name)
			require.Equal(t, authtypes.RoleTypeCustom, role.Type)
			return nil
		},
	}
	server := &Server{storage: storage}

	response := server.CreateRole(context.Background(), &apisecurity.Role{Name: "custom"})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestUpdateCustomRoleChangesDefinitionAndBindings(t *testing.T) {
	saved := &authtypes.Role{
		ID:       "custom-role-1",
		Name:     "before",
		Source:   "pole.io",
		Type:     authtypes.RoleTypeCustom,
		Comment:  "before",
		Metadata: map[string]string{"team": "before"},
	}
	storage := roleSystemTestStore{
		getRole: func(id string) (*authtypes.Role, error) {
			require.Equal(t, saved.ID, id)
			return saved, nil
		},
		updateRole: func(role *authtypes.Role) error {
			require.Equal(t, "after", role.Name)
			require.Equal(t, "after", role.Comment)
			require.Equal(t, map[string]string{"team": "after"}, role.Metadata)
			require.Equal(t, []authtypes.Principal{{PrincipalID: "user-1"}}, role.Users)
			return nil
		},
	}
	server := &Server{storage: storage}

	response := server.UpdateRole(context.Background(), &apisecurity.Role{
		Id:       saved.ID,
		Name:     "after",
		Source:   "pole.io",
		Comment:  "after",
		Metadata: map[string]string{"team": "after"},
		Users:    []*apisecurity.User{{Id: "user-1"}},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestUpdateSystemRoleDefinitionIsRejected(t *testing.T) {
	saved := authtypes.NewSystemRole(authtypes.SystemRoleDefinitions[0], "owner-1")
	storage := roleSystemTestStore{
		getRole: func(string) (*authtypes.Role, error) { return saved, nil },
	}
	server := &Server{storage: storage}

	response := server.UpdateRole(context.Background(), &apisecurity.Role{
		Id:   saved.ID,
		Name: "renamed-system-role",
	})

	require.Equal(t, uint32(apimodel.Code_BadRequest), response.GetCode())
}

func TestUpdateSystemRoleOnlyChangesPrincipalBindings(t *testing.T) {
	saved := &authtypes.Role{
		ID:       "pole-system-role-resource-reader",
		Name:     "resource-reader",
		Owner:    "owner-1",
		Source:   "pole-io",
		Type:     authtypes.RoleTypeSystem,
		Comment:  "读取全部业务资源",
		Metadata: map[string]string{"system-role": "resource-reader"},
	}
	storage := roleSystemTestStore{
		getRole: func(id string) (*authtypes.Role, error) {
			require.Equal(t, saved.ID, id)
			return saved, nil
		},
		updateRole: func(role *authtypes.Role) error {
			require.Equal(t, "resource-reader", role.Name)
			require.Equal(t, "pole-io", role.Source)
			require.Equal(t, "读取全部业务资源", role.Comment)
			require.Equal(t, map[string]string{"system-role": "resource-reader"}, role.Metadata)
			require.Equal(t, authtypes.RoleTypeSystem, role.Type)
			require.Equal(t, []authtypes.Principal{{PrincipalID: "user-1"}}, role.Users)
			require.Equal(t, []authtypes.Principal{{PrincipalID: "group-1"}}, role.UserGroups)
			return nil
		},
	}
	server := &Server{storage: storage}

	response := server.UpdateRole(context.Background(), &apisecurity.Role{
		Id:         saved.ID,
		Users:      []*apisecurity.User{{Id: "user-1"}},
		UserGroups: []*apisecurity.UserGroup{{Id: "group-1"}},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestDeleteSystemRoleIsRejected(t *testing.T) {
	storage := roleSystemTestStore{
		getRole: func(id string) (*authtypes.Role, error) {
			require.Equal(t, "pole-system-role-admin", id)
			return &authtypes.Role{ID: id, Type: authtypes.RoleTypeSystem}, nil
		},
	}
	server := &Server{storage: storage}

	response := server.DeleteRole(context.Background(), &apisecurity.Role{Id: "pole-system-role-admin"})

	require.Equal(t, uint32(apimodel.Code_BadRequest), response.GetCode())
}

func TestEnsureSystemRolesCreatesThreeRolesAndFixedPolicies(t *testing.T) {
	roles := map[string]*authtypes.Role{}
	policies := map[string]*authtypes.StrategyDetail{}
	storage := roleSystemTestStore{
		getMainUser: func() (*authtypes.User, error) {
			return &authtypes.User{ID: "owner-1"}, nil
		},
		getRole: func(id string) (*authtypes.Role, error) { return roles[id], nil },
		addRole: func(role *authtypes.Role) error {
			roles[role.ID] = role
			return nil
		},
		updateRole: func(role *authtypes.Role) error {
			roles[role.ID] = role
			return nil
		},
		getStrategy: func(id string) (*authtypes.StrategyDetail, error) { return policies[id], nil },
		startTx:     func() (storeapi.Tx, error) { return roleSystemTestTx{}, nil },
		addStrategy: func(_ storeapi.Tx, strategy *authtypes.StrategyDetail) error {
			policies[strategy.ID] = strategy
			return nil
		},
	}
	server := &Server{storage: storage}

	require.NoError(t, server.ensureSystemRoles())
	require.NoError(t, server.ensureSystemRoles())
	require.Len(t, roles, 3)
	require.Len(t, policies, 4)
	require.Equal(t, "admin", roles[authtypes.SystemRoleAdminID].Name)
	require.Equal(t, authtypes.RoleTypeSystem, roles[authtypes.SystemRoleResourceReaderID].Type)
	require.Equal(t, "owner-1", roles[authtypes.SystemRoleResourceWriterID].Owner)
	require.Equal(t, []authtypes.Principal{{
		PrincipalID:   authtypes.SystemRoleAdminID,
		PrincipalType: authtypes.PrincipalRole,
	}}, policies[authtypes.SystemRoleAdminPolicyID].Principals)
	require.ElementsMatch(t, []authtypes.Principal{
		{PrincipalID: authtypes.SystemRoleResourceWriterID, PrincipalType: authtypes.PrincipalRole},
		{PrincipalID: authtypes.SystemRoleResourceReaderID, PrincipalType: authtypes.PrincipalRole},
	}, policies[authtypes.SystemRoleResourceGuardPolicyID].Principals)
}
