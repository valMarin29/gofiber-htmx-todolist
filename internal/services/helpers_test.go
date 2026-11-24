package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

// Usuarios ficticios definidos en el Plan de Pruebas.
const (
	usuarioA int64 = 101
	usuarioB int64 = 202
)

// SQL que el componente debe solicitar. Se escribe aquí de forma literal para que
// la prueba falle si la consulta cambia o si los datos se concatenan en el texto.
const (
	sqlInsertarTarea = "INSERT INTO todos (title, description, created_by) VALUES (?, ?, ?)"
	sqlEditarTarea   = "UPDATE todos SET title = ?, description = ? WHERE id = ? AND created_by = ?"
	sqlCambiarEstado = "UPDATE todos SET completed = ? WHERE id = ? AND created_by = ?"
)

// nuevoStoreSimulado crea un TodoStore real conectado a un simulador SQL
// (go-sqlmock) que exige el texto exacto de cada consulta.
func nuevoStoreSimulado(t *testing.T) (*models.TodoStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("no se pudo crear el simulador SQL: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return models.NewTodoStore(db), mock
}

// verificarInteracciones falla si quedó alguna consulta esperada sin ejecutar.
// Las consultas no esperadas ya hacen fallar la operación dentro del simulador.
func verificarInteracciones(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("interacciones con la base de datos distintas a las esperadas: %v", err)
	}
}
