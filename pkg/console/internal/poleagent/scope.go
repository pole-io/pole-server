package poleagent

import (
	"fmt"
	"strings"
)

// NamespaceScopeMode 表示一次 Agent 对话允许访问的环境范围模式。
type NamespaceScopeMode string

const (
	// NamespaceScopeSingle 表示只允许访问一个明确的 Namespace。
	NamespaceScopeSingle NamespaceScopeMode = "single"
	// NamespaceScopeCrossEnvironment 表示用户显式选择了多个 Namespace 进行只读比较。
	NamespaceScopeCrossEnvironment NamespaceScopeMode = "cross_environment"
)

// NamespaceScope 是服务端校验后注入一次 Agent 对话的 Namespace 白名单。
type NamespaceScope struct {
	Mode       NamespaceScopeMode `json:"mode"`
	Namespaces []string           `json:"namespaces"`
}

// Validate 校验范围模式、数量以及 Namespace 值的唯一性。
func (s NamespaceScope) Validate() error {
	switch s.Mode {
	case NamespaceScopeSingle:
		if len(s.Namespaces) != 1 {
			return fmt.Errorf("single namespace scope requires exactly one namespace")
		}
	case NamespaceScopeCrossEnvironment:
		if len(s.Namespaces) < 2 {
			return fmt.Errorf("cross-environment scope requires at least two namespaces")
		}
	default:
		return fmt.Errorf("unsupported namespace scope mode %q", s.Mode)
	}

	seen := make(map[string]struct{}, len(s.Namespaces))
	for _, namespace := range s.Namespaces {
		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			return fmt.Errorf("namespace scope contains an empty namespace")
		}
		if _, duplicate := seen[namespace]; duplicate {
			return fmt.Errorf("namespace scope contains duplicate namespace %q", namespace)
		}
		seen[namespace] = struct{}{}
	}
	return nil
}

// Contains 判断指定 Namespace 是否位于本次对话范围内。
func (s NamespaceScope) Contains(namespace string) bool {
	namespace = strings.TrimSpace(namespace)
	for _, candidate := range s.Namespaces {
		if strings.TrimSpace(candidate) == namespace {
			return true
		}
	}
	return false
}

// IsSingle 判断当前范围是否为合法的单 Namespace 模式。
func (s NamespaceScope) IsSingle() bool {
	return s.Mode == NamespaceScopeSingle && len(s.Namespaces) == 1
}

// SingleNamespace 返回单 Namespace 模式中的目标环境，否则返回空字符串。
func (s NamespaceScope) SingleNamespace() string {
	if !s.IsSingle() {
		return ""
	}
	return strings.TrimSpace(s.Namespaces[0])
}
