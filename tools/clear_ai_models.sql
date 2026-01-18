-- 清除AI模型表数据的SQL脚本
-- 使用方法: sqlite3 data/data.db < clear_ai_models.sql

BEGIN TRANSACTION;

-- 显示AI模型表当前记录数
SELECT '当前AI模型记录数:' AS INFO, COUNT(*) AS COUNT FROM ai_models;

-- 备份AI模型表（可选）
-- .backup ai_models_backup_20260118.db

-- 删除所有AI模型记录
DELETE FROM ai_models;

-- 验证删除结果
SELECT '删除后AI模型记录数:' AS INFO, COUNT(*) AS COUNT FROM ai_models;

-- 显示其他相关表的记录数
SELECT '用户表记录数:' AS INFO, COUNT(*) AS COUNT FROM users;
SELECT '交易所表记录数:' AS INFO, COUNT(*) AS COUNT FROM exchanges;
SELECT '交易者表记录数:' AS INFO, COUNT(*) AS COUNT FROM traders;
SELECT '策略表记录数:' AS INFO, COUNT(*) AS COUNT FROM strategies;

COMMIT;

-- 完成
SELECT 'AI模型表数据清除完成' AS RESULT;