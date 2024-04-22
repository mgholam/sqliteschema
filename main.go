package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sqliteschema/sqlmap"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// syncSchema("ppp.db", "pp.json")
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}
	if args[0] == "sync" {
		if len(args) < 3 {
			printHelp()
			return
		}
		syncSchema(args[1], args[2])
		return
	} else if args[0] == "extract" {
		if len(args) < 2 {
			printHelp()
			return
		}
		extract(args[1])
		return
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("args:")
	fmt.Println("      sync    <dbname.db> <schema.json>")
	fmt.Println("      extract <dbname.db>")
}

type Column_Def struct {
	Cid        int    `json:"cid"`
	Dflt_value string `json:"dflt_value"`
	Name       string `json:"name"`
	Notnull    int    `json:"notnull"`
	Pk         int    `json:"pk"`
	CType      string `json:"type"`
}

type Schema map[string][]Column_Def

func syncSchema(dbpath, schemapath string) {
	sf, err := os.ReadFile(schemapath)
	if err != nil {
		log.Fatal(err)
		return
	}
	var schema Schema
	err = json.Unmarshal(sf, &schema)
	if err != nil {
		log.Fatal(err)
		return
	}
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal("error", err)
		return
	}
	defer db.Close()

	tables := getTables(db)
	// fmt.Println(tables)
	// fmt.Println(schema)

	for k, v := range schema {
		exists := tables[k]
		if !exists {
			createTable(db, k, v)
			continue
		}
		syncColumns(db, k, v)
	}
}

func syncColumns(db *sql.DB, tablename string, schema []Column_Def) {
	dbtable := getTableDef(db, tablename)

	for _, c := range schema {
		// fix: column default changed, column type changed
		if hasColumn(dbtable, c) {
			continue
		}

		addColumn(db, tablename, c)
	}
}

func hasColumn(dbc []map[string]any, newc Column_Def) bool {
	ret := false
	for _, i := range dbc {
		if i["name"].(string) == newc.Name {
			ret = true
		}
	}
	return ret
}

func addColumn(db *sql.DB, tablename string, coldef Column_Def) {
	sb := strings.Builder{}

	sb.WriteString(`alter table "`)
	sb.WriteString(tablename)
	sb.WriteString(`" add column "`)
	sb.WriteString(coldef.Name)
	sb.WriteString(`" `)
	sb.WriteString(coldef.CType)
	if coldef.Dflt_value != "" {
		sb.WriteString(" default ")
		sb.WriteString(coldef.Dflt_value)
	}

	// log.Println(sb.String())
	_, err := db.Exec(sb.String())
	if err != nil {
		log.Println(err)
	}
	log.Println("adding column", coldef.Name, "to table", tablename)
	sb.Reset()
}

func createTable(db *sql.DB, tablename string, schema []Column_Def) {
	log.Println("creating table", tablename)
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(`create table "%s" (`, tablename))
	var pk Column_Def
	pkfound := false
	for i, c := range schema {
		if c.Pk == 1 {
			pk = c
			pkfound = true
		}
		sb.WriteString(`"`)
		sb.WriteString(c.Name)
		sb.WriteString(`" `)
		sb.WriteString(c.CType)
		if c.Notnull == 1 {
			sb.WriteString(" not null ")
		}
		if c.Dflt_value != "" {
			sb.WriteString(" default ")
			sb.WriteString(c.Dflt_value)
		}
		if i < len(schema)-1 {
			sb.WriteString(", ")
		}
	}
	if pkfound {
		sb.WriteString(`, primary key ("`)
		sb.WriteString(pk.Name)
		sb.WriteString(`")`)
	}
	sb.WriteString(")")
	_, err := db.Exec(sb.String())
	if err != nil {
		log.Println(err)
	}
}

func getTables(db *sql.DB) map[string]bool {
	tables := make(map[string]bool)
	rows, _ := db.Query("pragma table_list")
	o, err := sqlmap.SelectScan(rows)
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, i := range o {
		tables[i["name"].(string)] = true
	}
	return tables
}

func getTableDef(db *sql.DB, tablename string) []map[string]any {
	log.Println("table name", tablename)
	rows, _ := db.Query("pragma table_info(" + tablename + ")")
	oo, _ := sqlmap.SelectScan(rows)
	return oo
}

func extract(dbpath string) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal("error", err)
		return
	}
	defer db.Close()
	rows, _ := db.Query("pragma table_list")
	o, err := sqlmap.SelectScan(rows)
	if err != nil {
		log.Println(err)
		return
	}
	output := make(map[string]any)
	for _, i := range o {
		name := i["name"].(string)
		if strings.HasPrefix(name, "sqlite_") {
			continue
		}
		output[name] = getTableDef(db, name)
	}
	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(b))
}
