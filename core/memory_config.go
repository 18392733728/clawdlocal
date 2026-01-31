package core

import "clawdlocal/config"

// ConvertConfigToMemoryConfig converts config.MemoryConfig to core.MemoryConfig
func ConvertConfigToMemoryConfig(cfg config.MemoryConfig) *MemoryConfig {
	return &MemoryConfig{
		ShortTermCapacity: cfg.ShortTerm.MaxEntries,
		LongTermFile:      cfg.LongTerm.StorageDir + "/long_term.json",
		CleanupInterval:   5, // 5 minutes
	}
}