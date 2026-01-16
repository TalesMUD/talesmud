package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/talesmud/talesmud/pkg/db"
)

type sqliteGenericRepo struct {
	db        *sql.DB
	table     string
	generator func() interface{}
}

func newSQLiteGenericRepo(dbConn *sql.DB, table string, gen func() interface{}) *sqliteGenericRepo {
	return &sqliteGenericRepo{
		db:        dbConn,
		table:     table,
		generator: gen,
	}
}

func (repo *sqliteGenericRepo) DropCollection() error {
	_, err := repo.db.Exec(fmt.Sprintf("DELETE FROM %s", repo.table))
	return err
}

func (repo *sqliteGenericRepo) FindByID(id string) (interface{}, error) {
	row := repo.db.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE id = ?", repo.table), id)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return nil, errors.New("entity not found")
	}
	entity := repo.generator()
	if err := json.Unmarshal([]byte(payload), entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (repo *sqliteGenericRepo) FindByField(key string, value string) (interface{}, error) {
	path := "$." + key
	row := repo.db.QueryRow(
		fmt.Sprintf("SELECT data FROM %s WHERE json_extract(data, ?) = ? LIMIT 1", repo.table),
		path,
		value,
	)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return nil, errors.New("entity not found")
	}
	entity := repo.generator()
	if err := json.Unmarshal([]byte(payload), entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (repo *sqliteGenericRepo) UpdateByField(item interface{}, key string, value string) error {
	id, err := extractEntityID(item)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	path := "$." + key
	_, err = repo.db.Exec(
		fmt.Sprintf("UPDATE %s SET id = ?, data = ? WHERE json_extract(data, ?) = ?", repo.table),
		id,
		string(payload),
		path,
		value,
	)
	return err
}

func (repo *sqliteGenericRepo) FindAllWithParam(params *db.QueryParams, collector func(element interface{})) error {
	where, args := buildWhere(params)
	query := fmt.Sprintf("SELECT data FROM %s", repo.table)
	if where != "" {
		query += " WHERE " + where
	}

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return err
		}
		elem := repo.generator()
		if err := json.Unmarshal([]byte(payload), elem); err != nil {
			continue
		}
		collector(elem)
	}
	return rows.Err()
}

func (repo *sqliteGenericRepo) FindAll(collector func(element interface{})) error {
	rows, err := repo.db.Query(fmt.Sprintf("SELECT data FROM %s", repo.table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return err
		}
		elem := repo.generator()
		if err := json.Unmarshal([]byte(payload), elem); err != nil {
			continue
		}
		collector(elem)
	}
	return rows.Err()
}

func (repo *sqliteGenericRepo) Store(entity interface{}) (interface{}, error) {
	id, err := extractEntityID(entity)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}
	_, err = repo.db.Exec(
		fmt.Sprintf("INSERT INTO %s (id, data) VALUES (?, ?)", repo.table),
		id,
		string(payload),
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (repo *sqliteGenericRepo) Delete(id string) error {
	_, err := repo.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", repo.table), id)
	return err
}

func (repo *sqliteGenericRepo) Update(item interface{}, id string) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(
		fmt.Sprintf("UPDATE %s SET data = ? WHERE id = ?", repo.table),
		string(payload),
		id,
	)
	return err
}

func buildWhere(params *db.QueryParams) (string, []interface{}) {
	if params == nil {
		return "", nil
	}
	clauses := []string{}
	args := []interface{}{}
	for _, param := range params.Params() {
		path := "$." + param.Key
		clauses = append(clauses, "json_extract(data, ?) = ?")
		args = append(args, path, param.Value)
	}
	return strings.Join(clauses, " AND "), args
}

func extractEntityID(entity interface{}) (string, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", errors.New("entity is nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", errors.New("entity is not a struct")
	}

	if field := v.FieldByName("ID"); field.IsValid() && field.Kind() == reflect.String {
		return field.String(), nil
	}

	if field := v.FieldByName("Entity"); field.IsValid() {
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			idField := field.Elem().FieldByName("ID")
			if idField.IsValid() && idField.Kind() == reflect.String {
				return idField.String(), nil
			}
		}
	}
	return "", errors.New("entity ID not found")
}
