import React, { useState, useEffect } from 'react';
import { bypassAPI, BypassSupport } from '../api/bypass';

interface ModelBypassToggleProps {
  modelId: string;
  onBypassChange: (modelId: string, bypassEnabled: boolean) => void;
  initialBypassState?: boolean;
}

const ModelBypassToggle: React.FC<ModelBypassToggleProps> = ({
  modelId,
  onBypassChange,
  initialBypassState = false,
}) => {
  const [bypassEnabled, setBypassEnabled] = useState<boolean>(initialBypassState);
  const [loading, setLoading] = useState<boolean>(false);
  const [bypassInfo, setBypassInfo] = useState<BypassSupport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBypassInfo = async () => {
      try {
        const supportList = await bypassAPI.getBypassSupport();
        const modelInfo = supportList.find(info => info.modelId === modelId);
        setBypassInfo(modelInfo || null);
      } catch (err) {
        setError('Failed to fetch bypass support info');
        console.error('Error fetching bypass support:', err);
      }
    };

    fetchBypassInfo();
  }, [modelId]);

  const handleToggle = async () => {
    setLoading(true);
    setError(null);

    try {
      const result = await bypassAPI.toggleModelBypass({
        modelId,
        enableBypass: !bypassEnabled,
      });

      const newBypassState = result.bypassEnabled;
      setBypassEnabled(newBypassState);
      onBypassChange(modelId, newBypassState);
    } catch (err) {
      setError('Failed to toggle bypass setting');
      console.error('Error toggling bypass:', err);
    } finally {
      setLoading(false);
    }
  };

  if (!bypassInfo || !bypassInfo.supportsBypass) {
    return null; // Don't render if model doesn't support bypass
  }

  return (
    <div className="model-bypass-toggle">
      <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg mb-2">
        <div className="flex items-center">
          <input
            type="checkbox"
            id={`bypass-${modelId}`}
            checked={bypassEnabled}
            onChange={handleToggle}
            disabled={loading}
            className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 focus:ring-2"
          />
          <label htmlFor={`bypass-${modelId}`} className="ml-2 text-sm font-medium text-gray-700">
            <span className="font-semibold">绕过API</span> - 使用浏览器自动化
          </label>
        </div>
        
        {loading && (
          <span className="text-xs text-blue-600">处理中...</span>
        )}
      </div>
      
      <div className="text-xs text-gray-500 ml-6 mb-2">
        {bypassEnabled 
          ? '✅ 启用浏览器自动化，绕过API调用' 
          : '📡 使用标准API调用'}
      </div>
      
      {error && (
        <div className="text-xs text-red-600 ml-6">{error}</div>
      )}
    </div>
  );
};

export default ModelBypassToggle;