-- 启动交易员的SQL脚本
-- 将所有交易员的is_running字段设置为1（运行状态）

UPDATE traders SET is_running = 1;

-- 验证更新结果
SELECT 'Updated trader status:' AS INFO;
SELECT id, name, is_running FROM traders;

-- 提示用户重启系统
SELECT '请重启后端系统以使更改生效' AS MESSAGE;