package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PromptsHandler struct {
	db *pgxpool.Pool
}

func NewPromptsHandler(db *pgxpool.Pool) *PromptsHandler {
	return &PromptsHandler{db: db}
}

type promptSummary struct {
	IDOrKey  string   `json:"id_or_key"`
	Name     string   `json:"name"`
	Origin   string   `json:"origin"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type promptDetail struct {
	IDOrKey      string   `json:"id_or_key"`
	Name         string   `json:"name"`
	Category     string   `json:"category,omitempty"`
	Origin       string   `json:"origin"`
	Template     string   `json:"template"`
	Variables    []string `json:"variables"`
	SystemPrompt *string  `json:"system_prompt,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Version      *string  `json:"version,omitempty"`
}

var (
	placeholderRe           = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}|\$\{([a-zA-Z0-9_]+)\}`)
	defaultPromptCategories = []string{"short_drama_narration"}
)

const shortDramaScriptSystemPrompt = "你是一位顶级的短剧解说up主，精通短视频创作的所有核心技巧。你必须严格按照JSON格式输出，绝不能包含任何其他文字、说明或代码块标记。"

const shortDramaScriptTemplate = `# 短剧解说脚本创作任务

## 任务目标
我是一位专业的短剧解说up主，需要为短剧《${drama_name}》创作一份高质量的解说脚本。目标是让观众在短时间内了解剧情精华，并产生强烈的继续观看欲望。

## 素材信息

### 剧情概述
<plot>
${plot_analysis}
</plot>

### 原始字幕（含精确时间戳）
<subtitles>
${subtitle_content}
</subtitles>

## 短剧解说创作核心要素

### 1. 黄金开场（3秒法则）
**开头3秒内必须制造强烈钩子，激发"想知道后续发展"的强烈好奇心**
- **悬念设置**：直接抛出最核心的冲突或疑问
  * 示例："身为一个名声恶臭的政客，他知道自己早晚会被暗杀"
  * 技巧：直接定性角色身份和处境，制造紧张感
- **冲突展示**：展现最激烈的对立关系
  * 示例："而这一天，就在他刚露头的时候..."
  * 技巧：用时间节点强调关键时刻的到来
- **情感共鸣**：触及观众内心的普遍情感
- **反转预告**：暗示即将发生的惊人转折
  * 技巧：使用"没想到"、"原来"、"竟然"等词汇预告反转

### 2. 主线提炼（去繁就简）
**快节奏解说，速度超越原剧，专注核心主线**
- 舍弃次要情节和配角，只保留推动主线的关键人物
- 突出核心矛盾冲突，每个片段都要推进主要故事线
- 快速跳过铺垫，直击剧情要害
- 确保每个解说片段都有明确的剧情推进作用
- **转折技巧**：大量使用"而这时"、"就在这时"、"没多久"等时间转折词

### 3. 爽点放大（情绪引爆）
**精准识别剧中"爽点"并用富有感染力的语言放大**
- **主角逆袭**：突出弱者变强、反败为胜的瞬间
- **反派被打脸**：强调恶人得到报应的痛快感
- **智商在线**：赞美角色的机智和策略
  * 示例："豺狼已经提前数日跟踪这名清洁工，并在他身上放了窃听器"
  * 技巧：展现角色的深谋远虑和专业能力
- **情感爆发**：放大感人、愤怒、震撼等强烈情绪
- 使用激昂语气和富有感染力的词汇调动观众情绪

### 4. 个性吐槽（增加趣味）
**以观众视角进行犀利点评，体现解说员独特人设**
- 避免单纯复述剧情，要有自己的观点和态度
- **"上帝视角"分析技巧**：
  * 揭示角色内心："他莫名地笑了一下"
  * 分析动机："豺狼的这几步都是事先算好的"
  * 预判后果："这又会有何代价呢"
- 适当吐槽剧情的套路或角色的愚蠢行为
- 用幽默、犀利的语言增加观看趣味
- 站在观众立场，说出观众想说的话
- **心理活动描述**：深入角色内心，增强代入感

### 5. 悬念预埋（引导互动）
**在关键节点和结尾处"卖关子"，激发互动欲望**
- 在剧情高潮前停止，留下"接下来会发生什么"的疑问
- **悬念设置技巧**：
  * 问题抛出："那么，UDC究竟是谁呢？"
  * 反转预告："而从这句话开始，所有的专业、体面和虚伪的平静都将分崩瓦解"
  * 时间悬念："几分钟后..."、"不久之后..."
- 提出引导性问题："你们觉得他会怎么做？"
- 预告后续精彩："更劲爆的还在后面"
- 为后续内容预热，激发评论、点赞、关注

### 6. 卡点配合（视听协调）
**考虑文案与画面、音乐的完美结合**
- 在情感高潮处预设BGM卡点
- 解说节奏要配合画面节奏
- 重要台词处保留原声，解说适时停顿
- 追求文案+画面+音乐的协同效应

## 专业解说语言技巧

### 1. 氛围营造技巧
**通过环境和细节描述增强画面感和代入感**
- **环境描述**："在这个距离，枪声都无法传到那边"
- **细节刻画**："他的床头有酒，身边的纸碟堆满烟头"
- **氛围渲染**："黑暗树林里有一间仓房"
- **情绪描述**："孤独又无助的豺狼，竟在这时露出了反常的一面"

### 2. 情感词汇运用
**使用富有感染力的词汇调动观众情绪**
- **紧张感**："名声恶臭"、"早晚会被暗杀"、"动用军警资源"
- **神秘感**："尘封的传奇"、"高度机密"、"暗藏玄机"
- **震撼感**："空前绝后的一枪"、"天衣无缝"、"神不知鬼不觉"
- **悲伤感**："目光非常悲伤"、"注定永远无法哀悼"

### 3. 节奏控制技巧
**通过语言节奏控制观众注意力**
- **快节奏推进**：使用短句，密集信息
- **慢节奏渲染**：使用长句，详细描述
- **停顿技巧**：在关键信息前适当停顿
- **重复强调**：重要信息适当重复

## 严格技术要求

### 时间戳管理（绝对不能违反）
- **时间戳绝对不能重叠**，确保剪辑后无重复画面
- **时间段必须连续且不交叉**，严格按时间顺序排列
- **每个时间戳都必须在原始字幕中找到对应范围**
- 可以拆分原时间片段，但必须保持时间连续性
- 时间戳的格式必须与原始字幕中的格式完全一致

### 时长控制（1/3原则）
- **解说视频总长度 = 原视频长度的 1/3**
- 精确控制节奏和密度，既不能过短也不能过长
- 合理分配解说和原声的时间比例

### 剧情连贯性
- **保持故事逻辑完整**，确保情节发展自然流畅
- **严格按照时间顺序**，禁止跳跃式叙述
- **符合因果逻辑**：先发生A，再发生B，A导致B

## 原声片段使用规范

### 原声片段格式要求
原声片段必须严格按照以下JSON格式：
{
  "_id": 序号,
  "timestamp": "开始时间-结束时间",
  "picture": "画面内容描述",
  "narration": "播放原片+序号",
  "OST": 1
}

### 原声片段插入策略

#### 1. 关键情绪爆发点
**在角色强烈情绪表达时必须保留原声，增强观众代入感**
- **愤怒爆发**：角色愤怒咆哮、情绪失控的瞬间
  * 参考："Come on, you bastard. Reaching."（愤怒对峙）
- **感动落泪**：角色感动哭泣、情感宣泄的时刻
- **震惊反应**：角色震惊、不敢置信的表情和台词
  * 参考："Are you sure about that?"（质疑震惊）
- **绝望崩溃**：角色绝望、崩溃的情感表达
  * 参考："Charles you're scaring me, what's wrong"（恐惧绝望）
- **狂欢庆祝**：角色兴奋、狂欢的情绪高潮

#### 2. 重要对白时刻
**保留推动剧情发展的关键台词和对话**
- **身份揭露**：揭示角色真实身份的重要台词
- **真相大白**：揭晓谜底、真相的关键对话
- **情感告白**：爱情告白、情感表达的重要台词
  * 参考："i'm really not good"（情感表达）
- **威胁警告**：反派威胁、警告的重要对白
  * 参考："You do not want to make enemies of these people"（威胁警告）
- **决定宣布**：角色做出重要决定的宣告

#### 3. 爽点瞬间
**在"爽点"时刻保留原声增强痛快感**
- **主角逆袭**：弱者反击、逆转局面的台词
- **反派被打脸**：恶人得到报应、被揭穿的瞬间
- **智商碾压**：主角展现智慧、碾压对手的台词
  * 参考："That is a fucking work of art guys"（技能展示）
- **正义伸张**：正义得到伸张、恶有恶报的时刻
- **实力展现**：主角展现真实实力、震撼全场

#### 4. 悬念节点
**在制造悬念或揭晓答案的关键时刻保留原声**
- **悬念制造**：制造悬念、留下疑问的台词
- **答案揭晓**：揭晓答案、解开谜团的对话
- **转折预告**：暗示即将发生转折的重要台词
- **危机降临**：危机来临、紧张时刻的对白

#### 5. 经典台词时刻
**保留具有强烈感染力和记忆点的经典台词**
- **哲理感悟**：角色的人生感悟和哲理思考
- **幽默调侃**：轻松幽默的对话增加趣味性
- **专业术语**：体现角色专业性的术语和对话
  * 参考："The scanner will pick up the metal components"（专业解释）
- **情感共鸣**：能引起观众共鸣的经典表达

### 原声片段技术规范

#### 格式规范
- **OST字段**：设置为1表示保留原声（解说片段设置为0）
- **narration格式**：严格使用"播放原片+序号"（如"播放原片26"）
- **picture字段**：详细描述画面内容，便于后期剪辑参考
- **时间戳精度**：必须与字幕中的重要对白时间精确匹配

#### 比例控制
- **原声与解说比例**：7:3（原声70%，解说30%）
- **分布均匀**：原声片段要在整个视频中均匀分布
- **长度适中**：单个原声片段时长控制在3-8秒
- **衔接自然**：原声片段与解说片段之间衔接自然流畅

#### 选择原则
- **情感优先**：优先选择情感强烈的台词和对话
- **剧情关键**：必须是推动剧情发展的重要内容
- **观众共鸣**：选择能引起观众共鸣的经典台词
- **视听效果**：考虑台词的声音效果和表演张力
- **代入感强**：选择能让观众产生强烈代入感的对话

## 输出格式要求

请严格按照以下JSON格式输出，绝不添加任何其他文字、说明或代码块标记：

{
  "items": [
    {
        "_id": 1,
        "timestamp": "00:00:01,000-00:00:05,500",
        "picture": "女主角林小雨慌张地道歉，男主角沈墨轩冷漠地看着她",
        "narration": "一个普通女孩的命运即将因为一杯咖啡彻底改变！她撞到的这个男人，竟然是...",
        "OST": 0
    },
    {
        "_id": 2,
        "timestamp": "00:00:05,500-00:00:08,000",
        "picture": "沈墨轩质问林小雨，语气冷厉威严",
        "narration": "播放原片2",
        "OST": 1
    },
    {
        "_id": 3,
        "timestamp": "00:00:08,000-00:00:12,000",
        "picture": "林小雨惊慌失措，沈墨轩眼中闪过一丝兴趣",
        "narration": "霸道总裁的经典开场！一杯咖啡引发的爱情故事就这样开始了...",
        "OST": 0
    }
  ]
}

## 质量标准

### 解说文案要求：
- **字数控制**：每段解说文案80-150字
- **语言风格**：生动有趣，富有感染力，符合短视频观众喜好
  * 参考风格："身为一个名声恶臭的政客，他知道自己早晚会被暗杀"
  * 直接定性，制造紧张感和代入感
- **情感调动**：能够有效调动观众情绪，产生代入感
  * 使用"而这时"、"没想到"、"原来"等转折词增强戏剧性
- **节奏把控**：快节奏但不失条理，紧凑但不混乱
  * 短句推进剧情，长句渲染氛围

### 技术规范：
- **解说与原片比例**：3:7（解说30%，原片70%）
- **原声片段标识**：OST=1表示原声，OST=0表示解说
- **原声格式规范**：narration字段必须使用"播放原片+序号"格式
- **关键情绪点**：必须保留原片原声，增强观众代入感
- **时间戳精度**：精确到毫秒级别，确保与字幕完美匹配
- **逻辑连贯性**：严格遵循剧情发展顺序

### 创作原则：
1. **只输出JSON内容**，不要任何说明性文字
2. **严格基于提供的剧情和字幕**，不虚构内容
3. **突出核心冲突**，舍弃无关细节
4. **强化观众体验**，始终考虑观看感受
5. **保持专业水准**，体现解说up主的专业素养
6. **融入经典解说技巧**：
   - 大量使用"上帝视角"分析
   - 适时插入心理活动描述
   - 运用悬念设置和反转技巧
   - 保持强烈的画面感和代入感

### 参考解说风格示例：
- **开场悬念**："身为一个名声恶臭的政客，他知道自己早晚会被暗杀"
- **转折技巧**："而这一天，就在他刚露头的时候..."
- **上帝视角**："豺狼已经提前数日跟踪这名清洁工"
- **情感渲染**："孤独又无助的豺狼，竟在这时露出了反常的一面"
- **悬念设置**："那么，UDC究竟是谁呢？"
- **反转预告**："而从这句话开始，所有的专业、体面和虚伪的平静都将分崩瓦解"

现在请基于以上要求，为短剧《${drama_name}》创作解说脚本：`

func getOfficialPromptDetail(key string) (promptDetail, bool) {
	if key != "short_drama_narration:script_generation" {
		return promptDetail{}, false
	}
	system := shortDramaScriptSystemPrompt
	return promptDetail{
		IDOrKey:      key,
		Name:         "短剧解说脚本生成",
		Category:     "short_drama_narration",
		Origin:       "official",
		Template:     shortDramaScriptTemplate,
		Variables:    []string{"drama_name", "plot_analysis", "subtitle_content"},
		SystemPrompt: &system,
		Tags: []string{
			"短剧", "解说脚本", "文案生成", "原声片段", "黄金开场", "爽点放大", "个性吐槽", "悬念预埋",
		},
		Version: strPtr("v2.0"),
	}, true
}

func strPtr(val string) *string {
	return &val
}

func (h *PromptsHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	rows, err := h.db.Query(r.Context(), `
		SELECT DISTINCT category
		FROM prompts
		WHERE user_id=$1
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	set := map[string]struct{}{}
	for _, c := range defaultPromptCategories {
		set[c] = struct{}{}
	}
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err == nil && strings.TrimSpace(cat) != "" {
			set[cat] = struct{}{}
		}
	}

	cats := make([]string, 0, len(set))
	for cat := range set {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	writeJSON(w, http.StatusOK, map[string]any{
		"data":      cats,
		"message":   "获取分类成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	category := strings.TrimSpace(r.URL.Query().Get("category"))

	query := `
		SELECT id, name, category, origin, tags
		FROM prompts
		WHERE user_id=$1
	`
	args := []any{userID}
	if category != "" {
		query += " AND category=$2"
		args = append(args, category)
	}
	query += " ORDER BY updated_at DESC"

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	out := make([]promptSummary, 0)
	for rows.Next() {
		var item promptSummary
		if err := rows.Scan(&item.IDOrKey, &item.Name, &item.Category, &item.Origin, &item.Tags); err == nil {
			out = append(out, item)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":      out,
		"message":   fmt.Sprintf("获取到 %d 条提示词", len(out)),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	key := chi.URLParam(r, "promptID")
	if item, ok := getOfficialPromptDetail(key); ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"data":      item,
			"message":   "获取提示词成功",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}
	var item promptDetail
	err := h.db.QueryRow(r.Context(), `
		SELECT id, name, category, origin, template, variables, system_prompt, tags, version
		FROM prompts
		WHERE user_id=$1 AND id=$2
	`, userID, key).Scan(
		&item.IDOrKey,
		&item.Name,
		&item.Category,
		&item.Origin,
		&item.Template,
		&item.Variables,
		&item.SystemPrompt,
		&item.Tags,
		&item.Version,
	)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":      item,
		"message":   "获取提示词成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) RenderPreview(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	promptID := chi.URLParam(r, "promptID")
	var payload struct {
		Variables map[string]any `json:"variables"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	item, err := h.getPromptDetail(r.Context(), userID, promptID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
		return
	}

	content := renderTemplate(item.Template, payload.Variables)
	messages := []map[string]string{}
	if item.SystemPrompt != nil && strings.TrimSpace(*item.SystemPrompt) != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": *item.SystemPrompt,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": content,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"messages": messages,
		},
		"message": "渲染成功",
	})
}

func (h *PromptsHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var payload struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		Template     string   `json:"template"`
		Variables    []string `json:"variables"`
		SystemPrompt *string  `json:"system_prompt"`
		Tags         []string `json:"tags"`
		Version      *string  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称不能为空"})
		return
	}
	if strings.TrimSpace(payload.Category) == "" {
		payload.Category = defaultPromptCategories[0]
	}

	vars := payload.Variables
	if len(vars) == 0 {
		vars = extractPlaceholders(payload.Template)
	}
	var id string
	err := h.db.QueryRow(r.Context(), `
		INSERT INTO prompts (user_id, name, category, origin, template, variables, system_prompt, tags, version)
		VALUES ($1, $2, $3, 'user', $4, $5, $6, $7, $8)
		RETURNING id
	`, userID, payload.Name, payload.Category, payload.Template, vars, payload.SystemPrompt, payload.Tags, payload.Version).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}

	item := promptDetail{
		IDOrKey:      id,
		Name:         payload.Name,
		Category:     payload.Category,
		Origin:       "user",
		Template:     payload.Template,
		Variables:    vars,
		SystemPrompt: payload.SystemPrompt,
		Tags:         payload.Tags,
		Version:      payload.Version,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      item,
		"message":   "创建成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	promptID := chi.URLParam(r, "promptID")
	var payload struct {
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		Template     string   `json:"template"`
		Variables    []string `json:"variables"`
		SystemPrompt *string  `json:"system_prompt"`
		Tags         []string `json:"tags"`
		Version      *string  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	vars := payload.Variables
	if len(vars) == 0 && strings.TrimSpace(payload.Template) != "" {
		vars = extractPlaceholders(payload.Template)
	}
	var varsParam any
	if len(vars) > 0 {
		varsParam = vars
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE prompts
		SET name=COALESCE(NULLIF($3, ''), name),
		    category=COALESCE(NULLIF($4, ''), category),
		    template=COALESCE(NULLIF($5, ''), template),
		    variables=CASE WHEN $6 IS NULL THEN variables ELSE $6 END,
		    system_prompt=$7,
		    tags=$8,
		    version=$9,
		    updated_at=NOW()
		WHERE user_id=$1 AND id=$2
	`, userID, promptID, payload.Name, payload.Category, payload.Template, varsParam, payload.SystemPrompt, payload.Tags, payload.Version)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	item := promptDetail{
		IDOrKey:      promptID,
		Name:         payload.Name,
		Category:     payload.Category,
		Origin:       "user",
		Template:     payload.Template,
		Variables:    vars,
		SystemPrompt: payload.SystemPrompt,
		Tags:         payload.Tags,
		Version:      payload.Version,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      item,
		"message":   "更新成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	promptID := chi.URLParam(r, "promptID")
	_, err := h.db.Exec(r.Context(), `
		DELETE FROM prompts WHERE user_id=$1 AND id=$2
	`, userID, promptID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"deleted": true,
		},
		"message": "删除成功",
	})
}

func (h *PromptsHandler) ValidateTemplate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Template     string   `json:"template"`
		RequiredVars []string `json:"required_vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	placeholders := extractPlaceholders(payload.Template)
	phSet := map[string]struct{}{}
	for _, p := range placeholders {
		phSet[p] = struct{}{}
	}
	reqSet := map[string]struct{}{}
	for _, r := range payload.RequiredVars {
		reqSet[r] = struct{}{}
	}
	missing := make([]string, 0)
	for _, req := range payload.RequiredVars {
		if _, ok := phSet[req]; !ok {
			missing = append(missing, req)
		}
	}
	extra := make([]string, 0)
	for _, p := range placeholders {
		if _, ok := reqSet[p]; !ok {
			extra = append(extra, p)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"placeholders": placeholders,
			"missing":      missing,
			"extra":        extra,
		},
		"message": "校验完成",
	})
}

func (h *PromptsHandler) SetProjectSelection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	var payload struct {
		FeatureKey string         `json:"feature_key"`
		Selection  map[string]any `json:"selection"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.FeatureKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	raw, _ := json.Marshal(payload.Selection)
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO prompt_selections (user_id, project_id, feature_key, selection, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, project_id, feature_key)
		DO UPDATE SET selection=EXCLUDED.selection, updated_at=NOW()
	`, userID, projectID, payload.FeatureKey, raw)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"prompt_selection": map[string]any{
				payload.FeatureKey: payload.Selection,
			},
		},
		"message": "更新成功",
	})
}

func (h *PromptsHandler) GetProjectSelection(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")
	rows, err := h.db.Query(r.Context(), `
		SELECT feature_key, selection
		FROM prompt_selections
		WHERE user_id=$1 AND project_id=$2
	`, userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	result := map[string]any{}
	for rows.Next() {
		var key string
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			continue
		}
		selection := map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &selection)
		}
		result[key] = selection
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":      result,
		"message":   "获取成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PromptsHandler) getPromptDetail(ctx context.Context, userID, promptID string) (promptDetail, error) {
	if item, ok := getOfficialPromptDetail(promptID); ok {
		return item, nil
	}
	var item promptDetail
	err := h.db.QueryRow(ctx, `
		SELECT id, name, category, origin, template, variables, system_prompt, tags, version
		FROM prompts
		WHERE user_id=$1 AND id=$2
	`, userID, promptID).Scan(
		&item.IDOrKey,
		&item.Name,
		&item.Category,
		&item.Origin,
		&item.Template,
		&item.Variables,
		&item.SystemPrompt,
		&item.Tags,
		&item.Version,
	)
	return item, err
}

func extractPlaceholders(template string) []string {
	matches := placeholderRe.FindAllStringSubmatch(template, -1)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		key := ""
		if len(match) > 1 && match[1] != "" {
			key = match[1]
		} else if len(match) > 2 {
			key = match[2]
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func renderTemplate(template string, vars map[string]any) string {
	if len(vars) == 0 {
		return template
	}
	return placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))
		if strings.HasPrefix(match, "${") {
			key = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}"))
		}
		if val, ok := vars[key]; ok {
			return fmt.Sprint(val)
		}
		return match
	})
}
