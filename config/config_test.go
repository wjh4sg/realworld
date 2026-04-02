package config

import (
	"os"
	"testing"
	"time"
)

// 保存原始环境变量，用于测试后恢复
var originalConfigPath string

// TestMain 用于测试前的设置和测试后的清理
func TestMain(m *testing.M) {
	// 保存原始环境变量
	originalConfigPath = os.Getenv("CONFIG_PATH")

	// 运行测试
	code := m.Run()

	// 恢复原始环境变量
	if originalConfigPath != "" {
		os.Setenv("CONFIG_PATH", originalConfigPath)
	} else {
		os.Unsetenv("CONFIG_PATH")
	}

	// 清理全局缓存
	appConfig = nil

	// 退出测试
	os.Exit(code)
}

// TestLoadConfig_ValidConfig 测试加载有效的配置文件
func TestLoadConfig_ValidConfig(t *testing.T) {
	// 创建临时配置文件
	configContent := `
mysql:
  addr: 192.168.1.1:3306
  username: testuser
  password: testpass
  database: testdb
  max-idle-connections: 50
  max-open-connections: 50
  max-connection-life-time: 30s
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量指向临时配置文件
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 加载配置
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if cfg.MySQL.Addr != "192.168.1.1:3306" {
		t.Errorf("Expected addr to be '192.168.1.1:3306', got '%s'", cfg.MySQL.Addr)
	}
	if cfg.MySQL.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", cfg.MySQL.Username)
	}
	if cfg.MySQL.Password != "testpass" {
		t.Errorf("Expected password to be 'testpass', got '%s'", cfg.MySQL.Password)
	}
	if cfg.MySQL.Database != "testdb" {
		t.Errorf("Expected database to be 'testdb', got '%s'", cfg.MySQL.Database)
	}
	if cfg.MySQL.MaxIdleConnections != 50 {
		t.Errorf("Expected max-idle-connections to be 50, got %d", cfg.MySQL.MaxIdleConnections)
	}
	if cfg.MySQL.MaxOpenConnections != 50 {
		t.Errorf("Expected max-open-connections to be 50, got %d", cfg.MySQL.MaxOpenConnections)
	}
	if cfg.MySQL.MaxConnectionLifeTime != 30*time.Second {
		t.Errorf("Expected max-connection-life-time to be 30s, got %s", cfg.MySQL.MaxConnectionLifeTime)
	}
}

// TestLoadConfig_ConfigFileNotFound 测试配置文件不存在的情况
func TestLoadConfig_ConfigFileNotFound(t *testing.T) {
	// 设置环境变量指向不存在的文件
	os.Setenv("CONFIG_PATH", "./non_existent_config.yaml")
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 加载配置，应该返回错误
	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("Expected error when config file not found, but got nil")
	}

	// 检查错误消息是否包含预期的前缀，不依赖于具体的系统错误消息格式
	if err.Error()[:28] != "failed to read config file: " {
		t.Errorf("Expected error message to start with 'failed to read config file: ', got '%s'", err.Error())
	}
}

// TestLoadConfig_InvalidYAML 测试配置文件格式错误的情况
func TestLoadConfig_InvalidYAML(t *testing.T) {
	// 创建真正格式错误的临时配置文件 - 缺少冒号
	invalidContent := `
mysql:
  addr 192.168.1.1:3306 # 缺少冒号，这是真正无效的YAML
  username: testuser
  password: testpass
  database: testdb
  max-idle-connections: 50
  max-open-connections: 50
  max-connection-life-time: 30s
`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量指向临时配置文件
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 加载配置，应该返回错误
	_, err = LoadConfig()
	if err == nil {
		t.Fatalf("Expected error when config file has invalid YAML, but got nil")
	}

	if err.Error()[:29] != "failed to parse config file: " {
		t.Errorf("Expected error message to start with 'failed to parse config file: ', got '%s'", err.Error())
	}
}

// TestApplyDefaults 测试默认值应用
func TestApplyDefaults(t *testing.T) {
	// 创建临时配置文件，只包含部分配置项
	partialConfig := `
mysql:
  addr: 192.168.1.1:3306
  username: testuser
  # 其他配置项缺失，应该应用默认值
`

	tmpFile, err := os.CreateTemp("", "partial-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(partialConfig); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量指向临时配置文件
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 加载配置
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证部分配置项
	if cfg.MySQL.Addr != "192.168.1.1:3306" {
		t.Errorf("Expected addr to be '192.168.1.1:3306', got '%s'", cfg.MySQL.Addr)
	}
	if cfg.MySQL.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", cfg.MySQL.Username)
	}

	// 验证默认值
	if cfg.MySQL.Password != "" {
		t.Errorf("Expected password to be empty (no default), got '%s'", cfg.MySQL.Password)
	}
	if cfg.MySQL.Database != "app" {
		t.Errorf("Expected database to be 'app' (default), got '%s'", cfg.MySQL.Database)
	}
	if cfg.MySQL.MaxIdleConnections != 100 {
		t.Errorf("Expected max-idle-connections to be 100 (default), got %d", cfg.MySQL.MaxIdleConnections)
	}
	if cfg.MySQL.MaxOpenConnections != 100 {
		t.Errorf("Expected max-open-connections to be 100 (default), got %d", cfg.MySQL.MaxOpenConnections)
	}
	if cfg.MySQL.MaxConnectionLifeTime != 10*time.Second {
		t.Errorf("Expected max-connection-life-time to be 10s (default), got %s", cfg.MySQL.MaxConnectionLifeTime)
	}
}

// TestGetConfig_Cache 测试 GetConfig 的缓存功能
func TestGetConfig_Cache(t *testing.T) {
	// 创建临时配置文件
	configContent := `
mysql:
  addr: 192.168.1.1:3306
  username: testuser
  password: testpass
  database: testdb
`

	tmpFile, err := os.CreateTemp("", "cache-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量指向临时配置文件
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 第一次调用 GetConfig
	cfg1 := GetConfig()
	if cfg1 == nil {
		t.Fatalf("Expected non-nil config, got nil")
	}

	// 第二次调用 GetConfig，应该返回相同的实例
	cfg2 := GetConfig()
	if cfg2 == nil {
		t.Fatalf("Expected non-nil config, got nil")
	}

	// 验证是否为同一实例
	if cfg1 != cfg2 {
		t.Errorf("Expected cfg1 and cfg2 to be the same instance, but they are different")
	}
}

// TestGetMySQLConfig 测试 GetMySQLConfig 函数
func TestGetMySQLConfig(t *testing.T) {
	// 创建临时配置文件
	configContent := `
mysql:
  addr: 192.168.1.1:3306
  username: testuser
  password: testpass
  database: testdb
  max-idle-connections: 50
  max-open-connections: 50
  max-connection-life-time: 30s
`

	tmpFile, err := os.CreateTemp("", "mysql-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	tmpFile.Close()

	// 设置环境变量指向临时配置文件
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// 清理全局缓存
	appConfig = nil

	// 加载配置
	mysqlConfig := GetMySQLConfig()

	// 验证 MySQL 配置
	if mysqlConfig.Addr != "192.168.1.1:3306" {
		t.Errorf("Expected addr to be '192.168.1.1:3306', got '%s'", mysqlConfig.Addr)
	}
	if mysqlConfig.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", mysqlConfig.Username)
	}
	if mysqlConfig.Password != "testpass" {
		t.Errorf("Expected password to be 'testpass', got '%s'", mysqlConfig.Password)
	}
	if mysqlConfig.Database != "testdb" {
		t.Errorf("Expected database to be 'testdb', got '%s'", mysqlConfig.Database)
	}
	if mysqlConfig.MaxIdleConnections != 50 {
		t.Errorf("Expected max-idle-connections to be 50, got %d", mysqlConfig.MaxIdleConnections)
	}
	if mysqlConfig.MaxOpenConnections != 50 {
		t.Errorf("Expected max-open-connections to be 50, got %d", mysqlConfig.MaxOpenConnections)
	}
	if mysqlConfig.MaxConnectionLifeTime != 30*time.Second {
		t.Errorf("Expected max-connection-life-time to be 30s, got %s", mysqlConfig.MaxConnectionLifeTime)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	configContent := `
server:
  port: "8080"
  rate-limit-rps: 100
mysql:
  addr: 127.0.0.1:3306
  username: yaml-user
  password: yaml-pass
  database: yaml-db
  max-idle-connections: 10
  max-open-connections: 20
  max-connection-life-time: 5s
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
jwt:
  secret: yaml-secret
`

	tmpFile, err := os.CreateTemp("", "env-override-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}
	tmpFile.Close()

	os.Setenv("CONFIG_PATH", tmpFile.Name())
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_RATE_LIMIT_RPS", "200")
	os.Setenv("MYSQL_ADDR", "mysql:3306")
	os.Setenv("MYSQL_USERNAME", "env-user")
	os.Setenv("MYSQL_PASSWORD", "env-pass")
	os.Setenv("MYSQL_DATABASE", "env-db")
	os.Setenv("MYSQL_MAX_IDLE_CONNECTIONS", "30")
	os.Setenv("MYSQL_MAX_OPEN_CONNECTIONS", "40")
	os.Setenv("MYSQL_MAX_CONNECTION_LIFE_TIME", "15s")
	os.Setenv("REDIS_ADDR", "redis:6379")
	os.Setenv("REDIS_PASSWORD", "env-redis-pass")
	os.Setenv("REDIS_DB", "3")
	os.Setenv("JWT_SECRET", "env-secret")
	defer os.Unsetenv("CONFIG_PATH")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("SERVER_RATE_LIMIT_RPS")
	defer os.Unsetenv("MYSQL_ADDR")
	defer os.Unsetenv("MYSQL_USERNAME")
	defer os.Unsetenv("MYSQL_PASSWORD")
	defer os.Unsetenv("MYSQL_DATABASE")
	defer os.Unsetenv("MYSQL_MAX_IDLE_CONNECTIONS")
	defer os.Unsetenv("MYSQL_MAX_OPEN_CONNECTIONS")
	defer os.Unsetenv("MYSQL_MAX_CONNECTION_LIFE_TIME")
	defer os.Unsetenv("REDIS_ADDR")
	defer os.Unsetenv("REDIS_PASSWORD")
	defer os.Unsetenv("REDIS_DB")
	defer os.Unsetenv("JWT_SECRET")

	appConfig = nil

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Fatalf("Expected overridden server port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.RateLimitRPS != 200 {
		t.Fatalf("Expected overridden rate limit 200, got %d", cfg.Server.RateLimitRPS)
	}
	if cfg.MySQL.Addr != "mysql:3306" || cfg.MySQL.Username != "env-user" || cfg.MySQL.Password != "env-pass" || cfg.MySQL.Database != "env-db" {
		t.Fatalf("Expected MySQL env overrides to apply, got %+v", cfg.MySQL)
	}
	if cfg.MySQL.MaxIdleConnections != 30 || cfg.MySQL.MaxOpenConnections != 40 || cfg.MySQL.MaxConnectionLifeTime != 15*time.Second {
		t.Fatalf("Expected MySQL pool env overrides to apply, got %+v", cfg.MySQL)
	}
	if cfg.Redis.Addr != "redis:6379" || cfg.Redis.Password != "env-redis-pass" || cfg.Redis.DB != 3 {
		t.Fatalf("Expected Redis env overrides to apply, got %+v", cfg.Redis)
	}
	if cfg.JWT.Secret != "env-secret" {
		t.Fatalf("Expected JWT secret override to apply, got %s", cfg.JWT.Secret)
	}
}
