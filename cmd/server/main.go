package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ai-novel/studio/internal/appearance"
	"github.com/ai-novel/studio/internal/domain/pipeline"
	"github.com/ai-novel/studio/internal/infrastructure/config"
	"github.com/ai-novel/studio/internal/infrastructure/database"
	"github.com/ai-novel/studio/internal/infrastructure/llm"
	"github.com/ai-novel/studio/internal/infrastructure/logging"
	"github.com/ai-novel/studio/internal/infrastructure/quota"
	"github.com/ai-novel/studio/internal/interfaces/api"
)

func main() {
	ctx := context.Background()

	// 1. 加载配置文件（API 密钥由 config.yaml 管理，禁止硬编码）
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 1.1 初始化文件日志
	logDir := cfg.Server.LogDir
	if logDir == "" {
		logDir = "log"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("警告: 创建日志目录失败: %v", err)
	}
	if err := logging.Init(logDir); err != nil {
		log.Printf("警告: 文件日志初始化失败: %v", err)
	}
	defer logging.Close()

	// 1.2 设置认证密码
	if cfg.Server.AuthPassword != "" {
		api.SetAuthConfig(cfg.Server.AuthPassword)
		log.Println("🔐 密码认证已启用")
	} else {
		log.Println("⚠️  未设置 auth_password，认证关闭（公网部署必须设置）")
	}

	// 1.3 初始化Cookie加密器（首次启动自动生成随机密钥并持久化到数据目录，防止所有安装共享同一密钥）
	// 先确保数据目录存在，否则首次启动写 cookie_key 会失败，导致每次重启都换密钥。
	dataDir := filepath.Dir(cfg.Server.SQLitePath)
	if dataDir != "" && dataDir != "." {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 创建数据目录失败: %v", err)
		}
	}
	cookieKey := cfg.Server.CookieEncryptKey
	if cookieKey == "" {
		keyFile := filepath.Join(dataDir, "cookie_key")
		if data, err := os.ReadFile(keyFile); err == nil && len(data) > 16 {
			cookieKey = string(data)
		} else {
			randomKey := make([]byte, 32)
			if _, err := rand.Read(randomKey); err != nil {
				log.Fatalf("生成随机加密密钥失败: %v", err)
			}
			cookieKey = hex.EncodeToString(randomKey)
			if err := os.WriteFile(keyFile, []byte(cookieKey), 0600); err != nil {
				log.Printf("警告: 保存加密密钥失败: %v", err)
			}
			log.Println("⚠️  已自动生成随机加密密钥（存储在 data/cookie_key），重装后请备份此文件")
		}
	}
	llm.InitCookieEncryptor(cookieKey)
	log.Println("🔒 Cookie加密已启用")

	// 2. 初始化 SQLite 数据库（零配置，自动建表迁移）
	db, err := database.Open(ctx, cfg.Server.SQLitePath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 3. 种子：限额参数 / 模型池 / 角色绑定
	if err := db.Seed(ctx, cfg); err != nil {
		log.Fatalf("数据库种子失败: %v", err)
	}

	store := database.NewStore(db)

	// 4. 种子：系统内置模板
	if err := store.SeedSystemTemplates(ctx, systemTemplates()); err != nil {
		log.Printf("警告: 系统模板种子失败: %v", err)
	}

	// 4.3 数据库版本迁移（幂等）
	if err := db.RunMigrations(ctx, []database.DBMigration{
		{Version: 1, Name: "models扩展字段(is_custom/context_limit等)", SQL: "SELECT 1"},
		{Version: 2, Name: "timeline表", SQL: "SELECT 1"},
		{Version: 3, Name: "chapters扩展(tags/synopsis)", SQL: "SELECT 1"},
		{Version: 4, Name: "chapters软删除(is_deleted/deleted_at)", SQL: "SELECT 1"},
		{Version: 5, Name: "style_samples扩展拆书素材类型(kind)已并入schema", SQL: "SELECT 1"},
	}); err != nil {
		log.Printf("警告: 数据库迁移失败: %v", err)
	}

	// 4.5 清理30天前的旧日志
	if err := store.CleanOldLogs(ctx); err != nil {
		log.Printf("警告: 日志清理失败: %v", err)
	}

	// 4.6 清理7天前的回收站章节
	purged, err := store.PurgeOldTrash(ctx)
	if err != nil {
		log.Printf("警告: 回收站清理失败: %v", err)
	} else if purged > 0 {
		log.Printf("🗑 已清理 %d 个过期回收站章节", purged)
	}

	// 4.7 字数统计迁移（幂等）：按统一算法重算存量章节，避免旧库字数不一致。
	if n, err := store.RecountAllWordCounts(ctx); err != nil {
		log.Printf("警告: 字数迁移失败: %v", err)
	} else if n > 0 {
		log.Printf("字数迁移: 重算 %d 个章节", n)
	}

	// 5. 构造模型适配层注册中心（角色 -> 有序模型列表）
	registry := llm.NewRegistry(store)

	// 6. 成本控制 / 限流器
	limiter := quota.NewLimiter(store)

	// 7. 调度中枢 Agent（核心）
	dispatcher := pipeline.NewDispatcher(registry, store, limiter)

	// 7.5 定时备份（每 4 小时一次，一致性快照）和回收站清理（每小时）
	go periodicBackup(db, cfg.Server.SQLitePath)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			n, _ := store.PurgeOldTrash(context.Background())
			if n > 0 {
				log.Printf("🗑 回收站自动清理: 删除了 %d 个过期章节", n)
			}
		}
	}()

	// 7.6 定时清理过期会话
	go cleanExpiredSessions()

	// 8. 启动 API 服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8082
	}
	listenAddr := cfg.Server.ListenAddr
	if listenAddr == "" {
		// 本地优先：默认只监听回环地址，避免把写作数据暴露到局域网。
		listenAddr = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", listenAddr, port)
	appSvc := appearance.NewService(nil)

	server := api.NewServer(store, registry, dispatcher, limiter, appSvc)
	server.SetLibraryDir(cfg.Server.LibraryDir)

	// 优雅退出
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("收到退出信号，正在关闭...")
		logging.Close()
		_ = db.Close()
		os.Exit(0)
	}()

	displayHost := listenAddr
	if displayHost == "0.0.0.0" || displayHost == "::" {
		displayHost = "localhost"
	}
	log.Printf("🚀 AI Novel Studio 后端已启动: http://%s:%d", displayHost, port)
	log.Printf("   SQLite: %s", cfg.Server.SQLitePath)
	log.Printf("   日志: %s", logDir)
	log.Printf("   模型池: %d 个，角色: thinker/worker/verifier/helper", len(cfg.Models))
	if cfg.Server.AuthPassword != "" {
		log.Printf("   认证: 已启用密码保护")
	} else if !isLoopbackAddr(listenAddr) {
		log.Printf("   ⚠️  当前监听 %s 且未设置 auth_password：同一网络内的设备可访问你的写作数据", listenAddr)
	}
	if err := server.Start(addr); err != nil {
		log.Fatalf("API Server 启动失败: %v", err)
	}
}

// periodicBackup 定时生成 SQLite 一致性快照（VACUUM INTO）。
// 替代裸文件复制：WAL 模式下 io.Copy 可能拿到不含 -wal 未 checkpoint 数据的不一致副本。
func periodicBackup(db *database.DB, dbPath string) {
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	// 启动时先备份一次（留存每日首份），之后每 4 小时
	doBackup(db, backupDir)
	for {
		time.Sleep(4 * time.Hour)
		doBackup(db, backupDir)
	}
}

func doBackup(db *database.DB, backupDir string) {
	store := database.NewStore(db)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("备份失败: 创建备份目录: %v", err)
		return
	}
	today := time.Now().Format("2006-01-02-150405")
	dst := filepath.Join(backupDir, fmt.Sprintf("ai-novel-%s.db", today))
	if err := store.BackupTo(context.Background(), dst); err != nil {
		log.Printf("备份失败: %v", err)
		return
	}
	log.Printf("✅ 数据库已备份（一致性快照）: %s", dst)
	// 保留最近 30 份快照（时间戳文件名按字典序即时间序，删除最旧的溢出部分）
	entries, _ := os.ReadDir(backupDir)
	if len(entries) > 30 {
		for i := 0; i < len(entries)-30; i++ {
			os.Remove(filepath.Join(backupDir, entries[i].Name()))
		}
	}
}

// cleanExpiredSessions 每小时清理一次过期会话
func cleanExpiredSessions() {
	for {
		time.Sleep(1 * time.Hour)
		// 会话过期由 middleware.go 中的 goroutine 处理即可
	}
}

// systemTemplates 内置系统提示词模板（is_system=1，前端可读不可删）
func systemTemplates() []database.Template {
	return []database.Template{
		{
			Name:     "标准小说续写",
			Category: "novel",
			Content:  "请基于世界观、人物卡与前文，规划本章剧情并续写正文。保持人设一致、叙事连贯，字数按用户要求执行。",
		},
		{
			Name:     "严谨论文撰写",
			Category: "strict",
			Content:  "请按学术规范搭建框架：核心论点、论据排布、事实核查。优先保证逻辑与事实准确，语言正式严谨。",
		},
		{
			Name:     "文艺氛围创作",
			Category: "art",
			Content:  "请强化细节描写、心理活动、画面感与情绪渲染，追求文字表现力与故事氛围感。",
		},
		{
			Name:     "段落摘要提取",
			Category: "light",
			Content:  "请提取所选文本的核心信息，输出简洁摘要，保留关键事实，不新增观点。",
		},
		{
			Name:     "局部润色改写",
			Category: "light",
			Content:  "请在保留原文核心剧情与逻辑的前提下，仅优化语言节奏与文字流畅度，不改动未涉及内容。",
		},
	}
}

// isLoopbackAddr 判断监听地址是否只对本机开放。
func isLoopbackAddr(addr string) bool {
	if addr == "" || addr == "localhost" {
		return true
	}
	ip := net.ParseIP(addr)
	return ip != nil && ip.IsLoopback()
}
