package pipeline

import "strings"

// ProseStyle v3.0 文笔风格预设。
type ProseStyle struct {
	ID    string
	Label string
	Brief string
}

// ProseStyles 前端下拉与后端校验共用。
var ProseStyles = []ProseStyle{
	{ID: "cold", Label: "冷峻克制", Brief: "短句为主，少形容词，把情绪落到动作与细节上，不写总结句和抒情旁白。"},
	{ID: "classical", Label: "古典雅致", Brief: "白话为底，带一点文言的节奏与意象，讲究用词，避开网络流行语和现代口语。"},
	{ID: "cinematic", Label: "电影感", Brief: "像镜头一样调度场景：先给空间与光影，再给动作与声音，段落之间有画面切换。"},
	{ID: "earthy", Label: "烟火气", Brief: "生活化细节、具体动作和口语对话，写食物、物件、价钱、气味，不悬浮抒情。"},
	{ID: "suspense", Label: "悬疑张力", Brief: "信息克制释放，长短句交替，少解释多留白，段尾留钩子。"},
}

// ProseStyleByID 查找预设；未知返回零值。
func ProseStyleByID(id string) ProseStyle {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, s := range ProseStyles {
		if s.ID == id {
			return s
		}
	}
	return ProseStyle{}
}

// ProseStyleDirective 生成注入创作提示词的文笔要求；空 ID 返回空串。
func ProseStyleDirective(id string) string {
	s := ProseStyleByID(id)
	if s.ID == "" {
		return ""
	}
	return "【文笔风格要求 · " + s.Label + "】\n" + s.Brief +
		"\n这条要求只改变文字层（句式、节奏、意象、用词），不得改变剧情、人设、对话原意；与用户明确指令冲突时以用户指令为准。"
}

// ProseStylePolishDirective 润色场景的文笔指令（更强调不改剧情）。
func ProseStylePolishDirective(id string) string {
	s := ProseStyleByID(id)
	if s.ID == "" {
		return ""
	}
	return "【本次润色的文笔方向 · " + s.Label + "】\n" + s.Brief +
		"\n只改文字层；情节、人物、对话信息量不变。不要为了风格化而增删事件。"
}
