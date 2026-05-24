package models

import "time"

type Novel struct {
	ID          uint64    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Author      string    `json:"author" db:"author"`
	Description string    `json:"description" db:"description"`
	CoverURL    string    `json:"cover_url" db:"cover_url"`
	Status      int       `json:"status" db:"status"`
	ChapterCount int      `json:"chapter_count" db:"chapter_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type NovelListReq struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
	Status   int `form:"status,default=-1"`
	Keyword  string `form:"keyword"`
}

type Chapter struct {
	ID           uint64    `json:"id" db:"id"`
	NovelID      uint64    `json:"novel_id" db:"novel_id"`
	Title        string    `json:"title" db:"title" binding:"required"`
	Content      string    `json:"content" db:"content"`
	WordCount    int       `json:"word_count" db:"word_count"`
	ChapterOrder int       `json:"chapter_order" db:"chapter_order"`
	Status       int       `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Character struct {
	ID            uint64    `json:"id" db:"id"`
	NovelID       uint64    `json:"novel_id" db:"novel_id"`
	Name          string    `json:"name" db:"name" binding:"required"`
	Alias         string    `json:"alias" db:"alias"`
	AvatarURL     string    `json:"avatar_url" db:"avatar_url"`
	Gender        int       `json:"gender" db:"gender"`
	Age           string    `json:"age" db:"age"`
	Description   string    `json:"description" db:"description"`
	Personality   string    `json:"personality" db:"personality"`
	Background    string    `json:"background" db:"background"`
	CharacterOrder int      `json:"character_order" db:"character_order"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Worldview struct {
	ID        uint64    `json:"id" db:"id"`
	NovelID   uint64    `json:"novel_id" db:"novel_id"`
	Category  string    `json:"category" db:"category" binding:"required"`
	Title     string    `json:"title" db:"title" binding:"required"`
	Content   string    `json:"content" db:"content"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Foreshadowing struct {
	ID               uint64    `json:"id" db:"id"`
	NovelID          uint64    `json:"novel_id" db:"novel_id"`
	Title            string    `json:"title" db:"title" binding:"required"`
	Description      string    `json:"description" db:"description"`
	PlantedChapterID *uint64   `json:"planted_chapter_id" db:"planted_chapter_id"`
	ResolvedChapterID *uint64  `json:"resolved_chapter_id" db:"resolved_chapter_id"`
	Status           int       `json:"status" db:"status"`
	Importance       int       `json:"importance" db:"importance" binding:"min=1,max=5"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// 关联信息
type ForeshadowingDetail struct {
	Foreshadowing
	PlantedChapterTitle string  `json:"planted_chapter_title"`
	ResolvedChapterTitle string `json:"resolved_chapter_title"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

type ContentVersion struct {
	ID            uint64    `json:"id"`
	EntityType    string    `json:"entity_type"`
	EntityID      uint64    `json:"entity_id"`
	Version       int       `json:"version"`
	Snapshot      string    `json:"snapshot"`
	ChangeSummary string    `json:"change_summary"`
	CreatedAt     time.Time `json:"created_at"`
}

type RollbackReq struct {
	ChangeSummary string `json:"change_summary"`
}
