package skillmarketplace

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func ValidateVersion(version string) error {
	if !semverPattern.MatchString(version) {
		return fmt.Errorf("version %q is not strict SemVer", version)
	}
	return nil
}

func CompareVersions(left, right string) int {
	l := semverPattern.FindStringSubmatch(left)
	r := semverPattern.FindStringSubmatch(right)
	if l == nil || r == nil {
		return strings.Compare(left, right)
	}
	for i := 1; i <= 3; i++ {
		ln, _ := strconv.ParseUint(l[i], 10, 64)
		rn, _ := strconv.ParseUint(r[i], 10, 64)
		if ln < rn {
			return -1
		}
		if ln > rn {
			return 1
		}
	}
	return comparePrerelease(l[4], r[4])
}

func comparePrerelease(left, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return 1
	}
	if right == "" {
		return -1
	}
	ls, rs := strings.Split(left, "."), strings.Split(right, ".")
	for i := 0; i < len(ls) && i < len(rs); i++ {
		if ls[i] == rs[i] {
			continue
		}
		ln, le := strconv.ParseUint(ls[i], 10, 64)
		rn, re := strconv.ParseUint(rs[i], 10, 64)
		switch {
		case le == nil && re == nil && ln < rn:
			return -1
		case le == nil && re == nil && ln > rn:
			return 1
		case le == nil:
			return -1
		case re == nil:
			return 1
		default:
			return strings.Compare(ls[i], rs[i])
		}
	}
	if len(ls) < len(rs) {
		return -1
	}
	return 1
}
