package repository

import (
	topicRepo "kun-galgame-api/internal/topic/repository"
	"time"

	"gorm.io/gorm"

	"kun-galgame-api/pkg/miniapp"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

type TopicRow struct {
	ID               int
	Title            string
	View             int
	Status           int
	LikeCount        int
	ReplyCount       int
	CommentCount     int
	StatusUpdateTime time.Time
	UserID           int
	IsNSFW           bool
	BestAnswerID     *int
	UpvoteTime       *time.Time
}

type TopicSectionRow struct {
	TopicID     int    `gorm:"column:topic_id"`
	SectionName string `gorm:"column:name"`
}

type ReplyRow struct {
	ID          int
	TopicID     int
	TopicTitle  string
	TopicUserID int
	Content     string
	Floor       int
	UserID      int
	Created     time.Time
}

type CommentRow struct {
	ID          int
	TopicID     int
	TopicTitle  string
	TopicUserID int
	Content     string
	UserID      int
	Created     time.Time
}

func (r *SearchRepository) SearchTopics(keywords []string, page, limit int, authenticated bool) (rows []TopicRow, total int64) {
	score, scoreArgs := topicRelevance(keywords)
	query := r.db.Table("topic t").
		Select(`t.id, t.title, t.view, t.status, t.like_count, t.reply_count,
			t.comment_count, t.status_update_time, t.user_id,
			t.is_nsfw, t.best_answer_id, t.upvote_time, `+score+` AS relevance`, scoreArgs...).
		Where("t.status != 1").
		Where(topicRepo.SharedListPredicate("t", authenticated))
	for _, kw := range keywords {
		like := "%" + kw + "%"
		query = query.Where("(t.title ILIKE ? OR t.content ILIKE ? OR t.category ILIKE ?)",
			like, like, like)
	}

	query.Count(&total)
	query.Order("relevance DESC, t.status_update_time DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

func (r *SearchRepository) FindTopicSections(topicIDs []int) []TopicSectionRow {
	if len(topicIDs) == 0 {
		return nil
	}
	var rows []TopicSectionRow
	r.db.Table("topic_section_relation tsr").
		Select("tsr.topic_id, ts.name").
		Joins("JOIN topic_section ts ON ts.id = tsr.topic_section_id").
		Where("tsr.topic_id IN ?", topicIDs).
		Find(&rows)
	return rows
}

func (r *SearchRepository) FindTopicMiniApps(topicIDs []int) map[int][]string {
	return miniapp.ByTopic(r.db, topicIDs)
}

// CountUserPosts is what a user result shows besides a name and a bio: how much
// of the forum this account actually wrote. OAuth's /users/search cannot answer
// it — the counts are local.
func (r *SearchRepository) CountUserPosts(userIDs []int, authenticated bool) (topics, replies map[int]int) {
	topics, replies = map[int]int{}, map[int]int{}
	if len(userIDs) == 0 {
		return
	}
	type countRow struct {
		UserID int   `gorm:"column:user_id"`
		Total  int64 `gorm:"column:total"`
	}
	for _, lane := range []struct {
		table string
		where string
		out   map[int]int
	}{
		{"topic", "status != 1 AND " + topicRepo.SharedListPredicate("", authenticated), topics},
		{"topic_reply", "status = 0", replies},
	} {
		var rows []countRow
		r.db.Table(lane.table).
			Select("user_id, count(*) AS total").
			Where("user_id IN ?", userIDs).
			Where(lane.where).
			Group("user_id").
			Scan(&rows)
		for _, row := range rows {
			lane.out[row.UserID] = int(row.Total)
		}
	}
	return
}

func (r *SearchRepository) SearchReplies(keywords []string, page, limit int, authenticated bool) (rows []ReplyRow, total int64) {
	snippet, snippetArgs := contentSnippet("r.content", keywords)
	score, scoreArgs := contentRelevance("r.content", keywords)

	query := r.db.Table("topic_reply r").
		Select(`r.id, r.topic_id, t.title AS topic_title, t.user_id AS topic_user_id, `+
			snippet+` AS content, r.floor, r.user_id, r.created, `+score+` AS relevance`,
			append(snippetArgs, scoreArgs...)...).
		Joins("LEFT JOIN topic t ON t.id = r.topic_id").
		Where("r.status = 0").
		Where("EXISTS (SELECT 1 FROM topic parent WHERE parent.id = r.topic_id AND parent.status != 1 AND " + topicRepo.SharedListPredicate("parent", authenticated) + ")")
	for _, kw := range keywords {
		query = query.Where("r.content ILIKE ?", "%"+kw+"%")
	}

	query.Count(&total)
	query.Order("relevance DESC, r.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

func (r *SearchRepository) SearchComments(keywords []string, page, limit int, authenticated bool) (rows []CommentRow, total int64) {
	snippet, snippetArgs := contentSnippet("c.content", keywords)
	score, scoreArgs := contentRelevance("c.content", keywords)

	query := r.db.Table("topic_comment c").
		Select(`c.id, c.topic_id, t.title AS topic_title, t.user_id AS topic_user_id, `+
			snippet+` AS content, c.user_id, c.created, `+score+` AS relevance`,
			append(snippetArgs, scoreArgs...)...).
		Joins("LEFT JOIN topic t ON t.id = c.topic_id").
		Where("c.status = 0").
		Where("EXISTS (SELECT 1 FROM topic parent WHERE parent.id = c.topic_id AND parent.status != 1 AND " + topicRepo.SharedListPredicate("parent", authenticated) + ")")
	for _, kw := range keywords {
		query = query.Where("c.content ILIKE ?", "%"+kw+"%")
	}

	query.Count(&total)
	query.Order("relevance DESC, c.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}
