package mysql

import (
	"database/sql"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"log"
)

type ConfigRepository struct {
	db *sqlx.DB
}

func NewConfigRepository(connection *sqlx.DB) ConfigRepository {
	return ConfigRepository{
		db: connection,
	}
}

func (r *ConfigRepository) GetValues() map[string]string {
	var configValues map[string]string

	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.From("config").
		Prepared(true).
		ToSQL()

	_ = r.db.Select(&configValues, query, args...)

	return configValues
}

func (r *ConfigRepository) Exist(key string) bool {
	var value string

	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.From("config").
		Prepared(true).
		Select("config_value").
		Where(goqu.C("config_key").Eq(key)).ToSQL()

	err := r.db.Get(&value, query, args...)

	if err == sql.ErrNoRows {
		return false
	}

	return true
}

func (r *ConfigRepository) GetValue(key string) (string, error) {
	var value string

	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.From("config").
		Prepared(true).
		Where(goqu.C("config_key").Eq(key)).ToSQL()

	err := r.db.Select(&value, query, args...)

	if err != nil {
		return "", err
	}

	return value, nil
}

func (r *ConfigRepository) SetValue(key string, value string) error {
	exist := r.Exist(key)
	dialect := goqu.Dialect("mysql")

	if exist {
		//update
		log.Println("Update key value", key, value)
		query, args, _ := dialect.Update("config").
			Prepared(true).
			Set(goqu.Record{"config_value": value}).
			Where(goqu.C("config_key").Eq(key)).ToSQL()

		_, err := r.db.Exec(query, args...)

		return err
	}

	//insert
	log.Println("Insert key value", value, key)

	query, args, _ := dialect.Insert("config").
		Prepared(true).
		Rows(goqu.Record{"config_key": key, "config_value": value}).
		ToSQL()

	_, err := r.db.Exec(query, args...)

	return err
}
