package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Usuarios ficticios definidos en el Plan de Pruebas.
const (
	usuarioA int64 = 101
	usuarioB int64 = 202
)

// nuevoSimulador crea una base de datos simulada (go-sqlmock) que exige el texto
// exacto de cada consulta. Así se comprueba que los datos van como parámetros
// y no concatenados dentro del SQL. No se usa ningún archivo SQLite real.
func nuevoSimulador(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("no se pudo crear el simulador SQL: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// verificarInteracciones falla si quedó alguna consulta esperada sin ejecutar.
// Las consultas no esperadas ya hacen fallar la operación dentro del simulador.
func verificarInteracciones(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("interacciones con la base de datos distintas a las esperadas: %v", err)
	}
}
