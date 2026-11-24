package models

// Plan de Pruebas 05 — Eliminación de tareas.
// Casos CP-ELI-01 a CP-ELI-10: eliminación por id y propietario (simulador SQL).

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// La eliminación debe restringirse por id de tarea Y por created_by.
const sqlEliminarTarea = "DELETE FROM todos WHERE id = ? AND created_by = ?"

func esperarEliminacion(mock sqlmock.Sqlmock, todoID, owner, filasAfectadas int64) {
	mock.ExpectExec(sqlEliminarTarea).
		WithArgs(todoID, owner).
		WillReturnResult(sqlmock.NewResult(0, filasAfectadas))
}

// CP-ELI-01 — Eliminación de una tarea propia pendiente (RU009).
func TestCP_ELI_01_EliminarTareaPendiente(t *testing.T) {
	db, mock := nuevoSimulador(t)
	esperarEliminacion(mock, 11, usuarioA, 1)

	if err := NewTodoStore(db).Delete(usuarioA, 11); err != nil {
		t.Fatalf("se esperaba una eliminación confirmada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-02 — Eliminación de una tarea propia completada (RU009).
func TestCP_ELI_02_EliminarTareaCompletada(t *testing.T) {
	db, mock := nuevoSimulador(t)
	esperarEliminacion(mock, 12, usuarioA, 1)

	if err := NewTodoStore(db).Delete(usuarioA, 12); err != nil {
		t.Fatalf("no se permitió eliminar una tarea completada: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-03 — Intento de eliminar una tarea inexistente (RU009).
func TestCP_ELI_03_EliminarTareaInexistente(t *testing.T) {
	db, mock := nuevoSimulador(t)
	esperarEliminacion(mock, 999, usuarioA, 0)

	err := NewTodoStore(db).Delete(usuarioA, 999)

	if !errors.Is(err, ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-04 — Intento de eliminar una tarea ajena (RU012).
func TestCP_ELI_04_EliminarTareaAjena(t *testing.T) {
	db, mock := nuevoSimulador(t)
	// Tarea 22 es del usuario B; el solicitante es A -> sin coincidencia.
	esperarEliminacion(mock, 22, usuarioA, 0)

	err := NewTodoStore(db).Delete(usuarioA, 22)

	if !errors.Is(err, ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock) // no hubo un segundo DELETE solo por id
}

// CP-ELI-05 — Verificación de una eliminación limitada al registro solicitado (RU009, RU012).
func TestCP_ELI_05_EliminacionLimitada(t *testing.T) {
	db, mock := nuevoSimulador(t)
	// El simulador solo acepta "DELETE ... WHERE id = ? AND created_by = ?" con (11, 101).
	esperarEliminacion(mock, 11, usuarioA, 1)

	if err := NewTodoStore(db).Delete(usuarioA, 11); err != nil {
		t.Fatalf("la eliminación no usó las dos restricciones esperadas: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-06 — Repetición de una solicitud de eliminación (RU009).
func TestCP_ELI_06_EliminacionRepetida(t *testing.T) {
	db, mock := nuevoSimulador(t)
	esperarEliminacion(mock, 11, usuarioA, 1)
	esperarEliminacion(mock, 11, usuarioA, 0)
	store := NewTodoStore(db)

	err1 := store.Delete(usuarioA, 11)
	err2 := store.Delete(usuarioA, 11)

	if err1 != nil {
		t.Fatalf("la primera eliminación debía confirmarse: %v", err1)
	}
	if !errors.Is(err2, ErrTareaNoEncontrada) {
		t.Fatalf("la segunda eliminación afirmó borrar de nuevo: %v", err2)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-07 — Manejo de un error técnico de eliminación (RU009).
func TestCP_ELI_07_ErrorTecnico(t *testing.T) {
	db, mock := nuevoSimulador(t)
	errTecnico := errors.New("base de datos bloqueada")
	mock.ExpectExec(sqlEliminarTarea).WithArgs(int64(11), usuarioA).WillReturnError(errTecnico)

	err := NewTodoStore(db).Delete(usuarioA, 11)

	if !errors.Is(err, errTecnico) {
		t.Fatalf("se esperaba el error técnico, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-08 — Manejo de un error al comprobar las filas afectadas (RU009).
func TestCP_ELI_08_ErrorAlContarFilas(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectExec(sqlEliminarTarea).WithArgs(int64(11), usuarioA).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("filas afectadas no disponibles")))

	err := NewTodoStore(db).Delete(usuarioA, 11)

	if !errors.Is(err, ErrResultadoNoConfirmado) {
		t.Fatalf("se esperaba ErrResultadoNoConfirmado, se obtuvo: %v", err)
	}
	if errors.Is(err, ErrTareaNoEncontrada) {
		t.Fatal("se afirmó que la tarea sigue intacta, pero no se puede saber")
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-09 — Independencia de eliminaciones entre usuarios (RU009, RU012).
func TestCP_ELI_09_UsuariosDiferentes(t *testing.T) {
	db, mock := nuevoSimulador(t)
	esperarEliminacion(mock, 11, usuarioA, 1)
	esperarEliminacion(mock, 22, usuarioB, 1)
	store := NewTodoStore(db)

	errA := store.Delete(usuarioA, 11)
	errB := store.Delete(usuarioB, 22)

	if errA != nil || errB != nil {
		t.Fatalf("errores inesperados: %v / %v", errA, errB)
	}
	verificarInteracciones(t, mock)
}

// CP-ELI-10 — Eliminación válida después de una operación fallida (RU009).
func TestCP_ELI_10_EliminacionDespuesDeFallo(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectExec(sqlEliminarTarea).WithArgs(int64(11), usuarioA).WillReturnError(errors.New("error temporal"))
	esperarEliminacion(mock, 12, usuarioA, 1)
	store := NewTodoStore(db)

	err1 := store.Delete(usuarioA, 11)
	err2 := store.Delete(usuarioA, 12)

	if err1 == nil {
		t.Fatal("la primera eliminación debía informar su error")
	}
	if err2 != nil {
		t.Fatalf("la segunda eliminación heredó el error anterior: %v", err2)
	}
	verificarInteracciones(t, mock)
}
