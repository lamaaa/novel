-- 小说长期记忆层表结构
-- 请先 USE novel_service; 再执行本文件。

CREATE TABLE IF NOT EXISTS chapter_summary (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节ID',
    summary TEXT COMMENT '章节摘要',
    key_events JSON COMMENT '关键事件列表',
    characters JSON COMMENT '登场/相关人物',
    locations JSON COMMENT '地点列表',
    timeline_position VARCHAR(255) DEFAULT '' COMMENT '时间线位置',
    plot_threads JSON COMMENT '剧情线进展',
    foreshadowing_changes JSON COMMENT '伏笔变化',
    character_changes JSON COMMENT '人物状态变化',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_chapter (chapter_id),
    KEY idx_novel_chapter (novel_id, chapter_id),
    CONSTRAINT fk_chapter_summary_novel FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE,
    CONSTRAINT fk_chapter_summary_chapter FOREIGN KEY (chapter_id) REFERENCES chapter(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节记忆摘要表';

CREATE TABLE IF NOT EXISTS character_state (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    character_id BIGINT UNSIGNED NOT NULL COMMENT '人物ID',
    current_state TEXT COMMENT '当前状态总述',
    location VARCHAR(255) DEFAULT '' COMMENT '当前位置',
    goal TEXT COMMENT '当前目标/动机',
    relationship_summary TEXT COMMENT '关系状态摘要',
    ability_state TEXT COMMENT '能力/等级/装备状态',
    knowledge_state TEXT COMMENT '该人物已知/未知信息',
    last_seen_chapter_id BIGINT UNSIGNED DEFAULT NULL COMMENT '最后出场章节',
    extra JSON COMMENT '扩展状态JSON',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_character (character_id),
    KEY idx_novel (novel_id),
    CONSTRAINT fk_character_state_novel FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE,
    CONSTRAINT fk_character_state_character FOREIGN KEY (character_id) REFERENCES `character`(id) ON DELETE CASCADE,
    CONSTRAINT fk_character_state_chapter FOREIGN KEY (last_seen_chapter_id) REFERENCES chapter(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人物当前状态表';

CREATE TABLE IF NOT EXISTS plot_memory (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    memory_type VARCHAR(50) NOT NULL DEFAULT 'fact' COMMENT '记忆类型: fact, secret, clue, rule, relationship, conflict, plan',
    title VARCHAR(255) NOT NULL COMMENT '记忆标题',
    content TEXT COMMENT '记忆内容',
    importance TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '重要程度1-5',
    chapter_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联章节',
    character_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联人物',
    tags JSON COMMENT '标签',
    status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0=有效 1=已过期 2=存疑',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_novel_type (novel_id, memory_type),
    KEY idx_novel_importance (novel_id, importance),
    KEY idx_status (status),
    CONSTRAINT fk_plot_memory_novel FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE,
    CONSTRAINT fk_plot_memory_chapter FOREIGN KEY (chapter_id) REFERENCES chapter(id) ON DELETE SET NULL,
    CONSTRAINT fk_plot_memory_character FOREIGN KEY (character_id) REFERENCES `character`(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='剧情事实记忆表';

CREATE TABLE IF NOT EXISTS novel_timeline (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    chapter_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联章节',
    sequence_no INT NOT NULL DEFAULT 0 COMMENT '时间线排序',
    event_time VARCHAR(255) DEFAULT '' COMMENT '故事内时间',
    title VARCHAR(255) NOT NULL COMMENT '事件标题',
    content TEXT COMMENT '事件内容',
    importance TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '重要程度1-5',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_novel_sequence (novel_id, sequence_no),
    KEY idx_novel_importance (novel_id, importance),
    CONSTRAINT fk_novel_timeline_novel FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE,
    CONSTRAINT fk_novel_timeline_chapter FOREIGN KEY (chapter_id) REFERENCES chapter(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='小说时间线表';
