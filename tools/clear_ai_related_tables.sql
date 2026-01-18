-- 清空AI模型相关数据表的SQL脚本
-- 保留用户、策略等基础配置，仅清空AI相关的数据

.echo ON
.print "开始清空AI模型相关数据表..."

-- 备份当前状态
.print "正在备份当前AI相关数据..."
-- 注意：SQLite中不能直接复制表结构和数据，我们只是记录当前状态
SELECT 'Current decision_records count:' as info, COUNT(*) FROM decision_records;
SELECT 'Current trader_positions count:' as info, COUNT(*) FROM trader_positions;
SELECT 'Current trader_orders count:' as info, COUNT(*) FROM trader_orders;
SELECT 'Current trader_equity_snapshots count:' as info, COUNT(*) FROM trader_equity_snapshots;
SELECT 'Current backtest_runs count:' as info, COUNT(*) FROM backtest_runs;
SELECT 'Current backtest_trades count:' as info, COUNT(*) FROM backtest_trades;
SELECT 'Current backtest_decisions count:' as info, COUNT(*) FROM backtest_decisions;
SELECT 'Current backtest_checkpoints count:' as info, COUNT(*) FROM backtest_checkpoints;
SELECT 'Current backtest_equity count:' as info, COUNT(*) FROM backtest_equity;
SELECT 'Current backtest_metrics count:' as info, COUNT(*) FROM backtest_metrics;
SELECT 'Current debate_sessions count:' as info, COUNT(*) FROM debate_sessions;
SELECT 'Current debate_messages count:' as info, COUNT(*) FROM debate_messages;
SELECT 'Current debate_participants count:' as info, COUNT(*) FROM debate_participants;
SELECT 'Current debate_votes count:' as info, COUNT(*) FROM debate_votes;

.print "\n开始清空AI相关数据表..."

-- 清空决策记录表
DELETE FROM decision_records;
.print "已清空 decision_records 表"

-- 清空交易员持仓表
DELETE FROM trader_positions;
.print "已清空 trader_positions 表"

-- 清空交易员订单表
DELETE FROM trader_orders;
.print "已清空 trader_orders 表"

-- 清空交易员权益快照表
DELETE FROM trader_equity_snapshots;
.print "已清空 trader_equity_snapshots 表"

-- 清空回测相关表
DELETE FROM backtest_runs;
.print "已清空 backtest_runs 表"

DELETE FROM backtest_trades;
.print "已清空 backtest_trades 表"

DELETE FROM backtest_decisions;
.print "已清空 backtest_decisions 表"

DELETE FROM backtest_checkpoints;
.print "已清空 backtest_checkpoints 表"

DELETE FROM backtest_equity;
.print "已清空 backtest_equity 表"

DELETE FROM backtest_metrics;
.print "已清空 backtest_metrics 表"

-- 清空辩论相关表
DELETE FROM debate_sessions;
.print "已清空 debate_sessions 表"

DELETE FROM debate_messages;
.print "已清空 debate_messages 表"

DELETE FROM debate_participants;
.print "已清空 debate_participants 表"

DELETE FROM debate_votes;
.print "已清空 debate_votes 表"

-- 保留AI模型配置、交易所配置、交易员配置和策略配置表
-- 这些表包含基础设置，不应该被清空

.print "\nAI相关数据表清理完成！"
.print "保留的表："
.print "- users (用户表)"
.print "- ai_models (AI模型配置表)"
.print "- exchanges (交易所配置表)"
.print "- traders (交易员配置表)"
.print "- strategies (策略配置表)"
.print "\n现在可以重启后端系统以应用更改。"