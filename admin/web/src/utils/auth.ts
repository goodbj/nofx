// 认证相关的工具函数

// 检查用户是否已认证
export const isAuthenticated = (): boolean => {
  const token = localStorage.getItem('admin_token');
  return !!token;
};

// 获取当前用户信息
export const getCurrentUser = (): any => {
  const token = localStorage.getItem('admin_token');
  if (!token) return null;

  try {
    // 解析JWT token payload
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );

    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error('解析用户信息失败:', error);
    return null;
  }
};

// 检查用户权限
export const hasPermission = (permission: string): boolean => {
  const user = getCurrentUser();
  if (!user) return false;

  // 实际权限检查应该从后端获取，这里简化处理
  // 在真实场景中，应该调用API获取用户的权限列表
  return true; // 简化处理，实际应检查权限
};

// 检查用户角色
export const hasRole = (role: string): boolean => {
  const user = getCurrentUser();
  if (!user) return false;

  return user.role === role;
};

// 检查是否为超级管理员
export const isSuperAdmin = (): boolean => {
  return hasRole('super_admin');
};

// 登出
export const logout = (): void => {
  localStorage.removeItem('admin_token');
  window.location.href = '/login';
};