/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package matchs

import (
	"strconv"
	"strings"

	regexp "github.com/dlclark/regexp2"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

const (
	// EmptyErrString empty error string
	EmptyErrString = "empty"
	// NilErrString null pointer error string
	NilErrString = "nil"
	// MatchAll rule match all service or namespace value
	MatchAll = "*"
)

// IsPrefixWildName 判断名字是否为通配名字，只支持前缀索引(名字最后为*)
func IsPrefixWildName(name string) bool {
	length := len(name)
	return length >= 1 && name[length-1:length] == "*"
}

// IsWildName 判断名字是否为通配名字，前缀或者后缀
func IsWildName(name string) bool {
	return IsPrefixWildName(name) || IsSuffixWildName(name)
}

// ParseWildNameForSql 如果 name 是通配字符串，将通配字符*替换为sql中的%
func ParseWildNameForSql(name string) string {
	if IsPrefixWildName(name) {
		name = name[:len(name)-1] + "%"
	}
	if IsSuffixWildName(name) {
		name = "%" + name[1:]
	}
	return name
}

// IsSuffixWildName 判断名字是否为通配名字，只支持后缀索引(名字第一个字符为*)
func IsSuffixWildName(name string) bool {
	length := len(name)
	return length >= 1 && name[0:1] == "*"
}

// ParseWildName 判断是否为格式化查询条件并且返回真正的查询信息
func ParseWildName(name string) (string, bool) {
	length := len(name)
	ok := length >= 1 && name[length-1:length] == "*"

	if ok {
		return name[:len(name)-1], ok
	}

	return name, false
}

// IsWildMatchIgnoreCase 判断 name 是否匹配 pattern，pattern 可以是前缀或者后缀，忽略大小写
func IsWildMatchIgnoreCase(name, pattern string) bool {
	return IsWildMatch(strings.ToLower(name), strings.ToLower(pattern))
}

// IsWildNotMatch .
func IsWildNotMatch(name, pattern string) bool {
	return !IsWildMatch(name, pattern)
}

// IsWildMatch 判断 name 是否匹配 pattern，pattern 可以是前缀或者后缀
func IsWildMatch(name, pattern string) bool {
	if IsPrefixWildName(pattern) {
		pattern = strings.TrimRight(pattern, "*")
		if strings.HasPrefix(name, pattern) {
			return true
		}
		if IsSuffixWildName(pattern) {
			pattern = strings.TrimLeft(pattern, "*")
			return strings.Contains(name, pattern)
		}
		return false
	} else if IsSuffixWildName(pattern) {
		pattern = strings.TrimLeft(pattern, "*")
		if strings.HasSuffix(name, pattern) {
			return true
		}
		return false
	}
	return pattern == name
}

func IsMatchAll(v string) bool {
	return v == "" || v == MatchAll
}

func MatchString(srcMetaValue string, matchValule *apimodel.MatchString, regexToPattern func(string) *regexp.Regexp) bool {
	rawMetaValue := matchValule.GetValue().GetValue()
	if IsMatchAll(rawMetaValue) {
		return true
	}

	switch matchValule.Type {
	case apimodel.MatchString_REGEX:
		matchExp := regexToPattern(rawMetaValue)
		if matchExp == nil {
			return false
		}
		match, err := matchExp.MatchString(srcMetaValue)
		if err != nil {
			return false
		}
		return match
	case apimodel.MatchString_NOT_EQUALS:
		return srcMetaValue != rawMetaValue
	case apimodel.MatchString_EXACT:
		return srcMetaValue == rawMetaValue
	case apimodel.MatchString_IN:
		find := false
		tokens := strings.Split(rawMetaValue, ",")
		for _, token := range tokens {
			if token == srcMetaValue {
				find = true
				break
			}
		}
		return find
	case apimodel.MatchString_NOT_IN:
		tokens := strings.Split(rawMetaValue, ",")
		for _, token := range tokens {
			if token == srcMetaValue {
				return false
			}
		}
		return true
	case apimodel.MatchString_RANGE:
		// range 模式只支持数字
		tokens := strings.Split(rawMetaValue, "~")
		if len(tokens) != 2 {
			return false
		}
		left, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			return false
		}
		right, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			return false
		}
		srcVal, err := strconv.ParseInt(srcMetaValue, 10, 64)
		if err != nil {
			return false
		}
		return srcVal >= left && srcVal <= right
	}
	return true
}
