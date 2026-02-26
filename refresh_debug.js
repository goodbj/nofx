// refresh_debug.js - 决策刷新调试工具
// 在浏览器控制台中运行此代码来测试刷新机制

(function debugRefresh() {
  console.log('🔍 开始调试决策刷新机制...');
  
  // 检查SWR状态
  const checkSWR = () => {
    console.log('=== SWR状态检查 ===');
    
    // 查找相关的SWR缓存
    const swrCache = {};
    for (let key in localStorage) {
      if (key.startsWith('swr')) {
        swrCache[key] = localStorage.getItem(key);
      }
    }
    
    console.log('发现的SWR缓存项:', Object.keys(swrCache));
    
    // 检查页面中的SWR实例
    if (window.__SWR) {
      console.log('页面中存在的SWR实例:', window.__SWR);
    }
    
    return swrCache;
  };
  
  // 模拟手动刷新
  const simulateRefresh = async () => {
    console.log('=== 模拟手动刷新 ===');
    
    try {
      const startTime = Date.now();
      const response = await fetch('/api/decisions/latest?trader_id=xxx&limit=10', {
        headers: {
          'Authorization': 'Bearer ' + localStorage.getItem('token')
        }
      });
      
      const endTime = Date.now();
      const data = await response.json();
      
      console.log('✅ API调用成功:', {
        耗时: `${endTime - startTime}ms`,
        数据长度: data.length,
        最新周期号: data[0]?.cycle_number,
        状态码: response.status
      });
      
      return data;
    } catch (error) {
      console.error('❌ API调用失败:', error);
      return null;
    }
  };
  
  // 检查React组件状态
  const checkReactState = () => {
    console.log('=== React组件状态检查 ===');
    
    // 查找相关的React组件
    const components = [];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_ELEMENT
    );
    
    let node;
    while (node = walker.nextNode()) {
      if (node.__reactFiber$) {
        const fiber = node.__reactFiber$;
        if (fiber.memoizedProps && fiber.memoizedProps.decisions) {
          components.push({
            element: node,
            props: fiber.memoizedProps,
            decisions: fiber.memoizedProps.decisions
          });
        }
      }
    }
    
    console.log('找到的决策组件:', components.length);
    if (components.length > 0) {
      console.log('组件props:', components[0].props);
    }
    
    return components;
  };
  
  // 运行所有检查
  const runAllChecks = async () => {
    console.log('🚀 开始全面检查...');
    
    checkSWR();
    await simulateRefresh();
    checkReactState();
    
    console.log('✅ 检查完成');
  };
  
  // 暴露到全局作用域
  window.debugRefresh = {
    checkSWR,
    simulateRefresh,
    checkReactState,
    runAllChecks
  };
  
  console.log('✅ 调试工具已加载，可使用以下命令:');
  console.log('window.debugRefresh.runAllChecks() // 运行所有检查');
  console.log('window.debugRefresh.simulateRefresh() // 模拟刷新');
  console.log('window.debugRefresh.checkSWR() // 检查SWR缓存');
  
  return {
    checkSWR,
    simulateRefresh,
    checkReactState,
    runAllChecks
  };
})();