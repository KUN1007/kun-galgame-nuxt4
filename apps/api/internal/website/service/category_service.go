package service

import (
	"strconv"

	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/model"
	"kun-galgame-api/internal/website/repository"
	"kun-galgame-api/pkg/errors"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
	websiteRepo  *repository.WebsiteRepository
	tagRepo      *repository.TagRepository
	cdnBase      string
}

func NewCategoryService(
	categoryRepo *repository.CategoryRepository,
	websiteRepo *repository.WebsiteRepository,
	tagRepo *repository.TagRepository,
	cdnBase string,
) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		websiteRepo:  websiteRepo,
		tagRepo:      tagRepo,
		cdnBase:      cdnBase,
	}
}

func (s *CategoryService) GetDetail(name string, isSFW bool) (*dto.WebsiteCategoryDetailResponse, *errors.AppError) {
	category, err := s.categoryRepo.FindByName(name)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该分类")
	}

	rows := s.websiteRepo.FindByCategoryID(category.ID, isSFW)
	websiteIDs := collectWebsiteIDs(rows)
	levelMap := s.tagRepo.LevelSumsByWebsiteIDs(websiteIDs)
	cards := websiteCardsFromRowsSingleCategory(rows, category.Name, levelMap, s.cdnBase)

	return &dto.WebsiteCategoryDetailResponse{
		ID:           category.ID,
		Name:         category.Name,
		Label:        category.Label,
		Description:  category.Description,
		SortOrder:    category.SortOrder,
		WebsiteCount: len(rows),
		Websites:     cards,
		Created:      category.CreatedAt,
		Updated:      category.UpdatedAt,
	}, nil
}

func (s *CategoryService) GetAll() []dto.WebsiteCategoryListItem {
	categories := s.categoryRepo.FindAll()
	out := make([]dto.WebsiteCategoryListItem, len(categories))
	for i, c := range categories {
		out[i] = dto.WebsiteCategoryListItem{
			ID:           c.ID,
			Name:         c.Name,
			Label:        c.Label,
			Description:  c.Description,
			SortOrder:    c.SortOrder,
			WebsiteCount: int(s.websiteRepo.CountByCategoryID(c.ID)),
		}
	}
	return out
}

func (s *CategoryService) Create(req *dto.CreateWebsiteCategoryRequest) *errors.AppError {
	if existing, err := s.categoryRepo.FindByName(req.Name); err == nil && existing != nil {
		return errors.ErrBadRequest("分类名称「" + req.Name + "」已存在")
	}
	category := &model.GalgameWebsiteCategory{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}
	if err := s.categoryRepo.Create(category); err != nil {
		return errors.ErrInternal("创建分类失败")
	}
	return nil
}

// galgame_website.category_id is ON DELETE RESTRICT, so a category with
// websites cannot be removed. Say which, instead of letting the FK surface as
// a 500.
func (s *CategoryService) Delete(id int) *errors.AppError {
	if count := s.websiteRepo.CountByCategoryID(id); count > 0 {
		return errors.ErrBadRequest("该分类下还有 " + strconv.FormatInt(count, 10) + " 个网站, 请先移动它们")
	}
	if err := s.categoryRepo.DeleteByID(id); err != nil {
		return errors.ErrInternal("删除分类失败")
	}
	return nil
}

func (s *CategoryService) Update(req *dto.UpdateWebsiteCategoryRequest) *errors.AppError {
	if err := s.categoryRepo.UpdateFields(req.CategoryID, map[string]any{
		"name":        req.Name,
		"label":       req.Label,
		"description": req.Description,
		"sort_order":  req.SortOrder,
	}); err != nil {
		return errors.ErrInternal("更新分类失败")
	}
	return nil
}
