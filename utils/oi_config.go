package utils

import (
	"nofx/config"
	"os"
	"strconv"
	"strings"
)

// IsOIFeatureEnabled检查OI功能是否启用
func IsOIFeatureEnabled() bool {
	//首先检查全局配置
	cfg := config.Get()
	if cfg != nil {
		return cfg.EnableOIFeature
	}

	// 如果配置不可用，检查环境变量
	envValue := os.Getenv("ENABLE_OI_FEATURE")
	if envValue != "" {
		return strings.ToLower(envValue) == "true"
	}

	// 默认启用
	return true
}

// GetMinOIThresholdMillions 获取最小OI阈值（百万美元）
func GetMinOIThresholdMillions() float64 {
	//首先检查全局配置
	cfg := config.Get()
	if cfg != nil {
		return cfg.MinOIThresholdMillions
	}

	// 如果配置不可用，检查环境变量
	envValue := os.Getenv("MIN_OI_THRESHOLD_MILLIONS")
	if envValue != "" {
		if threshold, err := strconv.ParseFloat(envValue, 64); err == nil {
			return threshold
		}
	}

	// 默认值：15百万美元
	return 15.0
}

// ShouldApplyOIFiltering检查是否应该应用OI过滤
func ShouldApplyOIFiltering() bool {
	// 如果OI功能被禁用，不应用过滤
	if !IsOIFeatureEnabled() {
		return false
	}

	// 如果阈值为0，不应用过滤
	threshold := GetMinOIThresholdMillions()
	return threshold > 0
}

// GetOIFilterThreshold 获取OI过滤阈值（百万美元）
func GetOIFilterThreshold() float64 {
	if !ShouldApplyOIFiltering() {
		return 0
	}
	return GetMinOIThresholdMillions()
}
