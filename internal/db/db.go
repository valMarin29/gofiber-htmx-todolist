// Package db abre la base de datos SQLite y crea las tablas si no existen (RU015).
package db

import (
	"database/sql"

	_ "modernc.org/sqlite" // driver SQLite en Go puro (no necesita CGO)
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	username VARCHAR(64)  NOT NULL,
	email    VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS todos (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       VARCHAR(64)  NOT NULL,
	description VARCHAR(255) NOT NULL DEFAULT '',
	completed   BOOLEAN      NOT NULL DEFAULT 0,
	created_by  INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE
);`

// Open abre (o crea) el archivo SQLite y aplica el esquema.
func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
