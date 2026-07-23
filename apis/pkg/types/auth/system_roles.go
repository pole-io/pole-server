package auth

const (
	SystemRoleAdminID                = "pole-system-role-admin"
	SystemRoleResourceReaderID       = "pole-system-role-resource-reader"
	SystemRoleResourceWriterID       = "pole-system-role-resource-writer"
	SystemRoleAdminPolicyID          = "pole-system-policy-admin"
	SystemRoleResourceReaderPolicyID = "pole-system-policy-resource-reader"
	SystemRoleResourceWriterPolicyID = "pole-system-policy-resource-writer"
	SystemRoleResourceGuardPolicyID  = "pole-system-policy-resource-guard"

	SystemRoleMetadataKey = "pole-system-role"
)

type SystemRoleDefinition struct {
	ID      string
	Name    string
	Comment string
}

func IsSystemRolePolicyID(id string) bool {
	switch id {
	case SystemRoleAdminPolicyID,
		SystemRoleResourceReaderPolicyID,
		SystemRoleResourceWriterPolicyID,
		SystemRoleResourceGuardPolicyID:
		return true
	default:
		return false
	}
}

var SystemRoleDefinitions = []SystemRoleDefinition{
	{ID: SystemRoleAdminID, Name: "admin", Comment: "管理控制面、权限与全部业务资源"},
	{ID: SystemRoleResourceReaderID, Name: "resource-reader", Comment: "读取全部业务资源"},
	{ID: SystemRoleResourceWriterID, Name: "resource-writer", Comment: "读取和写入全部业务资源"},
}

func GetSystemRoleDefinition(id string) (SystemRoleDefinition, bool) {
	for _, definition := range SystemRoleDefinitions {
		if definition.ID == id {
			return definition, true
		}
	}
	return SystemRoleDefinition{}, false
}

func NewSystemRole(definition SystemRoleDefinition, owner string) *Role {
	return &Role{
		ID:      definition.ID,
		Name:    definition.Name,
		Owner:   owner,
		Source:  "pole-io",
		Type:    RoleTypeSystem,
		Comment: definition.Comment,
		Metadata: map[string]string{
			SystemRoleMetadataKey: definition.Name,
		},
		Valid: true,
	}
}
