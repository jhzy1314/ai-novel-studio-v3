package api

import "net/http"

// ProductVersion 当前产品版本，供启动脚本与前端识别实例身份。
const ProductVersion = "3.0.0"

// HandleVersion GET /api/version —— 无需认证的实例身份探针。
func (s *Server) HandleVersion(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]interface{}{
		"product": "AI Novel Studio",
		"version": ProductVersion,
		"edition": "ai",
	})
}
