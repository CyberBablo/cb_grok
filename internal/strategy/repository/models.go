package repository

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"cb_grok/internal/strategy"
)

// Strategy represents a row in the strategy table
type Strategy struct {
	ID          int64        `db:"id"`
	Symbol      string       `db:"symbol"`
	Timeframe   string       `db:"timeframe"`
	Trials      int          `db:"trials"`
	Workers     int          `db:"workers"`
	ValSetDays  int          `db:"val_set_days"`
	TrainSetDay int          `db:"train_set_day"`
	WinRate     float64      `db:"win_rate"`
	Data        StrategyData `db:"data"`
	FromDt      int64        `db:"from_dt"`
	ToDt        int64        `db:"to_dt"`
	DT          time.Time    `db:"dt"`
}

// StrategyAcc represents a row in the strategy_acc table
type StrategyAcc struct {
	ID         int64     `db:"id"`
	StrategyID int64     `db:"strategy_id"`
	Symbol     string    `db:"symbol"`
	Timeframe  string    `db:"timeframe"`
	DTUpd      time.Time `db:"dt_upd"`
	DTCreate   time.Time `db:"dt_create"`
}

// StrategyData is a wrapper around StrategyParams for JSON serialization/deserialization
type StrategyData strategy.StrategyParams

// ParamsData is a JSON data type for additional parameters
type ParamsData map[string]interface{}

// Value implements the driver.Valuer interface for JSON serialization
func (d StrategyData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface for JSON deserialization
func (d *StrategyData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &d)
}

// Value implements the driver.Valuer interface for JSON serialization
func (d ParamsData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface for JSON deserialization
func (d *ParamsData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	// Handle null JSON values
	if len(b) == 0 {
		*d = make(ParamsData)
		return nil
	}

	return json.Unmarshal(b, &d)
}
