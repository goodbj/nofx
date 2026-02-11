import { API_BASE } from '../constants/api'

interface EnvVariable {
  key: string;
  value: string;
  category: string;
  description: string;
  isSecret?: boolean;
}

/**
 * 获取所有环境变量
 */
export const fetchEnvVariables = async (token?: string): Promise<EnvVariable[]> => {
  try {
    const response = await fetch(`${API_BASE}/api/env-variables`, {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch environment variables: ${response.status}`);
    }

    const data = await response.json();
    return data.variables || [];
  } catch (error) {
    console.error('Error fetching environment variables:', error);
    // Return empty array in case of error
    return [];
  }
};

/**
 * 更新环境变量
 */
export const updateEnvVariables = async (envVars: Partial<Record<string, string>>, token?: string) => {
  try {
    const response = await fetch(`${API_BASE}/api/env-variables`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: JSON.stringify({ variables: envVars }),
    });

    if (!response.ok) {
      throw new Error(`Failed to update environment variables: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error updating environment variables:', error);
    throw error;
  }
};