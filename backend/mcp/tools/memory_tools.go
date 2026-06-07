package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"novel-service/database"
	"novel-service/mcp/types"
)

func normalizeJSONArg(args map[string]interface{}, key string) interface{} {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if json.Valid([]byte(s)) {
			return s
		}
		b, _ := json.Marshal(s)
		return string(b)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

func nullableID(args map[string]interface{}, key string) interface{} {
	id, ok := getInt64Arg(args, key)
	if !ok || id <= 0 {
		return nil
	}
	return id
}

func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func likePattern(query string) string {
	return "%" + strings.TrimSpace(query) + "%"
}

func clipText(s string, limit int) string {
	if limit <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "..."
}

func snippetAround(text, query string, limit int) string {
	if limit <= 0 {
		limit = 300
	}
	if query == "" {
		return clipText(text, limit)
	}
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerText, lowerQuery)
	if idx < 0 {
		return clipText(text, limit)
	}
	start := idx - limit/3
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(text) {
		end = len(text)
	}
	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(text) {
		suffix = "..."
	}
	return prefix + text[start:end] + suffix
}

func memoryError(err error) *types.CallToolResult {
	if err == nil {
		return errorResult("unknown memory error")
	}
	if strings.Contains(err.Error(), "doesn't exist") || strings.Contains(err.Error(), "Unknown column") {
		return errorResult("小说记忆表尚未创建，请先执行 novel/sql/memory.sql: " + err.Error())
	}
	return errorResult(err.Error())
}

// ============================================================
// Chapter summary memory
// ============================================================

func UpdateChapterSummary(args map[string]interface{}) *types.CallToolResult {
	chapterID, ok := getInt64Arg(args, "chapter_id")
	if !ok {
		return errorResult("chapter_id is required")
	}

	var novelID int64
	if argNovelID, ok := getInt64Arg(args, "novel_id"); ok && argNovelID > 0 {
		novelID = argNovelID
	} else {
		if err := database.DB.QueryRow("SELECT novel_id FROM chapter WHERE id=?", chapterID).Scan(&novelID); err != nil {
			if err == sql.ErrNoRows {
				return errorResult("章节不存在")
			}
			return errorResult(err.Error())
		}
	}

	summary, _ := getStringArg(args, "summary")
	timelinePosition, _ := getStringArg(args, "timeline_position")

	_, err := database.DB.Exec(
		"INSERT INTO chapter_summary "+
			"(novel_id, chapter_id, summary, key_events, characters, locations, timeline_position, plot_threads, foreshadowing_changes, character_changes) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE novel_id=VALUES(novel_id), summary=VALUES(summary), key_events=VALUES(key_events), "+
			"characters=VALUES(characters), locations=VALUES(locations), timeline_position=VALUES(timeline_position), "+
			"plot_threads=VALUES(plot_threads), foreshadowing_changes=VALUES(foreshadowing_changes), "+
			"character_changes=VALUES(character_changes)",
		novelID, chapterID, summary,
		normalizeJSONArg(args, "key_events"),
		normalizeJSONArg(args, "characters"),
		normalizeJSONArg(args, "locations"),
		timelinePosition,
		normalizeJSONArg(args, "plot_threads"),
		normalizeJSONArg(args, "foreshadowing_changes"),
		normalizeJSONArg(args, "character_changes"),
	)
	if err != nil {
		return memoryError(err)
	}

	return successResult(map[string]interface{}{
		"message":    "章节记忆已更新",
		"novel_id":   novelID,
		"chapter_id": chapterID,
	})
}

func GetRecentChapterSummaries(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	limit, _ := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	rows, err := database.DB.Query(
		"SELECT cs.id, cs.novel_id, cs.chapter_id, c.title, c.chapter_order, cs.summary, cs.key_events, "+
			"cs.characters, cs.locations, cs.timeline_position, cs.plot_threads, cs.foreshadowing_changes, "+
			"cs.character_changes, cs.created_at, cs.updated_at "+
			"FROM chapter_summary cs JOIN chapter c ON cs.chapter_id=c.id "+
			"WHERE cs.novel_id=? ORDER BY c.chapter_order DESC LIMIT ?",
		novelID, limit)
	if err != nil {
		return memoryError(err)
	}
	defer rows.Close()

	type Summary struct {
		ID                   uint64 `json:"id"`
		NovelID              uint64 `json:"novel_id"`
		ChapterID            uint64 `json:"chapter_id"`
		ChapterTitle         string `json:"chapter_title"`
		ChapterOrder         int    `json:"chapter_order"`
		Summary              string `json:"summary"`
		KeyEvents            string `json:"key_events"`
		Characters           string `json:"characters"`
		Locations            string `json:"locations"`
		TimelinePosition     string `json:"timeline_position"`
		PlotThreads          string `json:"plot_threads"`
		ForeshadowingChanges string `json:"foreshadowing_changes"`
		CharacterChanges     string `json:"character_changes"`
		CreatedAt            string `json:"created_at"`
		UpdatedAt            string `json:"updated_at"`
	}

	list := make([]Summary, 0)
	for rows.Next() {
		var s Summary
		var summary, keyEvents, characters, locations, timelinePosition sql.NullString
		var plotThreads, foreshadowingChanges, characterChanges sql.NullString
		rows.Scan(&s.ID, &s.NovelID, &s.ChapterID, &s.ChapterTitle, &s.ChapterOrder,
			&summary, &keyEvents, &characters, &locations, &timelinePosition,
			&plotThreads, &foreshadowingChanges, &characterChanges, &s.CreatedAt, &s.UpdatedAt)
		s.Summary = nullStringValue(summary)
		s.KeyEvents = nullStringValue(keyEvents)
		s.Characters = nullStringValue(characters)
		s.Locations = nullStringValue(locations)
		s.TimelinePosition = nullStringValue(timelinePosition)
		s.PlotThreads = nullStringValue(plotThreads)
		s.ForeshadowingChanges = nullStringValue(foreshadowingChanges)
		s.CharacterChanges = nullStringValue(characterChanges)
		list = append(list, s)
	}

	return successResult(list)
}

// ============================================================
// Character state memory
// ============================================================

func UpdateCharacterState(args map[string]interface{}) *types.CallToolResult {
	characterID, ok := getInt64Arg(args, "character_id")
	if !ok {
		return errorResult("character_id is required")
	}

	var novelID int64
	if argNovelID, ok := getInt64Arg(args, "novel_id"); ok && argNovelID > 0 {
		novelID = argNovelID
	} else {
		if err := database.DB.QueryRow("SELECT novel_id FROM `character` WHERE id=?", characterID).Scan(&novelID); err != nil {
			if err == sql.ErrNoRows {
				return errorResult("人物不存在")
			}
			return errorResult(err.Error())
		}
	}

	currentState, _ := getStringArg(args, "current_state")
	location, _ := getStringArg(args, "location")
	goal, _ := getStringArg(args, "goal")
	relationshipSummary, _ := getStringArg(args, "relationship_summary")
	abilityState, _ := getStringArg(args, "ability_state")
	knowledgeState, _ := getStringArg(args, "knowledge_state")

	_, err := database.DB.Exec(
		"INSERT INTO character_state "+
			"(novel_id, character_id, current_state, location, goal, relationship_summary, ability_state, knowledge_state, last_seen_chapter_id, extra) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE novel_id=VALUES(novel_id), current_state=VALUES(current_state), location=VALUES(location), "+
			"goal=VALUES(goal), relationship_summary=VALUES(relationship_summary), ability_state=VALUES(ability_state), "+
			"knowledge_state=VALUES(knowledge_state), last_seen_chapter_id=VALUES(last_seen_chapter_id), extra=VALUES(extra)",
		novelID, characterID, currentState, location, goal, relationshipSummary,
		abilityState, knowledgeState, nullableID(args, "last_seen_chapter_id"), normalizeJSONArg(args, "extra"))
	if err != nil {
		return memoryError(err)
	}

	return successResult(map[string]interface{}{
		"message":      "人物状态已更新",
		"novel_id":     novelID,
		"character_id": characterID,
	})
}

func GetCharacterCurrentState(args map[string]interface{}) *types.CallToolResult {
	characterID, ok := getInt64Arg(args, "character_id")
	if !ok {
		return errorResult("character_id is required")
	}

	var s struct {
		ID                  uint64  `json:"id"`
		NovelID             uint64  `json:"novel_id"`
		CharacterID         uint64  `json:"character_id"`
		Name                string  `json:"name"`
		Alias               string  `json:"alias"`
		CurrentState        string  `json:"current_state"`
		Location            string  `json:"location"`
		Goal                string  `json:"goal"`
		RelationshipSummary string  `json:"relationship_summary"`
		AbilityState        string  `json:"ability_state"`
		KnowledgeState      string  `json:"knowledge_state"`
		LastSeenChapterID   *uint64 `json:"last_seen_chapter_id"`
		LastSeenChapter     string  `json:"last_seen_chapter_title"`
		Extra               string  `json:"extra"`
		UpdatedAt           string  `json:"updated_at"`
	}
	var lastTitle sql.NullString
	var currentState, location, goal, relationshipSummary sql.NullString
	var abilityState, knowledgeState, extra sql.NullString

	err := database.DB.QueryRow(
		"SELECT cs.id, cs.novel_id, cs.character_id, c.name, c.alias, cs.current_state, cs.location, "+
			"cs.goal, cs.relationship_summary, cs.ability_state, cs.knowledge_state, cs.last_seen_chapter_id, "+
			"ch.title, cs.extra, cs.updated_at "+
			"FROM character_state cs JOIN `character` c ON cs.character_id=c.id "+
			"LEFT JOIN chapter ch ON cs.last_seen_chapter_id=ch.id WHERE cs.character_id=?",
		characterID).Scan(&s.ID, &s.NovelID, &s.CharacterID, &s.Name, &s.Alias, &currentState,
		&location, &goal, &relationshipSummary, &abilityState, &knowledgeState,
		&s.LastSeenChapterID, &lastTitle, &extra, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return errorResult("人物当前状态不存在，请先调用 update_character_state 建立状态")
	}
	if err != nil {
		return memoryError(err)
	}
	if lastTitle.Valid {
		s.LastSeenChapter = lastTitle.String
	}
	s.CurrentState = nullStringValue(currentState)
	s.Location = nullStringValue(location)
	s.Goal = nullStringValue(goal)
	s.RelationshipSummary = nullStringValue(relationshipSummary)
	s.AbilityState = nullStringValue(abilityState)
	s.KnowledgeState = nullStringValue(knowledgeState)
	s.Extra = nullStringValue(extra)

	return successResult(s)
}

// ============================================================
// Plot memory and timeline
// ============================================================

func UpsertPlotMemory(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	title, _ := getStringArg(args, "title")
	content, _ := getStringArg(args, "content")
	if title == "" {
		return errorResult("title is required")
	}
	memoryType, _ := getStringArg(args, "memory_type")
	if memoryType == "" {
		memoryType = "fact"
	}
	importance, _ := getIntArg(args, "importance")
	if importance <= 0 {
		importance = 3
	}
	status, hasStatus := getIntArg(args, "status")
	if !hasStatus {
		status = 0
	}

	if id, ok := getInt64Arg(args, "memory_id"); ok && id > 0 {
		_, err := database.DB.Exec(
			"UPDATE plot_memory SET memory_type=?, title=?, content=?, importance=?, chapter_id=?, character_id=?, tags=?, status=? WHERE id=? AND novel_id=?",
			memoryType, title, content, importance, nullableID(args, "chapter_id"), nullableID(args, "character_id"),
			normalizeJSONArg(args, "tags"), status, id, novelID)
		if err != nil {
			return memoryError(err)
		}
		return successResult(map[string]interface{}{"message": "剧情记忆已更新", "id": id})
	}

	result, err := database.DB.Exec(
		"INSERT INTO plot_memory (novel_id, memory_type, title, content, importance, chapter_id, character_id, tags, status) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		novelID, memoryType, title, content, importance, nullableID(args, "chapter_id"),
		nullableID(args, "character_id"), normalizeJSONArg(args, "tags"), status)
	if err != nil {
		return memoryError(err)
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{"message": "剧情记忆已创建", "id": id})
}

func UpsertTimelineEvent(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	title, _ := getStringArg(args, "title")
	content, _ := getStringArg(args, "content")
	if title == "" {
		return errorResult("title is required")
	}
	sequenceNo, _ := getIntArg(args, "sequence_no")
	eventTime, _ := getStringArg(args, "event_time")
	importance, _ := getIntArg(args, "importance")
	if importance <= 0 {
		importance = 3
	}

	if id, ok := getInt64Arg(args, "timeline_id"); ok && id > 0 {
		_, err := database.DB.Exec(
			"UPDATE novel_timeline SET chapter_id=?, sequence_no=?, event_time=?, title=?, content=?, importance=? WHERE id=? AND novel_id=?",
			nullableID(args, "chapter_id"), sequenceNo, eventTime, title, content, importance, id, novelID)
		if err != nil {
			return memoryError(err)
		}
		return successResult(map[string]interface{}{"message": "时间线事件已更新", "id": id})
	}

	result, err := database.DB.Exec(
		"INSERT INTO novel_timeline (novel_id, chapter_id, sequence_no, event_time, title, content, importance) VALUES (?, ?, ?, ?, ?, ?, ?)",
		novelID, nullableID(args, "chapter_id"), sequenceNo, eventTime, title, content, importance)
	if err != nil {
		return memoryError(err)
	}
	id, _ := result.LastInsertId()

	return successResult(map[string]interface{}{"message": "时间线事件已创建", "id": id})
}

func ListTimelineEvents(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	limit, _ := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := database.DB.Query(
		"SELECT nt.id, nt.novel_id, nt.chapter_id, c.title, nt.sequence_no, nt.event_time, nt.title, nt.content, nt.importance, nt.updated_at "+
			"FROM novel_timeline nt LEFT JOIN chapter c ON nt.chapter_id=c.id "+
			"WHERE nt.novel_id=? ORDER BY nt.sequence_no ASC, nt.id ASC LIMIT ?",
		novelID, limit)
	if err != nil {
		return memoryError(err)
	}
	defer rows.Close()

	type Event struct {
		ID           uint64  `json:"id"`
		NovelID      uint64  `json:"novel_id"`
		ChapterID    *uint64 `json:"chapter_id"`
		ChapterTitle string  `json:"chapter_title"`
		SequenceNo   int     `json:"sequence_no"`
		EventTime    string  `json:"event_time"`
		Title        string  `json:"title"`
		Content      string  `json:"content"`
		Importance   int     `json:"importance"`
		UpdatedAt    string  `json:"updated_at"`
	}

	list := make([]Event, 0)
	for rows.Next() {
		var e Event
		var chapterTitle sql.NullString
		rows.Scan(&e.ID, &e.NovelID, &e.ChapterID, &chapterTitle, &e.SequenceNo, &e.EventTime,
			&e.Title, &e.Content, &e.Importance, &e.UpdatedAt)
		if chapterTitle.Valid {
			e.ChapterTitle = chapterTitle.String
		}
		list = append(list, e)
	}

	return successResult(list)
}

// ============================================================
// Retrieval context
// ============================================================

func SearchChapters(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	query, _ := getStringArg(args, "query")
	if strings.TrimSpace(query) == "" {
		return errorResult("query is required")
	}
	limit, _ := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}

	rows, err := database.DB.Query(
		"SELECT id, title, content, word_count, chapter_order, status, updated_at FROM chapter "+
			"WHERE novel_id=? AND (title LIKE ? OR content LIKE ?) ORDER BY chapter_order ASC LIMIT ?",
		novelID, likePattern(query), likePattern(query), limit)
	if err != nil {
		return errorResult(err.Error())
	}
	defer rows.Close()

	type Hit struct {
		ChapterID    uint64 `json:"chapter_id"`
		Title        string `json:"title"`
		Snippet      string `json:"snippet"`
		WordCount    int    `json:"word_count"`
		ChapterOrder int    `json:"chapter_order"`
		Status       int    `json:"status"`
		UpdatedAt    string `json:"updated_at"`
	}

	hits := make([]Hit, 0)
	for rows.Next() {
		var h Hit
		var content string
		rows.Scan(&h.ChapterID, &h.Title, &content, &h.WordCount, &h.ChapterOrder, &h.Status, &h.UpdatedAt)
		h.Snippet = snippetAround(content, query, 360)
		hits = append(hits, h)
	}

	return successResult(map[string]interface{}{"query": query, "results": hits})
}

func SearchNovelMemory(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	query, _ := getStringArg(args, "query")
	limit, _ := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}
	pattern := likePattern(query)

	results := make([]map[string]interface{}, 0)

	appendRows := func(source string, rows *sql.Rows, err error) error {
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id uint64
			var title, content, updatedAt string
			var importance int
			if err := rows.Scan(&id, &title, &content, &importance, &updatedAt); err != nil {
				return err
			}
			results = append(results, map[string]interface{}{
				"source":     source,
				"id":         id,
				"title":      title,
				"snippet":    snippetAround(content, query, 360),
				"importance": importance,
				"updated_at": updatedAt,
			})
		}
		return rows.Err()
	}

	rows, qerr := database.DB.Query(
		"SELECT id, title, content, importance, updated_at FROM plot_memory WHERE novel_id=? AND status=0 "+
			"AND (?='' OR title LIKE ? OR content LIKE ? OR memory_type LIKE ?) ORDER BY importance DESC, updated_at DESC LIMIT ?",
		novelID, strings.TrimSpace(query), pattern, pattern, pattern, limit)
	err := appendRows("plot_memory", rows, qerr)
	if err != nil {
		return memoryError(err)
	}
	rows, qerr = database.DB.Query(
		"SELECT cs.id, c.title, CONCAT(IFNULL(cs.summary,''),' ',IFNULL(cs.key_events,''),' ',IFNULL(cs.plot_threads,''),' ',IFNULL(cs.foreshadowing_changes,'')), "+
			"3, cs.updated_at FROM chapter_summary cs JOIN chapter c ON cs.chapter_id=c.id WHERE cs.novel_id=? "+
			"AND (?='' OR c.title LIKE ? OR cs.summary LIKE ? OR cs.key_events LIKE ? OR cs.plot_threads LIKE ? OR cs.foreshadowing_changes LIKE ?) "+
			"ORDER BY c.chapter_order DESC LIMIT ?",
		novelID, strings.TrimSpace(query), pattern, pattern, pattern, pattern, pattern, limit)
	err = appendRows("chapter_summary", rows, qerr)
	if err != nil {
		return memoryError(err)
	}
	rows, qerr = database.DB.Query(
		"SELECT cs.id, c.name, CONCAT(IFNULL(cs.current_state,''),' ',IFNULL(cs.location,''),' ',IFNULL(cs.goal,''),' ',IFNULL(cs.relationship_summary,''),' ',IFNULL(cs.ability_state,''),' ',IFNULL(cs.knowledge_state,'')), "+
			"3, cs.updated_at FROM character_state cs JOIN `character` c ON cs.character_id=c.id WHERE cs.novel_id=? "+
			"AND (?='' OR c.name LIKE ? OR c.alias LIKE ? OR cs.current_state LIKE ? OR cs.location LIKE ? OR cs.goal LIKE ? OR cs.relationship_summary LIKE ? OR cs.ability_state LIKE ? OR cs.knowledge_state LIKE ?) "+
			"ORDER BY cs.updated_at DESC LIMIT ?",
		novelID, strings.TrimSpace(query), pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, limit)
	err = appendRows("character_state", rows, qerr)
	if err != nil {
		return memoryError(err)
	}
	rows, qerr = database.DB.Query(
		"SELECT id, title, CONCAT(IFNULL(event_time,''),' ',IFNULL(content,'')), importance, updated_at FROM novel_timeline "+
			"WHERE novel_id=? AND (?='' OR title LIKE ? OR content LIKE ? OR event_time LIKE ?) ORDER BY sequence_no DESC, id DESC LIMIT ?",
		novelID, strings.TrimSpace(query), pattern, pattern, pattern, limit)
	err = appendRows("timeline", rows, qerr)
	if err != nil {
		return memoryError(err)
	}

	if len(results) > limit {
		results = results[:limit]
	}
	return successResult(map[string]interface{}{"query": query, "results": results, "count": len(results)})
}

func GetNovelContext(args map[string]interface{}) *types.CallToolResult {
	novelID, ok := getInt64Arg(args, "novel_id")
	if !ok {
		return errorResult("novel_id is required")
	}
	recentLimit, _ := getIntArg(args, "recent_limit")
	if recentLimit <= 0 {
		recentLimit = 5
	}
	query, _ := getStringArg(args, "query")

	var novel struct {
		ID           uint64 `json:"id"`
		Title        string `json:"title"`
		Author       string `json:"author"`
		Description  string `json:"description"`
		Status       int    `json:"status"`
		ChapterCount int    `json:"chapter_count"`
		UpdatedAt    string `json:"updated_at"`
	}
	if err := database.DB.QueryRow(
		"SELECT n.id, n.title, n.author, n.description, n.status, "+
			"(SELECT COUNT(*) FROM chapter c WHERE c.novel_id=n.id), n.updated_at FROM novel n WHERE n.id=?",
		novelID).Scan(&novel.ID, &novel.Title, &novel.Author, &novel.Description, &novel.Status, &novel.ChapterCount, &novel.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return errorResult("小说不存在")
		}
		return errorResult(err.Error())
	}

	recentSummaries := GetRecentChapterSummaries(map[string]interface{}{"novel_id": novelID, "limit": recentLimit})
	if recentSummaries.IsError {
		return recentSummaries
	}
	memoryHits := SearchNovelMemory(map[string]interface{}{"novel_id": novelID, "query": query, "limit": 10})
	if memoryHits.IsError {
		return memoryHits
	}

	characterRows, err := database.DB.Query(
		"SELECT cs.character_id, c.name, c.alias, cs.current_state, cs.location, cs.goal, cs.ability_state, cs.last_seen_chapter_id, cs.updated_at "+
			"FROM character_state cs JOIN `character` c ON cs.character_id=c.id WHERE cs.novel_id=? ORDER BY cs.updated_at DESC LIMIT 30",
		novelID)
	if err != nil {
		return memoryError(err)
	}
	defer characterRows.Close()

	type CharacterState struct {
		CharacterID       uint64  `json:"character_id"`
		Name              string  `json:"name"`
		Alias             string  `json:"alias"`
		CurrentState      string  `json:"current_state"`
		Location          string  `json:"location"`
		Goal              string  `json:"goal"`
		AbilityState      string  `json:"ability_state"`
		LastSeenChapterID *uint64 `json:"last_seen_chapter_id"`
		UpdatedAt         string  `json:"updated_at"`
	}
	characterStates := make([]CharacterState, 0)
	for characterRows.Next() {
		var cs CharacterState
		var currentState, location, goal, abilityState sql.NullString
		characterRows.Scan(&cs.CharacterID, &cs.Name, &cs.Alias, &currentState, &location,
			&goal, &abilityState, &cs.LastSeenChapterID, &cs.UpdatedAt)
		cs.CurrentState = nullStringValue(currentState)
		cs.Location = nullStringValue(location)
		cs.Goal = nullStringValue(goal)
		cs.AbilityState = nullStringValue(abilityState)
		characterStates = append(characterStates, cs)
	}

	unresolvedForeshadowings := ListForeshadowings(map[string]interface{}{"novel_id": novelID, "status": 0})
	timeline := ListTimelineEvents(map[string]interface{}{"novel_id": novelID, "limit": 30})

	return successResult(map[string]interface{}{
		"novel":                     novel,
		"recent_chapter_summaries":  recentSummaries.Content[0].Text,
		"character_states":          characterStates,
		"unresolved_foreshadowings": unresolvedForeshadowings.Content[0].Text,
		"timeline":                  timeline.Content[0].Text,
		"related_memory":            memoryHits.Content[0].Text,
		"note":                      fmt.Sprintf("recent_limit=%d, query=%q", recentLimit, query),
	})
}
