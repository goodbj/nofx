-- 自动修复数据库结构缺失字段的SQL脚本
-- 该脚本会检查并添加缺失的字段到各个表中

.echo ON

-- 检查并添加users表的字段（如需要）
-- 注意：SQLite不支持直接检查字段是否存在，所以我们创建临时表来处理

.print "检查并修复数据库结构..."

-- 为traders表添加缺失字段（如果它们不存在的话）
-- 由于SQLite的限制，我们需要逐一尝试添加（即使字段已存在也不会报错）

-- 为traders表添加可能缺失的字段
ALTER TABLE traders ADD COLUMN IF NOT EXISTS btc_eth_leverage INTEGER DEFAULT 5;
ALTER TABLE traders ADD COLUMN IF NOT EXISTS altcoin_leverage INTEGER DEFAULT 5;
ALTER TABLE traders ADD COLUMN IF NOT EXISTS trading_symbols TEXT DEFAULT "";
ALTER TABLE traders ADD COLUMN IF NOT EXISTS use_coin_pool NUMERIC DEFAULT FALSE;
ALTER TABLE traders ADD COLUMN IF NOT EXISTS use_oi_top NUMERIC DEFAULT FALSE;
ALTER TABLE traders ADD COLUMN IF NOT EXISTS custom_prompt TEXT DEFAULT "";
ALTER TABLE traders ADD COLUMN IF NOT EXISTS override_base_prompt NUMERIC DEFAULT FALSE;
ALTER TABLE traders ADD COLUMN IF NOT EXISTS system_prompt_template TEXT DEFAULT "default";

-- 为strategies表添加可能缺失的字段
ALTER TABLE strategies ADD COLUMN IF NOT EXISTS created_at DATETIME;
ALTER TABLE strategies ADD COLUMN IF NOT EXISTS updated_at DATETIME;

-- 为decision_records表添加可能缺失的字段
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS trader_id TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS symbol TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS action TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS reason TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS confidence REAL;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS market_condition TEXT;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS created_at DATETIME;
ALTER TABLE decision_records ADD COLUMN IF NOT EXISTS updated_at DATETIME;

-- 为trader_positions表添加可能缺失的字段
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS trader_id TEXT;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS symbol TEXT;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS side TEXT;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS quantity REAL;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS entry_price REAL;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS current_price REAL;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS unrealized_pnl REAL;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS leverage REAL;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS created_at DATETIME;
ALTER TABLE trader_positions ADD COLUMN IF NOT EXISTS updated_at DATETIME;

-- 为trader_orders表添加可能缺失的字段
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS trader_id TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS symbol TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS side TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS order_type TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS quantity REAL;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS price REAL;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS status TEXT;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS created_at DATETIME;
ALTER TABLE trader_orders ADD COLUMN IF NOT EXISTS updated_at DATETIME;

-- 为trader_equity_snapshots表添加可能缺失的字段
ALTER TABLE trader_equity_snapshots ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE trader_equity_snapshots ADD COLUMN IF NOT EXISTS trader_id TEXT;
ALTER TABLE trader_equity_snapshots ADD COLUMN IF NOT EXISTS equity REAL;
ALTER TABLE trader_equity_snapshots ADD COLUMN IF NOT EXISTS balance REAL;
ALTER TABLE trader_equity_snapshots ADD COLUMN IF NOT EXISTS timestamp DATETIME;

-- 为backtest_runs表添加可能缺失的字段
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS strategy_id TEXT;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS start_date DATETIME;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS end_date DATETIME;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS initial_balance REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS final_balance REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS total_return REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS sharpe_ratio REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS max_drawdown REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS win_rate REAL;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS total_trades INTEGER;
ALTER TABLE backtest_runs ADD COLUMN IF NOT EXISTS created_at DATETIME;

-- 为backtest_trades表添加可能缺失的字段
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS backtest_run_id TEXT;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS symbol TEXT;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS side TEXT;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS quantity REAL;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS entry_price REAL;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS exit_price REAL;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS pnl REAL;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS entry_time DATETIME;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS exit_time DATETIME;
ALTER TABLE backtest_trades ADD COLUMN IF NOT EXISTS fee REAL;

-- 显示所有表的结构以确认修复结果
.print "\n=== 修复完成，当前数据库结构 ==="

.print "\n--- traders 表结构 ---"
.schema traders

.print "\n--- strategies 表结构 ---"
.schema strategies

.print "\n--- decision_records 表结构 ---"
.schema decision_records

.print "\n--- trader_positions 表结构 ---"
.schema trader_positions

.print "\n--- trader_orders 表结构 ---"
.schema trader_orders

.print "\n--- trader_equity_snapshots 表结构 ---"
.schema trader_equity_snapshots

.print "\n--- backtest_runs 表结构 ---"
.schema backtest_runs

.print "\n--- backtest_trades 表结构 ---"
.schema backtest_trades

.print "\n--- ai_models 表结构 ---"
.schema ai_models

.print "\n--- exchanges 表结构 ---"
.schema exchanges

.print "\n--- users 表结构 ---"
.schema users

.print "\n自动数据库结构修复完成！"