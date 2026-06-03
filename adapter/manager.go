package adapter

import (
	"fmt"
	"sync"

	"github.com/jingyuexing/stocks/compiler"
	"github.com/jingyuexing/stocks/gateway"
)

// Manager 管理交易网关的生命周期：创建、连接、主备切换
type Manager struct {
	mu       sync.RWMutex
	primary  gateway.TradingGateway
	fallback gateway.TradingGateway
	config   compiler.AdapterConfig
}

// NewManager 创建 AdapterManager
func NewManager() *Manager {
	return &Manager{}
}

// Load 根据 AdapterConfig 加载并连接主网关
func (m *Manager) Load(cfg compiler.AdapterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg

	if cfg.Primary == "" {
		return fmt.Errorf("adapter primary is empty")
	}

	gw, err := gateway.Create(cfg.Primary)
	if err != nil {
		return fmt.Errorf("create primary gateway failed: %w", err)
	}

	// 解析 adapter_config 中的 JSON 或 key-value 为 map[string]any
	params := parseParams(cfg.Params)
	if err := gw.Connect(params); err != nil {
		return fmt.Errorf("connect primary gateway failed: %w", err)
	}
	m.primary = gw

	// 加载备用网关
	if cfg.Fallback != "" {
		fw, err := gateway.Create(cfg.Fallback)
		if err == nil {
			_ = fw.Connect(params)
			m.fallback = fw
		}
	}

	return nil
}

// Primary 返回主网关
func (m *Manager) Primary() gateway.TradingGateway {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primary
}

// Fallback 返回备用网关
func (m *Manager) Fallback() gateway.TradingGateway {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fallback
}

// Close 断开所有网关连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.primary != nil {
		_ = m.primary.Disconnect()
		m.primary = nil
	}
	if m.fallback != nil {
		_ = m.fallback.Disconnect()
		m.fallback = nil
	}
	return nil
}

// IsLoaded 是否已加载
func (m *Manager) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primary != nil
}

// parseParams 将参数字符串简单解析为 map[string]any（骨架，后续扩展 JSON 解析）
func parseParams(raw string) map[string]any {
	m := make(map[string]any)
	if raw == "" {
		return m
	}
	// 骨架：直接返回原始文本
	m["raw"] = raw
	return m
}
