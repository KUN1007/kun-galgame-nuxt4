package service

import (
	"context"
	"time"

	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/topic/dto"
	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/imageclient"
	"kun-galgame-api/pkg/perm"
	"kun-galgame-api/pkg/userclient"
)

func (s *LotteryService) GetEntrants(
	ctx context.Context,
	lotteryID int,
	userInfo *middleware.UserInfo,
) ([]dto.LotteryEntrantResponse, *errors.AppError) {
	lottery, err := s.lotteryRepo.FindByID(lotteryID)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该抽奖")
	}
	canModerate := false
	viewerID := 0
	if userInfo != nil {
		viewerID = userInfo.ID
		canModerate = perm.CanUser(userInfo.ID, userInfo.Roles, perm.LotteryViewRestricted)
	}
	if !lottery.ShowEntrants && lottery.UserID != viewerID && !canModerate {
		return []dto.LotteryEntrantResponse{}, nil
	}

	entries, err := s.lotteryRepo.FindEntries(lotteryID)
	if err != nil {
		return nil, errors.ErrInternal("获取参与名单失败")
	}
	uids := userclient.CollectIDs(entries, func(e topicModel.TopicLotteryEntry) int { return e.UserID })
	userMap := s.userClient.Hydrate(ctx, uids)

	out := make([]dto.LotteryEntrantResponse, 0, len(entries))
	for _, e := range entries {
		u := userMap[e.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		out = append(out, dto.LotteryEntrantResponse{
			User:    dto.KunUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Floor:   e.ReplyFloor,
			Created: e.CreatedAt,
		})
	}
	return out, nil
}

func (s *LotteryService) GetLotteriesByTopic(
	ctx context.Context,
	topicID int,
	userInfo *middleware.UserInfo,
) ([]dto.TopicLotteryResponse, *errors.AppError) {
	lotteries, err := s.lotteryRepo.FindByTopicID(topicID)
	if err != nil {
		return nil, errors.ErrInternal("获取抽奖失败")
	}
	if len(lotteries) == 0 {
		return []dto.TopicLotteryResponse{}, nil
	}

	ids := make([]int, len(lotteries))
	for i, l := range lotteries {
		ids[i] = l.ID
	}
	prizes, _ := s.lotteryRepo.FindPrizesForLotteries(ids)
	winners, _ := s.lotteryRepo.FindWinnersForLotteries(ids)
	codeCounts, _ := s.lotteryRepo.CountCodesForLotteries(ids)

	viewerID := 0
	if userInfo != nil {
		viewerID = userInfo.ID
	}
	myEntries, _ := s.lotteryRepo.FindEntriesForUser(ids, viewerID)

	uids := make([]int, 0, len(lotteries)+len(winners))
	for _, l := range lotteries {
		uids = append(uids, l.UserID)
	}
	for _, w := range winners {
		uids = append(uids, w.UserID)
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	prizeByLottery := map[int][]topicModel.TopicLotteryPrize{}
	prizeByID := map[int]topicModel.TopicLotteryPrize{}
	for _, p := range prizes {
		prizeByLottery[p.LotteryID] = append(prizeByLottery[p.LotteryID], p)
		prizeByID[p.ID] = p
	}
	winnerByLottery := map[int][]topicModel.TopicLotteryEntry{}
	for _, w := range winners {
		winnerByLottery[w.LotteryID] = append(winnerByLottery[w.LotteryID], w)
	}

	out := make([]dto.TopicLotteryResponse, 0, len(lotteries))
	for i := range lotteries {
		out = append(out, s.buildLotteryResponse(ctx, &lotteries[i],
			prizeByLottery[lotteries[i].ID], winnerByLottery[lotteries[i].ID],
			prizeByID, myEntries, userMap, codeCounts, viewerID))
	}
	return out, nil
}

func (s *LotteryService) buildLotteryResponse(
	ctx context.Context,
	lottery *topicModel.TopicLottery,
	prizes []topicModel.TopicLotteryPrize,
	winners []topicModel.TopicLotteryEntry,
	prizeByID map[int]topicModel.TopicLotteryPrize,
	myEntries map[int]topicModel.TopicLotteryEntry,
	userMap map[int]userclient.User,
	codeCounts map[int]int,
	viewerID int,
) dto.TopicLotteryResponse {
	author := userMap[lottery.UserID]
	if author.ID == 0 {
		author = userclient.Placeholder(lottery.UserID)
	}

	totalSlots := 0
	prizeResp := make([]dto.LotteryPrizeResponse, 0, len(prizes))
	for _, p := range prizes {
		totalSlots += p.Slots
		prizeResp = append(prizeResp, dto.LotteryPrizeResponse{
			ID: p.ID, Name: p.Name, Description: p.Description,
			ImageHash: p.ImageHash,
			ImageURL:  imageclient.ResolveURL(s.cdnBase, p.ImageHash, ""),
			Delivery:  p.Delivery, PointAmount: p.PointAmount, Slots: p.Slots,
			CodesLoaded: codeCounts[p.ID],
		})
	}

	winnerResp := make([]dto.LotteryWinnerResponse, 0, len(winners))
	for _, w := range winners {
		u := userMap[w.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		winnerResp = append(winnerResp, dto.LotteryWinnerResponse{
			EntryID: w.ID, PrizeID: w.PrizeID,
			PrizeName:  prizeByID[w.PrizeID].Name,
			User:       dto.KunUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			ReplyFloor: w.ReplyFloor, RankKey: w.RankKey,
			Fulfillment: w.Fulfillment,
			WonAt:       derefTime(w.WonAt),
		})
	}

	resp := dto.TopicLotteryResponse{
		ID: lottery.ID, TopicID: lottery.TopicID,
		Title: lottery.Title, Description: lottery.Description,
		EntryMode: lottery.EntryMode, FloorRule: lottery.FloorRule,
		DrawMode: lottery.DrawMode, DrawThreshold: lottery.DrawThreshold,
		Deadline:          lottery.Deadline,
		MinAccountAgeDays: lottery.MinAccountAgeDays,
		MinMoemoepoint:    lottery.MinMoemoepoint,
		ShowEntrants:      lottery.ShowEntrants,
		Status:            lottery.Status,
		SeedHash:          lottery.SeedHash,
		EntryCount:        lottery.EntryCount,
		TotalSlots:        totalSlots,
		DrawnAt:           lottery.DrawnAt,
		User:              dto.KunUser{ID: author.ID, Name: author.Name, Avatar: author.Avatar},
		Prizes:            prizeResp,
		Winners:           winnerResp,
		Created:           lottery.CreatedAt,
		Updated:           lottery.UpdatedAt,
	}
	// The seed is the other half of the commitment and is worthless to an
	// attacker once the winners are fixed, so it ships only after the draw.
	if lottery.Status == topicModel.LotteryStatusDrawn {
		resp.Seed = lottery.Seed
	}

	if viewerID != 0 {
		if appErr := s.entryBlocker(ctx, lottery, viewerID); appErr != nil {
			resp.EnterBlocked = appErr.Message
		} else {
			resp.CanEnter = true
		}
	}
	if mine, ok := myEntries[lottery.ID]; ok {
		resp.HasEntered = true
		resp.CanEnter = false
		resp.MyEntryID = mine.ID
		resp.MyPrizeID = mine.PrizeID
		resp.MyFulfillment = mine.Fulfillment
		if mine.PrizeID > 0 {
			resp.MyPrizeName = prizeByID[mine.PrizeID].Name
			resp.MyDelivery = prizeByID[mine.PrizeID].Delivery
			resp.MyCodeReady = mine.CodeID > 0
		}
	}
	return resp
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func parseLotteryDeadline(raw *string) *time.Time {
	if raw == nil || *raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil
	}
	return &t
}
