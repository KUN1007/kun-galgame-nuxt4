package service

import (
	"testing"

	"kun-galgame-api/internal/galgame/dto"
)

func TestSeriesDetailLeavesTheUnpublishedBucketEmpty(t *testing.T) {
	detail := &dto.SeriesDetail{UnpublishedGalgame: []dto.GalgameCard{}}
	if len(detail.UnpublishedGalgame) != 0 {
		t.Errorf("unpublished_galgame = %d, want empty — catalog membership is the page", len(detail.UnpublishedGalgame))
	}
}
