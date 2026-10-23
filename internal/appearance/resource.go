package appearance

import "path/filepath"

type Resource struct {
	ID       string
	URL      string
	FileName string
	MD5      string
	DestDir  string
}

func (r *Resource) LocalPath() string {
	return filepath.Join(r.DestDir, r.FileName)
}

type Preset struct {
	ID    string `yaml:"id"`
	URL   string `yaml:"url"`
	File  string `yaml:"file"`
	MD5   string `yaml:"md5"`
}

// 默认背景已全部随前端嵌入二进制（web/static/backgrounds），
// 运行期只保留用户上传/生成的自定义背景，不再从外部仓库预取。
