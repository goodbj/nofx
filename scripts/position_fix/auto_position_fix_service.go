package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type PositionRecord struct {
	ID        int64      `db:"id"`
	TraderID  string     `db:"trader_id"`
	Symbol    string     `db:"symbol"`
	Status    string     `db:"status"`
	OpenTime  time.Time  `db:"open_time"`
	CloseTime *time.Time `db:"close_time"`
}

type InconsistencyReport struct {
	TraderID        string
	Symbol          string
	PositionID      int64
	OpenTime        time.Time
	DiscrepancyType string
	AutoClosed      bool
	Reason          string
}

func main() {
	fmt.Println("=== 持仓状态自动检查服务启动 ===")

	// 设置定时检查（每小时执行一次）
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 立即执行一次检查
	fmt.Println("执行初始检查...")
	autoCheckAndFixPositions()

	// 定时执行
	for {
		select {
		case <-ticker.C:
			fmt.Println("定时检查执行中...")
			autoCheckAndFixPositions()
		}
	}
}

func autoCheckAndFixPositions() {
	// 数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=your_user password=your_password dbname=nofx sslmode=disable")
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		return
	}
	defer db.Close()

	// 执行检查
	inconsistencies := checkPositionConsistency(db)

	if len(inconsistencies) > 0 {
		fmt.Printf("发现 %d 条不一致记录，开始自动修复...\n", len(inconsistencies))
		fixedCount := executeAutomaticFix(db, inconsistencies)
		logResults(inconsistencies, fixedCount)
	} else {
		fmt.Println("未发现不一致记录")
	}
}

func checkPositionConsistency(db *sql.DB) []InconsistencyReport {
	var report []InconsistencyReport

	// 查询所有OPEN状态的持仓
	query := `
		SELECT id, trader_id, symbol, status, open_time, close_time
		FROM trader_positions 
		WHERE status = 'OPEN'
		ORDER BY trader_id, open_time
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return report
	}
	defer rows.Close()

	// 处理查询结果
	for rows.Next() {
		var record PositionRecord
		err := rows.Scan(&record.ID, &record.TraderID, &record.Symbol, &record.Status, &record.OpenTime, &record.CloseTime)
		if err != nil {
			log.Printf("扫描记录失败: %v", err)
			continue
		}

		// 检查是否应该自动关闭
		if shouldAutoClose(record) {
			report = append(report, InconsistencyReport{
				TraderID:        record.TraderID,
				Symbol:          record.Symbol,
				PositionID:      record.ID,
				OpenTime:        record.OpenTime,
				DiscrepancyType: "数据库有但实际无",
				AutoClosed:      true,
				Reason:          "系统自动检测到持仓状态不一致，且开仓时间超过24小时",
			})
		}
	}

	return report
}

func shouldAutoClose(position PositionRecord) bool {
	// 自动关闭条件：开仓时间超过24小时
	now := time.Now()
	duration := now.Sub(position.OpenTime)
	return duration > 24*time.Hour
}

func executeAutomaticFix(db *sql.DB, report []InconsistencyReport) int {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("开启事务失败: %v", err)
		return 0
	}
	defer tx.Rollback()

	fixedCount := 0
	for _, item := range report {
		if item.AutoClosed {
			updateQuery := `
				UPDATE trader_positions 
				SET status = 'CLOSED',
				    close_time = $1,
				    close_reason = $2,
				    updated_at = $1
				WHERE id = $3
			`

			_, err := tx.Exec(updateQuery, time.Now(), item.Reason, item.PositionID)
			if err != nil {
				log.Printf("更新仓位%d失败: %v", item.PositionID, err)
				continue
			}
			fixedCount++
		}
	}

	if fixedCount > 0 {
		err = tx.Commit()
		if err != nil {
			log.Printf("提交事务失败: %v", err)
			return 0
		}
	}

	return fixedCount
}

func logResults(report []InconsistencyReport, fixedCount int) {
	// 记录到日志文件
	logFile, err := openLogFile()
	if err != nil {
		log.Printf("无法打开日志文件: %v", err)
		return
	}
	defer logFile.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(logFile, "[%s] 自动检查完成 - 发现%d条不一致记录，修复%d条\n",
		timestamp, len(report), fixedCount)

	// 记录详细信息
	for _, item := range report {
		status := "已修复"
		if !item.AutoClosed {
			status = "需人工处理"
		}
		fmt.Fprintf(logFile, "  - 交易员:%s 币种:%s 状态:%s\n",
			item.TraderID, item.Symbol, status)
	}

	fmt.Printf("自动检查完成 - 修复%d条记录，详情请查看日志\n", fixedCount)
}

func openLogFile() (*os.File, error) {
	// 创建日志目录
	logDir := "logs"
	os.MkdirAll(logDir, 0755)

	// 按日期创建日志文件
	today := time.Now().Format("2006-01-02")
	logPath := fmt.Sprintf("%s/position_fix_%s.log", logDir, today)

	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}
