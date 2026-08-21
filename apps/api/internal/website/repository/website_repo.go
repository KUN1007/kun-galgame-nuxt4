package repository

import (
	"kun-galgame-api/internal/website/model"

	"gorm.io/gorm"
)

type WebsiteRepository struct {
	db *gorm.DB
}

func NewWebsiteRepository(db *gorm.DB) *WebsiteRepository {
	return &WebsiteRepository{db: db}
}

func (r *WebsiteRepository) DB() *gorm.DB { return r.db }

type WebsiteListRow struct {
	ID            int    `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	URL           string `gorm:"column:url"`
	Description   string `gorm:"column:description"`
	Icon          string `gorm:"column:icon"`
	IconImageHash string `gorm:"column:icon_image_hash"`
	AgeLimit      string `gorm:"column:age_limit"`
	Status        string `gorm:"column:status"`
	CategoryID    int    `gorm:"column:category_id"`
}

const websiteListColumns = "id, name, url, description, icon, icon_image_hash, age_limit, status, category_id"

func sfwScope(isSFW bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isSFW {
			return db.Where("age_limit = ?", "all")
		}
		return db
	}
}

func (r *WebsiteRepository) FindAll(isSFW bool) []WebsiteListRow {
	var rows []WebsiteListRow
	r.db.Table("galgame_website").
		Select(websiteListColumns).
		Scopes(sfwScope(isSFW)).
		Order("created DESC").
		Scan(&rows)
	return rows
}

func (r *WebsiteRepository) FindByCategoryID(categoryID int, isSFW bool) []WebsiteListRow {
	var rows []WebsiteListRow
	r.db.Table("galgame_website").
		Select(websiteListColumns).
		Where("category_id = ?", categoryID).
		Scopes(sfwScope(isSFW)).
		Scan(&rows)
	return rows
}

func (r *WebsiteRepository) FindByIDs(ids []int, isSFW bool) []WebsiteListRow {
	if len(ids) == 0 {
		return nil
	}
	var rows []WebsiteListRow
	r.db.Table("galgame_website").
		Select(websiteListColumns).
		Where("id IN ?", ids).
		Scopes(sfwScope(isSFW)).
		Scan(&rows)
	return rows
}

func (r *WebsiteRepository) FindByDomain(domain string) (*model.GalgameWebsite, error) {
	var website model.GalgameWebsite
	if err := r.db.Where("url = ?", domain).First(&website).Error; err != nil {
		return nil, err
	}
	return &website, nil
}

func (r *WebsiteRepository) IncrementView(id int) {
	r.db.Model(&model.GalgameWebsite{}).Where("id = ?", id).
		Update("view", gorm.Expr("view + 1"))
}

func (r *WebsiteRepository) Create(tx *gorm.DB, website *model.GalgameWebsite) error {
	return tx.Create(website).Error
}

func (r *WebsiteRepository) UpdateFields(tx *gorm.DB, id int, updates map[string]any) error {
	return tx.Model(&model.GalgameWebsite{}).Where("id = ?", id).Updates(updates).Error
}

func (r *WebsiteRepository) DeleteByID(id int) error {
	return r.db.Delete(&model.GalgameWebsite{}, id).Error
}

func (r *WebsiteRepository) AdjustLikeCount(tx *gorm.DB, id, delta int) error {
	return tx.Model(&model.GalgameWebsite{}).Where("id = ?", id).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *WebsiteRepository) AdjustFavoriteCount(tx *gorm.DB, id, delta int) error {
	return tx.Model(&model.GalgameWebsite{}).Where("id = ?", id).
		Update("favorite_count", gorm.Expr("favorite_count + ?", delta)).Error
}

func (r *WebsiteRepository) FindLike(tx *gorm.DB, userID, websiteID int) (*model.GalgameWebsiteLike, error) {
	var like model.GalgameWebsiteLike
	if err := tx.Where("user_id = ? AND website_id = ?", userID, websiteID).First(&like).Error; err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *WebsiteRepository) FindFavorite(tx *gorm.DB, userID, websiteID int) (*model.GalgameWebsiteFavorite, error) {
	var fav model.GalgameWebsiteFavorite
	if err := tx.Where("user_id = ? AND website_id = ?", userID, websiteID).First(&fav).Error; err != nil {
		return nil, err
	}
	return &fav, nil
}

func (r *WebsiteRepository) CountByCategoryID(categoryID int) int64 {
	var c int64
	r.db.Model(&model.GalgameWebsite{}).Where("category_id = ?", categoryID).Count(&c)
	return c
}

func (r *WebsiteRepository) FindConflict(name, url string, excludeID int) (*model.GalgameWebsite, error) {
	var website model.GalgameWebsite
	query := r.db.Where("name = ? OR url = ?", name, url)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&website).Error; err != nil {
		return nil, err
	}
	return &website, nil
}

func (r *WebsiteRepository) CreateLike(tx *gorm.DB, userID, websiteID int) error {
	return tx.Create(&model.GalgameWebsiteLike{UserID: userID, WebsiteID: websiteID}).Error
}

func (r *WebsiteRepository) DeleteLike(tx *gorm.DB, like *model.GalgameWebsiteLike) error {
	return tx.Delete(like).Error
}

func (r *WebsiteRepository) CreateFavorite(tx *gorm.DB, userID, websiteID int) error {
	return tx.Create(&model.GalgameWebsiteFavorite{UserID: userID, WebsiteID: websiteID}).Error
}

func (r *WebsiteRepository) DeleteFavorite(tx *gorm.DB, fav *model.GalgameWebsiteFavorite) error {
	return tx.Delete(fav).Error
}

func (r *WebsiteRepository) HasLike(userID, websiteID int) bool {
	var c int64
	r.db.Model(&model.GalgameWebsiteLike{}).
		Where("user_id = ? AND website_id = ?", userID, websiteID).Count(&c)
	return c > 0
}

func (r *WebsiteRepository) HasFavorite(userID, websiteID int) bool {
	var c int64
	r.db.Model(&model.GalgameWebsiteFavorite{}).
		Where("user_id = ? AND website_id = ?", userID, websiteID).Count(&c)
	return c > 0
}
