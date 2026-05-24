package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"novel-service/database"
	"novel-service/models"
)

// 实体类型常量
const (
	EntityNovel        = "novel"
	EntityChapter      = "chapter"
	EntityCharacter    = "character"
	EntityWorldview    = "worldview"
	EntityForeshadowing = "foreshadowing"
)

// saveVersionSnapshot 在更新前保存当前数据的快照
func saveVersionSnapshot(entityType, entityID string, changeSummary string) error {
	// 查询当前数据并转为 JSON
	snapshot, err := getEntitySnapshot(entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}
	if snapshot == nil {
		return nil // 实体不存在，跳过
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// 获取当前最大版本号
	var maxVersion sql.NullInt64
	database.DB.QueryRow(
		"SELECT MAX(version) FROM content_version WHERE entity_type=? AND entity_id=?",
		entityType, entityID).Scan(&maxVersion)

	nextVersion := 1
	if maxVersion.Valid {
		nextVersion = int(maxVersion.Int64) + 1
	}

	_, err = database.DB.Exec(
		"INSERT INTO content_version (entity_type, entity_id, version, snapshot, change_summary) VALUES (?, ?, ?, ?, ?)",
		entityType, entityID, nextVersion, string(snapshotJSON), changeSummary)
	return err
}

// getEntitySnapshot 根据实体类型查询当前数据转为 map
func getEntitySnapshot(entityType, entityID string) (map[string]interface{}, error) {
	switch entityType {
	case EntityNovel:
		var n models.Novel
		err := database.DB.QueryRow(
			"SELECT id, title, author, description, cover_url, status, created_at, updated_at FROM novel WHERE id=?", entityID).
			Scan(&n.ID, &n.Title, &n.Author, &n.Description, &n.CoverURL, &n.Status, &n.CreatedAt, &n.UpdatedAt)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return structToMap(n)

	case EntityChapter:
		var ch models.Chapter
		err := database.DB.QueryRow(
			"SELECT id, novel_id, title, content, word_count, chapter_order, status, created_at, updated_at FROM chapter WHERE id=?", entityID).
			Scan(&ch.ID, &ch.NovelID, &ch.Title, &ch.Content, &ch.WordCount, &ch.ChapterOrder, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return structToMap(ch)

	case EntityCharacter:
		var ch models.Character
		err := database.DB.QueryRow(
			"SELECT id, novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order, created_at, updated_at FROM `character` WHERE id=?", entityID).
			Scan(&ch.ID, &ch.NovelID, &ch.Name, &ch.Alias, &ch.AvatarURL, &ch.Gender, &ch.Age, &ch.Description, &ch.Personality, &ch.Background, &ch.CharacterOrder, &ch.CreatedAt, &ch.UpdatedAt)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return structToMap(ch)

	case EntityWorldview:
		var w models.Worldview
		err := database.DB.QueryRow(
			"SELECT id, novel_id, category, title, content, sort_order, created_at, updated_at FROM worldview WHERE id=?", entityID).
			Scan(&w.ID, &w.NovelID, &w.Category, &w.Title, &w.Content, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return structToMap(w)

	case EntityForeshadowing:
		var f models.Foreshadowing
		err := database.DB.QueryRow(
			"SELECT id, novel_id, title, description, planted_chapter_id, resolved_chapter_id, status, importance, created_at, updated_at FROM foreshadowing WHERE id=?", entityID).
			Scan(&f.ID, &f.NovelID, &f.Title, &f.Description, &f.PlantedChapterID, &f.ResolvedChapterID, &f.Status, &f.Importance, &f.CreatedAt, &f.UpdatedAt)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return structToMap(f)
	}

	return nil, fmt.Errorf("unknown entity type: %s", entityType)
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- 版本 API ---

func GetVersions(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")

	// 验证实体类型
	validTypes := map[string]bool{EntityNovel: true, EntityChapter: true, EntityCharacter: true, EntityWorldview: true, EntityForeshadowing: true}
	if !validTypes[entityType] {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: "无效的实体类型"})
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary, created_at "+
			"FROM content_version WHERE entity_type=? AND entity_id=? ORDER BY version DESC",
		entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	versions := make([]models.ContentVersion, 0)
	for rows.Next() {
		var v models.ContentVersion
		if err := rows.Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		versions = append(versions, v)
	}

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: versions})
}

func GetVersion(c *gin.Context) {
	id := c.Param("id")
	var v models.ContentVersion
	err := database.DB.QueryRow(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary, created_at FROM content_version WHERE id=?", id).
		Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "版本不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: v})
}

func RollbackVersion(c *gin.Context) {
	id := c.Param("id")

	// 获取版本快照
	var v models.ContentVersion
	err := database.DB.QueryRow(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary, created_at FROM content_version WHERE id=?", id).
		Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "版本不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}

	// 解析快照
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(v.Snapshot), &snapshot); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "快照数据解析失败"})
		return
	}

	// 解析请求中的变更摘要
	var req models.RollbackReq
	c.ShouldBindJSON(&req)
	summary := req.ChangeSummary
	if summary == "" {
		summary = fmt.Sprintf("回退到版本%d", v.Version)
	}

	// 先保存当前状态为版本
	if err := saveVersionSnapshot(v.EntityType, fmt.Sprintf("%d", v.EntityID), summary); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存当前版本失败: " + err.Error()})
		return
	}

	// 根据实体类型回退数据
	if err := applySnapshot(v.EntityType, v.EntityID, snapshot); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "回退失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "回退成功"})
}

// applySnapshot 将快照数据写回实体表
func applySnapshot(entityType string, entityID uint64, snapshot map[string]interface{}) error {
	idStr := fmt.Sprintf("%d", entityID)

	switch entityType {
	case EntityNovel:
		_, err := database.DB.Exec(
			"UPDATE novel SET title=?, author=?, description=?, cover_url=?, status=? WHERE id=?",
			snapshot["title"], snapshot["author"], snapshot["description"], snapshot["cover_url"], snapshot["status"], idStr)
		return err

	case EntityChapter:
		wordCount := 0
		if wc, ok := snapshot["word_count"]; ok && wc != nil {
			switch v := wc.(type) {
			case float64:
				wordCount = int(v)
			case json.Number:
				n, _ := v.Int64()
				wordCount = int(n)
			}
		}
		chapterOrder := 0
		if co, ok := snapshot["chapter_order"]; ok && co != nil {
			switch v := co.(type) {
			case float64:
				chapterOrder = int(v)
			case json.Number:
				n, _ := v.Int64()
				chapterOrder = int(n)
			}
		}
		status := 0
		if s, ok := snapshot["status"]; ok && s != nil {
			switch v := s.(type) {
			case float64:
				status = int(v)
			case json.Number:
				n, _ := v.Int64()
				status = int(n)
			}
		}
		_, err := database.DB.Exec(
			"UPDATE chapter SET title=?, content=?, word_count=?, chapter_order=?, status=? WHERE id=?",
			snapshot["title"], snapshot["content"], wordCount, chapterOrder, status, idStr)
		return err

	case EntityCharacter:
		gender := 0
		if g, ok := snapshot["gender"]; ok && g != nil {
			switch v := g.(type) {
			case float64:
				gender = int(v)
			case json.Number:
				n, _ := v.Int64()
				gender = int(n)
			}
		}
		charOrder := 0
		if co, ok := snapshot["character_order"]; ok && co != nil {
			switch v := co.(type) {
			case float64:
				charOrder = int(v)
			case json.Number:
				n, _ := v.Int64()
				charOrder = int(n)
			}
		}
		_, err := database.DB.Exec(
			"UPDATE `character` SET name=?, alias=?, avatar_url=?, gender=?, age=?, description=?, personality=?, background=?, character_order=? WHERE id=?",
			snapshot["name"], snapshot["alias"], snapshot["avatar_url"], gender, snapshot["age"],
			snapshot["description"], snapshot["personality"], snapshot["background"], charOrder, idStr)
		return err

	case EntityWorldview:
		sortOrder := 0
		if so, ok := snapshot["sort_order"]; ok && so != nil {
			switch v := so.(type) {
			case float64:
				sortOrder = int(v)
			case json.Number:
				n, _ := v.Int64()
				sortOrder = int(n)
			}
		}
		_, err := database.DB.Exec(
			"UPDATE worldview SET category=?, title=?, content=?, sort_order=? WHERE id=?",
			snapshot["category"], snapshot["title"], snapshot["content"], sortOrder, idStr)
		return err

	case EntityForeshadowing:
		status := 0
		if s, ok := snapshot["status"]; ok && s != nil {
			switch v := s.(type) {
			case float64:
				status = int(v)
			case json.Number:
				n, _ := v.Int64()
				status = int(n)
			}
		}
		importance := 3
		if imp, ok := snapshot["importance"]; ok && imp != nil {
			switch v := imp.(type) {
			case float64:
				importance = int(v)
			case json.Number:
				n, _ := v.Int64()
				importance = int(n)
			}
		}
		_, err := database.DB.Exec(
			"UPDATE foreshadowing SET title=?, description=?, planted_chapter_id=?, resolved_chapter_id=?, status=?, importance=? WHERE id=?",
			snapshot["title"], snapshot["description"], snapshot["planted_chapter_id"], snapshot["resolved_chapter_id"],
			status, importance, idStr)
		return err
	}

	return fmt.Errorf("unknown entity type: %s", entityType)
}
