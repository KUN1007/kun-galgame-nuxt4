package dto

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

type ListQuery struct {
	Page  int `query:"page" validate:"min=1"`
	Limit int `query:"limit" validate:"min=1,max=50"`
}

type TodoListQuery struct {
	Page   int  `query:"page" validate:"min=1"`
	Limit  int  `query:"limit" validate:"min=1,max=50"`
	Status *int `query:"status" validate:"omitempty,min=0,max=3"`
}

// Caps mirror apps/web/app/validations/update-log.ts + todo.ts. The FE
// schemas enforced these (version max=20, content_* max=1000, todo
// status 0..10) but the BE was silently accepting anything — so a
// malicious / mis-typed direct API call could land a 100k-char locale
// blob or status=99 into the DB. Tightened here.
type CreateHistoryRequest struct {
	Type        string `json:"type" validate:"required"`
	Version     string `json:"version" validate:"max=20"`
	ContentEnUS string `json:"content_en_us" validate:"max=1000"`
	ContentJaJP string `json:"content_ja_jp" validate:"max=1000"`
	ContentZhCN string `json:"content_zh_cn" validate:"max=1000"`
	ContentZhTW string `json:"content_zh_tw" validate:"max=1000"`
}

type DeleteHistoryRequest struct {
	ID int `query:"update_log_id" validate:"required,min=1"`
}

type CreateTodoRequest struct {
	Type        string `json:"type" validate:"required"`
	Status      int    `json:"status" validate:"min=0,max=10"`
	ContentEnUS string `json:"content_en_us" validate:"max=1000"`
	ContentJaJP string `json:"content_ja_jp" validate:"max=1000"`
	ContentZhCN string `json:"content_zh_cn" validate:"max=1000"`
	ContentZhTW string `json:"content_zh_tw" validate:"max=1000"`
}

type DeleteTodoRequest struct {
	ID int `query:"todo_id" validate:"required,min=1"`
}

type UpdateHistoryRequest struct {
	ID          int    `json:"update_log_id" validate:"required,min=1"`
	Type        string `json:"type" validate:"required"`
	Version     string `json:"version" validate:"max=20"`
	ContentEnUS string `json:"content_en_us" validate:"max=1000"`
	ContentJaJP string `json:"content_ja_jp" validate:"max=1000"`
	ContentZhCN string `json:"content_zh_cn" validate:"max=1000"`
	ContentZhTW string `json:"content_zh_tw" validate:"max=1000"`
}

type UpdateTodoRequest struct {
	ID          int    `json:"todo_id" validate:"required,min=1"`
	Type        string `json:"type" validate:"required"`
	Status      int    `json:"status" validate:"min=0,max=10"`
	ContentEnUS string `json:"content_en_us" validate:"max=1000"`
	ContentJaJP string `json:"content_ja_jp" validate:"max=1000"`
	ContentZhCN string `json:"content_zh_cn" validate:"max=1000"`
	ContentZhTW string `json:"content_zh_tw" validate:"max=1000"`
}
