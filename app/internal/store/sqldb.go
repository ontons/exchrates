package store

import (
	"database/sql"
	"exchrates/internal/provider"
)

type SqlDB struct {
	db *sql.DB
}

func NewSqlDB(db *sql.DB) *SqlDB {
	ret := &SqlDB{db: db}
	ret.Migrate()
	return ret
}

func (s *SqlDB) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS rates (
		currency VARCHAR(10) NOT NULL,
		value DECIMAL(18,8) NOT NULL,
		date DATETIME(6) NOT NULL,
		PRIMARY KEY (currency, date)
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *SqlDB) SaveRates(rates []provider.Rate) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO rates(currency, value, date) VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, rate := range rates {
		_, err := stmt.Exec(rate.Currency, rate.Value, rate.Date)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SqlDB) GetLatest() ([]provider.Rate, error) {
	query := `
        SELECT r.currency, r.value, r.date
        FROM rates r
        JOIN (
            SELECT currency, MAX(date) AS latest_date
            FROM rates
            GROUP BY currency
        ) t
        ON r.currency = t.currency AND r.date = t.latest_date;
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []provider.Rate

	for rows.Next() {
		var r provider.Rate
		if err := rows.Scan(&r.Currency, &r.Value, &r.Date); err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

func (s *SqlDB) GetHistory(currency string) ([]provider.Rate, error) {
	query := `
		SELECT currency, value, date
		FROM rates
		WHERE currency = ?
		ORDER BY date DESC
	`

	rows, err := s.db.Query(query, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []provider.Rate
	for rows.Next() {
		var rate provider.Rate
		if err := rows.Scan(&rate.Currency, &rate.Value, &rate.Date); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}
