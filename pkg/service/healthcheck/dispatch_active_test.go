package healthcheck

import (
	"testing"

	commonhash "github.com/pole-io/pole-server/pkg/common/utils/hash"
)

func TestSelectCheckerBucketsFallsBackWhenAllUnhealthy(t *testing.T) {
	healthy := map[commonhash.Bucket]bool{}
	fallback := map[commonhash.Bucket]bool{
		{Host: "127.0.0.1", Weight: 100}: true,
	}
	selected := selectCheckerBuckets(healthy, fallback)
	if len(selected) != 1 {
		t.Fatalf("selected bucket count = %d", len(selected))
	}
	for bucket := range selected {
		if bucket.Host != "127.0.0.1" {
			t.Fatalf("selected host = %s", bucket.Host)
		}
	}
}

func TestSelectCheckerBucketsPrefersHealthy(t *testing.T) {
	healthyBucket := commonhash.Bucket{Host: "10.0.0.1", Weight: 100}
	healthy := map[commonhash.Bucket]bool{healthyBucket: true}
	fallback := map[commonhash.Bucket]bool{
		healthyBucket:                   true,
		{Host: "10.0.0.2", Weight: 100}: true,
	}
	selected := selectCheckerBuckets(healthy, fallback)
	if len(selected) != 1 || !selected[healthyBucket] {
		t.Fatalf("unexpected selected buckets: %+v", selected)
	}
}
