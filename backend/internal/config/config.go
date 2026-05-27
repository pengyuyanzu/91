package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  Server  `yaml:"server"`
	Storage Storage `yaml:"storage"`
	Scanner Scanner `yaml:"scanner"`
	Preview Preview `yaml:"preview"`
	Nightly Nightly `yaml:"nightly"`
	Drives  []Drive `yaml:"drives"`
}

type Server struct {
	Listen        string `yaml:"listen"`
	Admin         Admin  `yaml:"admin"`
	SessionSecret string `yaml:"session_secret"`
	// AllowedOrigins 是允许跨源访问的前端 Origin 白名单（如 "https://video.example.com"）。
	// 默认空 → 不开启 CORS 跨源；同源部署（前后端在同一个域名 + 端口下）不需要配置此项。
	// 浏览器对不在列表里的 Origin 不会拿到 Access-Control-Allow-Origin 头，自然就读不到响应。
	// 不要写 "*"；带 cookie 的 CORS 必须是具体 Origin。
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type Admin struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Storage struct {
	DBPath          string `yaml:"db_path"`
	LocalPreviewDir string `yaml:"local_preview_dir"`
}

type Scanner struct {
	// IntervalSeconds 已废弃。旧版每天 02:00–07:00 窗口内按这个间隔重复扫盘；
	// 新版统一由 nightly.cron_hour 调度，此字段被忽略，保留仅为兼容旧 yaml。
	IntervalSeconds int      `yaml:"interval_seconds"`
	MaxDepth        int      `yaml:"max_depth"`
	VideoExtensions []string `yaml:"video_extensions"`
}

type Preview struct {
	Enabled         bool   `yaml:"enabled"`
	FFmpegPath      string `yaml:"ffmpeg_path"`
	FFprobePath     string `yaml:"ffprobe_path"`
	DurationSeconds int    `yaml:"duration_seconds"`
	Width           int    `yaml:"width"`
	Segments        int    `yaml:"segments"`
}

// Nightly 是凌晨流水线（扫盘 → 91 爬虫 → 迁移）的调度配置。
//
// 一个进程只跑一条 nightly 流水线；该 cron 时间到达且当天还没跑过时触发，
// 也可被管理后台「立即跑全流程」按钮手动触发。MaxDuration 是软超时，超过
// 后当前 phase 完成、后续 phase 不再启动。
type Nightly struct {
	// CronHour 是每日触发整点（0–23）；默认 1 表示 01:00。
	CronHour int `yaml:"cron_hour"`
	// MaxDuration 是单次流水线总耗时上限；默认 6h。
	MaxDuration time.Duration `yaml:"max_duration"`
}

// Drive 配置项中的敏感字段（Cookie / RefreshToken 等）最终由管理后台写入 DB
// 这里保留 yaml 中的静态定义，用于启动时预置盘。生产建议只在 DB 里维护。
type Drive struct {
	ID     string            `yaml:"id"`
	Kind   string            `yaml:"kind"` // quark / p115 / pikpak / wopan / onedrive
	Name   string            `yaml:"name"`
	RootID string            `yaml:"root_id"`
	Params map[string]string `yaml:"params,omitempty"`
}

// Load 读取配置；若不存在则从 config.example.yaml 复制一份并返回
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		example := filepath.Join(filepath.Dir(path), "config.example.yaml")
		data, err := os.ReadFile(example)
		if err != nil {
			return nil, fmt.Errorf("config not found and example missing: %w", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, fmt.Errorf("write default config: %w", err)
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	c.applyDefaults()
	return &c, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Listen == "" {
		c.Server.Listen = ":8080"
	}
	if c.Storage.DBPath == "" {
		c.Storage.DBPath = "./data/video-site.db"
	}
	if c.Storage.LocalPreviewDir == "" {
		c.Storage.LocalPreviewDir = "./data/previews"
	}
	if c.Scanner.MaxDepth == 0 {
		c.Scanner.MaxDepth = 5
	}
	if len(c.Scanner.VideoExtensions) == 0 {
		c.Scanner.VideoExtensions = []string{".mp4", ".mkv", ".mov", ".webm", ".avi"}
	}
	if c.Preview.FFmpegPath == "" {
		c.Preview.FFmpegPath = "ffmpeg"
	}
	if c.Preview.FFprobePath == "" {
		c.Preview.FFprobePath = "ffprobe"
	}
	if c.Preview.DurationSeconds != 3 {
		c.Preview.DurationSeconds = 3
	}
	if c.Preview.Width == 0 {
		c.Preview.Width = 480
	}
	if c.Preview.Segments == 0 {
		c.Preview.Segments = 3
	}
	// Nightly defaults。CronHour=0 是合法值（午夜），没法用 zero-value 单独
	// 区分"未设"和"显式 0"。把整个 nightly 块当 sentinel —— MaxDuration==0
	// 视为整个块缺失，重置成 (cron_hour=1, max_duration=6h)。代价：用户想配
	// CronHour=0（午夜）必须同时显式写 max_duration（任何 >0 的值即可）。
	// 收益：默认部署（yaml 没 nightly 块）得到 01:00 + 6h，与用户预期一致。
	if c.Nightly.MaxDuration <= 0 {
		c.Nightly.CronHour = 1
		c.Nightly.MaxDuration = 6 * time.Hour
	} else if c.Nightly.CronHour < 0 || c.Nightly.CronHour > 23 {
		c.Nightly.CronHour = 1
	}
}
