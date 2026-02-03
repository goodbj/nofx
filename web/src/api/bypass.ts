export interface BypassSupport {
  modelId: string;
  modelName: string;
  supportsBypass: boolean;
  defaultBypass: boolean;
  description: string;
}

export interface ToggleBypassRequest {
  modelId: string;
  enableBypass: boolean;
  configParams?: Record<string, string>;
}

export interface ToggleBypassResponse {
  modelId: string;
  bypassEnabled: boolean;
  message: string;
  configParams?: Record<string, string>;
}

const API_BASE = import.meta.env.VITE_API_BASE || '';

/**
 * 获取支持API旁路的模型列表
 */
export const bypassAPI = {
  getBypassSupport: async (): Promise<BypassSupport[]> => {
    const response = await fetch(`${API_BASE}/api/bypass/support`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Get bypass support failed: ${response.status}`);
    }

    return response.json();
  },

  /**
   * 切换模型的API旁路设置
   */
  toggleModelBypass: async (
    request: ToggleBypassRequest
  ): Promise<ToggleBypassResponse> => {
    const response = await fetch(`${API_BASE}/api/bypass/toggle`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Toggle bypass failed: ${response.status}`);
    }

    return response.json();
  },
};