import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const TestPage: React.FC = () => {
  const { token } = useAuth();
  const [apiCredentials, setApiCredentials] = useState({
    apiKey: '53zcowwoNVUiRWyg46RQb2nRwSQhUNwYeBEHetyJsjB0MLms1rqxABxWjXClkYoA1',
    secretKey: 'ZkvgnmrJiyOjuix0QxSa1eeQBwvlBENcwY5xBVp9uS085oDUpQHQcxHNfk0dppuj1',
    apiUrl: 'https://testnet.binancefuture.com',
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
    if (!token) {
      setTestResults('Error: Not authenticated. Please log in first.');
      return;
    }
    
    setLoading(true);
    setTestResults('');
    
    try {
      const response = await fetch('/api/test/run', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
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
    <div 
      className="min-h-screen" 
      style={{ background: '#0B0E11', color: '#EAECEF' }}
    >
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div 
          className="rounded-lg p-6" 
          style={{ 
            background: '#181A20', 
            border: '1px solid #2B3139'
          }}
        >
          <h1 className="text-2xl font-bold mb-6" style={{ color: '#EAECEF' }}>
            API测试页面
          </h1>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                API密钥:
              </label>
              <input
                type="password"
                name="apiKey"
                value={apiCredentials.apiKey}
                onChange={handleInputChange}
                className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入API密钥"
                style={{
                  background: '#2B3139',
                  border: '1px solid #3D444D',
                  color: '#EAECEF'
                }}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                密钥:
              </label>
              <input
                type="password"
                name="secretKey"
                value={apiCredentials.secretKey}
                onChange={handleInputChange}
                className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入密钥"
                style={{
                  background: '#2B3139',
                  border: '1px solid #3D444D',
                  color: '#EAECEF'
                }}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                API URL:
              </label>
              <input
                type="text"
                name="apiUrl"
                value={apiCredentials.apiUrl}
                onChange={handleInputChange}
                className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="输入API URL"
                style={{
                  background: '#2B3139',
                  border: '1px solid #3D444D',
                  color: '#EAECEF'
                }}
              />
            </div>
            
            <button
              onClick={runTest}
              disabled={loading}
              className="px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              style={{
                background: '#2B3139',
                color: '#EAECEF',
                border: '1px solid #3D444D'
              }}
              onMouseEnter={(e) => {
                if (!loading) {
                  e.currentTarget.style.background = '#3D444D';
                }
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = '#2B3139';
              }}
            >
              {loading ? '测试中...' : '运行测试'}
            </button>
          </div>
          
          {testResults && (
            <div className="mt-6">
              <h2 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>
                测试结果:
              </h2>
              <pre 
                className="p-4 rounded-md text-sm overflow-auto max-h-96 font-mono" 
                style={{ 
                  background: '#1E293B', 
                  border: '1px solid #334155',
                  color: '#E2E8F0',
                  fontFamily: 'monospace',
                  lineHeight: '1.5'
                }}
              >
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