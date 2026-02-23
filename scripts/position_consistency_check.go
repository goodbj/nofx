package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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
	ShouldClose     bool
	Reason          string
}

func main() {
	// 数据库连接配置
	db, err := sql.Open("postgres", "host=localhost port=5432 user=your_user password=your_password dbname=nofx sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 执行一致性检查
	report := checkPositionConsistency(db)

	// 显示检查结果
	displayReport(report)

	// 询问是否执行修复
	if len(report) > 0 {
		fmt.Println("\n发现", len(report), "条不一致记录")
		fmt.Print("是否执行自动修复? (y/N): ")

		var input string
		fmt.Scanln(&input)

		if input == "y" || input == "Y" {
			executeFix(db, report)
		}
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
		log.Fatal("查询失败:", err)
	}
	defer rows.Close()

	// 按交易员分组检查
	traderPositions := make(map[string][]PositionRecord)

	for rows.Next() {
		var record PositionRecord
		err := rows.Scan(&record.ID, &record.TraderID, &record.Symbol, &record.Status, &record.OpenTime, &record.CloseTime)
		if err != nil {
			log.Printf("扫描记录失败: %v", err)
			continue
		}
		traderPositions[record.TraderID] = append(traderPositions[record.TraderID], record)
	}

	// 对每个交易员检查持仓一致性
	for traderID, positions := range traderPositions {
		inconsistencies := checkTraderConsistency(traderID, positions)
		report = append(report, inconsistencies...)
	}

	return report
}

func checkTraderConsistency(traderID string, positions []PositionRecord) []InconsistencyReport {
	var inconsistencies []InconsistencyReport

	// 模拟获取交易所实际持仓（这里需要替换为真实的API调用）
	actualHoldings := getActualHoldingsFromExchange(traderID)

	// 创建实际持仓的映射
	actualSymbols := make(map[string]bool)
	for _, symbol := range actualHoldings {
		actualSymbols[symbol] = true
	}

	// 检查数据库中的每条记录
	for _, position := range positions {
		// 如果数据库中有但实际没有持仓
		if !actualSymbols[position.Symbol] {
			shouldClose := shouldAutoClose(position)
			reason := determineCloseReason(position, shouldClose)

			inconsistencies = append(inconsistencies, InconsistencyReport{
				TraderID:        traderID,
				Symbol:          position.Symbol,
				PositionID:      position.ID,
				OpenTime:        position.OpenTime,
				DiscrepancyType: "数据库有但实际无",
				ShouldClose:     shouldClose,
				Reason:          reason,
			})
		}
	}

	return inconsistencies
}

func getActualHoldingsFromExchange(traderID string) []string {
	// TODO: 这里需要实现真实的交易所API调用
	// 暂时返回空数组模拟无持仓状态
	return []string{}
}

func shouldAutoClose(position PositionRecord) bool {
	// 自动关闭的条件：
	// 1. 开仓时间超过24小时
	// 2. 不是近期的交易（避免误操作刚开的仓位）
	// 3. 没有异常标记

	now := time.Now()
	duration := now.Sub(position.OpenTime)

	// 开仓超过24小时且不是今天开的仓位
	return duration > 24*time.Hour && position.OpenTime.Day() != now.Day()
}

func determineCloseReason(position PositionRecord, autoClose bool) string {
	if autoClose {
		return "系统自动检测到持仓状态不一致，且开仓时间超过24小时"
	}
	return "需要人工审核确认"
}

func displayReport(report []InconsistencyReport) {
	fmt.Println("\n=== 持仓状态一致性检查报告 ===")
	fmt.Printf("%-15s %-10s %-15s %-20s %-10s %s\n",
		"交易员ID", "币种", "仓位ID", "开仓时间", "自动关闭", "原因")
	fmt.Println(strings.Repeat("-", 80))

	for _, item := range report {
		autoClose := "否"
		if item.ShouldClose {
			autoClose = "是"
		}
		fmt.Printf("%-15s %-10s %-15d %-20s %-10s %s\n",
			item.TraderID,
			item.Symbol,
			item.PositionID,
			item.OpenTime.Format("2006-01-02 15:04"),
			autoClose,
			item.Reason)
	}
}

func executeFix(db *sql.DB, report []InconsistencyReport) {
	fmt.Println("\n开始执行自动修复...")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("开启事务失败:", err)
	}
	defer tx.Rollback()

	fixedCount := 0
	for _, item := range report {
		if item.ShouldClose {
			// 更新状态为CLOSED
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

			fmt.Printf("✓ 已关闭交易员%s的%s仓位 (ID: %d)\n",
				item.TraderID, item.Symbol, item.PositionID)
			fixedCount++
		}
	}

	if fixedCount > 0 {
		err = tx.Commit()
		if err != nil {
			log.Fatal("提交事务失败:", err)
		}
		fmt.Printf("\n修复完成！共更新%d条记录\n", fixedCount)
	} else {
		fmt.Println("没有需要自动修复的记录")
	}
}
