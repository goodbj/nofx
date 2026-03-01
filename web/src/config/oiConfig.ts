// OI功能配置
export const oiConfig = {
  //检查OI功能是否启用
  isFeatureEnabled: (): boolean => {
    const envValue = import.meta.env.VITE_ENABLE_OI_FEATURE;
    if (envValue !== undefined) {
      return envValue === 'true';
    }
    // 默认启用
    return true;
  },

  // 获取最小OI阈值（百万美元）
  getMinThresholdMillions: (): number => {
    const envValue = import.meta.env.VITE_MIN_OI_THRESHOLD_MILLIONS;
    if (envValue) {
      const threshold = parseFloat(envValue);
      if (!isNaN(threshold)) {
        return threshold;
      }
    }
    // 默认值：15百万美元
    return 15.0;
  },

  //检查是否应该应用OI过滤
  shouldApplyFiltering: (): boolean => {
    // 如果OI功能被禁用，不应用过滤
    if (!oiConfig.isFeatureEnabled()) {
      return false;
    }
    
    // 如果阈值为0，不应用过滤
    const threshold = oiConfig.getMinThresholdMillions();
    return threshold > 0;
  },

  // 获取OI过滤阈值（百万美元）
  getFilterThreshold: (): number => {
    if (!oiConfig.shouldApplyFiltering()) {
      return 0;
    }
    return oiConfig.getMinThresholdMillions();
  }
};

//导出常用的检查函数
export const isOIFeatureEnabled = oiConfig.isFeatureEnabled;
export const shouldApplyOIFiltering = oiConfig.shouldApplyFiltering;
export const getOIFilterThreshold = oiConfig.getFilterThreshold;