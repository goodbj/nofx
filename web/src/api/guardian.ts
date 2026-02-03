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
}