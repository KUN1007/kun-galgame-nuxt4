package service

import (
	"slices"
	"testing"
)

// A local row catalog has no work for stays NULL forever, so without the memo
// the ten-minute pass re-asks about the same orphans on every tick, and the
// nightly sweep is the only thing that ever clears the answer.
func TestContentLimitSyncSkipsKnownOrphans(t *testing.T) {
	s := NewGalgameContentLimitSync(nil, nil)
	asked := []int{1, 2, 3}

	s.rememberUnresolved(asked, map[int]string{1: "sfw", 3: "nsfw"})
	if got := s.resolvable(asked); !slices.Equal(got, []int{1, 3}) {
		t.Fatalf("resolvable = %v, want [1 3]", got)
	}

	s.rememberUnresolved([]int{2}, map[int]string{2: "sfw"})
	if got := s.resolvable(asked); !slices.Equal(got, asked) {
		t.Fatalf("an adopted orphan stays skipped: resolvable = %v, want %v", got, asked)
	}
}

func TestGroupByContentLimitDropsUnknownVerdicts(t *testing.T) {
	got := groupByContentLimit(map[int]string{1: "sfw", 2: "nsfw", 3: "", 4: "all"})
	if !slices.Equal(got["sfw"], []int{1}) || !slices.Equal(got["nsfw"], []int{2}) {
		t.Fatalf("groupByContentLimit = %v", got)
	}
	if len(got) != 2 {
		t.Fatalf("groupByContentLimit kept %d buckets, want only sfw and nsfw", len(got))
	}
}
