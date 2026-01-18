-- 重置可能存在问题的表的SQL脚本
-- 使用方法: sqlite3 data/data.db < reset_problematic_tables.sql

BEGIN TRANSACTION;

-- 显示各表当前记录数
SELECT '当前AI模型记录数:' AS INFO, COUNT(*) AS COUNT FROM ai_models;
SELECT '当前交易所记录数:' AS INFO, COUNT(*) AS COUNT FROM exchanges;
SELECT '当前交易者记录数:' AS INFO, COUNT(*) AS COUNT FROM traders;
SELECT '当前决策记录数:' AS INFO, COUNT(*) AS COUNT FROM decision_records;

-- 删除AI模型表的所有记录
DELETE FROM ai_models;

-- 删除交易所表的所有记录
DELETE FROM exchanges;

-- 删除交易者表的所有记录
DELETE FROM traders;

-- 删除决策记录表的所有记录
DELETE FROM decision_records;

-- 验证删除结果
SELECT '删除后AI模型记录数:' AS INFO, COUNT(*) AS COUNT FROM ai_models;
SELECT '删除后交易所记录数:' AS INFO, COUNT(*) AS COUNT FROM exchanges;
SELECT '删除后交易者记录数:' AS INFO, COUNT(*) AS COUNT FROM traders;
SELECT '删除后决策记录数:' AS INFO, COUNT(*) AS COUNT FROM decision_records;

-- 显示其他表的记录数以确认数据库连接正常
SELECT '用户表记录数:' AS INFO, COUNT(*) AS COUNT FROM users;
SELECT '策略表记录数:' AS INFO, COUNT(*) AS COUNT FROM strategies;
SELECT '仓位表记录数:' AS INFO, COUNT(*) AS COUNT FROM trader_positions;
SELECT '订单表记录数:' AS INFO, COUNT(*) AS COUNT FROM trader_orders;

COMMIT;

-- 完成
SELECT '问题表数据清除完成' AS RESULT;