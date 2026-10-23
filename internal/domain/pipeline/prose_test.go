package pipeline

import (
	"strings"
	"testing"
)

func TestAnalyzeProseBasic(t *testing.T) {
	text := `雨下得更密了。林云把伞往他那边偏了偏，自己的肩头很快湿了一片。惊鸿看见了，没说话，只把书往怀里收了收。
“你冷吗？”她问。
“不冷。”他答得很快，像怕她再问。
操场边的梧桐叶子被雨打得发亮，空气里有一股潮湿的土腥味。远处广播还在放那首老歌，声音隔着雨幕，忽远忽近。`
	r := AnalyzeProse(text)
	if r.Score < 0 || r.Score > 100 {
		t.Fatalf("score out of range: %d", r.Score)
	}
	if r.Words == 0 || r.Sentences == 0 || r.Paragraphs == 0 {
		t.Fatalf("unexpected zero stats: %+v", r)
	}
	if len(r.Metrics) != 7 {
		t.Fatalf("expected 7 metrics, got %d", len(r.Metrics))
	}
	if r.Summary == "" || r.Level == "" {
		t.Fatalf("missing summary/level: %+v", r)
	}
}

func TestAnalyzeProseEmpty(t *testing.T) {
	r := AnalyzeProse("   ")
	if r.Score != 0 || r.Level != "样本为空" {
		t.Fatalf("unexpected empty report: %+v", r)
	}
}

func TestProseStyleDirective(t *testing.T) {
	if ProseStyleDirective("") != "" {
		t.Fatal("empty style should not inject a directive")
	}
	d := ProseStyleDirective("cold")
	if !strings.Contains(d, "冷峻克制") || !strings.Contains(d, "不得改变剧情") {
		t.Fatalf("unexpected directive: %s", d)
	}
	if ProseStyleByID("unknown").ID != "" {
		t.Fatal("unknown style should be empty")
	}
}
