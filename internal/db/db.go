// Package db отвечает за хранение истории расчётов пользователя в SQLite.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"examplan/internal/calc"
)

// DB — обёртка над соединением с SQLite.
type DB struct {
	conn *sql.DB
}

// Record — одна сохранённая запись истории расчётов.
type Record struct {
	ID        int64       `json:"id"`
	Subject   string      `json:"subject"`
	CreatedAt time.Time   `json:"createdAt"`
	Input     calc.Input  `json:"input"`
	Result    calc.Result `json:"result"`
}

// Open открывает (или создаёт) файл базы данных и накатывает схему.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("открытие БД: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("проверка соединения с БД: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subject TEXT NOT NULL,
		days_left INTEGER NOT NULL,
		free_hours REAL NOT NULL,
		difficulty INTEGER NOT NULL,
		current_level REAL NOT NULL,
		optimal_hours REAL NOT NULL,
		predicted_result REAL NOT NULL,
		burnout_risk TEXT NOT NULL,
		result_json TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := conn.Exec(schema); err != nil {
		return nil, fmt.Errorf("создание схемы: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close закрывает соединение с базой.
func (d *DB) Close() error {
	return d.conn.Close()
}

// SavePlan сохраняет результат расчёта в историю и возвращает ID записи.
func (d *DB) SavePlan(subject string, in calc.Input, res calc.Result) (int64, error) {
	resultJSON, err := json.Marshal(res)
	if err != nil {
		return 0, fmt.Errorf("сериализация результата: %w", err)
	}

	stmt := `
	INSERT INTO plans (subject, days_left, free_hours, difficulty, current_level,
		optimal_hours, predicted_result, burnout_risk, result_json)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res2, err := d.conn.Exec(stmt, subject, in.DaysLeft, in.FreeHours, in.Difficulty, in.CurrentLevel,
		res.OptimalHours, res.PredictedResult, res.BurnoutRisk, string(resultJSON))
	if err != nil {
		return 0, fmt.Errorf("вставка записи: %w", err)
	}
	return res2.LastInsertId()
}

// ListPlans возвращает последние N записей истории, от новых к старым.
func (d *DB) ListPlans(limit int) ([]Record, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := d.conn.Query(`
		SELECT id, subject, days_left, free_hours, difficulty, current_level,
		       result_json, created_at
		FROM plans
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("запрос истории: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		var resultJSON string
		var difficulty int
		if err := rows.Scan(&r.ID, &r.Subject, &r.Input.DaysLeft, &r.Input.FreeHours,
			&difficulty, &r.Input.CurrentLevel, &resultJSON, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("чтение строки истории: %w", err)
		}
		r.Input.Difficulty = calc.Difficulty(difficulty)
		if err := json.Unmarshal([]byte(resultJSON), &r.Result); err != nil {
			return nil, fmt.Errorf("разбор результата: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// DeletePlan удаляет запись истории по ID.
func (d *DB) DeletePlan(id int64) error {
	_, err := d.conn.Exec(`DELETE FROM plans WHERE id = ?`, id)
	return err
}
