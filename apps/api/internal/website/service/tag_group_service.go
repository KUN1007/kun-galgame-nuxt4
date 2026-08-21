package service

import (
	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/model"
	"kun-galgame-api/internal/website/repository"
	"kun-galgame-api/pkg/errors"
)

type TagGroupService struct {
	tagGroupRepo *repository.TagGroupRepository
}

func NewTagGroupService(tagGroupRepo *repository.TagGroupRepository) *TagGroupService {
	return &TagGroupService{tagGroupRepo: tagGroupRepo}
}

func (s *TagGroupService) GetAll() []dto.WebsiteTagGroupBrief {
	groups := s.tagGroupRepo.FindAll()
	counts := s.tagGroupRepo.TagCounts()
	out := make([]dto.WebsiteTagGroupBrief, len(groups))
	for i, g := range groups {
		out[i] = dto.WebsiteTagGroupBrief{
			ID:          g.ID,
			Name:        g.Name,
			Label:       g.Label,
			Description: g.Description,
			SortOrder:   g.SortOrder,
			MultiSelect: g.MultiSelect,
			TagCount:    counts[g.ID],
		}
	}
	return out
}

func (s *TagGroupService) Create(req *dto.CreateWebsiteTagGroupRequest) *errors.AppError {
	group := &model.GalgameWebsiteTagGroup{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		MultiSelect: req.MultiSelect,
	}
	if err := s.tagGroupRepo.Create(group); err != nil {
		return errors.ErrBadRequest("创建标签分组失败, 分组名称可能已存在")
	}
	return nil
}

func (s *TagGroupService) Update(req *dto.UpdateWebsiteTagGroupRequest) *errors.AppError {
	if err := s.tagGroupRepo.UpdateFields(req.GroupID, map[string]any{
		"name":         req.Name,
		"label":        req.Label,
		"description":  req.Description,
		"sort_order":   req.SortOrder,
		"multi_select": req.MultiSelect,
	}); err != nil {
		return errors.ErrInternal("更新标签分组失败")
	}
	return nil
}

func (s *TagGroupService) Delete(id int) *errors.AppError {
	if err := s.tagGroupRepo.DeleteByID(id); err != nil {
		return errors.ErrInternal("删除标签分组失败")
	}
	return nil
}
