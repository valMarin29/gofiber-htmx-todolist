package models

// Pruebas complementarias: NO forman parte de los 60 casos del Plan de Pruebas.
// Cubren código de apoyo (detección del error UNIQUE de SQLite y lectura de una
// tarea para la interfaz) y aportan a la métrica de cobertura.

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const sqlInsertarUsuario = "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"

func TestComplementaria_CreateUser_DetectaCorreoDuplicado(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectExec(sqlInsertarUsuario).
		WithArgs("usuario_b", "usuario.a@example.com", "hash").
		WillReturnError(errors.New("constraint failed: UNIQUE constraint failed: users.email (2067)"))

	err := NewUserStore(db).CreateUser("usuario_b", "usuario.a@example.com", "hash")

	if !errors.Is(err, ErrCorreoDuplicado) {
		t.Fatalf("se esperaba ErrCorreoDuplicado, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

func TestComplementaria_CreateUser_Exitoso(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectExec(sqlInsertarUsuario).
		WithArgs("usuario_a", "usuario.a@example.com", "hash").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := NewUserStore(db).CreateUser("usuario_a", "usuario.a@example.com", "hash"); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	verificarInteracciones(t, mock)
}

func TestComplementaria_GetByUser(t *testing.T) {
	const q = "SELECT id, title, description, completed, created_by FROM todos WHERE id = ? AND created_by = ?"

	t.Run("tarea propia", func(t *testing.T) {
		db, mock := nuevoSimulador(t)
		mock.ExpectQuery(q).WithArgs(int64(11), usuarioA).
			WillReturnRows(sqlmock.NewRows(columnasTarea).AddRow(int64(11), "Título", "", true, usuarioA))

		td, err := NewTodoStore(db).GetByUser(usuarioA, 11)
		if err != nil || td.ID != 11 || !td.Completed {
			t.Fatalf("resultado inesperado: %+v, %v", td, err)
		}
		verificarInteracciones(t, mock)
	})

	t.Run("tarea ajena o inexistente", func(t *testing.T) {
		db, mock := nuevoSimulador(t)
		mock.ExpectQuery(q).WithArgs(int64(22), usuarioA).WillReturnRows(sqlmock.NewRows(columnasTarea))

		if _, err := NewTodoStore(db).GetByUser(usuarioA, 22); !errors.Is(err, ErrTareaNoEncontrada) {
			t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
		}
		verificarInteracciones(t, mock)
	})
}
