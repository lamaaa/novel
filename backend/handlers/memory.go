package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"novel-service/database"
	"novel-service/models"
)

func memoryTableError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "Unknown column") {
		c.JSON(http.StatusOK, models.Response{
			Code:    1,
			Message: "长期记忆表尚未创建，请先执行 novel/sql/memory.sql",
		})
		return true
	}
	c.JSON(http.StatusInternalServerError, models.Response{Code: 1, Message: msg})
	return true
}

func strOrEmpty(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func uintPtrOrNil(v sql.NullInt64) *uint64 {
	if !v.Valid {
		return nil
	}
	n := uint64(v.Int64)
	return &n
}

func GetNovelMemory(c *gin.Context) {
	novelID := c.Param("id")
	limit := parseLimit(c.Query("limit"), 20, 100)

	summaries, err := queryChapterSummaries(novelID, limit)
	if err != nil {
		memoryTableError(c, err)
		return
	}
	characters, err := queryCharacterStates(novelID)
	if err != nil {
		memoryTableError(c, err)
		return
	}
	memories, err := queryPlotMemories(novelID, "", limit)
	if err != nil {
		memoryTableError(c, err)
		return
	}
	timeline, err := queryTimelineEvents(novelID, limit)
	if err != nil {
		memoryTableError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: gin.H{
		"summaries":  summaries,
		"characters": characters,
		"memories":   memories,
		"timeline":   timeline,
		"counts": gin.H{
			"summaries":  len(summaries),
			"characters": len(characters),
			"memories":   len(memories),
			"timeline":   len(timeline),
		},
	}})
}

func SearchNovelMemory(c *gin.Context) {
	novelID := c.Param("id")
	query := strings.TrimSpace(c.Query("query"))
	limit := parseLimit(c.Query("limit"), 20, 100)

	results := make([]gin.H, 0)
	appendRows := func(source string, rows *sql.Rows, err error) error {
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id uint64
			var title string
			var content sql.NullString
			var importance int
			var updatedAt string
			if err := rows.Scan(&id, &title, &content, &importance, &updatedAt); err != nil {
				return err
			}
			results = append(results, gin.H{
				"source":     source,
				"id":         id,
				"title":      title,
				"content":    strOrEmpty(content),
				"snippet":    makeSnippet(strOrEmpty(content), query, 320),
				"importance": importance,
				"updated_at": updatedAt,
			})
		}
		return rows.Err()
	}

	pattern := "%" + query + "%"
	rows, err := database.DB.Query(
		"SELECT id, title, content, importance, updated_at FROM plot_memory WHERE novel_id=? AND status=0 "+
			"AND (?='' OR title LIKE ? OR content LIKE ? OR memory_type LIKE ?) ORDER BY importance DESC, updated_at DESC LIMIT ?",
		novelID, query, pattern, pattern, pattern, limit)
	if err := appendRows("plot_memory", rows, err); err != nil {
		memoryTableError(c, err)
		return
	}
	rows, err = database.DB.Query(
		"SELECT cs.id, c.title, CONCAT(IFNULL(cs.summary,''),' ',IFNULL(cs.key_events,''),' ',IFNULL(cs.plot_threads,''),' ',IFNULL(cs.foreshadowing_changes,'')), "+
			"3, cs.updated_at FROM chapter_summary cs JOIN chapter c ON cs.chapter_id=c.id WHERE cs.novel_id=? "+
			"AND (?='' OR c.title LIKE ? OR cs.summary LIKE ? OR cs.key_events LIKE ? OR cs.plot_threads LIKE ? OR cs.foreshadowing_changes LIKE ?) "+
			"ORDER BY c.chapter_order DESC LIMIT ?",
		novelID, query, pattern, pattern, pattern, pattern, pattern, limit)
	if err := appendRows("chapter_summary", rows, err); err != nil {
		memoryTableError(c, err)
		return
	}
	rows, err = database.DB.Query(
		"SELECT cs.id, c.name, CONCAT(IFNULL(cs.current_state,''),' ',IFNULL(cs.location,''),' ',IFNULL(cs.goal,''),' ',IFNULL(cs.relationship_summary,''),' ',IFNULL(cs.ability_state,''),' ',IFNULL(cs.knowledge_state,'')), "+
			"3, cs.updated_at FROM character_state cs JOIN `character` c ON cs.character_id=c.id WHERE cs.novel_id=? "+
			"AND (?='' OR c.name LIKE ? OR c.alias LIKE ? OR cs.current_state LIKE ? OR cs.location LIKE ? OR cs.goal LIKE ? OR cs.relationship_summary LIKE ? OR cs.ability_state LIKE ? OR cs.knowledge_state LIKE ?) "+
			"ORDER BY cs.updated_at DESC LIMIT ?",
		novelID, query, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, limit)
	if err := appendRows("character_state", rows, err); err != nil {
		memoryTableError(c, err)
		return
	}
	rows, err = database.DB.Query(
		"SELECT id, title, CONCAT(IFNULL(event_time,''),' ',IFNULL(content,'')), importance, updated_at FROM novel_timeline "+
			"WHERE novel_id=? AND (?='' OR title LIKE ? OR content LIKE ? OR event_time LIKE ?) ORDER BY sequence_no DESC, id DESC LIMIT ?",
		novelID, query, pattern, pattern, pattern, limit)
	if err := appendRows("timeline", rows, err); err != nil {
		memoryTableError(c, err)
		return
	}

	if len(results) > limit {
		results = results[:limit]
	}
	c.JSON(http.StatusOK, models.Response{Code: 0, Message: "success", Data: gin.H{
		"query":   query,
		"results": results,
	}})
}

func parseLimit(raw string, fallback, max int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	if n > max {
		return max
	}
	return n
}

func makeSnippet(text, query string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	if query == "" {
		return string(runes[:limit]) + "..."
	}
	lowerText := strings.ToLower(string(runes))
	lowerQuery := strings.ToLower(query)
	byteIdx := strings.Index(lowerText, lowerQuery)
	if byteIdx < 0 {
		return string(runes[:limit]) + "..."
	}
	matchRuneIdx := len([]rune(lowerText[:byteIdx]))
	start := matchRuneIdx - limit/3
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(runes) {
		end = len(runes)
	}
	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(runes) {
		suffix = "..."
	}
	return prefix + string(runes[start:end]) + suffix
}

func queryChapterSummaries(novelID string, limit int) ([]gin.H, error) {
	rows, err := database.DB.Query(
		"SELECT cs.id, cs.chapter_id, c.title, c.chapter_order, cs.summary, cs.key_events, cs.characters, "+
			"cs.locations, cs.timeline_position, cs.plot_threads, cs.foreshadowing_changes, cs.character_changes, cs.updated_at "+
			"FROM chapter_summary cs JOIN chapter c ON cs.chapter_id=c.id WHERE cs.novel_id=? "+
			"ORDER BY c.chapter_order DESC LIMIT ?",
		novelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id, chapterID uint64
		var chapterTitle, timelinePosition, updatedAt string
		var chapterOrder int
		var summary, keyEvents, characters, locations, plotThreads, foreshadowingChanges, characterChanges sql.NullString
		if err := rows.Scan(&id, &chapterID, &chapterTitle, &chapterOrder, &summary, &keyEvents, &characters,
			&locations, &timelinePosition, &plotThreads, &foreshadowingChanges, &characterChanges, &updatedAt); err != nil {
			return nil, err
		}
		list = append(list, gin.H{
			"id":                    id,
			"chapter_id":            chapterID,
			"chapter_title":         chapterTitle,
			"chapter_order":         chapterOrder,
			"summary":               strOrEmpty(summary),
			"key_events":            strOrEmpty(keyEvents),
			"characters":            strOrEmpty(characters),
			"locations":             strOrEmpty(locations),
			"timeline_position":     timelinePosition,
			"plot_threads":          strOrEmpty(plotThreads),
			"foreshadowing_changes": strOrEmpty(foreshadowingChanges),
			"character_changes":     strOrEmpty(characterChanges),
			"updated_at":            updatedAt,
		})
	}
	return list, rows.Err()
}

func queryCharacterStates(novelID string) ([]gin.H, error) {
	rows, err := database.DB.Query(
		"SELECT cs.id, cs.character_id, c.name, c.alias, cs.current_state, cs.location, cs.goal, "+
			"cs.relationship_summary, cs.ability_state, cs.knowledge_state, cs.last_seen_chapter_id, ch.title, cs.extra, cs.updated_at "+
			"FROM character_state cs JOIN `character` c ON cs.character_id=c.id "+
			"LEFT JOIN chapter ch ON cs.last_seen_chapter_id=ch.id WHERE cs.novel_id=? ORDER BY cs.updated_at DESC",
		novelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id, characterID uint64
		var name, alias, location, updatedAt string
		var currentState, goal, relationship, ability, knowledge, extra, lastChapterTitle sql.NullString
		var lastChapterID sql.NullInt64
		if err := rows.Scan(&id, &characterID, &name, &alias, &currentState, &location, &goal,
			&relationship, &ability, &knowledge, &lastChapterID, &lastChapterTitle, &extra, &updatedAt); err != nil {
			return nil, err
		}
		list = append(list, gin.H{
			"id":                      id,
			"character_id":            characterID,
			"name":                    name,
			"alias":                   alias,
			"current_state":           strOrEmpty(currentState),
			"location":                location,
			"goal":                    strOrEmpty(goal),
			"relationship_summary":    strOrEmpty(relationship),
			"ability_state":           strOrEmpty(ability),
			"knowledge_state":         strOrEmpty(knowledge),
			"last_seen_chapter_id":    uintPtrOrNil(lastChapterID),
			"last_seen_chapter_title": strOrEmpty(lastChapterTitle),
			"extra":                   strOrEmpty(extra),
			"updated_at":              updatedAt,
		})
	}
	return list, rows.Err()
}

func queryPlotMemories(novelID, query string, limit int) ([]gin.H, error) {
	pattern := "%" + query + "%"
	rows, err := database.DB.Query(
		"SELECT pm.id, pm.memory_type, pm.title, pm.content, pm.importance, pm.chapter_id, c.title, "+
			"pm.character_id, ch.name, pm.tags, pm.status, pm.updated_at "+
			"FROM plot_memory pm LEFT JOIN chapter c ON pm.chapter_id=c.id LEFT JOIN `character` ch ON pm.character_id=ch.id "+
			"WHERE pm.novel_id=? AND (?='' OR pm.title LIKE ? OR pm.content LIKE ? OR pm.memory_type LIKE ?) "+
			"ORDER BY pm.status ASC, pm.importance DESC, pm.updated_at DESC LIMIT ?",
		novelID, query, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id uint64
		var memoryType, title, updatedAt string
		var content, chapterTitle, characterName, tags sql.NullString
		var importance, status int
		var chapterID, characterID sql.NullInt64
		if err := rows.Scan(&id, &memoryType, &title, &content, &importance, &chapterID, &chapterTitle,
			&characterID, &characterName, &tags, &status, &updatedAt); err != nil {
			return nil, err
		}
		list = append(list, gin.H{
			"id":             id,
			"memory_type":    memoryType,
			"title":          title,
			"content":        strOrEmpty(content),
			"importance":     importance,
			"chapter_id":     uintPtrOrNil(chapterID),
			"chapter_title":  strOrEmpty(chapterTitle),
			"character_id":   uintPtrOrNil(characterID),
			"character_name": strOrEmpty(characterName),
			"tags":           strOrEmpty(tags),
			"status":         status,
			"updated_at":     updatedAt,
		})
	}
	return list, rows.Err()
}

func queryTimelineEvents(novelID string, limit int) ([]gin.H, error) {
	rows, err := database.DB.Query(
		"SELECT nt.id, nt.chapter_id, c.title, nt.sequence_no, nt.event_time, nt.title, nt.content, nt.importance, nt.updated_at "+
			"FROM novel_timeline nt LEFT JOIN chapter c ON nt.chapter_id=c.id WHERE nt.novel_id=? "+
			"ORDER BY nt.sequence_no ASC, nt.id ASC LIMIT ?",
		novelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id uint64
		var chapterID sql.NullInt64
		var chapterTitle, content sql.NullString
		var sequenceNo, importance int
		var eventTime, title, updatedAt string
		if err := rows.Scan(&id, &chapterID, &chapterTitle, &sequenceNo, &eventTime, &title, &content, &importance, &updatedAt); err != nil {
			return nil, err
		}
		list = append(list, gin.H{
			"id":            id,
			"chapter_id":    uintPtrOrNil(chapterID),
			"chapter_title": strOrEmpty(chapterTitle),
			"sequence_no":   sequenceNo,
			"event_time":    eventTime,
			"title":         title,
			"content":       strOrEmpty(content),
			"importance":    importance,
			"updated_at":    updatedAt,
		})
	}
	return list, rows.Err()
}
