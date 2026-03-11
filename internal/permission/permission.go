package permission

import (
	"os"
	"sync"

	"mcp-agent/pkg/logger"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type PermissionConfig struct {
	Permissions []RolePermission `yaml:"permissions"`
}

type RolePermission struct {
	Role  string   `yaml:"role"`
	Tools []string `yaml:"tools"`
}

type Manager struct {
	mu          sync.RWMutex
	permissions map[string]map[string]bool // role -> tool -> allowed
	configPath  string
}

func NewManager(configPath string) *Manager {
	m := &Manager{
		permissions: make(map[string]map[string]bool),
		configPath:  configPath,
	}
	if err := m.Load(); err != nil {
		logger.Warn("failed to load permissions config, using defaults", zap.Error(err))
	}
	return m
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var cfg PermissionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.permissions = make(map[string]map[string]bool)
	for _, perm := range cfg.Permissions {
		toolMap := make(map[string]bool)
		for _, tool := range perm.Tools {
			toolMap[tool] = true
		}
		m.permissions[perm.Role] = toolMap
	}

	logger.Info("permissions loaded", zap.Int("roles", len(cfg.Permissions)))
	return nil
}

func (m *Manager) CanAccess(role, toolName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools, ok := m.permissions[role]
	if !ok {
		return false
	}

	if tools["all"] {
		return true
	}

	return tools[toolName]
}

func (m *Manager) GetRoleTools(role string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools, ok := m.permissions[role]
	if !ok {
		return nil
	}

	var result []string
	for tool := range tools {
		result = append(result, tool)
	}
	return result
}
