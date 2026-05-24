package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"novel-service/database"
	"novel-service/mcp/types"
)

// Helper: success result
func successResult(data interface{}) *types.CallToolResult {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return &types.CallToolResult{
		Content: []types.ContentItem{
			{Type: "text", Text: string(jsonBytes)},
		},
	}
}

// Helper: error result
func errorResult(msg string) *types.CallToolResult {
	return &types.CallToolResult{
		Content: []types.ContentItem{
			{Type: "text", Text: fmt.Sprintf("Error: %s", msg)},
		},
		IsError: true,
	}
}

// Helper: get int arg
func getIntArg(args map[string]interface{}, key string) (int, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return int(val), true
	case int:
		return val, true
	case json.Number:
		n, _ := val.Int64()
		return int(n), true
	case string:
		n, err := strconv.Atoi(val)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

// Helper: get string arg
func getStringArg(args map[string]interface{}, key string) (string, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// Helper: get int64 arg (for IDs)
func getInt64Arg(args map[string]interface{}, key string) (int64, bool) {
	n, ok := getIntArg(args, key)
	return int64(n), ok
}

// ============================================================
// Novel
// ============================================================

func ListNovels(args map[string]interface{}) *types.CallToolResult {
	keyword, _ := getStringArg(args, "keyword")
	status, hasStatus := getIntArg(args, "status")
	page, _ := getIntArg(args, "page")
	pageSize, _ := getIntArg(args, "page_size")
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "WHERE 1=1"
	sqlArgs := []interface{}{}
	if hasStatus && status >= 0 {
		where += " AND status = ?"
		sqlArgs = append(sqlArgs, status)
	}
	if keyword != "" {
		where += " AND (title LIKE ? OR author LIKE ?)"
		sqlArgs = append(sqlArgs, "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM novel "+where, sqlArgs...).Scan(&total)

	query := "SELECT n.id, n.title, n.author, n.description, n.cover_url, n.status, " +
		"(SELECT COUNT(*) FROM chapter c WHERE c.novel_id = n.id AND c.status = 1) as chapter_count, " +
		"n.created_at, n.updated_at FROM novel n " + where +
		" ORDER BY n.updated_at DESC LIMIT ? OFFSET ?"
	sqlArgs = append(sqlArgs, pageSize, offset)

	rows, err := database.DB.Query(query, sqlArgs...)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Novel struct {
		ID           uint64 `json:"id"`
		Title        string `json:"title"`
		Author       string `json:"author"`
		Description  string `json:"description"`
		CoverURL     string `json:"cover_url"`
		Status       int    `json:"status"`
		ChapterCount int    `json:"chapter_count"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}

	novels := make([]Novel, 0)
	for rows.Next() {
		var n Novel
		rows.Scan(&n.ID, &n.Title, &n.Author, &n.Description, &n.CoverURL,
			&n.Status, &n.ChapterCount, &n.CreatedAt, &n.UpdatedAt)
		novels = append(novels, n)
	}

	return successResult(map[string]interface{}{
		"list":  novels,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func GetNovel(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}

	var n struct {
		ID           uint64 `json:"id"`
		Title        string `json:"title"`
		Author       string `json:"author"`
		Description  string `json:"description"`
		CoverURL     string `json:"cover_url"`
		Status       int    `json:"status"`
		ChapterCount int    `json:"chapter_count"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}

	err := database.DB.QueryRow(
		"SELECT n.id, n.title, n.author, n.description, n.cover_url, n.status, "+
			"(SELECT COUNT(*) FROM chapter c WHERE c.novel_id = n.id AND c.status = 1) as chapter_count, "+
			"n.created_at, n.updated_at FROM novel n WHERE n.id = ?", id).
		Scan(&n.ID, &n.Title, &n.Author, &n.Description, &n.CoverURL,
			&n.Status, &n.ChapterCount, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return errorResult("小说不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(n)
}

func CreateNovel(args map[string]interface{}) *types.CallToolResult {
	title, _ := getStringArg(args, "title")
	if title == "" {
		return errorResult("title is required")
	}
	author, _ := getStringArg(args, "author")
	description, _ := getStringArg(args, "description")
	coverURL, _ := getStringArg(args, "cover_url")
	status, _ := getIntArg(args, "status")

	result, err := database.DB.Exec(
		"INSERT INTO novel (title, author, description, cover_url, status) VALUES (?, ?, ?, ?, ?)",
		title, author, description, coverURL, status)
	if err != nil {
		return errorResult(err.Error())
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{
		"id":       id,
		"title":    title,
		"author":   author,
		"status":   status,
		"message":  "小说创建成功",
	})
}

func UpdateNovel(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	idStr := strconv.FormatInt(id, 10)

	// Save version snapshot
	if err := saveVersionSnapshot("novel", idStr, "更新小说信息"); err != nil {
		return errorResult("保存版本失败: " + err.Error())
	}

	title, _ := getStringArg(args, "title")
	author, _ := getStringArg(args, "author")
	description, _ := getStringArg(args, "description")
	coverURL, _ := getStringArg(args, "cover_url")
	status, hasStatus := getIntArg(args, "status")

	// Build dynamic UPDATE
	setClauses := []string{}
	setArgs := []interface{}{}
	if title != "" {
		setClauses = append(setClauses, "title=?")
		setArgs = append(setArgs, title)
	}
	if author != "" {
		setClauses = append(setClauses, "author=?")
		setArgs = append(setArgs, author)
	}
	if description != "" {
		setClauses = append(setClauses, "description=?")
		setArgs = append(setArgs, description)
	}
	if coverURL != "" {
		setClauses = append(setClauses, "cover_url=?")
		setArgs = append(setArgs, coverURL)
	}
	if hasStatus {
		setClauses = append(setClauses, "status=?")
		setArgs = append(setArgs, status)
	}
	if len(setClauses) == 0 {
		return errorResult("没有需要更新的字段")
	}

	setArgs = append(setArgs, idStr)
	query := "UPDATE novel SET " + joinStrings(setClauses, ", ") + " WHERE id=?"
	_, err := database.DB.Exec(query, setArgs...)
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(map[string]interface{}{"message": "更新成功"})
}

func DeleteNovel(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	_, err := database.DB.Exec("DELETE FROM novel WHERE id=?", id)
	if err != nil {
		return errorResult(err.Error())
	}
	return successResult(map[string]interface{}{"message": "删除成功"})
}

// ============================================================
// Chapter
// ============================================================

func ListChapters(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}

	rows, err := database.DB.Query(
		"SELECT id, novel_id, title, word_count, chapter_order, status, created_at, updated_at "+
			"FROM chapter WHERE novel_id=? ORDER BY chapter_order ASC", novelID)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Chapter struct {
		ID           uint64 `json:"id"`
		NovelID      uint64 `json:"novel_id"`
		Title        string `json:"title"`
		WordCount    int    `json:"word_count"`
		ChapterOrder int    `json:"chapter_order"`
		Status       int    `json:"status"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}

	chapters := make([]Chapter, 0)
	for rows.Next() {
		var ch Chapter
		rows.Scan(&ch.ID, &ch.NovelID, &ch.Title, &ch.WordCount, &ch.ChapterOrder, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt)
		chapters = append(chapters, ch)
	}

	return successResult(chapters)
}

func GetChapter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "chapter_id")
	if !ok {
		return errorResult("chapter_id is required")
	}

	var ch struct {
		ID           uint64 `json:"id"`
		NovelID      uint64 `json:"novel_id"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		WordCount    int    `json:"word_count"`
		ChapterOrder int    `json:"chapter_order"`
		Status       int    `json:"status"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}

	err := database.DB.QueryRow(
		"SELECT id, novel_id, title, content, word_count, chapter_order, status, created_at, updated_at "+
			"FROM chapter WHERE id=?", id).
		Scan(&ch.ID, &ch.NovelID, &ch.Title, &ch.Content, &ch.WordCount,
			&ch.ChapterOrder, &ch.Status, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		return errorResult("章节不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(ch)
}

func CreateChapter(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	title, _ := getStringArg(args, "title")
	if title == "" {
		return errorResult("title is required")
	}
	content, _ := getStringArg(args, "content")
	chapterOrder, _ := getIntArg(args, "chapter_order")
	status, _ := getIntArg(args, "status")

	// Auto chapter order
	if chapterOrder == 0 {
		var maxOrder sql.NullInt64
		database.DB.QueryRow("SELECT MAX(chapter_order) FROM chapter WHERE novel_id=?", novelID).Scan(&maxOrder)
		if maxOrder.Valid {
			chapterOrder = int(maxOrder.Int64) + 1
		} else {
			chapterOrder = 1
		}
	}

	wordCount := len([]rune(content))

	result, err := database.DB.Exec(
		"INSERT INTO chapter (novel_id, title, content, word_count, chapter_order, status) VALUES (?, ?, ?, ?, ?, ?)",
		novelID, title, content, wordCount, chapterOrder, status)
	if err != nil {
		return errorResult(err.Error())
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{
		"id":            id,
		"novel_id":      novelID,
		"title":         title,
		"word_count":    wordCount,
		"chapter_order": chapterOrder,
		"message":       "章节创建成功",
	})
}

func UpdateChapter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "chapter_id")
	if !ok {
		return errorResult("chapter_id is required")
	}
	idStr := strconv.FormatInt(id, 10)

	// Save version snapshot
	if err := saveVersionSnapshot("chapter", idStr, "更新章节"); err != nil {
		return errorResult("保存版本失败: " + err.Error())
	}

	title, _ := getStringArg(args, "title")
	content, _ := getStringArg(args, "content")
	chapterOrder, hasOrder := getIntArg(args, "chapter_order")
	status, hasStatus := getIntArg(args, "status")

	setClauses := []string{}
	setArgs := []interface{}{}
	if title != "" {
		setClauses = append(setClauses, "title=?")
		setArgs = append(setArgs, title)
	}
	if content != "" {
		wordCount := len([]rune(content))
		setClauses = append(setClauses, "content=?", "word_count=?")
		setArgs = append(setArgs, content, wordCount)
	}
	if hasOrder {
		setClauses = append(setClauses, "chapter_order=?")
		setArgs = append(setArgs, chapterOrder)
	}
	if hasStatus {
		setClauses = append(setClauses, "status=?")
		setArgs = append(setArgs, status)
	}
	if len(setClauses) == 0 {
		return errorResult("没有需要更新的字段")
	}

	setArgs = append(setArgs, idStr)
	query := "UPDATE chapter SET " + joinStrings(setClauses, ", ") + " WHERE id=?"
	_, err := database.DB.Exec(query, setArgs...)
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(map[string]interface{}{"message": "更新成功"})
}

func DeleteChapter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "chapter_id")
	if !ok {
		return errorResult("chapter_id is required")
	}
	_, err := database.DB.Exec("DELETE FROM chapter WHERE id=?", id)
	if err != nil {
		return errorResult(err.Error())
	}
	return successResult(map[string]interface{}{"message": "删除成功"})
}

// ============================================================
// Character
// ============================================================

func ListCharacters(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}

	rows, err := database.DB.Query(
		"SELECT id, novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order, created_at, updated_at "+
			"FROM `character` WHERE novel_id=? ORDER BY character_order ASC", novelID)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Character struct {
		ID             uint64 `json:"id"`
		NovelID        uint64 `json:"novel_id"`
		Name           string `json:"name"`
		Alias          string `json:"alias"`
		AvatarURL      string `json:"avatar_url"`
		Gender         int    `json:"gender"`
		Age            string `json:"age"`
		Description    string `json:"description"`
		Personality    string `json:"personality"`
		Background     string `json:"background"`
		CharacterOrder int    `json:"character_order"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
	}

	characters := make([]Character, 0)
	for rows.Next() {
		var ch Character
		rows.Scan(&ch.ID, &ch.NovelID, &ch.Name, &ch.Alias, &ch.AvatarURL,
			&ch.Gender, &ch.Age, &ch.Description, &ch.Personality, &ch.Background,
			&ch.CharacterOrder, &ch.CreatedAt, &ch.UpdatedAt)
		characters = append(characters, ch)
	}

	return successResult(characters)
}

func GetCharacter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "character_id")
	if !ok {
		return errorResult("character_id is required")
	}

	var ch struct {
		ID             uint64 `json:"id"`
		NovelID        uint64 `json:"novel_id"`
		Name           string `json:"name"`
		Alias          string `json:"alias"`
		AvatarURL      string `json:"avatar_url"`
		Gender         int    `json:"gender"`
		Age            string `json:"age"`
		Description    string `json:"description"`
		Personality    string `json:"personality"`
		Background     string `json:"background"`
		CharacterOrder int    `json:"character_order"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
	}

	err := database.DB.QueryRow(
		"SELECT id, novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order, created_at, updated_at "+
			"FROM `character` WHERE id=?", id).
		Scan(&ch.ID, &ch.NovelID, &ch.Name, &ch.Alias, &ch.AvatarURL,
			&ch.Gender, &ch.Age, &ch.Description, &ch.Personality, &ch.Background,
			&ch.CharacterOrder, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		return errorResult("人物不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(ch)
}

func CreateCharacter(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	name, _ := getStringArg(args, "name")
	if name == "" {
		return errorResult("name is required")
	}
	alias, _ := getStringArg(args, "alias")
	avatarURL, _ := getStringArg(args, "avatar_url")
	gender, _ := getIntArg(args, "gender")
	age, _ := getStringArg(args, "age")
	description, _ := getStringArg(args, "description")
	personality, _ := getStringArg(args, "personality")
	background, _ := getStringArg(args, "background")

	charOrder := 0
	// Auto order
	var maxOrder sql.NullInt64
	database.DB.QueryRow("SELECT MAX(character_order) FROM `character` WHERE novel_id=?", novelID).Scan(&maxOrder)
	if maxOrder.Valid {
		charOrder = int(maxOrder.Int64) + 1
	} else {
		charOrder = 1
	}

	result, err := database.DB.Exec(
		"INSERT INTO `character` (novel_id, name, alias, avatar_url, gender, age, description, personality, background, character_order) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		novelID, name, alias, avatarURL, gender, age, description, personality, background, charOrder)
	if err != nil {
		return errorResult(err.Error())
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{
		"id":       id,
		"novel_id": novelID,
		"name":     name,
		"message":  "人物创建成功",
	})
}

func UpdateCharacter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "character_id")
	if !ok {
		return errorResult("character_id is required")
	}
	idStr := strconv.FormatInt(id, 10)

	// Save version snapshot
	if err := saveVersionSnapshot("character", idStr, "更新人物"); err != nil {
		return errorResult("保存版本失败: " + err.Error())
	}

	name, _ := getStringArg(args, "name")
	alias, _ := getStringArg(args, "alias")
	avatarURL, _ := getStringArg(args, "avatar_url")
	gender, hasGender := getIntArg(args, "gender")
	age, _ := getStringArg(args, "age")
	description, _ := getStringArg(args, "description")
	personality, _ := getStringArg(args, "personality")
	background, _ := getStringArg(args, "background")

	setClauses := []string{}
	setArgs := []interface{}{}
	if name != "" {
		setClauses = append(setClauses, "name=?")
		setArgs = append(setArgs, name)
	}
	if alias != "" {
		setClauses = append(setClauses, "alias=?")
		setArgs = append(setArgs, alias)
	}
	if avatarURL != "" {
		setClauses = append(setClauses, "avatar_url=?")
		setArgs = append(setArgs, avatarURL)
	}
	if hasGender {
		setClauses = append(setClauses, "gender=?")
		setArgs = append(setArgs, gender)
	}
	if age != "" {
		setClauses = append(setClauses, "age=?")
		setArgs = append(setArgs, age)
	}
	if description != "" {
		setClauses = append(setClauses, "description=?")
		setArgs = append(setArgs, description)
	}
	if personality != "" {
		setClauses = append(setClauses, "personality=?")
		setArgs = append(setArgs, personality)
	}
	if background != "" {
		setClauses = append(setClauses, "background=?")
		setArgs = append(setArgs, background)
	}
	if len(setClauses) == 0 {
		return errorResult("没有需要更新的字段")
	}

	setArgs = append(setArgs, idStr)
	query := "UPDATE `character` SET " + joinStrings(setClauses, ", ") + " WHERE id=?"
	_, err := database.DB.Exec(query, setArgs...)
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(map[string]interface{}{"message": "更新成功"})
}

func DeleteCharacter(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "character_id")
	if !ok {
		return errorResult("character_id is required")
	}
	_, err := database.DB.Exec("DELETE FROM `character` WHERE id=?", id)
	if err != nil {
		return errorResult(err.Error())
	}
	return successResult(map[string]interface{}{"message": "删除成功"})
}

// ============================================================
// Worldview
// ============================================================

func ListWorldviews(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	category, _ := getStringArg(args, "category")

	query := "SELECT id, novel_id, category, title, content, sort_order, created_at, updated_at FROM worldview WHERE novel_id=?"
	sqlArgs := []interface{}{novelID}
	if category != "" {
		query += " AND category=?"
		sqlArgs = append(sqlArgs, category)
	}
	query += " ORDER BY category, sort_order ASC"

	rows, err := database.DB.Query(query, sqlArgs...)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Worldview struct {
		ID        uint64 `json:"id"`
		NovelID   uint64 `json:"novel_id"`
		Category  string `json:"category"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		SortOrder int    `json:"sort_order"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	worldviews := make([]Worldview, 0)
	for rows.Next() {
		var w Worldview
		rows.Scan(&w.ID, &w.NovelID, &w.Category, &w.Title, &w.Content, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt)
		worldviews = append(worldviews, w)
	}

	// Get categories
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

	return successResult(map[string]interface{}{
		"list":       worldviews,
		"categories": categories,
	})
}

func GetWorldview(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "worldview_id")
	if !ok {
		return errorResult("worldview_id is required")
	}

	var w struct {
		ID        uint64 `json:"id"`
		NovelID   uint64 `json:"novel_id"`
		Category  string `json:"category"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		SortOrder int    `json:"sort_order"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	err := database.DB.QueryRow(
		"SELECT id, novel_id, category, title, content, sort_order, created_at, updated_at FROM worldview WHERE id=?", id).
		Scan(&w.ID, &w.NovelID, &w.Category, &w.Title, &w.Content, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return errorResult("世界观设定不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(w)
}

func CreateWorldview(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	category, _ := getStringArg(args, "category")
	if category == "" {
		category = "其他"
	}
	title, _ := getStringArg(args, "title")
	if title == "" {
		return errorResult("title is required")
	}
	content, _ := getStringArg(args, "content")
	sortOrder, _ := getIntArg(args, "sort_order")

	if sortOrder == 0 {
		var maxOrder sql.NullInt64
		database.DB.QueryRow("SELECT MAX(sort_order) FROM worldview WHERE novel_id=? AND category=?", novelID, category).Scan(&maxOrder)
		if maxOrder.Valid {
			sortOrder = int(maxOrder.Int64) + 1
		} else {
			sortOrder = 1
		}
	}

	result, err := database.DB.Exec(
		"INSERT INTO worldview (novel_id, category, title, content, sort_order) VALUES (?, ?, ?, ?, ?)",
		novelID, category, title, content, sortOrder)
	if err != nil {
		return errorResult(err.Error())
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{
		"id":       id,
		"novel_id": novelID,
		"category": category,
		"title":    title,
		"message":  "世界观设定创建成功",
	})
}

func UpdateWorldview(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "worldview_id")
	if !ok {
		return errorResult("worldview_id is required")
	}
	idStr := strconv.FormatInt(id, 10)

	// Save version snapshot
	if err := saveVersionSnapshot("worldview", idStr, "更新世界观设定"); err != nil {
		return errorResult("保存版本失败: " + err.Error())
	}

	category, _ := getStringArg(args, "category")
	title, _ := getStringArg(args, "title")
	content, _ := getStringArg(args, "content")
	sortOrder, hasSortOrder := getIntArg(args, "sort_order")

	setClauses := []string{}
	setArgs := []interface{}{}
	if category != "" {
		setClauses = append(setClauses, "category=?")
		setArgs = append(setArgs, category)
	}
	if title != "" {
		setClauses = append(setClauses, "title=?")
		setArgs = append(setArgs, title)
	}
	if content != "" {
		setClauses = append(setClauses, "content=?")
		setArgs = append(setArgs, content)
	}
	if hasSortOrder {
		setClauses = append(setClauses, "sort_order=?")
		setArgs = append(setArgs, sortOrder)
	}
	if len(setClauses) == 0 {
		return errorResult("没有需要更新的字段")
	}

	setArgs = append(setArgs, idStr)
	query := "UPDATE worldview SET " + joinStrings(setClauses, ", ") + " WHERE id=?"
	_, err := database.DB.Exec(query, setArgs...)
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(map[string]interface{}{"message": "更新成功"})
}

func DeleteWorldview(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "worldview_id")
	if !ok {
		return errorResult("worldview_id is required")
	}
	_, err := database.DB.Exec("DELETE FROM worldview WHERE id=?", id)
	if err != nil {
		return errorResult(err.Error())
	}
	return successResult(map[string]interface{}{"message": "删除成功"})
}

// ============================================================
// Foreshadowing
// ============================================================

func ListForeshadowings(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	status, hasStatus := getIntArg(args, "status")

	query := "SELECT f.id, f.novel_id, f.title, f.description, f.planted_chapter_id, f.resolved_chapter_id, " +
		"f.status, f.importance, f.created_at, f.updated_at, " +
		"pc.title as planted_chapter_title, rc.title as resolved_chapter_title " +
		"FROM foreshadowing f " +
		"LEFT JOIN chapter pc ON f.planted_chapter_id = pc.id " +
		"LEFT JOIN chapter rc ON f.resolved_chapter_id = rc.id " +
		"WHERE f.novel_id=?"
	sqlArgs := []interface{}{novelID}

	if hasStatus {
		query += " AND f.status=?"
		sqlArgs = append(sqlArgs, status)
	}
	query += " ORDER BY f.importance DESC, f.created_at ASC"

	rows, err := database.DB.Query(query, sqlArgs...)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Foreshadowing struct {
		ID                   uint64  `json:"id"`
		NovelID              uint64  `json:"novel_id"`
		Title                string  `json:"title"`
		Description          string  `json:"description"`
		PlantedChapterID     *uint64 `json:"planted_chapter_id"`
		ResolvedChapterID    *uint64 `json:"resolved_chapter_id"`
		Status               int     `json:"status"`
		Importance           int     `json:"importance"`
		CreatedAt            string  `json:"created_at"`
		UpdatedAt            string  `json:"updated_at"`
		PlantedChapterTitle  string  `json:"planted_chapter_title"`
		ResolvedChapterTitle string  `json:"resolved_chapter_title"`
	}

	list := make([]Foreshadowing, 0)
	for rows.Next() {
		var f Foreshadowing
		var plantedTitle, resolvedTitle sql.NullString
		rows.Scan(&f.ID, &f.NovelID, &f.Title, &f.Description,
			&f.PlantedChapterID, &f.ResolvedChapterID, &f.Status, &f.Importance,
			&f.CreatedAt, &f.UpdatedAt, &plantedTitle, &resolvedTitle)
		if plantedTitle.Valid {
			f.PlantedChapterTitle = plantedTitle.String
		}
		if resolvedTitle.Valid {
			f.ResolvedChapterTitle = resolvedTitle.String
		}
		list = append(list, f)
	}

	return successResult(list)
}

func GetForeshadowing(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "foreshadowing_id")
	if !ok {
		return errorResult("foreshadowing_id is required")
	}

	var f struct {
		ID                   uint64  `json:"id"`
		NovelID              uint64  `json:"novel_id"`
		Title                string  `json:"title"`
		Description          string  `json:"description"`
		PlantedChapterID     *uint64 `json:"planted_chapter_id"`
		ResolvedChapterID    *uint64 `json:"resolved_chapter_id"`
		Status               int     `json:"status"`
		Importance           int     `json:"importance"`
		CreatedAt            string  `json:"created_at"`
		UpdatedAt            string  `json:"updated_at"`
		PlantedChapterTitle  string  `json:"planted_chapter_title"`
		ResolvedChapterTitle string  `json:"resolved_chapter_title"`
	}
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
		return errorResult("伏笔不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}
	if plantedTitle.Valid {
		f.PlantedChapterTitle = plantedTitle.String
	}
	if resolvedTitle.Valid {
		f.ResolvedChapterTitle = resolvedTitle.String
	}

	return successResult(f)
}

func CreateForeshadowing(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	title, _ := getStringArg(args, "title")
	if title == "" {
		return errorResult("title is required")
	}
	description, _ := getStringArg(args, "description")
	status, _ := getIntArg(args, "status")
	importance, _ := getIntArg(args, "importance")
	if importance <= 0 {
		importance = 3
	}

	// Handle nullable chapter IDs
	var plantedChapterID, resolvedChapterID interface{}
	if pid, ok := getInt64Arg(args, "planted_chapter_id"); ok && pid > 0 {
		plantedChapterID = pid
	}
	if rid, ok := getInt64Arg(args, "resolved_chapter_id"); ok && rid > 0 {
		resolvedChapterID = rid
	}

	result, err := database.DB.Exec(
		"INSERT INTO foreshadowing (novel_id, title, description, planted_chapter_id, resolved_chapter_id, status, importance) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		novelID, title, description, plantedChapterID, resolvedChapterID, status, importance)
	if err != nil {
		return errorResult(err.Error())
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{
		"id":       id,
		"novel_id": novelID,
		"title":    title,
		"message":  "伏笔创建成功",
	})
}

func UpdateForeshadowing(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "foreshadowing_id")
	if !ok {
		return errorResult("foreshadowing_id is required")
	}
	idStr := strconv.FormatInt(id, 10)

	// Save version snapshot
	if err := saveVersionSnapshot("foreshadowing", idStr, "更新伏笔"); err != nil {
		return errorResult("保存版本失败: " + err.Error())
	}

	title, _ := getStringArg(args, "title")
	description, _ := getStringArg(args, "description")
	status, hasStatus := getIntArg(args, "status")
	importance, hasImportance := getIntArg(args, "importance")

	setClauses := []string{}
	setArgs := []interface{}{}
	if title != "" {
		setClauses = append(setClauses, "title=?")
		setArgs = append(setArgs, title)
	}
	if description != "" {
		setClauses = append(setClauses, "description=?")
		setArgs = append(setArgs, description)
	}
	if pid, ok := getInt64Arg(args, "planted_chapter_id"); ok {
		setClauses = append(setClauses, "planted_chapter_id=?")
		if pid > 0 {
			setArgs = append(setArgs, pid)
		} else {
			setArgs = append(setArgs, nil)
		}
	}
	if rid, ok := getInt64Arg(args, "resolved_chapter_id"); ok {
		setClauses = append(setClauses, "resolved_chapter_id=?")
		if rid > 0 {
			setArgs = append(setArgs, rid)
		} else {
			setArgs = append(setArgs, nil)
		}
	}
	if hasStatus {
		setClauses = append(setClauses, "status=?")
		setArgs = append(setArgs, status)
	}
	if hasImportance {
		setClauses = append(setClauses, "importance=?")
		setArgs = append(setArgs, importance)
	}
	if len(setClauses) == 0 {
		return errorResult("没有需要更新的字段")
	}

	setArgs = append(setArgs, idStr)
	query := "UPDATE foreshadowing SET " + joinStrings(setClauses, ", ") + " WHERE id=?"
	_, err := database.DB.Exec(query, setArgs...)
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(map[string]interface{}{"message": "更新成功"})
}

func DeleteForeshadowing(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "foreshadowing_id")
	if !ok {
		return errorResult("foreshadowing_id is required")
	}
	_, err := database.DB.Exec("DELETE FROM foreshadowing WHERE id=?", id)
	if err != nil {
		return errorResult(err.Error())
	}
	return successResult(map[string]interface{}{"message": "删除成功"})
}

// ============================================================
// Version
// ============================================================

func ListVersions(args map[string]interface{}) *types.CallToolResult {
	entityType, _ := getStringArg(args, "entity_type")
	entityID, ok := getInt64Arg(args, "entity_id")
	if !ok || entityType == "" {
		return errorResult("entity_type and entity_id are required")
	}

	validTypes := map[string]bool{"novel": true, "chapter": true, "character": true, "worldview": true, "foreshadowing": true}
	if !validTypes[entityType] {
		return errorResult("无效的实体类型，可选：novel, chapter, character, worldview, foreshadowing")
	}

	rows, err := database.DB.Query(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary, created_at "+
			"FROM content_version WHERE entity_type=? AND entity_id=? ORDER BY version DESC",
		entityType, entityID)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Version struct {
		ID            uint64 `json:"id"`
		EntityType    string `json:"entity_type"`
		EntityID      uint64 `json:"entity_id"`
		Version       int    `json:"version"`
		Snapshot      string `json:"snapshot"`
		ChangeSummary string `json:"change_summary"`
		CreatedAt     string `json:"created_at"`
	}

	versions := make([]Version, 0)
	for rows.Next() {
		var v Version
		rows.Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt)
		versions = append(versions, v)
	}

	return successResult(versions)
}

func GetVersion(args map[string]interface{}) *types.CallToolResult {
	id, ok := getInt64Arg(args, "version_id")
	if !ok {
		return errorResult("version_id is required")
	}

	var v struct {
		ID            uint64 `json:"id"`
		EntityType    string `json:"entity_type"`
		EntityID      uint64 `json:"entity_id"`
		Version       int    `json:"version"`
		Snapshot      string `json:"snapshot"`
		ChangeSummary string `json:"change_summary"`
		CreatedAt     string `json:"created_at"`
	}

	err := database.DB.QueryRow(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary, created_at FROM content_version WHERE id=?", id).
		Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return errorResult("版本不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	return successResult(v)
}

func RollbackVersion(args map[string]interface{}) *types.CallToolResult {
	versionID, ok := getInt64Arg(args, "version_id")
	if !ok {
		return errorResult("version_id is required")
	}

	// Get the target version
	var v struct {
		ID            uint64 `json:"id"`
		EntityType    string `json:"entity_type"`
		EntityID      uint64 `json:"entity_id"`
		Version       int    `json:"version"`
		Snapshot      string `json:"snapshot"`
		ChangeSummary string `json:"change_summary"`
	}
	err := database.DB.QueryRow(
		"SELECT id, entity_type, entity_id, version, snapshot, change_summary FROM content_version WHERE id=?", versionID).
		Scan(&v.ID, &v.EntityType, &v.EntityID, &v.Version, &v.Snapshot, &v.ChangeSummary)
	if err == sql.ErrNoRows {
		return errorResult("版本不存在")
	}
	if err != nil {
		return errorResult(err.Error())
	}

	// Parse snapshot
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(v.Snapshot), &snapshot); err != nil {
		return errorResult("快照数据解析失败")
	}

	changeSummary, _ := getStringArg(args, "change_summary")
	if changeSummary == "" {
		changeSummary = fmt.Sprintf("回退到版本%d", v.Version)
	}

	// Save current state as new version first
	entityIDStr := strconv.FormatUint(v.EntityID, 64)
	if err := saveVersionSnapshot(v.EntityType, entityIDStr, changeSummary); err != nil {
		return errorResult("保存当前版本失败: " + err.Error())
	}

	// Apply snapshot to restore
	if err := applySnapshot(v.EntityType, v.EntityID, snapshot); err != nil {
		return errorResult("回退失败: " + err.Error())
	}

	return successResult(map[string]interface{}{
		"message":         "回退成功",
		"rolled_back_to":  v.Version,
		"entity_type":     v.EntityType,
		"entity_id":       v.EntityID,
	})
}

// ============================================================
// Shared version helpers (reimplemented for MCP tools without gin context)
// ============================================================

func saveVersionSnapshot(entityType, entityID string, changeSummary string) error {
	snapshot, err := getEntitySnapshot(entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}
	if snapshot == nil {
		return nil
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

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

func getEntitySnapshot(entityType, entityID string) (map[string]interface{}, error) {
	switch entityType {
	case "novel":
		var n struct {
			ID          uint64 `json:"id"`
			Title       string `json:"title"`
			Author      string `json:"author"`
			Description string `json:"description"`
			CoverURL    string `json:"cover_url"`
			Status      int    `json:"status"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}
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

	case "chapter":
		var ch struct {
			ID           uint64 `json:"id"`
			NovelID      uint64 `json:"novel_id"`
			Title        string `json:"title"`
			Content      string `json:"content"`
			WordCount    int    `json:"word_count"`
			ChapterOrder int    `json:"chapter_order"`
			Status       int    `json:"status"`
			CreatedAt    string `json:"created_at"`
			UpdatedAt    string `json:"updated_at"`
		}
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

	case "character":
		var ch struct {
			ID             uint64 `json:"id"`
			NovelID        uint64 `json:"novel_id"`
			Name           string `json:"name"`
			Alias          string `json:"alias"`
			AvatarURL      string `json:"avatar_url"`
			Gender         int    `json:"gender"`
			Age            string `json:"age"`
			Description    string `json:"description"`
			Personality    string `json:"personality"`
			Background     string `json:"background"`
			CharacterOrder int    `json:"character_order"`
			CreatedAt      string `json:"created_at"`
			UpdatedAt      string `json:"updated_at"`
		}
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

	case "worldview":
		var w struct {
			ID        uint64 `json:"id"`
			NovelID   uint64 `json:"novel_id"`
			Category  string `json:"category"`
			Title     string `json:"title"`
			Content   string `json:"content"`
			SortOrder int    `json:"sort_order"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}
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

	case "foreshadowing":
		var f struct {
			ID                uint64  `json:"id"`
			NovelID           uint64  `json:"novel_id"`
			Title             string  `json:"title"`
			Description       string  `json:"description"`
			PlantedChapterID  *uint64 `json:"planted_chapter_id"`
			ResolvedChapterID *uint64 `json:"resolved_chapter_id"`
			Status            int     `json:"status"`
			Importance        int     `json:"importance"`
			CreatedAt         string  `json:"created_at"`
			UpdatedAt         string  `json:"updated_at"`
		}
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

func applySnapshot(entityType string, entityID uint64, snapshot map[string]interface{}) error {
	idStr := strconv.FormatUint(entityID, 64)

	switch entityType {
	case "novel":
		_, err := database.DB.Exec(
			"UPDATE novel SET title=?, author=?, description=?, cover_url=?, status=? WHERE id=?",
			snapshot["title"], snapshot["author"], snapshot["description"], snapshot["cover_url"], snapshot["status"], idStr)
		return err

	case "chapter":
		wordCount := inttoFloat64(snapshot["word_count"])
		chapterOrder := inttoFloat64(snapshot["chapter_order"])
		status := inttoFloat64(snapshot["status"])
		_, err := database.DB.Exec(
			"UPDATE chapter SET title=?, content=?, word_count=?, chapter_order=?, status=? WHERE id=?",
			snapshot["title"], snapshot["content"], wordCount, chapterOrder, status, idStr)
		return err

	case "character":
		gender := inttoFloat64(snapshot["gender"])
		charOrder := inttoFloat64(snapshot["character_order"])
		_, err := database.DB.Exec(
			"UPDATE `character` SET name=?, alias=?, avatar_url=?, gender=?, age=?, description=?, personality=?, background=?, character_order=? WHERE id=?",
			snapshot["name"], snapshot["alias"], snapshot["avatar_url"], gender, snapshot["age"],
			snapshot["description"], snapshot["personality"], snapshot["background"], charOrder, idStr)
		return err

	case "worldview":
		sortOrder := inttoFloat64(snapshot["sort_order"])
		_, err := database.DB.Exec(
			"UPDATE worldview SET category=?, title=?, content=?, sort_order=? WHERE id=?",
			snapshot["category"], snapshot["title"], snapshot["content"], sortOrder, idStr)
		return err

	case "foreshadowing":
		status := inttoFloat64(snapshot["status"])
		importance := inttoFloat64(snapshot["importance"])
		_, err := database.DB.Exec(
			"UPDATE foreshadowing SET title=?, description=?, planted_chapter_id=?, resolved_chapter_id=?, status=?, importance=? WHERE id=?",
			snapshot["title"], snapshot["description"], snapshot["planted_chapter_id"], snapshot["resolved_chapter_id"],
			status, importance, idStr)
		return err
	}

	return fmt.Errorf("unknown entity type: %s", entityType)
}

func inttoFloat64(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	}
	return 0
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
