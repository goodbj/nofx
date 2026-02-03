interface GuardianCallRequest {
  provider: string
  systemPrompt: string
  userPrompt: string
  targetUrl?: string
}

interface GuardianCallResponse {
  success: boolean
  response: string
  provider: string
}

interface GuardianProvidersResponse {
  providers: string[]
}

interface GuardianLoginCheckRequest {
  provider: string
  targetUrl?: string
}

interface GuardianLoginCheckResponse {
  success: boolean
  isLoggedIn: boolean
  provider: string
  targetUrl: string
  description: string
}

const API_BASE = import.meta.env.VITE_API_BASE || ''

export const guardianAPI = {
  // Call a Guardian AI provider
  callGuardianAI: async (request: GuardianCallRequest): Promise<GuardianCallResponse> => {
    const response = await fetch(`${API_BASE}/api/guardian/call`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      throw new Error(`Guardian AI call failed: ${response.status}`)
    }

    return response.json()
  },

  // Get list of available Guardian AI providers
  getGuardianProviders: async (): Promise<GuardianProvidersResponse> => {
    const response = await fetch(`${API_BASE}/api/guardian/providers`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`Get Guardian providers failed: ${response.status}`)
    }

    return response.json()
  },

  // Check login status for a Guardian AI provider
  checkLoginStatus: async (request: GuardianLoginCheckRequest): Promise<GuardianLoginCheckResponse> => {
    const response = await fetch(`${API_BASE}/api/guardian/check-login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      throw new Error(`Check Guardian login status failed: ${response.status}`)
    }

    return response.json()
  },

  // Get target URL for a provider
  getTargetUrl: (provider: string): string => {
    const urls: Record<string, string> = {
      'deepseek-browser': 'https://chat.deepseek.com',
      'chatgpt-browser': 'https://chat.openai.com',
      'claude-browser': 'https://claude.ai',
      'qwen-browser': 'https://tongyi.aliyun.com/qwen/',
      'gemini-browser': 'https://gemini.google.com',
      'grok-browser': 'https://grok.x.ai',
      'kimi-browser': 'https://kimi.moonshot.cn',
      'ollama-browser': 'http://localhost:11434',
      'guardian-ai': 'https://chat.deepseek.com'
    };
    return urls[provider] || 'https://chat.deepseek.com';
  },
}