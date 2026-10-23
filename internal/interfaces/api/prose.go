package api

import (
	"net/http"
	"strings"

	"github.com/ai-novel/studio/internal/domain/pipeline"
)

// HandleProseAnalyze POST /api/prose/analyze
// 请求体：{ "content": "..." }
// 响应：{ score, level, summary, metrics, issues, ... }
func (s *Server) HandleProseAnalyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "content 不能为空")
		return
	}
	if len([]rune(req.Content)) > 60000 {
		writeError(w, http.StatusBadRequest, "单次分析上限 60000 字，请分章分析")
		return
	}
	writeOK(w, pipeline.AnalyzeProse(req.Content))
}
