package common

import (
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

// getRealWorldMySQLConfig 获取 realworld 数据库的实际配置
func getRealWorldMySQLConfig() *MySQLOptions {
	return &MySQLOptions{
		Addr:                  "127.0.0.1:3306",
		Username:              "realworld",
		Password:              "test-password",
		Database:              "realworld",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: 10 * time.Second,
		Logger:                logger.Default,
	}
}

// 测试 setMySQLDefaults 函数
func TestSetMySQLDefaults(t *testing.T) {
	// 测试用例 1：所有字段为空时的默认值设置
	t.Run("All fields empty", func(t *testing.T) {
		opts := &MySQLOptions{}
		setMySQLDefaults(opts)

		// 验证默认值
		if opts.Addr != "127.0.0.1:3306" {
			t.Errorf("Expected Addr to be '127.0.0.1:3306', got '%s'", opts.Addr)
		}
		if opts.MaxIdleConnections != 100 {
			t.Errorf("Expected MaxIdleConnections to be 100, got %d", opts.MaxIdleConnections)
		}
		if opts.MaxOpenConnections != 100 {
			t.Errorf("Expected MaxOpenConnections to be 100, got %d", opts.MaxOpenConnections)
		}
		if opts.MaxConnectionLifeTime != 10*time.Second {
			t.Errorf("Expected MaxConnectionLifeTime to be 10s, got %v", opts.MaxConnectionLifeTime)
		}
		if opts.Logger == nil {
			t.Error("Expected Logger to be set, got nil")
		}
	})

	// 测试用例 2：部分字段为空时的默认值设置
	t.Run("Partial fields empty", func(t *testing.T) {
		customAddr := "custom-host:3306"
		opts := &MySQLOptions{
			Addr: customAddr, // 非空字段
			// 其他字段为空
		}
		setMySQLDefaults(opts)

		// 验证非空字段保持不变
		if opts.Addr != customAddr {
			t.Errorf("Expected Addr to remain '%s', got '%s'", customAddr, opts.Addr)
		}
		// 验证空字段被设置默认值
		if opts.MaxIdleConnections != 100 {
			t.Errorf("Expected MaxIdleConnections to be 100, got %d", opts.MaxIdleConnections)
		}
	})
}

// 测试 DSN 方法
func TestDSN(t *testing.T) {
	// 使用 realworld 数据库配置
	opts := getRealWorldMySQLConfig()

	expectedDSN := "realworld:test-password@tcp(127.0.0.1:3306)/realworld?charset=utf8&parseTime=true&loc=Local"
	actualDSN := opts.DSN()

	if actualDSN != expectedDSN {
		t.Errorf("Expected DSN to be '%s', got '%s'", expectedDSN, actualDSN)
	}
}

// 测试 MustRawDB 函数
func TestMustRawDB(t *testing.T) {
	// 使用 realworld 数据库配置
	opts := getRealWorldMySQLConfig()

	db, err := NewMySQL(opts)
	if err != nil {
		t.Skipf("Skipping test: failed to create database connection: %v", err)
	}

	// 测试 MustRawDB
	sqlDB := MustRawDB(db)
	if sqlDB == nil {
		t.Error("Expected sqlDB to be non-nil, got nil")
	}
}

// 集成测试：使用 realworld 数据库测试完整的连接流程
func TestNewMySQL(t *testing.T) {
	// 使用 realworld 数据库配置
	opts := getRealWorldMySQLConfig()

	db, err := NewMySQL(opts)
	if err != nil {
		t.Skipf("Skipping integration-style database connection test: %v", err)
	}

	// 验证连接成功
	if db == nil {
		t.Error("Expected db to be non-nil, got nil")
	}

	// 验证连接池参数设置正确
	sqlDB := MustRawDB(db)
	if sqlDB == nil {
		t.Error("Expected sqlDB to be non-nil, got nil")
	}

	// 验证连接可用性：执行简单查询
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		t.Errorf("Expected query to succeed, got error: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected query result to be 1, got %d", result)
	}
}
