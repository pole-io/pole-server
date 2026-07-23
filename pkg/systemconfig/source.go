package systemconfig

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type SourceIndex map[string]SourceDescriptor

var environmentReferencePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

func SourceIndexFromFile(path string) (SourceIndex, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return SourceIndexFromYAML(content)
}

func SourceIndexFromYAML(content []byte) (SourceIndex, error) {
	var root any
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse system configuration sources: %w", err)
	}
	index := SourceIndex{}
	collectSources("", root, index)
	return index, nil
}

func (s SourceIndex) Source(path string) SourceDescriptor {
	if source, ok := s[path]; ok {
		return source
	}

	// YAML sequences are indexed by their leaf paths (for example,
	// toolAllowlist.0). A setting definition describes the sequence itself, so
	// fold the descendant sources back into one deterministic descriptor.
	prefix := strings.TrimSuffix(path, ".") + "."
	var descendants []SourceDescriptor
	for childPath, source := range s {
		if strings.HasPrefix(childPath, prefix) {
			descendants = append(descendants, source)
		}
	}
	if len(descendants) > 0 {
		return mergeSourceDescriptors(descendants)
	}
	return SourceDescriptor{Kind: SourceCompiledDefault}
}

func mergeSourceDescriptors(sources []SourceDescriptor) SourceDescriptor {
	selected := SourceCompiledDefault
	for _, source := range sources {
		if sourceKindRank(source.Kind) > sourceKindRank(selected) {
			selected = source.Kind
		}
	}

	referenceSet := map[string]struct{}{}
	for _, source := range sources {
		if source.Kind != selected || source.Reference == "" {
			continue
		}
		for _, reference := range strings.Split(source.Reference, ",") {
			reference = strings.TrimSpace(reference)
			if reference != "" {
				referenceSet[reference] = struct{}{}
			}
		}
	}
	references := make([]string, 0, len(referenceSet))
	for reference := range referenceSet {
		references = append(references, reference)
	}
	sort.Strings(references)
	return SourceDescriptor{Kind: selected, Reference: strings.Join(references, ", ")}
}

func sourceKindRank(kind SourceKind) int {
	switch kind {
	case SourceCommandLine:
		return 4
	case SourceEnvironment:
		return 3
	case SourceStaticFile:
		return 2
	case SourceCompiledDefault:
		return 1
	default:
		return 0
	}
}

func (s SourceIndex) Clone() SourceIndex {
	result := make(SourceIndex, len(s))
	for path, source := range s {
		result[path] = source
	}
	return result
}

func (s SourceIndex) MarkCompiledDefault(paths ...string) {
	for _, path := range paths {
		delete(s, path)
	}
}

func (s SourceIndex) Subtree(prefix string) SourceIndex {
	prefix = strings.TrimSuffix(prefix, ".") + "."
	result := SourceIndex{}
	for path, source := range s {
		if strings.HasPrefix(path, prefix) {
			result[strings.TrimPrefix(path, prefix)] = source
		}
	}
	return result
}

func collectSources(prefix string, value any, result SourceIndex) {
	switch typed := value.(type) {
	case map[interface{}]interface{}:
		for key, child := range typed {
			name := fmt.Sprint(key)
			path := name
			if prefix != "" {
				path = prefix + "." + name
			}
			collectSources(path, child, result)
		}
	case []interface{}:
		if len(typed) == 0 && prefix != "" {
			result[prefix] = SourceDescriptor{Kind: SourceStaticFile}
		}
		for index, child := range typed {
			collectSources(fmt.Sprintf("%s.%d", prefix, index), child, result)
		}
	default:
		source := SourceDescriptor{Kind: SourceStaticFile}
		if text, ok := typed.(string); ok {
			matches := environmentReferencePattern.FindAllStringSubmatch(text, -1)
			if len(matches) > 0 {
				references := make([]string, 0, len(matches))
				seen := map[string]struct{}{}
				for _, match := range matches {
					name := match[1]
					if name == "" {
						name = match[2]
					}
					if _, exists := seen[name]; exists {
						continue
					}
					seen[name] = struct{}{}
					references = append(references, "env:"+name)
				}
				source = SourceDescriptor{Kind: SourceEnvironment, Reference: strings.Join(references, ", ")}
			}
		}
		if prefix != "" {
			result[prefix] = source
		}
	}
}
