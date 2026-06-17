package config

import (
	"context"
	"modelcraft/pkg/logfacade"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig     `mapstructure:"server"`
	Database DatabaseConfig   `mapstructure:"database"` // 设计时数据库（平台数据库）
	Redis    RedisConfig      `mapstructure:"redis"`
	JWT      JWTConfig        `mapstructure:"jwt"`
	Auth     AuthConfig       `mapstructure:"auth"` // 认证配置
	Logger   logfacade.Config `mapstructure:"logger"`
	Crypto   CryptoConfig     `mapstructure:"crypto"` // 加密配置
}

// 解析配置到结构体
var Cfg Config

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type             string `json:"type" mapstructure:"type"`           // 数据库类型: mysql
	FilePath         string `json:"file_path" mapstructure:"file_path"` // SQLite文件路径
	Host             string `json:"host" mapstructure:"host"`
	Port             int    `json:"port" mapstructure:"port"`
	Username         string `json:"username" mapstructure:"username"`
	Password         string `json:"password" mapstructure:"password"`
	Database         string `json:"database" mapstructure:"database"`
	Charset          string `json:"charset" mapstructure:"charset"`
	MaxIdleConns     int    `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxOpenConns     int    `json:"max_open_conns" mapstructure:"max_open_conns"`
	ConnMaxLifetime  int    `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`   // 秒
	LogLevel         string `json:"log_level" mapstructure:"log_level"`                   // silent, error, warn, info
	MigrateOnStartup bool   `json:"migrate_on_startup" mapstructure:"migrate_on_startup"` // 是否在启动时自动迁移数据库
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`      // Legacy HMAC secret (kept for migration)
	PrivateKey string        `mapstructure:"private_key"` // ES256 PEM private key (preferred)
	Expiration time.Duration `mapstructure:"expiration"`  // Access token expiration time
	Issuer     string        `mapstructure:"issuer"`      // JWT issuer claim
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	AESKey string `mapstructure:"aes_key"` // AES-256密钥，必须是32字节
}

// CookieConfig 刷新令牌 Set-Cookie 属性配置
type CookieConfig struct {
	Domain   string `mapstructure:"domain"`    // Cookie 域名（留空则使用请求主机）
	Secure   bool   `mapstructure:"secure"`    // Secure 标志；生产环境应设为 true（HTTPS）
	SameSite string `mapstructure:"same_site"` // "strict" | "lax" | "none"
}

// AuthConfig 认证配置
type AuthConfig struct {
	Cookie        CookieConfig      `mapstructure:"cookie"`         // 刷新令牌 Cookie 配置
	Design        DesignAuthConfig  `mapstructure:"design"`         // 设计时API认证配置
	Runtime       RuntimeAuthConfig `mapstructure:"runtime"`        // 运行时API认证配置
}

// DesignAuthConfig 设计时API认证配置
type DesignAuthConfig struct {
	Enabled             bool   `mapstructure:"enabled"`               // 是否启用认证
	JWTPublicKeyPath    string `mapstructure:"jwt_public_key_path"`   // JWT 公钥路径
	JWTPublicKey        string `mapstructure:"jwt_public_key"`        // JWT 公钥内容
	AcceptModelcraftJWT bool   `mapstructure:"accept_modelcraft_jwt"` // 是否接受 ModelCraft JWT (migration flag)
}

// RuntimeAuthConfig 运行态认证配置
type RuntimeAuthConfig struct {
	Enabled             bool     `mapstructure:"enabled"`
	OptionalForProjects []string `mapstructure:"optional_for_projects"`
}

// ConfigOptions 配置选项
type ConfigOptions struct {
	ConfigFile string // 配置文件路径
	EnvFile    string // 环境变量文件路径
}

// LoadConfig 使用默认配置文件加载配置
func LoadConfig(ctx context.Context) *Config {
	return LoadConfigWithOptions(ctx, ConfigOptions{
		ConfigFile: "config.yaml",
		EnvFile:    ".env",
	})
}

// LoadConfigWithFile 使用指定配置文件加载配置
func LoadConfigWithFile(ctx context.Context, configFile string) *Config {
	return LoadConfigWithOptions(ctx, ConfigOptions{
		ConfigFile: configFile,
		EnvFile:    ".env",
	})
}

// LoadConfigWithOptions 使用配置选项加载配置
func LoadConfigWithOptions(ctx context.Context, opts ConfigOptions) *Config {
	logger := logfacade.GetLogger(ctx)
	// 创建新的 Viper 实例
	v := viper.New()

	// 加载配置文件
	if opts.ConfigFile != "" {
		loadConfigFile(ctx, v, opts.ConfigFile)
	}

	// 加载环境变量文件（使用 godotenv）
	if opts.EnvFile != "" {
		loadEnvFile(ctx, opts.EnvFile)
	}

	// 设置环境变量绑定
	setupEnvBindings(v)

	// 解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Fatal(ctx, "❌ 配置解析失败", logfacade.Err(err))
	}

	logger.Infof(ctx, "✅ 配置加载完成: 服务端口=%s, 数据库=%s:%d/%s",
		cfg.Server.Port, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

	return &cfg
}

// setupEnvBindings 设置环境变量绑定
func setupEnvBindings(v *viper.Viper) {
	// 启用自动环境变量
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定具体的环境变量（错误可忽略，环境变量绑定为可选配置）
	_ = v.BindEnv("server.port", "PORT")
	_ = v.BindEnv("server.mode", "GIN_MODE")
	_ = v.BindEnv("database.type", "DB_TYPE")
	_ = v.BindEnv("database.file_path", "DB_FILE_PATH")
	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.username", "DB_USERNAME")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.database", "DB_DATABASE")
	_ = v.BindEnv("database.charset", "DB_CHARSET")
	_ = v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	_ = v.BindEnv("database.log_level", "DB_LOG_LEVEL")
	_ = v.BindEnv("database.migrate_on_startup", "DB_MIGRATE_ON_STARTUP")
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.private_key", "JWT_PRIVATE_KEY")
	_ = v.BindEnv("jwt.expiration", "JWT_EXPIRATION")
	_ = v.BindEnv("jwt.issuer", "JWT_ISSUER")
	_ = v.BindEnv("crypto.aes_key", "CRYPTO_AES_KEY")

	// Logger配置环境变量绑定
	_ = v.BindEnv("logger.level", "LOG_LEVEL")
	_ = v.BindEnv("logger.output_path", "LOG_OUTPUT_PATH")
	_ = v.BindEnv("logger.max_size", "LOG_MAX_SIZE")
	_ = v.BindEnv("logger.max_backups", "LOG_MAX_BACKUPS")
	_ = v.BindEnv("logger.max_age", "LOG_MAX_AGE")
	_ = v.BindEnv("logger.compress", "LOG_COMPRESS")

	// 认证配置环境变量绑定
	_ = v.BindEnv("auth.internal_token", "INTERNAL_TOKEN")
	_ = v.BindEnv("auth.cookie.domain", "AUTH_COOKIE_DOMAIN")
	_ = v.BindEnv("auth.cookie.secure", "AUTH_COOKIE_SECURE")
	_ = v.BindEnv("auth.cookie.same_site", "AUTH_COOKIE_SAME_SITE")
	_ = v.BindEnv("auth.design.enabled", "AUTH_DESIGN_ENABLED")
	_ = v.BindEnv("auth.design.jwt_public_key_path", "AUTH_JWT_PUBLIC_KEY_PATH")
	_ = v.BindEnv("auth.design.jwt_public_key", "AUTH_JWT_PUBLIC_KEY")
	_ = v.BindEnv("auth.design.skip_jwt_validation", "AUTH_SKIP_JWT_VALIDATION")
	_ = v.BindEnv("auth.design.accept_modelcraft_jwt", "AUTH_ACCEPT_MODELCRAFT_JWT")
	_ = v.BindEnv("auth.runtime.enabled", "AUTH_RUNTIME_ENABLED")
}

// loadEnvFile 使用 godotenv 加载环境变量文件
func loadEnvFile(ctx context.Context, envFile string) {
	logger := logfacade.GetLogger(ctx)
	// 检查文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		logger.Infof(ctx, "⚠️  环境变量文件 %s 不存在，跳过加载", envFile)
		return
	}

	// 使用 godotenv 加载环境变量文件
	if err := godotenv.Load(envFile); err != nil {
		logger.Infof(ctx, "⚠️  读取环境变量文件 %s 时出错: %v", envFile, err)
	} else {
		logger.Infof(ctx, "✅ 环境变量文件 %s 加载成功", envFile)
	}
}

// loadConfigFile 加载配置文件
func loadConfigFile(ctx context.Context, v *viper.Viper, configFile string) {
	logger := logfacade.GetLogger(ctx)
	// 解析文件路径和扩展名
	dir := filepath.Dir(configFile)
	filename := filepath.Base(configFile)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// 设置配置文件信息
	v.SetConfigName(name)
	if ext != "" {
		v.SetConfigType(strings.TrimPrefix(ext, "."))
	}

	// 添加搜索路径
	if dir != "." && dir != "" {
		v.AddConfigPath(dir)
	}
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// 读取配置文件
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Infof(ctx, "⚠️  配置文件 %s 不存在，使用环境变量和默认配置", configFile)
		} else {
			logger.Infof(ctx, "⚠️  读取配置文件 %s 时出错: %v", configFile, err)
		}
	} else {
		logger.Infof(ctx, "✅ 配置文件加载成功: %s", v.ConfigFileUsed())
	}
}
