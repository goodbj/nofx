package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"nofx/store"

	_ "github.com/glebarez/go-sqlite" // Pure Go SQLite driver - same as main application
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Starting database field verification and repair...")

	// 1. Check database file
	dbPath := "data/data.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("❌ Database file does not exist: %s", dbPath)
	}

	// 2. Open database with GORM
	db, err := store.InitGorm(dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open database with GORM: %v", err)
	}

	// 3. Run the repair function
	if err := repairDatabaseFields(db); err != nil {
		log.Fatalf("❌ Database repair failed: %v", err)
	}

	log.Println("✅ Database field verification and repair completed!")
}

func repairDatabaseFields(gdb *gorm.DB) error {
	log.Println("🔄 Checking and repairing database fields...")

	// Convert to *sql.DB for direct queries
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Check each table for missing fields
	tableChecks := []func(*sql.DB) error{
		checkUsersTable,
		checkAIModelsTable,
		checkExchangesTable,
		checkTradersTable,
		checkStrategiesTable,
		checkPositionsTable,
		checkOrdersTable,
		checkBacktestTables,
		checkEquityTable,
	}

	for _, checkFunc := range tableChecks {
		if err := checkFunc(sqlDB); err != nil {
			log.Printf("⚠️ Error checking table: %v", err)
			// Continue with other checks even if one fails
		}
	}

	return nil
}

func checkUsersTable(db *sql.DB) error {
	log.Println("🔍 Checking users table...")

	// Get existing columns
	rows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		return fmt.Errorf("failed to get users table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	// Check for missing columns and add them
	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "TEXT PRIMARY KEY"},
		{"email", "TEXT NOT NULL"},
		{"password_hash", "TEXT NOT NULL"},
		{"otp_secret", "TEXT"},
		{"otp_verified", "INTEGER DEFAULT 0"},
		{"created_at", "TEXT"},
		{"updated_at", "TEXT"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to users table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to users table", col.name)
			}
		}
	}

	// Ensure unique index on email
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	if err != nil {
		log.Printf("⚠️ Could not create email index: %v", err)
	}

	log.Println("✅ Users table check completed")
	return nil
}

func checkAIModelsTable(db *sql.DB) error {
	log.Println("🔍 Checking ai_models table...")

	rows, err := db.Query("PRAGMA table_info(ai_models)")
	if err != nil {
		return fmt.Errorf("failed to get ai_models table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "TEXT PRIMARY KEY"},
		{"user_id", "TEXT NOT NULL DEFAULT 'default'"},
		{"name", "TEXT NOT NULL"},
		{"provider", "TEXT NOT NULL"},
		{"enabled", "INTEGER DEFAULT 0"},
		{"api_key", "TEXT DEFAULT ''"},
		{"custom_api_url", "TEXT DEFAULT ''"},
		{"custom_model_name", "TEXT DEFAULT ''"},
		{"created_at", "TEXT"},
		{"updated_at", "TEXT"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE ai_models ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to ai_models table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to ai_models table", col.name)
			}
		}
	}

	log.Println("✅ AI Models table check completed")
	return nil
}

func checkExchangesTable(db *sql.DB) error {
	log.Println("🔍 Checking exchanges table...")

	rows, err := db.Query("PRAGMA table_info(exchanges)")
	if err != nil {
		return fmt.Errorf("failed to get exchanges table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "TEXT PRIMARY KEY"},
		{"exchange_type", "TEXT DEFAULT ''"},
		{"account_name", "TEXT DEFAULT ''"},
		{"user_id", "TEXT NOT NULL DEFAULT 'default'"},
		{"name", "TEXT NOT NULL"},
		{"type", "TEXT NOT NULL"},
		{"enabled", "INTEGER DEFAULT 0"},
		{"api_key", "TEXT DEFAULT ''"},
		{"secret_key", "TEXT DEFAULT ''"},
		{"passphrase", "TEXT DEFAULT ''"},
		{"testnet", "INTEGER DEFAULT 0"},
		{"custom_api_url", "TEXT DEFAULT ''"},
		{"hyperliquid_wallet_addr", "TEXT DEFAULT ''"},
		{"aster_user", "TEXT DEFAULT ''"},
		{"aster_signer", "TEXT DEFAULT ''"},
		{"aster_private_key", "TEXT DEFAULT ''"},
		{"lighter_wallet_addr", "TEXT DEFAULT ''"},
		{"lighter_private_key", "TEXT DEFAULT ''"},
		{"lighter_api_key_private_key", "TEXT DEFAULT ''"},
		{"lighter_api_key_index", "INTEGER DEFAULT 0"},
		{"created_at", "TEXT"},
		{"updated_at", "TEXT"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE exchanges ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to exchanges table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to exchanges table", col.name)
			}
		}
	}

	log.Println("✅ Exchanges table check completed")
	return nil
}

func checkTradersTable(db *sql.DB) error {
	log.Println("🔍 Checking traders table...")

	rows, err := db.Query("PRAGMA table_info(traders)")
	if err != nil {
		return fmt.Errorf("failed to get traders table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "TEXT PRIMARY KEY"},
		{"user_id", "TEXT NOT NULL DEFAULT 'default'"},
		{"name", "TEXT NOT NULL"},
		{"ai_model_id", "TEXT NOT NULL"},
		{"exchange_id", "TEXT NOT NULL"},
		{"strategy_id", "TEXT DEFAULT ''"},
		{"initial_balance", "REAL NOT NULL"},
		{"scan_interval_minutes", "INTEGER DEFAULT 3"},
		{"is_running", "INTEGER DEFAULT 0"},
		{"is_cross_margin", "INTEGER DEFAULT 1"},
		{"show_in_competition", "INTEGER DEFAULT 1"},
		{"created_at", "TEXT"},
		{"updated_at", "TEXT"},
		{"btc_eth_leverage", "INTEGER DEFAULT 5"},
		{"altcoin_leverage", "INTEGER DEFAULT 5"},
		{"trading_symbols", "TEXT DEFAULT ''"},
		{"use_coin_pool", "INTEGER DEFAULT 0"},
		{"use_oi_top", "INTEGER DEFAULT 0"},
		{"custom_prompt", "TEXT DEFAULT ''"},
		{"override_base_prompt", "INTEGER DEFAULT 0"},
		{"system_prompt_template", "TEXT DEFAULT 'default'"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE traders ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to traders table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to traders table", col.name)
			}
		}
	}

	log.Println("✅ Traders table check completed")
	return nil
}

func checkStrategiesTable(db *sql.DB) error {
	log.Println("🔍 Checking strategies table...")

	rows, err := db.Query("PRAGMA table_info(strategies)")
	if err != nil {
		return fmt.Errorf("failed to get strategies table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "TEXT PRIMARY KEY"},
		{"user_id", "TEXT NOT NULL DEFAULT ''"},
		{"name", "TEXT NOT NULL"},
		{"description", "TEXT DEFAULT ''"},
		{"is_active", "INTEGER DEFAULT 0"},
		{"is_default", "INTEGER DEFAULT 0"},
		{"is_public", "INTEGER DEFAULT 0"},
		{"config_visible", "INTEGER DEFAULT 1"},
		{"config", "TEXT NOT NULL DEFAULT '{}'"},
		{"created_at", "TEXT"},
		{"updated_at", "TEXT"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE strategies ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to strategies table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to strategies table", col.name)
			}
		}
	}

	log.Println("✅ Strategies table check completed")
	return nil
}

func checkPositionsTable(db *sql.DB) error {
	log.Println("🔍 Checking trader_positions table...")

	rows, err := db.Query("PRAGMA table_info(trader_positions)")
	if err != nil {
		return fmt.Errorf("failed to get trader_positions table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "INTEGER PRIMARY KEY"},
		{"trader_id", "TEXT NOT NULL"},
		{"exchange_id", "TEXT NOT NULL DEFAULT ''"},
		{"exchange_type", "TEXT NOT NULL DEFAULT ''"},
		{"exchange_position_id", "TEXT NOT NULL DEFAULT ''"},
		{"symbol", "TEXT NOT NULL"},
		{"side", "TEXT NOT NULL"},
		{"entry_quantity", "REAL DEFAULT 0"},
		{"quantity", "REAL NOT NULL"},
		{"entry_price", "REAL NOT NULL"},
		{"entry_order_id", "TEXT DEFAULT ''"},
		{"entry_time", "INTEGER NOT NULL"},
		{"exit_price", "REAL DEFAULT 0"},
		{"exit_order_id", "TEXT DEFAULT ''"},
		{"exit_time", "INTEGER"},
		{"realized_pnl", "REAL DEFAULT 0"},
		{"fee", "REAL DEFAULT 0"},
		{"leverage", "INTEGER DEFAULT 1"},
		{"status", "TEXT DEFAULT 'OPEN'"},
		{"close_reason", "TEXT DEFAULT ''"},
		{"source", "TEXT DEFAULT 'system'"},
		{"created_at", "INTEGER"},
		{"updated_at", "INTEGER"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE trader_positions ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to trader_positions table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to trader_positions table", col.name)
			}
		}
	}

	log.Println("✅ Positions table check completed")
	return nil
}

func checkOrdersTable(db *sql.DB) error {
	log.Println("🔍 Checking trader_orders table...")

	rows, err := db.Query("PRAGMA table_info(trader_orders)")
	if err != nil {
		return fmt.Errorf("failed to get trader_orders table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "INTEGER PRIMARY KEY"},
		{"trader_id", "TEXT NOT NULL"},
		{"exchange_id", "TEXT NOT NULL DEFAULT ''"},
		{"exchange_type", "TEXT NOT NULL DEFAULT ''"},
		{"exchange_order_id", "TEXT NOT NULL"},
		{"client_order_id", "TEXT DEFAULT ''"},
		{"symbol", "TEXT NOT NULL"},
		{"side", "TEXT NOT NULL"},
		{"position_side", "TEXT DEFAULT ''"},
		{"type", "TEXT NOT NULL"},
		{"time_in_force", "TEXT DEFAULT 'GTC'"},
		{"quantity", "REAL NOT NULL"},
		{"price", "REAL DEFAULT 0"},
		{"stop_price", "REAL DEFAULT 0"},
		{"status", "TEXT NOT NULL DEFAULT 'NEW'"},
		{"filled_quantity", "REAL DEFAULT 0"},
		{"avg_fill_price", "REAL DEFAULT 0"},
		{"commission", "REAL DEFAULT 0"},
		{"commission_asset", "TEXT DEFAULT 'USDT'"},
		{"leverage", "INTEGER DEFAULT 1"},
		{"reduce_only", "INTEGER DEFAULT 0"},
		{"close_position", "INTEGER DEFAULT 0"},
		{"working_type", "TEXT DEFAULT 'CONTRACT_PRICE'"},
		{"price_protect", "INTEGER DEFAULT 0"},
		{"order_action", "TEXT DEFAULT ''"},
		{"related_position_id", "INTEGER DEFAULT 0"},
		{"created_at", "INTEGER"},
		{"updated_at", "INTEGER"},
		{"filled_at", "INTEGER"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE trader_orders ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to trader_orders table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to trader_orders table", col.name)
			}
		}
	}

	// Also check trader_fills table
	log.Println("🔍 Checking trader_fills table...")

	rows, err = db.Query("PRAGMA table_info(trader_fills)")
	if err != nil {
		return fmt.Errorf("failed to get trader_fills table info: %w", err)
	}
	defer rows.Close()

	columns = make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols = []struct {
		name string
		def  string
	}{
		{"id", "INTEGER PRIMARY KEY"},
		{"trader_id", "TEXT NOT NULL"},
		{"exchange_id", "TEXT NOT NULL DEFAULT ''"},
		{"exchange_type", "TEXT NOT NULL DEFAULT ''"},
		{"order_id", "INTEGER NOT NULL"},
		{"exchange_order_id", "TEXT NOT NULL"},
		{"exchange_trade_id", "TEXT NOT NULL"},
		{"symbol", "TEXT NOT NULL"},
		{"side", "TEXT NOT NULL"},
		{"price", "REAL NOT NULL"},
		{"quantity", "REAL NOT NULL"},
		{"quote_quantity", "REAL NOT NULL"},
		{"commission", "REAL NOT NULL"},
		{"commission_asset", "TEXT NOT NULL"},
		{"realized_pnl", "REAL DEFAULT 0"},
		{"is_maker", "INTEGER DEFAULT 0"},
		{"created_at", "INTEGER"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE trader_fills ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to trader_fills table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to trader_fills table", col.name)
			}
		}
	}

	// Create indexes for orders and fills
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_trader_id ON trader_orders(trader_id)")
	if err != nil {
		log.Printf("⚠️ Could not create trader_orders index: %v", err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_fills_trader_id ON trader_fills(trader_id)")
	if err != nil {
		log.Printf("⚠️ Could not create trader_fills index: %v", err)
	}

	log.Println("✅ Orders and Fills tables check completed")
	return nil
}

func checkBacktestTables(db *sql.DB) error {
	log.Println("🔍 Checking backtest tables...")

	tables := []string{
		"backtest_runs", "backtest_checkpoints", "backtest_equity",
		"backtest_trades", "backtest_metrics", "backtest_decisions",
	}

	for _, tableName := range tables {
		log.Printf("🔍 Checking %s table...", tableName)

		rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		if err != nil {
			log.Printf("⚠️ Could not get %s table info: %v", tableName, err)
			continue
		}
		defer rows.Close()

		columns := make(map[string]bool)
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull int
			var dflt_value *string
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
				continue
			}
			columns[name] = true
		}

		// Define expected columns for each table
		var missingCols []struct {
			name string
			def  string
		}

		switch tableName {
		case "backtest_runs":
			missingCols = []struct {
				name string
				def  string
			}{
				{"run_id", "TEXT PRIMARY KEY"},
				{"user_id", "TEXT NOT NULL DEFAULT ''"},
				{"config_json", "BLOB"},
				{"state", "TEXT NOT NULL DEFAULT 'created'"},
				{"label", "TEXT DEFAULT ''"},
				{"symbol_count", "INTEGER DEFAULT 0"},
				{"decision_tf", "TEXT DEFAULT ''"},
				{"processed_bars", "INTEGER DEFAULT 0"},
				{"progress_pct", "REAL DEFAULT 0"},
				{"equity_last", "REAL DEFAULT 0"},
				{"max_drawdown_pct", "REAL DEFAULT 0"},
				{"liquidated", "INTEGER DEFAULT 0"},
				{"liquidation_note", "TEXT DEFAULT ''"},
				{"prompt_template", "TEXT DEFAULT ''"},
				{"custom_prompt", "TEXT DEFAULT ''"},
				{"override_prompt", "INTEGER DEFAULT 0"},
				{"ai_provider", "TEXT DEFAULT ''"},
				{"ai_model", "TEXT DEFAULT ''"},
				{"last_error", "TEXT DEFAULT ''"},
				{"created_at", "TEXT"},
				{"updated_at", "TEXT"},
			}
		case "backtest_checkpoints":
			missingCols = []struct {
				name string
				def  string
			}{
				{"run_id", "TEXT PRIMARY KEY"},
				{"payload", "BLOB NOT NULL"},
				{"updated_at", "TEXT"},
			}
		case "backtest_equity":
			missingCols = []struct {
				name string
				def  string
			}{
				{"id", "INTEGER PRIMARY KEY"},
				{"run_id", "TEXT NOT NULL"},
				{"ts", "INTEGER NOT NULL"},
				{"equity", "REAL NOT NULL"},
				{"available", "REAL NOT NULL"},
				{"pnl", "REAL NOT NULL"},
				{"pnl_pct", "REAL NOT NULL"},
				{"dd_pct", "REAL NOT NULL"},
				{"cycle", "INTEGER NOT NULL"},
			}
		case "backtest_trades":
			missingCols = []struct {
				name string
				def  string
			}{
				{"id", "INTEGER PRIMARY KEY"},
				{"run_id", "TEXT NOT NULL"},
				{"ts", "INTEGER NOT NULL"},
				{"symbol", "TEXT NOT NULL"},
				{"action", "TEXT NOT NULL"},
				{"side", "TEXT DEFAULT ''"},
				{"qty", "REAL DEFAULT 0"},
				{"price", "REAL DEFAULT 0"},
				{"fee", "REAL DEFAULT 0"},
				{"slippage", "REAL DEFAULT 0"},
				{"order_value", "REAL DEFAULT 0"},
				{"realized_pnl", "REAL DEFAULT 0"},
				{"leverage", "INTEGER DEFAULT 0"},
				{"cycle", "INTEGER DEFAULT 0"},
				{"position_after", "REAL DEFAULT 0"},
				{"liquidation", "INTEGER DEFAULT 0"},
				{"note", "TEXT DEFAULT ''"},
			}
		case "backtest_metrics":
			missingCols = []struct {
				name string
				def  string
			}{
				{"run_id", "TEXT PRIMARY KEY"},
				{"payload", "BLOB NOT NULL"},
				{"updated_at", "TEXT"},
			}
		case "backtest_decisions":
			missingCols = []struct {
				name string
				def  string
			}{
				{"id", "INTEGER PRIMARY KEY"},
				{"run_id", "TEXT NOT NULL"},
				{"cycle", "INTEGER NOT NULL"},
				{"payload", "BLOB NOT NULL"},
				{"created_at", "TEXT"},
			}
		}

		for _, col := range missingCols {
			if !columns[col.name] {
				query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, col.name, col.def)
				if _, err := db.Exec(query); err != nil {
					log.Printf("❌ Failed to add column %s to %s table: %v", col.name, tableName, err)
				} else {
					log.Printf("✅ Added column %s to %s table", col.name, tableName)
				}
			}
		}
	}

	log.Println("✅ Backtest tables check completed")
	return nil
}

func checkEquityTable(db *sql.DB) error {
	log.Println("🔍 Checking trader_equity_snapshots table...")

	rows, err := db.Query("PRAGMA table_info(trader_equity_snapshots)")
	if err != nil {
		return fmt.Errorf("failed to get trader_equity_snapshots table info: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			continue
		}
		columns[name] = true
	}

	missingCols := []struct {
		name string
		def  string
	}{
		{"id", "INTEGER PRIMARY KEY"},
		{"trader_id", "TEXT NOT NULL"},
		{"timestamp", "TEXT NOT NULL"},
		{"total_equity", "REAL NOT NULL DEFAULT 0"},
		{"balance", "REAL NOT NULL DEFAULT 0"},
		{"unrealized_pnl", "REAL NOT NULL DEFAULT 0"},
		{"position_count", "INTEGER DEFAULT 0"},
		{"margin_used_pct", "REAL DEFAULT 0"},
		{"created_at", "TEXT"},
	}

	for _, col := range missingCols {
		if !columns[col.name] {
			query := fmt.Sprintf("ALTER TABLE trader_equity_snapshots ADD COLUMN %s %s", col.name, col.def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("❌ Failed to add column %s to trader_equity_snapshots table: %v", col.name, err)
			} else {
				log.Printf("✅ Added column %s to trader_equity_snapshots table", col.name)
			}
		}
	}

	log.Println("✅ Equity table check completed")
	return nil
}
