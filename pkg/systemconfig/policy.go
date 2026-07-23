package systemconfig

// EditPolicy is assigned explicitly for every registered setting. Keeping the
// policy beside the registration prevents a newly-added YAML field from
// becoming remotely editable by accident.
type EditPolicy struct {
	Editable    bool
	Reason      string
	Description string
	Validation  Validation
}

func ApplyEditPolicies(registrations []Registration, policies map[string]EditPolicy) {
	for index := range registrations {
		policy, ok := policies[registrations[index].Definition.Key]
		if !ok {
			registrations[index].Definition.Editable = false
			registrations[index].Definition.EditReason = "尚未完成字段级编辑评审"
			continue
		}
		registrations[index].Definition.Editable = policy.Editable
		registrations[index].Definition.EditReason = policy.Reason
		registrations[index].Definition.Description = policy.Description
		registrations[index].Definition.Validation = policy.Validation
	}
}
