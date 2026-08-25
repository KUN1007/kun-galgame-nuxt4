package service

import (
	"context"
	"errors"
	"log/slog"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/catalogclient"
)

func (s *GalgameService) hydrateCoverVotes(ctx context.Context, gid int, accessToken string, covers []dto.GalgameCover) {
	if s.catalog == nil || len(covers) == 0 {
		return
	}
	ids, appErr := s.galgameClient.CatalogWorkIDs(ctx, []int{gid})
	if appErr != nil {
		return
	}
	workID, ok := ids[gid]
	if !ok {
		return
	}
	var (
		tallies []catalogclient.CoverTally
		err     error
	)
	if accessToken != "" {
		tallies, err = s.catalog.WorkCoversUser(ctx, accessToken, workID)
		if errors.Is(err, catalogclient.ErrInsufficientScope) {
			tallies, err = s.catalog.WorkCoverVotes(ctx, workID)
		}
	} else {
		tallies, err = s.catalog.WorkCoverVotes(ctx, workID)
	}
	if err != nil {
		slog.Warn("galgame detail: cover vote tallies unavailable", "gid", gid, "error", err)
		return
	}
	byHash := make(map[string]int, len(tallies))
	for i, t := range tallies {
		if t.ImageHash != "" {
			byHash[t.ImageHash] = i
		}
	}
	for i := range covers {
		t, ok := byHash[covers[i].ImageHash]
		if !ok {
			continue
		}
		covers[i].ID = tallies[t].ID
		covers[i].VoteCount = tallies[t].VoteCount
		covers[i].Voted = tallies[t].Voted
	}
}
