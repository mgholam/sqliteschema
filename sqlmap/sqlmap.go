// Package sqlmap provides functions for querying directly into
// map[string]interface{}.
//
// In developing really simple api endpoints, I found the boilerplate needed
// to take the results of a database query and output them as JSON to be
// really fucking annoying; make a custom struct, scan into that struct, if
// there are multiple rows do the whole rows.Next() song and dance, and if
// anything changes update the three spots each column of the result is now
// dependent on. Even when using libraries like sqlx, there's still a lot of
// extraneous code that needs to be written.
//
// This package makes that kind of thing considerably easier. Instead of using
// a custom struct, it just scans into a generic map[string]interface{}. These
// maps generally serialize to JSON pretty well, so it's about as direct a
// conversion as is possible.
//
// Occasionally (especially when dealing with the claims database which
// inexplicably uses decimal columns for storing ints) you may need to push
// some of your display logic into the query itself to get the output to look
// right. This is because the results are converted directly from the
// underlying sql data type; if you have a `foo_flag` column that's an int,
// it's going to show up as a number in your JSON. If you want it to show up
// as a boolean, you need to `CAST(foo_flag AS BIT)`. For decimal columns you
// want to show up as numbers, this means `CAST(<column> AS INT)`. This can
// get complicated, but doing it in the query is usually easier than doing it
// in code where you'll need to do some hairy type assertions.
//
// I'll mention that sqlx has functionality similar to this, but the api was
// more annoying to use in part because of the assumption that I'm using their
// sqlx datastructures. This version is faster and simpler.
package sqlmap

import "database/sql"

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func Select(db Queryer, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return SelectScan(rows)
}

func SelectScan(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	numColumns := len(columns)

	values := make([]interface{}, numColumns)
	for i := range values {
		values[i] = new(interface{})
	}

	var results []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}

		dest := make(map[string]interface{}, numColumns)
		for i, column := range columns {
			dest[column] = *(values[i].(*interface{}))
		}
		results = append(results, dest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func Get(db Queryer, query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return GetScan(rows)
}

func GetScan(rows *sql.Rows) (map[string]interface{}, error) {
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	numColumns := len(columns)

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	values := make([]interface{}, numColumns)
	for i := range values {
		values[i] = new(interface{})
	}

	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, numColumns)
	for i, column := range columns {
		result[column] = *(values[i].(*interface{}))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
