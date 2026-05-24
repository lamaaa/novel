package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"novel-service/database"
	"novel-service/models"
)

// --- 小说 ---

func GetNovels(c *gin.Context) {
	var req models.NovelListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}

	offset := (req.Page - 1) * req.PageSize

	where := "WHERE 1=1"
	args := []interface{}{}

	if req.Status >= 0 {
		where += " AND status = ?"
		args = append(args, req.Status)
	}
	if req.Keyword != "" {
		where += " AND (title LIKE ? OR author LIKE ?)"
		args = append(args, "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM novel "+where, args...).Scan(&total)

	query := "SELECT n.id, n.title, n.author, n.description, n.cover_url, n.status, " +
		"(SELECT COUNT(*) FROM chapter c WHERE c.novel_id = n.id AND c.status = 1) as chapter_count, " +
		"n.created_at, n.updated_at FROM novel n " + where +
		" ORDER BY n.updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	novels := make([]models.Novel, 0)
	for rows.Next() {
		var n models.Novel
		if err := rows.Scan(&n.ID, &n.Title, &n.Author, &n.Description, &n.CoverURL,
			&n.Status, &n.ChapterCount, &n.CreatedAt, &n.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		novels = append(novels, n)
	}

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: models.PageResponse{
		List: novels, Total: total, Page: req.Page, Size: req.PageSize,
	}})
}

func GetNovel(c *gin.Context) {
	id := c.Param("id")
	var n models.Novel
	err := database.DB.QueryRow(
		"SELECT n.id, n.title, n.author, n.description, n.cover_url, n.status, "+
			"(SELECT COUNT(*) FROM chapter c WHERE c.novel_id = n.id AND c.status = 1) as chapter_count, "+
			"n.created_at, n.updated_at FROM novel n WHERE n.id = ?", id).
		Scan(&n.ID, &n.Title, &n.Author, &n.Description, &n.CoverURL,
			&n.Status, &n.ChapterCount, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "小说不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: n})
}

func CreateNovel(c *gin.Context) {
	var n models.Novel
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	result, err := database.DB.Exec(
		"INSERT INTO novel (title, author, description, cover_url, status) VALUES (?, ?, ?, ?, ?)",
		n.Title, n.Author, n.Description, n.CoverURL, n.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	n.ID = uint64(id)
	c.JSON(http.StatusCreated, models.Response{Code: 0, Message: "创建成功", Data: n})
}

func UpdateNovel(c *gin.Context) {
	id := c.Param("id")
	var n models.Novel
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	// 保存版本快照
	if err := saveVersionSnapshot(EntityNovel, id, "更新小说信息"); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存版本失败: " + err.Error()})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE novel SET title=?, author=?, description=?, cover_url=?, status=? WHERE id=?",
		n.Title, n.Author, n.Description, n.CoverURL, n.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "更新成功"})
}

func DeleteNovel(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM novel WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "删除成功"})
}

// --- 章节 ---

func GetChapters(c *gin.Context) {
	novelID := c.Param("id")
	rows, err := database.DB.Query(
		"SELECT id, novel_id, title, word_count, chapter_order, status, created_at, updated_at "+
			"FROM chapter WHERE novel_id=? ORDER BY chapter_order ASC", novelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	chapters := make([]models.Chapter, 0)
	for rows.Next() {
		var ch models.Chapter
		if err := rows.Scan(&ch.ID, &ch.NovelID, &ch.Title, &ch.WordCount,
			&ch.ChapterOrder, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		chapters = append(chapters, ch)
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: chapters})
}

func GetChapter(c *gin.Context) {
	id := c.Param("id")
	var ch models.Chapter
	err := database.DB.QueryRow(
		"SELECT id, novel_id, title, content, word_count, chapter_order, status, created_at, updated_at "+
			"FROM chapter WHERE id=?", id).
		Scan(&ch.ID, &ch.NovelID, &ch.Title, &ch.Content, &ch.WordCount,
			&ch.ChapterOrder, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "章节不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}

	// 获取上一章/下一章
	var prevID, nextID uint64
	database.DB.QueryRow("SELECT id FROM chapter WHERE novel_id=? AND chapter_order < ? ORDER BY chapter_order DESC LIMIT 1",
		ch.NovelID, ch.ChapterOrder).Scan(&prevID)
	database.DB.QueryRow("SELECT id FROM chapter WHERE novel_id=? AND chapter_order > ? ORDER BY chapter_order ASC LIMIT 1",
		ch.NovelID, ch.ChapterOrder).Scan(&nextID)

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: gin.H{
		"chapter": ch, "prev_id": prevID, "next_id": nextID,
	}})
}

func CreateChapter(c *gin.Context) {
	novelID := c.Param("id")
	var ch models.Chapter
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}

	// 自动计算章节序号
	if ch.ChapterOrder == 0 {
		var maxOrder sql.NullInt64
		database.DB.QueryRow("SELECT MAX(chapter_order) FROM chapter WHERE novel_id=?", novelID).Scan(&maxOrder)
		if maxOrder.Valid {
			ch.ChapterOrder = int(maxOrder.Int64) + 1
		} else {
			ch.ChapterOrder = 1
		}
	}

	// 计算字数
	ch.WordCount = len([]rune(ch.Content))

	result, err := database.DB.Exec(
		"INSERT INTO chapter (novel_id, title, content, word_count, chapter_order, status) VALUES (?, ?, ?, ?, ?, ?)",
		novelID, ch.Title, ch.Content, ch.WordCount, ch.ChapterOrder, ch.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	ch.ID = uint64(id)
	ch.NovelID, _ = strconv.ParseUint(novelID, 10, 64)
	c.JSON(http.StatusCreated, models.Response{Code: 0, Message: "创建成功", Data: ch})
}

func UpdateChapter(c *gin.Context) {
	id := c.Param("id")
	var ch models.Chapter
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	ch.WordCount = len([]rune(ch.Content))
	// 保存版本快照
	if err := saveVersionSnapshot(EntityChapter, id, "更新章节"); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存版本失败: " + err.Error()})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE chapter SET title=?, content=?, word_count=?, chapter_order=?, status=? WHERE id=?",
		ch.Title, ch.Content, ch.WordCount, ch.ChapterOrder, ch.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "更新成功"})
}

func DeleteChapter(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM chapter WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "删除成功"})
}

// --- 人物 ---

func GetCharacters(c *gin.Context) {
	novelID := c.Param("id")
	rows, err := database.DB.Query(
		"SELECT id, novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order, created_at, updated_at "+
			"FROM `character` WHERE novel_id=? ORDER BY character_order ASC", novelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	characters := make([]models.Character, 0)
	for rows.Next() {
		var ch models.Character
		if err := rows.Scan(&ch.ID, &ch.NovelID, &ch.Name, &ch.Alias, &ch.AvatarURL,
			&ch.Gender, &ch.Age, &ch.Description, &ch.Personality, &ch.Background,
			&ch.CharacterOrder, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		characters = append(characters, ch)
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: characters})
}

func GetCharacter(c *gin.Context) {
	id := c.Param("id")
	var ch models.Character
	err := database.DB.QueryRow(
		"SELECT id, novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order, created_at, updated_at "+
			"FROM `character` WHERE id=?", id).
		Scan(&ch.ID, &ch.NovelID, &ch.Name, &ch.Alias, &ch.AvatarURL,
			&ch.Gender, &ch.Age, &ch.Description, &ch.Personality, &ch.Background,
			&ch.CharacterOrder, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "人物不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: ch})
}

func CreateCharacter(c *gin.Context) {
	novelID := c.Param("id")
	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}

	if ch.CharacterOrder == 0 {
		var maxOrder sql.NullInt64
		database.DB.QueryRow("SELECT MAX(character_order) FROM `character` WHERE novel_id=?", novelID).Scan(&maxOrder)
		if maxOrder.Valid {
			ch.CharacterOrder = int(maxOrder.Int64) + 1
		} else {
			ch.CharacterOrder = 1
		}
	}

	result, err := database.DB.Exec(
		"INSERT INTO `character` (novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		novelID, ch.Name, ch.Alias, ch.AvatarURL, ch.Gender, ch.Age,
		ch.Description, ch.Personality, ch.Background, ch.CharacterOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	ch.ID = uint64(id)
	ch.NovelID, _ = strconv.ParseUint(novelID, 10, 64)
	c.JSON(http.StatusCreated, models.Response{Code: 0, Message: "创建成功", Data: ch})
}

func UpdateCharacter(c *gin.Context) {
	id := c.Param("id")
	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	// 保存版本快照
	if err := saveVersionSnapshot(EntityCharacter, id, "更新人物"); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存版本失败: " + err.Error()})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE `character` SET name=?, alias=?, avatar_url=?, gender=?, age=?, description=?, personality=?, background=?, character_order=? WHERE id=?",
		ch.Name, ch.Alias, ch.AvatarURL, ch.Gender, ch.Age,
		ch.Description, ch.Personality, ch.Background, ch.CharacterOrder, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "更新成功"})
}

func DeleteCharacter(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM `character` WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "删除成功"})
}

// --- 世界观 ---

func GetWorldviews(c *gin.Context) {
	novelID := c.Param("id")
	category := c.Query("category")

	query := "SELECT id, novel_id, category, title, content, sort_order, created_at, updated_at FROM worldview WHERE novel_id=?"
	args := []interface{}{novelID}

	if category != "" {
		query += " AND category=?"
		args = append(args, category)
	}
	query += " ORDER BY category, sort_order ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	worldviews := make([]models.Worldview, 0)
	for rows.Next() {
		var w models.Worldview
		if err := rows.Scan(&w.ID, &w.NovelID, &w.Category, &w.Title, &w.Content,
			&w.SortOrder, &w.CreatedAt, &w.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		worldviews = append(worldviews, w)
	}

	// 获取分类列表
	categories := make([]string, 0)
	catRows, _ := database.DB.Query("SELECT DISTINCT category FROM worldview WHERE novel_id=? ORDER BY category", novelID)
	if catRows != nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat string
			catRows.Scan(&cat)
			categories = append(categories, cat)
		}
	}

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: gin.H{
		"list":       worldviews,
		"categories": categories,
	}})
}

func GetWorldview(c *gin.Context) {
	id := c.Param("id")
	var w models.Worldview
	err := database.DB.QueryRow(
		"SELECT id, novel_id, category, title, content, sort_order, created_at, updated_at FROM worldview WHERE id=?", id).
		Scan(&w.ID, &w.NovelID, &w.Category, &w.Title, &w.Content, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "世界观设定不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: w})
}

func CreateWorldview(c *gin.Context) {
	novelID := c.Param("id")
	var w models.Worldview
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}

	if w.SortOrder == 0 {
		var maxOrder sql.NullInt64
		database.DB.QueryRow("SELECT MAX(sort_order) FROM worldview WHERE novel_id=? AND category=?", novelID, w.Category).Scan(&maxOrder)
		if maxOrder.Valid {
			w.SortOrder = int(maxOrder.Int64) + 1
		} else {
			w.SortOrder = 1
		}
	}

	result, err := database.DB.Exec(
		"INSERT INTO worldview (novel_id, category, title, content, sort_order) VALUES (?, ?, ?, ?, ?)",
		novelID, w.Category, w.Title, w.Content, w.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	w.ID = uint64(id)
	w.NovelID, _ = strconv.ParseUint(novelID, 10, 64)
	c.JSON(http.StatusCreated, models.Response{Code: 0, Message: "创建成功", Data: w})
}

func UpdateWorldview(c *gin.Context) {
	id := c.Param("id")
	var w models.Worldview
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	// 保存版本快照
	if err := saveVersionSnapshot(EntityWorldview, id, "更新世界观设定"); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存版本失败: " + err.Error()})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE worldview SET category=?, title=?, content=?, sort_order=? WHERE id=?",
		w.Category, w.Title, w.Content, w.SortOrder, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "更新成功"})
}

func DeleteWorldview(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM worldview WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "删除成功"})
}

// --- 伏笔 ---

func GetForeshadowings(c *gin.Context) {
	novelID := c.Param("id")
	status := c.Query("status")

	query := "SELECT f.id, f.novel_id, f.title, f.description, f.planted_chapter_id, f.resolved_chapter_id, " +
		"f.status, f.importance, f.created_at, f.updated_at, " +
		"pc.title as planted_chapter_title, rc.title as resolved_chapter_title " +
		"FROM foreshadowing f " +
		"LEFT JOIN chapter pc ON f.planted_chapter_id = pc.id " +
		"LEFT JOIN chapter rc ON f.resolved_chapter_id = rc.id " +
		"WHERE f.novel_id=?"
	args := []interface{}{novelID}

	if status != "" {
		query += " AND f.status=?"
		args = append(args, status)
	}
	query += " ORDER BY f.importance DESC, f.created_at ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	defer rows.Close()

	list := make([]models.ForeshadowingDetail, 0)
	for rows.Next() {
		var f models.ForeshadowingDetail
		var plantedTitle, resolvedTitle sql.NullString
		if err := rows.Scan(&f.ID, &f.NovelID, &f.Title, &f.Description,
			&f.PlantedChapterID, &f.ResolvedChapterID, &f.Status, &f.Importance,
			&f.CreatedAt, &f.UpdatedAt, &plantedTitle, &resolvedTitle); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
			return
		}
		if plantedTitle.Valid {
			f.PlantedChapterTitle = plantedTitle.String
		}
		if resolvedTitle.Valid {
			f.ResolvedChapterTitle = resolvedTitle.String
		}
		list = append(list, f)
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: list})
}

func GetForeshadowing(c *gin.Context) {
	id := c.Param("id")
	var f models.ForeshadowingDetail
	var plantedTitle, resolvedTitle sql.NullString
	err := database.DB.QueryRow(
		"SELECT f.id, f.novel_id, f.title, f.description, f.planted_chapter_id, f.resolved_chapter_id, "+
			"f.status, f.importance, f.created_at, f.updated_at, "+
			"pc.title, rc.title "+
			"FROM foreshadowing f "+
			"LEFT JOIN chapter pc ON f.planted_chapter_id = pc.id "+
			"LEFT JOIN chapter rc ON f.resolved_chapter_id = rc.id "+
			"WHERE f.id=?", id).
		Scan(&f.ID, &f.NovelID, &f.Title, &f.Description,
			&f.PlantedChapterID, &f.ResolvedChapterID, &f.Status, &f.Importance,
			&f.CreatedAt, &f.UpdatedAt, &plantedTitle, &resolvedTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{Code: 1, Message: "伏笔不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	if plantedTitle.Valid {
		f.PlantedChapterTitle = plantedTitle.String
	}
	if resolvedTitle.Valid {
		f.ResolvedChapterTitle = resolvedTitle.String
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: f})
}

func CreateForeshadowing(c *gin.Context) {
	novelID := c.Param("id")
	var f models.Foreshadowing
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	result, err := database.DB.Exec(
		"INSERT INTO foreshadowing (novel_id, title, description, planted_chapter_id, resolved_chapter_id, status, importance) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		novelID, f.Title, f.Description, f.PlantedChapterID, f.ResolvedChapterID, f.Status, f.Importance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	f.ID = uint64(id)
	f.NovelID, _ = strconv.ParseUint(novelID, 10, 64)
	c.JSON(http.StatusCreated, models.Response{Code: 0, Message: "创建成功", Data: f})
}

func UpdateForeshadowing(c *gin.Context) {
	id := c.Param("id")
	var f models.Foreshadowing
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Code: 1, Message: err.Error()})
		return
	}
	// 保存版本快照
	if err := saveVersionSnapshot(EntityForeshadowing, id, "更新伏笔"); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: "保存版本失败: " + err.Error()})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE foreshadowing SET title=?, description=?, planted_chapter_id=?, resolved_chapter_id=?, status=?, importance=? WHERE id=?",
		f.Title, f.Description, f.PlantedChapterID, f.ResolvedChapterID, f.Status, f.Importance, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "更新成功"})
}

func DeleteForeshadowing(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM foreshadowing WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "删除成功"})
}
