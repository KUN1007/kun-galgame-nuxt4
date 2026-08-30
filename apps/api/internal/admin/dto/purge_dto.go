package dto

type UserContentStats struct {
	Topics           int64 `json:"topics"`
	Replies          int64 `json:"replies"`
	TopicComments    int64 `json:"topic_comments"`
	Ratings          int64 `json:"ratings"`
	Resources        int64 `json:"resources"`
	Websites         int64 `json:"websites"`
	Toolsets         int64 `json:"toolsets"`
	ToolsetResources int64 `json:"toolset_resources"`
	Polls            int64 `json:"polls"`
	Lotteries        int64 `json:"lotteries"`
	Drafts           int64 `json:"drafts"`
	Quizzes          int64 `json:"quizzes"`
	Collections      int64 `json:"collections"`
	Todos            int64 `json:"todos"`
	ChatMessages     int64 `json:"chat_messages"`
	Messages         int64 `json:"messages"`
	Interactions     int64 `json:"interactions"`
	Total            int64 `json:"total"`
	CommunityPosts   int64 `json:"community_posts"`
}

type PurgeResult struct {
	UserContentStats
	CommunityPostsPurged      int64 `json:"community_posts_purged"`
	CommunityReactionsDeleted int64 `json:"community_reactions_deleted"`
}
