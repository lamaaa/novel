package mcp

import (
	"novel-service/mcp/tools"
	"novel-service/mcp/types"
)

// getTools returns all MCP tools exposed by this server
func getTools() []Tool {
	return []Tool{
		// ============================================================
		// Novel
		// ============================================================
		{
			Name:        "list_novels",
			Description: "获取小说列表。支持按关键词搜索标题或作者，按连载状态筛选，分页返回。返回结果包含每本小说的 id、标题、作者、简介、封面、状态、章节数等信息。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"keyword":   {Type: "string", Description: "搜索关键词，模糊匹配小说标题或作者名。不传则返回全部", Examples: []interface{}{"斗破", "番茄"}},
					"status":    {Type: "integer", Description: "按连载状态筛选：0=连载中，1=已完结，-1=不筛选（返回全部）。默认 -1", Default: -1, Examples: []interface{}{0, 1, -1}},
					"page":      {Type: "integer", Description: "页码，从1开始。默认 1", Default: 1, Examples: []interface{}{1, 2, 3}},
					"page_size": {Type: "integer", Description: "每页返回数量，最大100。默认 20", Default: 20, Examples: []interface{}{10, 20, 50}},
				},
			},
		},
		{
			Name:        "get_novel",
			Description: "根据小说ID获取单本小说的详细信息，包括标题、作者、简介、封面URL、连载状态、已发布章节数、创建和更新时间。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID（正整数），可通过 list_novels 获取", Examples: []interface{}{1, 42}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "create_novel",
			Description: "创建一本新小说。只需提供标题即可创建，其他字段可选。创建成功后返回新小说的ID。示例：{\"title\":\"斗破苍穹\",\"author\":\"天蚕土豆\",\"description\":\"三十年河东三十年河西\",\"status\":0}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"title":       {Type: "string", Description: "小说标题（必填），最长200字符", Examples: []interface{}{"斗破苍穹", "诡秘之主"}},
					"author":      {Type: "string", Description: "作者名，不填则默认为空", Examples: []interface{}{"天蚕土豆", "爱潜水的乌贼"}},
					"description": {Type: "string", Description: "小说简介/大纲，建议200字以内概述核心设定和故事线", Examples: []interface{}{"三十年河东三十年河西，莫欺少年穷！一个天赋尽失的少年，在绝境中获得了传说中的功法……"}},
					"cover_url":   {Type: "string", Description: "封面图片URL，需为完整的HTTP链接", Examples: []interface{}{"https://example.com/cover.jpg"}},
					"status":      {Type: "integer", Description: "连载状态：0=连载中（默认），1=已完结。新建小说通常设为0", Default: 0, Examples: []interface{}{0, 1}},
				},
				Required: []string{"title"},
			},
		},
		{
			Name:        "update_novel",
			Description: "更新小说信息。只需传入要修改的字段，未传的字段保持不变。修改前会自动保存当前数据为版本快照（可通过 rollback_version 回退）。示例：{\"novel_id\":1,\"status\":1} 表示将小说标记为已完结",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":    {Type: "integer", Description: "小说ID（必填，指定要更新的小说）", Examples: []interface{}{1}},
					"title":       {Type: "string", Description: "新标题，不传则不修改", Examples: []interface{}{"斗破苍穹（修订版）"}},
					"author":      {Type: "string", Description: "新作者名，不传则不修改", Examples: []interface{}{"天蚕土豆", "猫腻"}},
					"description": {Type: "string", Description: "新简介，不传则不修改", Examples: []interface{}{"一个少年的修仙之路…"}},
					"cover_url":   {Type: "string", Description: "新封面URL，不传则不修改", Examples: []interface{}{"https://example.com/new-cover.jpg"}},
					"status":      {Type: "integer", Description: "新的连载状态：0=连载中，1=已完结。小说完结时设为1。不传则不修改", Examples: []interface{}{0, 1}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "delete_novel",
			Description: "删除一本小说及其所有关联数据（包括所有章节、人物、世界观设定、伏笔），此操作不可恢复，请谨慎调用。建议删除前先用 list_versions 确认无需保留。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "要删除的小说ID", Examples: []interface{}{1}},
				},
				Required: []string{"novel_id"},
			},
		},

		// ============================================================
		// Chapter
		// ============================================================
		{
			Name:        "list_chapters",
			Description: "获取某本小说的全部章节列表，按章节序号升序排列。返回章节的 id、标题、字数、序号、状态（草稿/已发布）等，不含正文内容。如需正文请用 get_chapter。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "get_chapter",
			Description: "根据章节ID获取章节详情，包含完整的正文内容（content字段），以及字数统计、序号、发布状态等。用于阅读章节内容或编辑前获取当前内容。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"chapter_id": {Type: "integer", Description: "章节ID（正整数），可通过 list_chapters 获取", Examples: []interface{}{1, 10}},
				},
				Required: []string{"chapter_id"},
			},
		},
		{
			Name:        "create_chapter",
			Description: "为小说添加新章节。chapter_order 设为0或不传时自动递增（接在最后一章之后）；字数根据 content 自动计算。新建章节默认为草稿状态(status=0)，内容完善后再改为已发布(status=1)。示例：{\"novel_id\":1,\"title\":\"第一章 陨落的天才\",\"content\":\"那一年……\",\"status\":0}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":      {Type: "integer", Description: "小说ID（必填，指定章节归属哪本小说）", Examples: []interface{}{1}},
					"title":         {Type: "string", Description: "章节标题（必填），如\"第一章 陨落的天才\"", Examples: []interface{}{"第一章 陨落的天才", "第三十二章 决战"}},
					"content":       {Type: "string", Description: "章节正文内容，支持换行。字数会自动根据此字段计算", Examples: []interface{}{"那一年，萧炎十五岁。\n\n斗之气，三段！\n\n……"}},
					"chapter_order": {Type: "integer", Description: "章节排序序号（正整数）。设为0或不传则自动排在最后一章之后。如果要在中间插入章节，可指定具体序号如3表示排第3章。默认 0（自动）", Default: 0, Examples: []interface{}{0, 1, 5}},
					"status":        {Type: "integer", Description: "发布状态：0=草稿（默认，仅保存不对外展示），1=已发布（正式可见）。建议写作时先设为0，完成后再改为1", Default: 0, Examples: []interface{}{0, 1}},
				},
				Required: []string{"novel_id", "title"},
			},
		},
		{
			Name:        "update_chapter",
			Description: "更新章节内容。只需传入要修改的字段，未传的字段保持不变。修改前自动保存版本快照。如果修改了 content，字数会自动重新计算。典型用法：1) 编辑正文内容 2) 将草稿改为已发布 3) 调整章节顺序。示例：{\"chapter_id\":5,\"status\":1} 表示发布该章节",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"chapter_id":    {Type: "integer", Description: "章节ID（必填，指定要更新的章节）", Examples: []interface{}{5}},
					"title":         {Type: "string", Description: "新章节标题，不传则不修改", Examples: []interface{}{"第一章 陨落的天才（修订）", "终章 归来"}},
					"content":       {Type: "string", Description: "新正文内容，修改后字数自动重算。不传则不修改", Examples: []interface{}{"那一年，他十五岁。\n\n斗之气，三段！\n\n望着测验魔石碑上那刺眼的五个大字…"}},
					"chapter_order": {Type: "integer", Description: "新的章节序号，用于调整章节顺序。不传则不修改", Examples: []interface{}{1, 3, 10}},
					"status":        {Type: "integer", Description: "发布状态：0=草稿，1=已发布。不传则不修改。写作完成后改为1", Examples: []interface{}{0, 1}},
				},
				Required: []string{"chapter_id"},
			},
		},
		{
			Name:        "delete_chapter",
			Description: "删除一个章节，此操作不可恢复。删除后其他章节的序号不会自动调整，如需调整请手动 update_chapter 的 chapter_order。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"chapter_id": {Type: "integer", Description: "要删除的章节ID", Examples: []interface{}{5}},
				},
				Required: []string{"chapter_id"},
			},
		},

		// ============================================================
		// Character
		// ============================================================
		{
			Name:        "list_characters",
			Description: "获取某本小说的全部人物列表，按人物排序号升序排列。返回人物的姓名、别名、性别、年龄、描述、性格、背景等信息。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "get_character",
			Description: "根据人物ID获取人物详情，包含完整的描述、性格特点、背景故事等。用于查看人物设定或编辑前获取当前数据。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"character_id": {Type: "integer", Description: "人物ID（正整数），可通过 list_characters 获取", Examples: []interface{}{1, 3}},
				},
				Required: []string{"character_id"},
			},
		},
		{
			Name:        "create_character",
			Description: "为小说创建新人物。只需提供姓名即可，其他信息可选但建议尽量完善，越详细AI写作时越能保持人物一致性。示例：{\"novel_id\":1,\"name\":\"萧炎\",\"alias\":\"炎帝\",\"gender\":1,\"age\":\"16\",\"description\":\"乌坦城萧家子弟，曾为天才后沦为废柴\",\"personality\":\"坚韧不屈、重情重义\",\"background\":\"十一岁成为斗者，此后修为倒退至斗之气三段\"}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":    {Type: "integer", Description: "小说ID（必填，指定人物归属哪本小说）", Examples: []interface{}{1}},
					"name":        {Type: "string", Description: "人物姓名（必填）", Examples: []interface{}{"萧炎", "林动", "克莱恩"}},
					"alias":       {Type: "string", Description: "别名/外号/称号，如角色有多个身份时可填写", Examples: []interface{}{"炎帝", "愚者", "暗影刺客"}},
					"avatar_url":  {Type: "string", Description: "人物头像图片URL，需为完整HTTP链接", Examples: []interface{}{"https://example.com/avatar.jpg"}},
					"gender":      {Type: "integer", Description: "性别：0=未知/未设定，1=男，2=女。默认 0", Default: 0, Examples: []interface{}{0, 1, 2}},
					"age":         {Type: "string", Description: "年龄或年龄段，字符串类型支持灵活描述", Examples: []interface{}{"16", "约20岁", "未知（外表约25）"}},
					"description": {Type: "string", Description: "人物外貌和身份描述，包括长相、穿着、身份等", Examples: []interface{}{"黑发黑眸，身材修长，常穿黑色长袍。乌坦城萧家三少爷"}},
					"personality": {Type: "string", Description: "性格特点描述，包括性格优缺点、行为模式等", Examples: []interface{}{"坚韧不屈、重情重义、偶尔冲动。面对强敌从不退缩，但容易为亲友涉险"}},
					"background":  {Type: "string", Description: "人物背景故事，包括身世、经历、动机等", Examples: []interface{}{"十一岁突破斗者，被誉为天才。此后三年修为倒退至斗之气三段，遭受冷眼。实因药老寄宿导致修为无法凝聚"}},
				},
				Required: []string{"novel_id", "name"},
			},
		},
		{
			Name:        "update_character",
			Description: "更新人物设定。只需传入要修改的字段，未传的字段保持不变。修改前自动保存版本快照。典型场景：完善人物性格描述、更新人物状态变化（如年龄增长）。示例：{\"character_id\":1,\"age\":\"18\",\"background\":\"十六岁恢复修为，开始修炼焚诀…\"}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"character_id": {Type: "integer", Description: "人物ID（必填，指定要更新的人物）", Examples: []interface{}{1}},
					"name":         {Type: "string", Description: "新姓名，不传则不修改", Examples: []interface{}{"萧炎", "克莱恩·莫雷蒂"}},
					"alias":        {Type: "string", Description: "新别名/外号，不传则不修改", Examples: []interface{}{"炎帝", "愚者"}},
					"avatar_url":   {Type: "string", Description: "新头像URL，不传则不修改", Examples: []interface{}{"https://example.com/new-avatar.jpg"}},
					"gender":       {Type: "integer", Description: "性别：0=未知，1=男，2=女。不传则不修改", Examples: []interface{}{0, 1, 2}},
					"age":          {Type: "string", Description: "新年龄，不传则不修改", Examples: []interface{}{"18", "约25岁"}},
					"description":  {Type: "string", Description: "新人物描述，不传则不修改", Examples: []interface{}{"黑发黑眸，气质沉稳。已成为斗皇强者"}},
					"personality":  {Type: "string", Description: "新性格特点，不传则不修改", Examples: []interface{}{"成熟稳重、不再冲动"}},
					"background":   {Type: "string", Description: "新背景故事，不传则不修改", Examples: []interface{}{"经过三年历练，从废柴成长为斗皇"}},
				},
				Required: []string{"character_id"},
			},
		},
		{
			Name:        "delete_character",
			Description: "删除一个人物，此操作不可恢复。如果伏笔中引用了该人物，请先确认伏笔是否需要调整。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"character_id": {Type: "integer", Description: "要删除的人物ID", Examples: []interface{}{1}},
				},
				Required: []string{"character_id"},
			},
		},

		// ============================================================
		// Worldview
		// ============================================================
		{
			Name:        "list_worldviews",
			Description: "获取某本小说的世界观设定列表，按分类和排序返回。可按分类筛选（如只看\"魔法体系\"分类）。同时返回该小说已有的所有分类列表，方便了解世界观的整体结构。常见分类：地理、历史、种族、魔法体系、势力、政治、经济、宗教等。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"category": {Type: "string", Description: "按分类筛选，不传则返回全部分类。常见分类：地理、历史、种族、魔法体系、势力、政治、经济、宗教、科技、文化", Examples: []interface{}{"魔法体系", "势力", "种族"}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "get_worldview",
			Description: "根据ID获取单条世界观设定的详情，包含分类、标题和完整内容。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"worldview_id": {Type: "integer", Description: "世界观设定ID（正整数），可通过 list_worldviews 获取", Examples: []interface{}{1, 5}},
				},
				Required: []string{"worldview_id"},
			},
		},
		{
			Name:        "create_worldview",
			Description: "为小说添加世界观设定。每条设定属于一个分类(category)，同一分类下按 sort_order 排序。分类可自定义，建议使用有意义的分类名以保持结构清晰。示例：{\"novel_id\":1,\"category\":\"魔法体系\",\"title\":\"斗气等级\",\"content\":\"斗之气→斗者→斗师→大斗师→斗灵→斗王→斗皇→斗宗→斗尊→斗圣→斗帝\"}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":   {Type: "integer", Description: "小说ID（必填）", Examples: []interface{}{1}},
					"category":   {Type: "string", Description: "分类名称（必填），建议使用规范分类名保持一致性。常见分类：地理、历史、种族、魔法体系、势力、政治、经济、宗教、科技、文化。默认\"其他\"", Default: "其他", Examples: []interface{}{"魔法体系", "势力", "种族", "地理", "历史"}},
					"title":      {Type: "string", Description: "设定标题（必填），简明扼要描述这条设定是什么", Examples: []interface{}{"斗气等级", "大陆地图", "主要势力"}},
					"content":    {Type: "string", Description: "设定内容，详细描述世界观的具体规则、设定或背景。支持换行，可写长文", Examples: []interface{}{"斗之气→斗者→斗师→大斗师→斗灵→斗王→斗皇→斗宗→斗尊→斗圣→斗帝\n\n每个大等级之间实力差距巨大，每突破一个大等级都是质的飞跃。"}},
					"sort_order": {Type: "integer", Description: "分类内排序序号。0或不传则自动排在该分类末尾。默认 0（自动）", Default: 0, Examples: []interface{}{0, 1, 2}},
				},
				Required: []string{"novel_id", "category", "title"},
			},
		},
		{
			Name:        "update_worldview",
			Description: "更新世界观设定。只需传入要修改的字段，未传的字段保持不变。修改前自动保存版本快照。示例：{\"worldview_id\":3,\"content\":\"更新后的斗气体系说明…\"}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"worldview_id": {Type: "integer", Description: "世界观设定ID（必填，指定要更新的设定）", Examples: []interface{}{3}},
					"category":     {Type: "string", Description: "新分类名，可用于将设定移到其他分类。不传则不修改", Examples: []interface{}{"魔法体系", "历史"}},
					"title":        {Type: "string", Description: "新标题，不传则不修改", Examples: []interface{}{"斗气等级（修订版）", "大陆势力分布"}},
					"content":      {Type: "string", Description: "新内容，不传则不修改", Examples: []interface{}{"斗之气→斗者→斗师…（更新后的完整体系）"}},
					"sort_order":   {Type: "integer", Description: "新排序序号，不传则不修改", Examples: []interface{}{0, 1, 3}},
				},
				Required: []string{"worldview_id"},
			},
		},
		{
			Name:        "delete_worldview",
			Description: "删除一条世界观设定，此操作不可恢复。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"worldview_id": {Type: "integer", Description: "要删除的世界观设定ID", Examples: []interface{}{3}},
				},
				Required: []string{"worldview_id"},
			},
		},

		// ============================================================
		// Foreshadowing
		// ============================================================
		{
			Name:        "list_foreshadowings",
			Description: "获取某本小说的伏笔列表，默认按重要程度降序排列。返回伏笔的标题、描述、埋设/回收章节、状态、重要程度等。可通过 status 筛选特定状态的伏笔，例如只看未回收的伏笔(status=0)。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"status":   {Type: "integer", Description: "按伏笔状态筛选：0=已埋设（尚未回收），1=已回收（已在后续章节揭示），2=已放弃（不再使用该伏笔）。不传则返回全部", Examples: []interface{}{0, 1, 2}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "get_foreshadowing",
			Description: "根据ID获取单条伏笔的详情，包含描述、埋设章节、回收章节（如有）、状态、重要程度等。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"foreshadowing_id": {Type: "integer", Description: "伏笔ID（正整数），可通过 list_foreshadowings 获取", Examples: []interface{}{1, 5}},
				},
				Required: []string{"foreshadowing_id"},
			},
		},
		{
			Name:        "create_foreshadowing",
			Description: "为小说添加一条伏笔。伏笔用于规划和追踪小说中的悬念、暗示和前后呼应。planted_chapter_id 表示伏笔埋设在哪一章，resolved_chapter_id 表示伏笔在哪一章回收（揭示）。新伏笔默认状态为\"已埋设\"(status=0)。示例：{\"novel_id\":1,\"title\":\"神秘戒指的真正用途\",\"description\":\"药老寄宿的戒指实际上是远古强者的传承之物\",\"planted_chapter_id\":1,\"importance\":5}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":            {Type: "integer", Description: "小说ID（必填）", Examples: []interface{}{1}},
					"title":               {Type: "string", Description: "伏笔标题（必填），简明描述这个伏笔是什么", Examples: []interface{}{"神秘戒指的真正用途", "萧家灭门真相", "药老的过去"}},
					"description":         {Type: "string", Description: "伏笔的详细描述，包括你打算如何埋设和回收这个伏笔", Examples: []interface{}{"药老寄宿的戒指实际上是远古强者陀舍古帝的传承之物，第三章首次出现时只是普通戒指"}},
					"planted_chapter_id":  {Type: "integer", Description: "伏笔埋设的章节ID，即这个伏笔首次出现/暗示的章节。可通过 list_chapters 获取章节ID。不传则不指定", Examples: []interface{}{1, 3}},
					"resolved_chapter_id": {Type: "integer", Description: "伏笔回收的章节ID，即伏笔被揭示/解答的章节。新伏笔通常不填此字段，等写作到回收章节时再通过 update_foreshadowing 补充。不传则不指定", Examples: []interface{}{50, 100}},
					"status":              {Type: "integer", Description: "伏笔状态：0=已埋设（已埋下伏笔，尚未回收，默认），1=已回收（伏笔已在后续章节揭示），2=已放弃（决定不再使用此伏笔）。新建伏笔通常设为0", Default: 0, Examples: []interface{}{0, 1, 2}},
					"importance":          {Type: "integer", Description: "重要程度（1-5），5最重要。1=可有可无的细节，3=推动支线剧情，5=影响主线走向的核心伏笔。默认 3", Default: 3, Examples: []interface{}{1, 3, 5}},
				},
				Required: []string{"novel_id", "title"},
			},
		},
		{
			Name:        "update_foreshadowing",
			Description: "更新伏笔信息。只需传入要修改的字段，未传的字段保持不变。修改前自动保存版本快照。典型用法：1) 写到回收章节时设置 resolved_chapter_id 和 status=1 2) 调整伏笔重要程度 3) 放弃某条伏笔设 status=2。示例：{\"foreshadowing_id\":2,\"resolved_chapter_id\":50,\"status\":1} 表示在第50章回收了该伏笔",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"foreshadowing_id":    {Type: "integer", Description: "伏笔ID（必填，指定要更新的伏笔）", Examples: []interface{}{2}},
					"title":               {Type: "string", Description: "新标题，不传则不修改", Examples: []interface{}{"神秘戒指的真正来历（更新）", "萧家灭门幕后黑手"}},
					"description":         {Type: "string", Description: "新描述，不传则不修改", Examples: []interface{}{"药老的戒指其实是陀舍古帝的传承之物，蕴含焚诀功法"}},
					"planted_chapter_id":  {Type: "integer", Description: "新的埋设章节ID，不传则不修改。传0或负数表示取消指定", Examples: []interface{}{1, 3}},
					"resolved_chapter_id": {Type: "integer", Description: "新的回收章节ID，不传则不修改。传0或负数表示取消指定。通常在写到回收章节时设置", Examples: []interface{}{50, 100}},
					"status":              {Type: "integer", Description: "新状态：0=已埋设，1=已回收，2=已放弃。不传则不修改。伏笔回收时设为1", Examples: []interface{}{0, 1, 2}},
					"importance":          {Type: "integer", Description: "新的重要程度（1-5），不传则不修改", Examples: []interface{}{1, 3, 5}},
				},
				Required: []string{"foreshadowing_id"},
			},
		},
		{
			Name:        "delete_foreshadowing",
			Description: "删除一条伏笔，此操作不可恢复。建议优先考虑将伏笔状态改为\"已放弃\"(status=2)而非删除，以保留规划记录。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"foreshadowing_id": {Type: "integer", Description: "要删除的伏笔ID", Examples: []interface{}{2}},
				},
				Required: []string{"foreshadowing_id"},
			},
		},

		// ============================================================
		// Version
		// ============================================================
		{
			Name:        "list_versions",
			Description: "获取某个实体的版本历史列表。所有支持版本管理的实体（小说、章节、人物、世界观、伏笔）在每次更新前都会自动保存当前数据为版本快照。版本按版本号降序排列（最新的在前）。返回每个版本的ID、版本号、变更摘要、创建时间等。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"entity_type": {Type: "string", Description: "实体类型（必填），可选值：\"novel\"=小说, \"chapter\"=章节, \"character\"=人物, \"worldview\"=世界观, \"foreshadowing\"=伏笔", Enum: []string{"novel", "chapter", "character", "worldview", "foreshadowing"}, Examples: []interface{}{"chapter", "novel"}},
					"entity_id":   {Type: "integer", Description: "实体ID（必填），如章节ID、小说ID等", Examples: []interface{}{1, 5}},
				},
				Required: []string{"entity_type", "entity_id"},
			},
		},
		{
			Name:        "get_version",
			Description: "获取某个历史版本的详情，包含完整的快照数据（snapshot字段为JSON字符串，包含该版本时实体的全部字段值）。用于查看历史版本的具体内容，对比变更。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"version_id": {Type: "integer", Description: "版本ID（正整数），可通过 list_versions 获取", Examples: []interface{}{1, 10}},
				},
				Required: []string{"version_id"},
			},
		},
		{
			Name:        "rollback_version",
			Description: "回退到指定版本。回退前会自动将当前状态保存为新的版本快照，因此不会丢失数据。回退操作的本质是：1) 保存当前数据为新版本 2) 用目标版本的快照数据覆盖当前数据。示例：{\"version_id\":5,\"change_summary\":\"第5版内容更好，回退\"}",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"version_id":     {Type: "integer", Description: "要回退到的版本ID（必填）。该版本的快照数据将覆盖当前数据", Examples: []interface{}{3, 5}},
					"change_summary": {Type: "string", Description: "回退说明，记录为什么要回退，便于日后查看。默认\"手动回退\"", Default: "手动回退", Examples: []interface{}{"上一版内容更好", "误操作回退", "编辑方向调整"}},
				},
				Required: []string{"version_id"},
			},
		},

		// ============================================================
		// Novel Memory
		// ============================================================
		{
			Name:        "get_novel_context",
			Description: "获取写作前的小说长期记忆上下文，聚合小说基本信息、最近章节摘要、人物当前状态、未回收伏笔、时间线和相关记忆。写新章或审校前建议优先调用。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id":     {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"recent_limit": {Type: "integer", Description: "返回最近章节摘要数量，默认5，最大50", Default: 5, Examples: []interface{}{5, 10}},
					"query":        {Type: "string", Description: "可选检索词，用于召回相关剧情记忆、人物状态、时间线事件", Examples: []interface{}{"神秘戒指", "女主知道真相了吗"}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "update_chapter_summary",
			Description: "创建或更新章节长期记忆摘要。写完/改完一章后调用，用于记录本章概要、关键事件、人物变化、伏笔变化和剧情线进展。JSON字段可传数组/对象或JSON字符串。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"chapter_id":            {Type: "integer", Description: "章节ID", Examples: []interface{}{12}},
					"novel_id":              {Type: "integer", Description: "小说ID；不传时根据章节自动推断", Examples: []interface{}{1}},
					"summary":               {Type: "string", Description: "本章摘要，建议200-500字", Examples: []interface{}{"萧炎在魔兽山脉遭遇狼群，首次使用新功法脱险。"}},
					"key_events":            {Type: "string", Description: "关键事件JSON数组或对象", Examples: []interface{}{"[\"主角获得线索\",\"反派现身\"]"}},
					"characters":            {Type: "string", Description: "登场/相关人物JSON数组", Examples: []interface{}{"[\"萧炎\",\"药老\"]"}},
					"locations":             {Type: "string", Description: "地点JSON数组", Examples: []interface{}{"[\"魔兽山脉\"]"}},
					"timeline_position":     {Type: "string", Description: "故事内时间线位置", Examples: []interface{}{"离开乌坦城后的第三天夜晚"}},
					"plot_threads":          {Type: "string", Description: "剧情线进展JSON数组或对象", Examples: []interface{}{"{\"主线\":\"获得戒指线索\"}"}},
					"foreshadowing_changes": {Type: "string", Description: "伏笔变化JSON数组或对象", Examples: []interface{}{"[{\"title\":\"神秘戒指\",\"change\":\"再次暗示\"}]"}},
					"character_changes":     {Type: "string", Description: "人物状态变化JSON数组或对象", Examples: []interface{}{"[{\"name\":\"萧炎\",\"change\":\"学会新招式\"}]"}},
				},
				Required: []string{"chapter_id"},
			},
		},
		{
			Name:        "get_recent_chapter_summaries",
			Description: "获取最近若干章的长期记忆摘要，按章节倒序返回。适合写下一章前快速回顾前文，而不是读取大量完整正文。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"limit":    {Type: "integer", Description: "返回数量，默认10，最大50", Default: 10, Examples: []interface{}{5, 10, 20}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "update_character_state",
			Description: "创建或更新人物当前状态。写完章节后用于维护人物位置、目标、关系、能力、已知信息和最后出场章节。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"character_id":         {Type: "integer", Description: "人物ID", Examples: []interface{}{1}},
					"novel_id":             {Type: "integer", Description: "小说ID；不传时根据人物自动推断", Examples: []interface{}{1}},
					"current_state":        {Type: "string", Description: "当前状态总述", Examples: []interface{}{"刚突破斗者，身体疲惫但信心增强。"}},
					"location":             {Type: "string", Description: "当前位置", Examples: []interface{}{"魔兽山脉外围"}},
					"goal":                 {Type: "string", Description: "当前目标/动机", Examples: []interface{}{"寻找药材并隐藏实力"}},
					"relationship_summary": {Type: "string", Description: "重要关系变化摘要", Examples: []interface{}{"对药老更加信任，但仍保留戒心。"}},
					"ability_state":        {Type: "string", Description: "能力、等级、装备状态", Examples: []interface{}{"斗者一星，掌握八极崩雏形。"}},
					"knowledge_state":      {Type: "string", Description: "该人物已知/未知信息", Examples: []interface{}{"知道戒指与药老有关，不知道萧家危机将至。"}},
					"last_seen_chapter_id": {Type: "integer", Description: "最后出场章节ID", Examples: []interface{}{12}},
					"extra":                {Type: "string", Description: "扩展JSON对象", Examples: []interface{}{"{\"health\":\"轻伤\"}"}},
				},
				Required: []string{"character_id"},
			},
		},
		{
			Name:        "get_character_current_state",
			Description: "获取人物当前状态，包含位置、目标、关系、能力、已知信息和最后出场章节。写涉及该人物的情节前建议调用。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"character_id": {Type: "integer", Description: "人物ID", Examples: []interface{}{1}},
				},
				Required: []string{"character_id"},
			},
		},
		{
			Name:        "upsert_plot_memory",
			Description: "创建或更新剧情事实记忆，用于记录跨章节重要事实、秘密、线索、规则、关系、冲突和后续计划。memory_id 存在时更新，不传则创建。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"memory_id":    {Type: "integer", Description: "剧情记忆ID；传入则更新，不传则创建", Examples: []interface{}{3}},
					"novel_id":     {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"memory_type":  {Type: "string", Description: "记忆类型：fact/secret/clue/rule/relationship/conflict/plan，默认fact", Default: "fact", Examples: []interface{}{"fact", "secret", "clue"}},
					"title":        {Type: "string", Description: "记忆标题", Examples: []interface{}{"女主尚不知道戒指真相"}},
					"content":      {Type: "string", Description: "记忆内容", Examples: []interface{}{"第12章主角发现戒指秘密，但女主没有在场，后续不能让她直接知道。"}},
					"importance":   {Type: "integer", Description: "重要程度1-5，默认3", Default: 3, Examples: []interface{}{3, 5}},
					"chapter_id":   {Type: "integer", Description: "关联章节ID", Examples: []interface{}{12}},
					"character_id": {Type: "integer", Description: "关联人物ID", Examples: []interface{}{2}},
					"tags":         {Type: "string", Description: "标签JSON数组", Examples: []interface{}{"[\"戒指\",\"信息差\"]"}},
					"status":       {Type: "integer", Description: "状态：0=有效，1=已过期，2=存疑。默认0", Default: 0, Examples: []interface{}{0, 1, 2}},
				},
				Required: []string{"novel_id", "title"},
			},
		},
		{
			Name:        "search_novel_memory",
			Description: "检索小说长期记忆，覆盖剧情事实、章节摘要、人物状态和时间线。适合查找长篇前文细节，避免遗忘和矛盾。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"query":    {Type: "string", Description: "检索词；为空时返回重要/近期记忆", Examples: []interface{}{"戒指真相", "女主知道什么"}},
					"limit":    {Type: "integer", Description: "最多返回结果数，默认10，最大30", Default: 10, Examples: []interface{}{10}},
				},
				Required: []string{"novel_id"},
			},
		},
		{
			Name:        "search_chapters",
			Description: "按关键词检索章节正文和标题，返回命中的章节片段，不返回完整正文。用于需要追溯原文细节时定位章节。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"query":    {Type: "string", Description: "检索关键词", Examples: []interface{}{"戒指", "三年前"}},
					"limit":    {Type: "integer", Description: "最多返回结果数，默认10，最大30", Default: 10, Examples: []interface{}{10}},
				},
				Required: []string{"novel_id", "query"},
			},
		},
		{
			Name:        "upsert_timeline_event",
			Description: "创建或更新时间线事件。timeline_id 存在时更新，不传则创建。用于维护全书事件顺序，避免时间线混乱。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"timeline_id": {Type: "integer", Description: "时间线事件ID；传入则更新，不传则创建", Examples: []interface{}{1}},
					"novel_id":    {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"chapter_id":  {Type: "integer", Description: "关联章节ID", Examples: []interface{}{12}},
					"sequence_no": {Type: "integer", Description: "时间线排序序号", Examples: []interface{}{120}},
					"event_time":  {Type: "string", Description: "故事内时间", Examples: []interface{}{"离开乌坦城第三天夜"}},
					"title":       {Type: "string", Description: "事件标题", Examples: []interface{}{"主角首次进入魔兽山脉"}},
					"content":     {Type: "string", Description: "事件内容", Examples: []interface{}{"主角为寻找药材进入魔兽山脉，遭遇狼群。"}},
					"importance":  {Type: "integer", Description: "重要程度1-5，默认3", Default: 3, Examples: []interface{}{3, 5}},
				},
				Required: []string{"novel_id", "title"},
			},
		},
		{
			Name:        "list_timeline_events",
			Description: "按时间线顺序列出小说事件。用于检查故事时间顺序和跨章节因果。",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]PropertyDef{
					"novel_id": {Type: "integer", Description: "小说ID", Examples: []interface{}{1}},
					"limit":    {Type: "integer", Description: "返回数量，默认50，最大200", Default: 50, Examples: []interface{}{50}},
				},
				Required: []string{"novel_id"},
			},
		},
	}
}

// toolDispatcher maps tool names to handler functions
var toolDispatcher = map[string]func(map[string]interface{}) *types.CallToolResult{
	"list_novels":                  tools.ListNovels,
	"get_novel":                    tools.GetNovel,
	"create_novel":                 tools.CreateNovel,
	"update_novel":                 tools.UpdateNovel,
	"delete_novel":                 tools.DeleteNovel,
	"list_chapters":                tools.ListChapters,
	"get_chapter":                  tools.GetChapter,
	"create_chapter":               tools.CreateChapter,
	"update_chapter":               tools.UpdateChapter,
	"delete_chapter":               tools.DeleteChapter,
	"list_characters":              tools.ListCharacters,
	"get_character":                tools.GetCharacter,
	"create_character":             tools.CreateCharacter,
	"update_character":             tools.UpdateCharacter,
	"delete_character":             tools.DeleteCharacter,
	"list_worldviews":              tools.ListWorldviews,
	"get_worldview":                tools.GetWorldview,
	"create_worldview":             tools.CreateWorldview,
	"update_worldview":             tools.UpdateWorldview,
	"delete_worldview":             tools.DeleteWorldview,
	"list_foreshadowings":          tools.ListForeshadowings,
	"get_foreshadowing":            tools.GetForeshadowing,
	"create_foreshadowing":         tools.CreateForeshadowing,
	"update_foreshadowing":         tools.UpdateForeshadowing,
	"delete_foreshadowing":         tools.DeleteForeshadowing,
	"list_versions":                tools.ListVersions,
	"get_version":                  tools.GetVersion,
	"rollback_version":             tools.RollbackVersion,
	"get_novel_context":            tools.GetNovelContext,
	"update_chapter_summary":       tools.UpdateChapterSummary,
	"get_recent_chapter_summaries": tools.GetRecentChapterSummaries,
	"update_character_state":       tools.UpdateCharacterState,
	"get_character_current_state":  tools.GetCharacterCurrentState,
	"upsert_plot_memory":           tools.UpsertPlotMemory,
	"search_novel_memory":          tools.SearchNovelMemory,
	"search_chapters":              tools.SearchChapters,
	"upsert_timeline_event":        tools.UpsertTimelineEvent,
	"list_timeline_events":         tools.ListTimelineEvents,
}
