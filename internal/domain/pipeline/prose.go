package pipeline

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ai-novel/studio/internal/infrastructure/quality"
)

// ProseMetric 单项文笔指标（供前端画进度条）。
type ProseMetric struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Value   float64 `json:"value"`
	Display string  `json:"display"`
	Target  string  `json:"target"`
	Score   int     `json:"score"`
}

// ProseIssue 文笔问题项。
type ProseIssue struct {
	Severity    string `json:"severity"` // warning / info
	Category    string `json:"category"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
	Sentence    string `json:"sentence,omitempty"`
}

// ProseReport 文笔实验室报告。
type ProseReport struct {
	Score           int           `json:"score"`
	Level           string        `json:"level"`
	Summary         string        `json:"summary"`
	Words           int           `json:"words"`
	Sentences       int           `json:"sentences"`
	Paragraphs      int           `json:"paragraphs"`
	DialogueRatio   float64       `json:"dialogue_ratio"`
	AvgSentenceLen  float64       `json:"avg_sentence_len"`
	SentenceLenStd  float64       `json:"sentence_len_std"`
	SensoryDensity  float64       `json:"sensory_density"`
	AdverbDensity   float64       `json:"adverb_density"`
	Metrics         []ProseMetric `json:"metrics"`
	Issues          []ProseIssue  `json:"issues"`
}

var dialogueRe = regexp.MustCompile(`[“"「『]([^”"」』]+)[”"」』]`)

var sensoryWords = []string{
	"看", "望", "盯", "瞥", "目光", "光", "影", "色", "亮", "暗", "闪",
	"听", "声", "响", "噪", "静", "嗡", "轰",
	"闻", "香", "臭", "味", "腥", "甜", "苦", "酸", "辣",
	"冷", "热", "烫", "凉", "疼", "痛", "湿", "干", "粗糙", "柔软",
}

var adverbWords = []string{
	"很", "非常", "十分", "极其", "忽然", "突然", "竟然", "居然",
	"仿佛", "似乎", "不禁", "忍不住", "微微", "缓缓", "轻轻", "悄悄",
	"默默", "渐渐", "下意识", "不由自主", "刻意", "显然",
}

// AnalyzeProse 规则式文笔分析：零 token、可重复、给前端可执行的修改建议。
func AnalyzeProse(content string) ProseReport {
	text := strings.TrimSpace(content)
	report := ProseReport{Metrics: []ProseMetric{}, Issues: []ProseIssue{}}
	if text == "" {
		report.Level = "样本为空"
		report.Summary = "先在编辑器里写一段正文，再点「分析本章」。"
		return report
	}

	report.Words = quality.WordCount(text)
	chars := nonSpaceRunes(text)
	report.Sentences = len(splitSentences(text))
	report.Paragraphs = len(splitParagraphs(text))
	if report.Sentences == 0 {
		report.Sentences = 1
	}
	if report.Paragraphs == 0 {
		report.Paragraphs = 1
	}
	report.DialogueRatio = dialogueRatio(text, chars)
	report.AvgSentenceLen, report.SentenceLenStd = sentenceStats(text)
	report.SensoryDensity = density(text, chars, sensoryWords)
	report.AdverbDensity = density(text, chars, adverbWords)

	if report.Words < 80 {
		report.Level = "样本太短"
		report.Summary = fmt.Sprintf("当前约 %d 字，样本偏短，指标仅供参考；写到 300 字以上会更准。", report.Words)
		report.Issues = append(report.Issues, ProseIssue{
			Severity: "info", Category: "样本量", Description: "文本较短，句长/对话/感官等统计波动较大。",
			Suggestion: "至少写满一个完整场景（300 字以上）再分析。",
		})
	}

	metrics := []ProseMetric{
		sentenceLengthMetric(report.AvgSentenceLen),
		rhythmMetric(report.SentenceLenStd),
		dialogueMetric(report.DialogueRatio),
		sensoryMetric(report.SensoryDensity),
		adverbMetric(report.AdverbDensity),
		paragraphMetric(float64(chars) / float64(report.Paragraphs)),
		aiTellMetric(len(AnalyzeAITells(text))),
	}
	report.Metrics = metrics

	weighted, weight, penalty := 0.0, 0.0, 0
	weights := []float64{0.16, 0.22, 0.10, 0.20, 0.14, 0.08, 0.10}
	for i, m := range metrics {
		weighted += float64(m.Score) * weights[i]
		weight += weights[i]
		if m.Score < 55 {
			penalty += 5
		} else if m.Score < 70 {
			penalty += 2
		}
		if m.Score < 70 {
			report.Issues = append(report.Issues, ProseIssue{
				Severity: "warning", Category: m.Label,
				Description: fmt.Sprintf("%s 当前 %s（理想 %s）。", m.Label, m.Display, m.Target),
				Suggestion:  metricSuggestion(m.Key),
			})
		}
	}
	if weight > 0 {
		report.Score = int(math.Round(weighted/weight)) - penalty
		if report.Score < 0 {
			report.Score = 0
		}
	}
	if report.Words < 80 {
		report.Score = minInt(report.Score, 60)
	}

	// 规则式 AI 味问题：最多取 5 条，避免面板过长。
	for i, issue := range AnalyzeAITells(text) {
		if i >= 5 {
			break
		}
		report.Issues = append(report.Issues, ProseIssue{
			Severity: issue.Severity, Category: issue.Category,
			Description: issue.Description, Suggestion: issue.Suggestion,
		})
	}

	switch {
	case report.Score >= 85:
		report.Level = "出色"
	case report.Score >= 70:
		report.Level = "良好"
	case report.Score >= 55:
		report.Level = "可打磨"
	default:
		report.Level = "需要精修"
	}
	report.Summary = buildProseSummary(report)
	return report
}

func nonSpaceRunes(s string) int {
	n := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}

func dialogueRatio(text string, chars int) float64 {
	if chars == 0 {
		return 0
	}
	inside := 0
	for _, m := range dialogueRe.FindAllStringSubmatch(text, -1) {
		if len(m) > 1 {
			inside += utf8.RuneCountInString(strings.TrimSpace(m[1]))
		}
	}
	return math.Min(1, float64(inside)/float64(chars))
}

func sentenceStats(text string) (avg, std float64) {
	sents := splitSentences(text)
	lengths := make([]float64, 0, len(sents))
	for _, s := range sents {
		n := nonSpaceRunes(s)
		if n > 0 {
			lengths = append(lengths, float64(n))
		}
	}
	if len(lengths) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, n := range lengths {
		sum += n
	}
	avg = sum / float64(len(lengths))
	if len(lengths) < 2 {
		return avg, 0
	}
	variance := 0.0
	for _, n := range lengths {
		d := n - avg
		variance += d * d
	}
	std = math.Sqrt(variance / float64(len(lengths)))
	return avg, std
}

func density(text string, chars int, words []string) float64 {
	if chars == 0 {
		return 0
	}
	hits := 0
	for _, w := range words {
		hits += strings.Count(text, w)
	}
	return float64(hits) * 1000 / float64(chars)
}

func sentenceLengthMetric(avg float64) ProseMetric {
	score := 100
	if avg < 10 {
		score = 55
	} else if avg < 14 {
		score = 80
	} else if avg > 42 {
		score = 50
	} else if avg > 34 {
		score = 78
	}
	return ProseMetric{Key: "sentence_length", Label: "平均句长", Value: avg, Display: fmt.Sprintf("%.1f 字/句", avg), Target: "14–34 字", Score: score}
}

func rhythmMetric(std float64) ProseMetric {
	score := 100
	if std < 2.5 {
		score = 45
	} else if std < 4.5 {
		score = 68
	} else if std < 6 {
		score = 84
	} else if std > 24 {
		score = 62
	}
	return ProseMetric{Key: "rhythm", Label: "句长波动", Value: std, Display: fmt.Sprintf("±%.1f 字", std), Target: "6–24", Score: score}
}

func dialogueMetric(ratio float64) ProseMetric {
	pct := ratio * 100
	score := 100
	if pct < 5 {
		score = 55
	} else if pct < 12 {
		score = 78
	} else if pct > 65 {
		score = 65
	}
	return ProseMetric{Key: "dialogue", Label: "对话占比", Value: pct, Display: fmt.Sprintf("%.0f%%", pct), Target: "12–65%", Score: score}
}

func sensoryMetric(d float64) ProseMetric {
	score := int(math.Min(100, d/8*100))
	return ProseMetric{Key: "sensory", Label: "感官密度", Value: d, Display: fmt.Sprintf("%.1f/千字", d), Target: "≥ 6/千字", Score: score}
}

func adverbMetric(d float64) ProseMetric {
	score := 100
	if d > 20 {
		score = 40
	} else if d > 14 {
		score = 62
	} else if d > 9 {
		score = 80
	}
	return ProseMetric{Key: "adverb", Label: "副词密度", Value: d, Display: fmt.Sprintf("%.1f/千字", d), Target: "≤ 9/千字", Score: score}
}

func paragraphMetric(avg float64) ProseMetric {
	score := 100
	if avg > 260 {
		score = 45
	} else if avg > 190 {
		score = 68
	} else if avg < 24 {
		score = 70
	}
	return ProseMetric{Key: "paragraph", Label: "段落均长", Value: avg, Display: fmt.Sprintf("%.0f 字/段", avg), Target: "40–190 字", Score: score}
}

func aiTellMetric(n int) ProseMetric {
	score := 100 - n*12
	if score < 0 {
		score = 0
	}
	return ProseMetric{Key: "ai_tells", Label: "AI 味规则", Value: float64(n), Display: fmt.Sprintf("%d 处", n), Target: "≤ 2 处", Score: score}
}

func metricSuggestion(key string) string {
	switch key {
	case "sentence_length":
		return "长句拆短、短句适当合并；一段里至少留一句有节奏感的短句。"
	case "rhythm":
		return "刻意制造长短句落差：动作句短，景物/心理句稍长，避免每句一样长。"
	case "dialogue":
		return "增加人物对话或把叙述改成对话交锋；对话要带潜台词，不要只交代信息。"
	case "sensory":
		return "每个场景至少落 1–2 种具体感官：看到什么、听到什么、闻到/触到什么。"
	case "adverb":
		return "删掉「很/非常/忽然/仿佛/忍不住」等副词，用动作和细节替代。"
	case "paragraph":
		return "按手机阅读习惯拆段：一个动作、一个情绪转折或一次对话就分段。"
	case "ai_tells":
		return "对照 AI 味检测逐条改写：少总结、少解释、少「仿佛/竟然」、少对称收尾。"
	default:
		return "按提示微调文字层，不要改动剧情与人物。"
	}
}

func buildProseSummary(r ProseReport) string {
	if r.Words < 80 {
		return r.Summary
	}
	var good, bad []string
	for _, m := range r.Metrics {
		if m.Score >= 82 {
			good = append(good, m.Label)
		} else if m.Score < 65 {
			bad = append(bad, m.Label)
		}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("文笔分 %d（%s）。", r.Score, r.Level))
	if len(good) > 0 {
		b.WriteString(strings.Join(good, "、") + " 表现不错；")
	}
	if len(bad) > 0 {
		b.WriteString(strings.Join(bad, "、") + " 是最值得先改的地方。")
	} else {
		b.WriteString("整体没有明显短板，可以再读一遍抓语感。")
	}
	return b.String()
}


