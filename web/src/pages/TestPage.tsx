import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const TestPage: React.FC = () => {
  const [apiCredentials, setApiCredentials] = useState({
    apiKey: '',
    secretKey: '',
    apiUrl: '',
  });
  const [testResults, setTestResults] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setApiCredentials(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const runTest = async () => {
    setLoading(true);
    setTestResults('');
    
    try {
      const response = await fetch('/api/test/run', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          command: 'go run test/test_binance_testnet.go',
          apiKey: apiCredentials.apiKey,
          secretKey: apiCredentials.secretKey,
          apiUrl: apiCredentials.apiUrl,
        }),
      });

      const data = await response.json();
      setTestResults(`Status: ${data.status}\nOutput: ${data.output}`);
    } catch (error) {
      setTestResults(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        <div className="bg-white rounded-lg shadow-md p-6">
          <h1 className="text-2xl font-bold text-gray-800 mb-6">API测试页面</h1>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API密钥:
              </label>
              <input
                type="password"
                name="apiKey"
                value={apiCredentials.apiKey}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入API密钥"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                密钥:
              </label>
              <input
                type="password"
                name="secretKey"
                value={apiCredentials.secretKey}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入密钥"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API URL:
              </label>
              <input
                type="text"
                name="apiUrl"
                value={apiCredentials.apiUrl}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入API URL"
              />
            </div>
            
            <button
              onClick={runTest}
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {loading ? '测试中...' : '运行测试'}
            </button>
          </div>
          
          {testResults && (
            <div className="mt-6">
              <h2 className="text-lg font-semibold text-gray-800 mb-2">测试结果:</h2>
              <pre className="bg-gray-100 p-4 rounded-md text-sm overflow-auto max-h-96">
                {testResults}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TestPage;