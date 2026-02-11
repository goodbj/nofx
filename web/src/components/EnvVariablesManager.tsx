import { useState, useEffect } from 'react';
import { Settings, Eye, EyeOff, Copy, Save, AlertCircle, Loader2 } from 'lucide-react';
import { notify } from '../lib/notify';
import { useAuth } from '../contexts/AuthContext';
import { fetchEnvVariables, updateEnvVariables } from '../api/envApi';

interface EnvVariable {
  key: string;
  value: string;
  category: string;
  description: string;
  isSecret?: boolean;
}

interface Category {
  id: string;
  name: string;
  description: string;
}

const CATEGORIES: Category[] = [
  {
    id: 'server',
    name: 'Server Configuration',
    description: '服务器端口和基本配置'
  },
  {
    id: 'auth',
    name: 'Authentication',
    description: '认证和安全相关配置'
  },
  {
    id: 'encryption',
    name: 'Encryption Keys',
    description: '加密密钥和安全配置'
  },
  {
    id: 'database',
    name: 'Database',
    description: '数据库连接配置'
  },
  {
    id: 'ai',
    name: 'AI Models & Pricing',
    description: 'AI模型和定价配置'
  },
  {
    id: 'guardian',
    name: 'Guardian Automation',
    description: 'Guardian浏览器自动化配置'
  },
  {
    id: 'other',
    name: 'Other Settings',
    description: '其他配置选项'
  }
];

export function EnvVariablesManager() {
  const { token } = useAuth();
  const [envVars, setEnvVars] = useState<EnvVariable[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showSecrets, setShowSecrets] = useState(false);
  const [filteredCategory, setFilteredCategory] = useState<string>('all');
  
  useEffect(() => {
    const loadEnvVariables = async () => {
      try {
        const variables = await fetchEnvVariables(token || undefined);
        setEnvVars(variables);
      } catch (error) {
        console.error('Error loading environment variables:', error);
        notify.error('加载环境变量失败');
      } finally {
        setLoading(false);
      }
    };
    
    loadEnvVariables();
  }, [token]);

  const filteredEnvVars = filteredCategory === 'all' 
    ? envVars 
    : envVars.filter(envVar => envVar.category === filteredCategory);

  const handleCopy = (value: string) => {
    navigator.clipboard.writeText(value);
    notify.success('已复制到剪贴板');
  };

  const handleSaveChanges = async () => {
    setSaving(true);
    try {
      // In a real implementation, this would send the changes to backend
      // For now, just show success message
      await updateEnvVariables(
        envVars.reduce((acc, curr) => {
          acc[curr.key] = curr.value;
          return acc;
        }, {} as Record<string, string>),
        token || undefined
      );
      
      notify.success('环境变量已保存');
    } catch (error) {
      console.error('Error saving environment variables:', error);
      notify.error('保存环境变量失败');
    } finally {
      setSaving(false);
    }
  };

  const toggleSecretVisibility = () => {
    setShowSecrets(!showSecrets);
  };

  const renderValue = (value: string, isSecret?: boolean) => {
    if (isSecret && !showSecrets) {
      return '••••••••••••••••';
    }
    return value;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2 flex items-center gap-3">
          <Settings className="w-8 h-8 text-blue-400" />
          环境变量管理
        </h1>
        <p className="text-gray-400">集中管理和查看所有环境变量配置</p>
      </div>

      <div className="bg-gray-800/50 rounded-xl border border-gray-700 p-6 mb-6">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6">
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => setFilteredCategory('all')}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                filteredCategory === 'all'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              全部
            </button>
            {CATEGORIES.map(category => (
              <button
                key={category.id}
                onClick={() => setFilteredCategory(category.id)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  filteredCategory === category.id
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                {category.name}
              </button>
            ))}
          </div>
          
          <div className="flex items-center gap-3">
            <button
              onClick={toggleSecretVisibility}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-gray-700 text-gray-300 hover:bg-gray-600 transition-colors text-sm"
            >
              {showSecrets ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              {showSecrets ? '隐藏敏感信息' : '显示敏感信息'}
            </button>
            
            <button
              onClick={handleSaveChanges}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-500 transition-colors text-sm"
            >
              <Save className="w-4 h-4" />
              保存更改
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4">
          {filteredEnvVars.map((envVar, index) => (
            <div 
              key={index} 
              className="bg-gray-900/50 rounded-lg border border-gray-700 p-4 hover:border-gray-600 transition-colors"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <code className="text-blue-400 font-mono text-sm break-all">{envVar.key}</code>
                    {envVar.isSecret && (
                      <span className="text-xs px-2 py-1 rounded-full bg-red-900/50 text-red-400">
                        敏感
                      </span>
                    )}
                  </div>
                  <p className="text-gray-400 text-sm mb-2">{envVar.description}</p>
                  
                  <div className="flex items-center gap-2">
                    <code className="text-gray-300 font-mono text-sm break-all bg-gray-800/50 px-2 py-1 rounded">
                      {renderValue(envVar.value, envVar.isSecret)}
                    </code>
                    <button
                      onClick={() => handleCopy(envVar.value)}
                      className="p-1.5 rounded hover:bg-gray-700 transition-colors text-gray-400 hover:text-white"
                      title="复制值"
                    >
                      <Copy className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                
                <div className="text-xs px-3 py-1 rounded-full bg-gray-800 text-gray-400 capitalize">
                  {CATEGORIES.find(cat => cat.id === envVar.category)?.name}
                </div>
              </div>
            </div>
          ))}
        </div>
        
        {filteredEnvVars.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <AlertCircle className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>未找到匹配的环境变量</p>
          </div>
        )}
      </div>

      <div className="bg-blue-900/20 border border-blue-700/50 rounded-lg p-4">
        <h3 className="text-blue-300 font-semibold mb-2 flex items-center gap-2">
          <AlertCircle className="w-5 h-5" />
          重要提示
        </h3>
        <ul className="text-blue-200 text-sm space-y-1">
          <li>• 环境变量控制着系统的各种配置，请谨慎修改</li>
          <li>• 敏感信息如密钥等，建议定期更换以确保安全</li>
          <li>• 修改环境变量后可能需要重启服务才能生效</li>
          <li>• 在生产环境中，建议使用安全的密钥管理系统</li>
        </ul>
      </div>
    </div>
  );
}