package api

import (
	"sync"
	"time"
)

// AccountCache 账户信息缓存结构
type AccountCache struct {
	data       map[string]interface{}
	timestamp  time.Time
	expireTime time.Time
}

// AccountCacheManager 账户缓存管理器
type AccountCacheManager struct {
	cache    map[string]*AccountCache
	mu       sync.RWMutex
	duration time.Duration
}

// NewAccountCacheManager 创建新的账户缓存管理器
func NewAccountCacheManager(cacheDuration time.Duration) *AccountCacheManager {
	return &AccountCacheManager{
		cache:    make(map[string]*AccountCache),
		duration: cacheDuration,
	}
}

// Get 从缓存获取账户信息
func (acm *AccountCacheManager) Get(traderID string) (map[string]interface{}, bool) {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	cache, exists := acm.cache[traderID]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(cache.expireTime) {
		// 缓存过期，删除并返回不存在
		delete(acm.cache, traderID)
		return nil, false
	}

	// 返回缓存数据的副本
	result := make(map[string]interface{})
	for k, v := range cache.data {
		result[k] = v
	}

	return result, true
}

// Set 设置账户信息缓存
func (acm *AccountCacheManager) Set(traderID string, data map[string]interface{}) {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	acm.cache[traderID] = &AccountCache{
		data:       data,
		timestamp:  time.Now(),
		expireTime: time.Now().Add(acm.duration),
	}
}

// Clear 清除指定交易员的缓存
func (acm *AccountCacheManager) Clear(traderID string) {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	delete(acm.cache, traderID)
}

// ClearAll 清除所有缓存
func (acm *AccountCacheManager) ClearAll() {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	acm.cache = make(map[string]*AccountCache)
}

// GetCacheStats 获取缓存统计信息
func (acm *AccountCacheManager) GetCacheStats() map[string]interface{} {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_entries"] = len(acm.cache)
	stats["cache_duration"] = acm.duration.String()

	// 计算过期条目数
	expiredCount := 0
	now := time.Now()
	for _, cache := range acm.cache {
		if now.After(cache.expireTime) {
			expiredCount++
		}
	}
	stats["expired_entries"] = expiredCount

	return stats
}
