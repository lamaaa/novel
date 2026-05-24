-- 小说服务数据库建表语句

CREATE DATABASE IF NOT EXISTS novel_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE novel_service;

-- 小说表
CREATE TABLE IF NOT EXISTS novel (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL COMMENT '小说标题',
    author VARCHAR(255) NOT NULL DEFAULT '' COMMENT '作者',
    description TEXT COMMENT '简介',
    cover_url VARCHAR(512) DEFAULT '' COMMENT '封面图URL',
    status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0=连载中 1=已完结',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='小说表';

-- 章节表
CREATE TABLE IF NOT EXISTS chapter (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    title VARCHAR(255) NOT NULL COMMENT '章节标题',
    content LONGTEXT COMMENT '章节内容',
    word_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '字数',
    chapter_order INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '章节排序(从1开始)',
    status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0=草稿 1=已发布',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_novel_id (novel_id),
    INDEX idx_novel_order (novel_id, chapter_order),
    FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节表';

-- 人物表
CREATE TABLE IF NOT EXISTS `character` (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    name VARCHAR(255) NOT NULL COMMENT '姓名',
    alias VARCHAR(255) DEFAULT '' COMMENT '别名/外号',
    avatar_url VARCHAR(512) DEFAULT '' COMMENT '头像URL',
    gender TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '性别: 0=未知 1=男 2=女',
    age VARCHAR(50) DEFAULT '' COMMENT '年龄',
    description TEXT COMMENT '人物描述',
    personality TEXT COMMENT '性格特点',
    background TEXT COMMENT '背景故事',
    character_order INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_novel_id (novel_id),
    FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人物表';

-- 世界观设定表
CREATE TABLE IF NOT EXISTS worldview (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    category VARCHAR(100) NOT NULL DEFAULT '其他' COMMENT '分类(如:地理,历史,种族,魔法体系,势力,规则等)',
    title VARCHAR(255) NOT NULL COMMENT '标题',
    content TEXT COMMENT '内容',
    sort_order INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_novel_id (novel_id),
    INDEX idx_category (category),
    FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='世界观设定表';

-- 伏笔设定表
CREATE TABLE IF NOT EXISTS foreshadowing (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    novel_id BIGINT UNSIGNED NOT NULL COMMENT '小说ID',
    title VARCHAR(255) NOT NULL COMMENT '伏笔标题',
    description TEXT COMMENT '伏笔描述',
    planted_chapter_id BIGINT UNSIGNED DEFAULT NULL COMMENT '埋设章节ID',
    resolved_chapter_id BIGINT UNSIGNED DEFAULT NULL COMMENT '回收章节ID',
    status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态: 0=已埋设 1=已回收 2=已放弃',
    importance TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '重要程度: 1-5',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_novel_id (novel_id),
    INDEX idx_status (status),
    FOREIGN KEY (novel_id) REFERENCES novel(id) ON DELETE CASCADE,
    FOREIGN KEY (planted_chapter_id) REFERENCES chapter(id) ON DELETE SET NULL,
    FOREIGN KEY (resolved_chapter_id) REFERENCES chapter(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='伏笔设定表';

-- 内容版本表 (通用的版本快照，支持章节/人物/世界观/伏笔的版本管理)
CREATE TABLE IF NOT EXISTS content_version (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    entity_type VARCHAR(50) NOT NULL COMMENT '实体类型: novel, chapter, character, worldview, foreshadowing',
    entity_id BIGINT UNSIGNED NOT NULL COMMENT '实体ID',
    version INT UNSIGNED NOT NULL COMMENT '版本号(从1开始递增)',
    snapshot JSON NOT NULL COMMENT '快照数据(更新前的完整数据JSON)',
    change_summary VARCHAR(255) DEFAULT '' COMMENT '变更摘要',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_entity (entity_type, entity_id),
    UNIQUE INDEX idx_entity_version (entity_type, entity_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容版本表';
