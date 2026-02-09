import axios from 'axios';

// 定义类型接口
export interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: 'super_admin' | 'admin' | 'moderator';
  permissions: string[];
  status: 'active' | 'locked' | 'inactive';
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  id: string;
  user_id: string;
  username: string;
  action: string;
  resource: string;
  ip_address: string;
  user_agent: string;
  timestamp: string;
  details: Record<string, any>;
}

export interface SystemConfig {
  key: string;
  value: string;
  description: string;
  updated_at: string;
}

export interface DashboardStats {
  total_users: number;
  active_users: number;
  locked_users: number;
  recent_logs: AuditLog[];
  system_status: {
    uptime: string;
    cpu_usage: number;
    memory_usage: number;
  };
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 创建axios实例
const api = axios.create({
  baseURL: '/api', // 代理到后端API
  timeout: 10000,
});

// 请求拦截器 - 添加认证令牌
api.interceptors.request.use(
  (config: any) => {
    const token = localStorage.getItem('admin_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理认证错误
api.interceptors.response.use(
  (response: any) => {
    return response;
  },
  (error: any) => {
    if (error.response?.status === 401) {
      // 认证失败，清除令牌并跳转到登录页
      localStorage.removeItem('admin_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// 认证相关API
export const authAPI = {
  login: (credentials: { username: string; password: string }) =>
    api.post('/auth/login', credentials),
  
  refreshToken: () => 
    api.post('/auth/refresh'),
};

// 用户管理API
export const userAPI = {
  getUsers: (params?: { page?: number; limit?: number; search?: string; status?: string; role?: string }) =>
    api.get('/users', { params }),
  
  getUser: (id: string) =>
    api.get(`/users/${id}`),
  
  createUser: (userData: { username: string; email: string; password: string; role: string; status: string }) =>
    api.post('/users', userData),
  
  updateUser: (id: string, userData: Partial<{ username: string; email: string; password: string; role: string; status: string }>) =>
    api.put(`/users/${id}`, userData),
  
  deleteUser: (id: string) =>
    api.delete(`/users/${id}`),
  
  lockUser: (id: string) =>
    api.post(`/users/${id}/lock`),
  
  unlockUser: (id: string) =>
    api.post(`/users/${id}/unlock`),
};

// 权限管理API
export const permissionAPI = {
  getPermissions: (params?: { page?: number; limit?: number; user_id?: string }) =>
    api.get('/permissions', { params }),
  
  grantPermission: (data: { user_id: string; permission: string; reason?: string }) =>
    api.post('/permissions', data),
  
  revokePermission: (id: string) =>
    api.delete(`/permissions/${id}`),
  
  assignRole: (data: { user_id: string; role: string; reason?: string }) =>
    api.post('/permissions/assign-role', data),
};

// 仪表板API
export const dashboardAPI = {
  getDashboard: () =>
    api.get('/dashboard'),
  
  getProfile: () =>
    api.get('/profile'),
  
  updateProfile: (data: { email?: string; password?: string }) =>
    api.put('/profile', data),
  
  changePassword: (data: { old_password: string; new_password: string }) =>
    api.put('/profile/change-password', data),
};

// 系统配置API
export const systemAPI = {
  getConfig: () =>
    api.get('/system/config'),
  
  updateConfig: (data: any) =>
    api.put('/system/config', data),
};

// 监控API
export const monitoringAPI = {
  getStatus: () =>
    api.get('/monitoring/status'),
  
  getMetrics: () =>
    api.get('/monitoring/metrics'),
};

// 审计日志API
export const auditAPI = {
  getLogs: (params?: { page?: number; limit?: number; start_date?: string; end_date?: string; user_id?: string; username?: string; action?: string; ip_address?: string }) =>
    api.get('/audit', { params }),
};