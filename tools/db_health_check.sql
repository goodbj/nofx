-- 数据库健康检查脚本
-- 检查数据库结构完整性并验证数据

.print "=== NOFX 数据库健康检查 ==="
.print ""

-- 检查各表记录数
.print "--- 表记录统计 ---"
SELECT 'users' AS table_name, COUNT(*) AS record_count FROM users
UNION ALL
SELECT 'ai_models' AS table_name, COUNT(*) AS record_count FROM ai_models
UNION ALL
SELECT 'exchanges' AS table_name, COUNT(*) AS record_count FROM exchanges
UNION ALL
SELECT 'traders' AS table_name, COUNT(*) AS record_count FROM traders
UNION ALL
SELECT 'strategies' AS table_name, COUNT(*) AS record_count FROM strategies
UNION ALL
SELECT 'decision_records' AS table_name, COUNT(*) AS record_count FROM decision_records
UNION ALL
SELECT 'trader_positions' AS table_name, COUNT(*) AS record_count FROM trader_positions
UNION ALL
SELECT 'trader_orders' AS table_name, COUNT(*) AS record_count FROM trader_orders;

.print ""
.print "--- 交易员状态检查 ---"
SELECT id, name, CASE WHEN is_running = 1 THEN 'RUNNING' ELSE 'STOPPED' END AS status, created_at FROM traders ORDER BY name;

.print ""
.print "--- 策略配置检查 ---"
SELECT id, name, user_id, LENGTH(config) AS config_length FROM strategies ORDER BY name;

.print ""
.print "--- AI模型配置检查 ---"
SELECT id, name, type, LENGTH(config) AS config_length FROM ai_models ORDER BY name;

.print ""
.print "--- 交易所配置检查 ---"
SELECT id, name, type, LENGTH(config) AS config_length FROM exchanges ORDER BY name;

.print ""
.print "--- 最近的决策记录 ---"
SELECT id, trader_id, symbol, action, reason, created_at FROM decision_records ORDER BY created_at DESC LIMIT 5;

.print ""
.print "--- 数据库健康检查完成 ---"
.print "如需修复数据库，请运行: UPDATE traders SET is_running = 1 WHERE is_running = 0;"