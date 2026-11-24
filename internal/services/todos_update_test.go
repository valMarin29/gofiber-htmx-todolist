package services

// Plan de Pruebas 04 — Actualización de tareas y cambio de estado.
// Casos CP-ACT-01 a CP-ACT-14. La edición de contenido pasa por el servicio
// UpdateTodo (que valida el título); el cambio de estado recibe un valor
// explícito (completada = true / pendiente = false) en TodoStore.SetCompleted.

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

const (
	tareaA int64 = 11 // tarea del usuario A
	tareaB int64 = 22 // tarea del usuario B
)

func esperarEdicion(mock sqlmock.Sqlmock, title, desc string, todoID, owner, filasAfectadas int64) {
	mock.ExpectExec(sqlEditarTarea).
		WithArgs(title, desc, todoID, owner).
		WillReturnResult(sqlmock.NewResult(0, filasAfectadas))
}

func esperarCambioEstado(mock sqlmock.Sqlmock, completed bool, todoID, owner, filasAfectadas int64) {
	mock.ExpectExec(sqlCambiarEstado).
		WithArgs(completed, todoID, owner).
		WillReturnResult(sqlmock.NewResult(0, filasAfectadas))
}

// CP-ACT-01 — Modificación válida del título y la descripción (RU007).
func TestCP_ACT_01_EdicionValida(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarEdicion(mock, "Preparar exposición final", "Incluir ejemplos prácticos", tareaA, usuarioA, 1)

	err := UpdateTodo(store, usuarioA, tareaA, "Preparar exposición final", "Incluir ejemplos prácticos")

	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-02 — Modificación del título conservando la descripción (RU007).
func TestCP_ACT_02_CambiaTituloConservaDescripcion(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	// La operación envía ambos campos: se proporciona la descripción original.
	esperarEdicion(mock, "Nuevo título", "Revisar material", tareaA, usuarioA, 1)

	if err := UpdateTodo(store, usuarioA, tareaA, "Nuevo título", "Revisar material"); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-03 — Modificación de la descripción conservando el título (RU007).
func TestCP_ACT_03_CambiaDescripcionConservaTitulo(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarEdicion(mock, "Preparar exposición", "Nueva descripción", tareaA, usuarioA, 1)

	if err := UpdateTodo(store, usuarioA, tareaA, "Preparar exposición", "Nueva descripción"); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-04 — Eliminación del contenido de la descripción (RU005, RU007).
func TestCP_ACT_04_VaciarDescripcion(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarEdicion(mock, "Preparar exposición", "", tareaA, usuarioA, 1)

	if err := UpdateTodo(store, usuarioA, tareaA, "Preparar exposición", ""); err != nil {
		t.Fatalf("no se permitió vaciar la descripción: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-05 — Rechazo de una edición con título vacío (RU005, RU007).
func TestCP_ACT_05_EdicionConTituloVacio(t *testing.T) {
	for nombre, titulo := range map[string]string{"cadena vacía": "", "solo espacios": "  "} {
		t.Run(nombre, func(t *testing.T) {
			store, mock := nuevoStoreSimulado(t) // sin expectativas: cualquier escritura falla

			err := UpdateTodo(store, usuarioA, tareaA, titulo, "Revisar material")

			if !errors.Is(err, ErrTituloObligatorio) {
				t.Fatalf("se esperaba ErrTituloObligatorio, se obtuvo: %v", err)
			}
			verificarInteracciones(t, mock)
		})
	}
}

// CP-ACT-06 — Tratamiento seguro de caracteres especiales durante la edición (RU007, RU012).
func TestCP_ACT_06_CaracteresEspecialesEnEdicion(t *testing.T) {
	subcasos := map[string][2]string{
		"unicode":        {"Revisión de mañana", "Información añadida"},
		"comillas":       {`Revisar "Go"`, "Libro de O'Reilly"},
		"apariencia SQL": {"Prueba'); DROP TABLE todos; --", "' OR '1'='1"},
	}
	for nombre, datos := range subcasos {
		t.Run(nombre, func(t *testing.T) {
			store, mock := nuevoStoreSimulado(t)
			esperarEdicion(mock, datos[0], datos[1], tareaA, usuarioA, 1)

			if err := UpdateTodo(store, usuarioA, tareaA, datos[0], datos[1]); err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			verificarInteracciones(t, mock)
		})
	}
}

// CP-ACT-07 — Intento de editar una tarea inexistente (RU007).
func TestCP_ACT_07_EditarTareaInexistente(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarEdicion(mock, "Título", "Desc", 999, usuarioA, 0) // cero filas afectadas

	err := UpdateTodo(store, usuarioA, 999, "Título", "Desc")

	if !errors.Is(err, models.ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock) // no se intentó crear la tarea
}

// CP-ACT-08 — Intento de editar una tarea de otro usuario (RU012).
func TestCP_ACT_08_EditarTareaAjena(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	// Filtro conjunto: id = 22 AND created_by = 101 -> sin coincidencia.
	esperarEdicion(mock, "Intento", "Ajeno", tareaB, usuarioA, 0)

	err := UpdateTodo(store, usuarioA, tareaB, "Intento", "Ajeno")

	if !errors.Is(err, models.ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock) // no hubo una segunda actualización solo por id
}

// CP-ACT-09 — Cambio de pendiente a completada (RU008).
func TestCP_ACT_09_PendienteACompletada(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarCambioEstado(mock, true, tareaA, usuarioA, 1)

	if err := store.SetCompleted(usuarioA, tareaA, true); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	verificarInteracciones(t, mock) // solo cambia completed, no título ni descripción
}

// CP-ACT-10 — Cambio de completada a pendiente (RU008).
func TestCP_ACT_10_CompletadaAPendiente(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarCambioEstado(mock, false, 12, usuarioA, 1)

	if err := store.SetCompleted(usuarioA, 12, false); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-11 — Intento de cambiar el estado de una tarea inexistente (RU008).
func TestCP_ACT_11_EstadoDeTareaInexistente(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarCambioEstado(mock, true, 999, usuarioA, 0)

	err := store.SetCompleted(usuarioA, 999, true)

	if !errors.Is(err, models.ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-12 — Intento de cambiar el estado de una tarea ajena (RU012).
func TestCP_ACT_12_EstadoDeTareaAjena(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarCambioEstado(mock, true, tareaB, usuarioA, 0)

	err := store.SetCompleted(usuarioA, tareaB, true)

	if !errors.Is(err, models.ErrTareaNoEncontrada) {
		t.Fatalf("se esperaba ErrTareaNoEncontrada, se obtuvo: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-ACT-13 — Manejo de errores técnicos de actualización (RU007, RU008).
func TestCP_ACT_13_ErroresTecnicos(t *testing.T) {
	errTecnico := errors.New("base de datos bloqueada")

	t.Run("edición de contenido", func(t *testing.T) {
		store, mock := nuevoStoreSimulado(t)
		mock.ExpectExec(sqlEditarTarea).WithArgs("Título", "Desc", tareaA, usuarioA).WillReturnError(errTecnico)

		err := UpdateTodo(store, usuarioA, tareaA, "Título", "Desc")

		if !errors.Is(err, errTecnico) {
			t.Fatalf("se esperaba el error técnico, se obtuvo: %v", err)
		}
		verificarInteracciones(t, mock)
	})

	t.Run("cambio de estado", func(t *testing.T) {
		store, mock := nuevoStoreSimulado(t)
		mock.ExpectExec(sqlCambiarEstado).WithArgs(true, tareaA, usuarioA).WillReturnError(errTecnico)

		err := store.SetCompleted(usuarioA, tareaA, true)

		if !errors.Is(err, errTecnico) {
			t.Fatalf("se esperaba el error técnico, se obtuvo: %v", err)
		}
		verificarInteracciones(t, mock)
	})
}

// CP-ACT-14 — Independencia de actualizaciones entre solicitudes (RU007, RU008, RU012).
func TestCP_ACT_14_SolicitudesIndependientes(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarEdicion(mock, "Edición de A", "Datos de A", tareaA, usuarioA, 1)
	esperarCambioEstado(mock, true, tareaB, usuarioB, 1)

	errA := UpdateTodo(store, usuarioA, tareaA, "Edición de A", "Datos de A")
	errB := store.SetCompleted(usuarioB, tareaB, true)

	if errA != nil || errB != nil {
		t.Fatalf("errores inesperados: %v / %v", errA, errB)
	}
	verificarInteracciones(t, mock)
}
