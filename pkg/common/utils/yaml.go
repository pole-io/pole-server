package utils

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// LoadYAML 加载配置
func LoadYAML[C any](filePath string, v C) (C, error) {
	if filePath == "" {
		err := errors.New("invalid config file path")
		fmt.Printf("[ERROR] %v\n", err)
		return v, err
	}

	fmt.Printf("[INFO] load config from %v\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return v, err
	}
	defer func() {
		_ = file.Close()
	}()
	buf, err := os.ReadFile(filePath)
	if nil != err {
		return v, fmt.Errorf("read file %s error", filePath)
	}

	if err = ParseYamlContent(string(buf), v); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return v, err
	}

	return v, nil
}

// FlattenYaml 将嵌套的yaml结构扁平化为map[string]string
func FlattenYaml(prefix string, m map[string]any, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			FlattenYaml(key, val, out)
		default:
			out[key] = fmt.Sprintf("%v", val)
		}
	}
}

// ParseYamlContent parses the YAML content and replaces environment variables.
func ParseYamlContent[C any](content string, c C) error {
	if err := yaml.Unmarshal([]byte(replaceEnv(content)), c); nil != err {
		return fmt.Errorf("parse yaml %s error:%w", content, err)
	}
	return nil
}

// replaceEnv replace holder by env list
func replaceEnv(configContent string) string {
	return os.ExpandEnv(configContent)
}
