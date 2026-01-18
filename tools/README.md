# NOFX 工具脚本说明

这个目录包含各种实用工具脚本，用于管理和维护NOFX系统。

## 脚本清单

### 1. clear_ai_models.sql
- **功能**: 清除AI模型表中的所有数据
- **使用方法**: `sqlite3 data/data.db < clear_ai_models.sql`

### 2. clear_ai_models.bat
- **功能**: Windows批处理脚本，自动备份数据库并清除AI模型数据
- **使用方法**: 双击运行或在命令行中执行

### 3. reset_problematic_tables.sql
- **功能**: 清除多个可能存在问题的表（AI模型、交易所、交易者、决策记录）
- **使用方法**: `sqlite3 data/data.db < reset_problematic_tables.sql`

## 注意事项

1. **数据备份**: 在执行任何清理操作之前，系统会自动创建数据库备份
2. **谨慎操作**: 这些脚本会删除数据库中的数据，请在执行前确认操作的必要性
3. **适用场景**: 主要用于解决数据库中可能存在错误数据的情况

## 使用建议

- 通常情况下，只需要使用 `clear_ai_models.bat` 即可解决AI模型相关的问题
- 如果需要更广泛的清理，可以使用 `reset_problematic_tables.sql`
- 所有的清理操作都会在执行前显示当前数据量，便于评估影响范围