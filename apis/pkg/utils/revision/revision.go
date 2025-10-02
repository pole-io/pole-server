package revision

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"strconv"
)

func ComputeRevisionBySlice(h hash.Hash, slice []string) (string, error) {
	for _, revision := range slice {
		if _, err := h.Write([]byte(revision)); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ComputeRevisionUint64(slice []uint64) (string, error) {
	if len(slice) == 1 {
		return strconv.Itoa(int(slice[0])), nil
	}

	h := sha1.New()

	b := make([]byte, 8)
	for _, revision := range slice {
		binary.LittleEndian.PutUint64(b, revision)
		if _, err := h.Write(b); err != nil {
			return "", err
		}
		b = b[:0] // Reset the slice to avoid reallocation
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompositeComputeRevision 将多个 revision 合并计算为一个
func CompositeComputeRevision(revisions []string) (string, error) {
	if len(revisions) == 1 {
		return revisions[0], nil
	}

	h := sha1.New()
	for i := range revisions {
		if _, err := h.Write([]byte(revisions[i])); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
