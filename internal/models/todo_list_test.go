package models

// Plan de Pruebas 03 — Consulta de tareas.
// Casos CP-CON-01 a CP-CON-12: listado de tareas por usuario (simulador SQL).

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// La consulta debe filtrar por created_by y ordenar de la más reciente (id mayor)
// a la más antigua. El esquema no tiene fecha de creación.
const sqlListarTareas = "SELECT id, title, description, completed, created_by FROM todos WHERE created_by = ? ORDER BY id DESC"

var columnasTarea = []string{"id", "title", "description", "completed", "created_by"}

func filas() *sqlmock.Rows { return sqlmock.NewRows(columnasTarea) }

// CP-CON-01 — Consulta de un usuario sin tareas (RU006).
func TestCP_CON_01_UsuarioSinTareas(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).WillReturnRows(filas())

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("la ausencia de tareas se trató como error: %v", err)
	}
	if todos == nil || len(todos) != 0 {
		t.Fatalf("se esperaba un listado vacío, se obtuvo %#v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-02 — Recuperación de una única tarea (RU006).
func TestCP_CON_02_UnaTarea(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().AddRow(int64(11), "Preparar exposición", "Revisar material", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	esperado := []Todo{{ID: 11, Title: "Preparar exposición", Description: "Revisar material", Completed: false, CreatedBy: usuarioA}}
	if !reflect.DeepEqual(todos, esperado) {
		t.Fatalf("listado = %+v, se esperaba %+v", todos, esperado)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-03 — Recuperación de múltiples tareas (RU006).
func TestCP_CON_03_VariasTareas(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().
			AddRow(int64(13), "Tercera", "C", false, usuarioA).
			AddRow(int64(12), "Segunda", "B", true, usuarioA).
			AddRow(int64(11), "Primera", "A", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	esperado := []Todo{
		{ID: 13, Title: "Tercera", Description: "C", CreatedBy: usuarioA},
		{ID: 12, Title: "Segunda", Description: "B", Completed: true, CreatedBy: usuarioA},
		{ID: 11, Title: "Primera", Description: "A", CreatedBy: usuarioA},
	}
	if !reflect.DeepEqual(todos, esperado) {
		t.Fatalf("listado = %+v, se esperaba %+v", todos, esperado)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-04 — Ordenamiento de la más reciente a la más antigua (RU006).
func TestCP_CON_04_OrdenDescendente(t *testing.T) {
	db, mock := nuevoSimulador(t)
	// El orden lo aplica SQL: el simulador exige "ORDER BY id DESC" en la consulta
	// y devuelve las filas en ese orden.
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().
			AddRow(int64(30), "Nueva", "", false, usuarioA).
			AddRow(int64(20), "Media", "", false, usuarioA).
			AddRow(int64(10), "Antigua", "", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	var ids []int64
	for _, td := range todos {
		ids = append(ids, td.ID)
	}
	if !reflect.DeepEqual(ids, []int64{30, 20, 10}) {
		t.Fatalf("orden = %v, se esperaba [30 20 10]", ids)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-05 — Inclusión de tareas pendientes y completadas (RU006, RU008).
func TestCP_CON_05_PendientesYCompletadas(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().
			AddRow(int64(12), "Completada", "", true, usuarioA).
			AddRow(int64(11), "Pendiente", "", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("se esperaban 2 tareas, se obtuvieron %d", len(todos))
	}
	if !todos[0].Completed || todos[1].Completed {
		t.Fatalf("no se conservó el estado de cada tarea: %+v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-06 — Recuperación de una descripción vacía (RU005, RU006).
func TestCP_CON_06_DescripcionVacia(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().
			AddRow(int64(12), "Con descripción", "Tiene texto", false, usuarioA).
			AddRow(int64(11), "Sin descripción", "", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(todos) != 2 || todos[1].Description != "" {
		t.Fatalf("la descripción vacía no se conservó: %+v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-07 — Correspondencia de caracteres especiales (RU006).
func TestCP_CON_07_CaracteresEspeciales(t *testing.T) {
	db, mock := nuevoSimulador(t)
	titulo, desc := "Revisión de mañana", `Consultar "Go" y O'Reilly`
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().AddRow(int64(11), titulo, desc, false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(todos) != 1 || todos[0].Title != titulo || todos[0].Description != desc {
		t.Fatalf("se alteraron los caracteres especiales: %+v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-08 — Aplicación del filtro por propietario (RU012).
func TestCP_CON_08_FiltroPorPropietario(t *testing.T) {
	db, mock := nuevoSimulador(t)
	// Solo se acepta la consulta con "WHERE created_by = ?" y el valor 101.
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).WillReturnRows(filas())

	if _, err := NewTodoStore(db).ListByUser(usuarioA); err != nil {
		t.Fatalf("la consulta no incluyó el filtro por propietario esperado: %v", err)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-09 — Independencia de consultas entre dos usuarios (RU006, RU012).
func TestCP_CON_09_DosUsuarios(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().AddRow(int64(11), "Tarea exclusiva A", "", false, usuarioA))
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioB).
		WillReturnRows(filas().AddRow(int64(22), "Tarea exclusiva B", "", false, usuarioB))
	store := NewTodoStore(db)

	deA, errA := store.ListByUser(usuarioA)
	deB, errB := store.ListByUser(usuarioB)

	if errA != nil || errB != nil {
		t.Fatalf("errores inesperados: %v / %v", errA, errB)
	}
	if len(deA) != 1 || deA[0].Title != "Tarea exclusiva A" || deA[0].CreatedBy != usuarioA {
		t.Fatalf("listado de A incorrecto: %+v", deA)
	}
	if len(deB) != 1 || deB[0].Title != "Tarea exclusiva B" || deB[0].CreatedBy != usuarioB {
		t.Fatalf("listado de B incorrecto (¿se reutilizó el de A?): %+v", deB)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-10 — Manejo de un error inicial de consulta (RU006).
func TestCP_CON_10_ErrorInicialDeConsulta(t *testing.T) {
	db, mock := nuevoSimulador(t)
	errTecnico := errors.New("no se pudo abrir la base de datos")
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).WillReturnError(errTecnico)

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if !errors.Is(err, errTecnico) {
		t.Fatalf("se esperaba el error de consulta, se obtuvo: %v", err)
	}
	if todos != nil {
		t.Fatalf("el fallo se confundió con un usuario sin tareas: %#v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-11 — Manejo de un error al convertir una fila (RU006).
func TestCP_CON_11_ErrorDeConversion(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().AddRow("no-es-un-numero", "Título", "", false, usuarioA))

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if err == nil {
		t.Fatal("no se detectó el error de conversión")
	}
	if todos != nil {
		t.Fatalf("se devolvió una tarea con valores por defecto: %#v", todos)
	}
	verificarInteracciones(t, mock)
}

// CP-CON-12 — Manejo de un error durante la iteración (RU006).
func TestCP_CON_12_ErrorDuranteIteracion(t *testing.T) {
	db, mock := nuevoSimulador(t)
	errLectura := errors.New("fallo de lectura")
	mock.ExpectQuery(sqlListarTareas).WithArgs(usuarioA).
		WillReturnRows(filas().
			AddRow(int64(12), "Válida", "", false, usuarioA).
			AddRow(int64(11), "Nunca se lee", "", false, usuarioA).
			RowError(1, errLectura)).
		RowsWillBeClosed()

	todos, err := NewTodoStore(db).ListByUser(usuarioA)

	if !errors.Is(err, errLectura) {
		t.Fatalf("se esperaba el error de iteración, se obtuvo: %v", err)
	}
	if todos != nil {
		t.Fatalf("se presentó un listado parcial como completo: %+v", todos)
	}
	verificarInteracciones(t, mock) // también comprueba que las filas se cerraron
}
