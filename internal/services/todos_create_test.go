package services

// Plan de Pruebas 02 — Creación de tareas.
// Casos CP-CRE-01 a CP-CRE-12: validación del servicio + inserción SQL (simulador).

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

// esperarInsercion configura una inserción exitosa con los datos indicados.
func esperarInsercion(mock sqlmock.Sqlmock, title, desc string, owner, nuevoID int64) {
	mock.ExpectExec(sqlInsertarTarea).
		WithArgs(title, desc, owner).
		WillReturnResult(sqlmock.NewResult(nuevoID, 1))
}

// CP-CRE-01 — Creación de una tarea con datos válidos (RU005).
func TestCP_CRE_01_CreacionValida(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarInsercion(mock, "Preparar exposición", "Revisar los temas asignados", usuarioA, 1)

	todo, err := CreateTodo(store, usuarioA, "Preparar exposición", "Revisar los temas asignados")

	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo: %v", err)
	}
	esperado := models.Todo{ID: 1, Title: "Preparar exposición", Description: "Revisar los temas asignados", CreatedBy: usuarioA}
	if todo != esperado {
		t.Fatalf("tarea creada = %+v, se esperaba %+v", todo, esperado)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-02 — Creación de una tarea sin descripción (RU005).
func TestCP_CRE_02_SinDescripcion(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarInsercion(mock, "Comprar materiales", "", usuarioA, 2)

	todo, err := CreateTodo(store, usuarioA, "Comprar materiales", "")

	if err != nil {
		t.Fatalf("la falta de descripción impidió crear la tarea: %v", err)
	}
	if todo.Title != "Comprar materiales" || todo.CreatedBy != usuarioA || todo.Description != "" {
		t.Fatalf("datos de la tarea alterados: %+v", todo)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-03 — Rechazo de una tarea con título vacío (RU005).
func TestCP_CRE_03_TituloVacio(t *testing.T) {
	variantes := map[string]string{
		"cadena vacía":  "",
		"solo espacios": "   ",
	}
	for nombre, titulo := range variantes {
		t.Run(nombre, func(t *testing.T) {
			store, mock := nuevoStoreSimulado(t) // sin expectativas: cualquier escritura falla

			todo, err := CreateTodo(store, usuarioA, titulo, "Revisar los temas asignados")

			if !errors.Is(err, ErrTituloObligatorio) {
				t.Fatalf("se esperaba ErrTituloObligatorio, se obtuvo: %v", err)
			}
			if todo != (models.Todo{}) {
				t.Fatalf("se devolvió una tarea aunque fue rechazada: %+v", todo)
			}
			verificarInteracciones(t, mock)
		})
	}
}

// CP-CRE-04 — Creación con un título de un carácter (RU005).
func TestCP_CRE_04_TituloDeUnCaracter(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarInsercion(mock, "A", "", usuarioA, 3)

	todo, err := CreateTodo(store, usuarioA, "A", "")

	if err != nil {
		t.Fatalf("se rechazó un título de un carácter: %v", err)
	}
	if todo.Title != "A" {
		t.Fatalf("título enviado = %q, se esperaba \"A\"", todo.Title)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-05 — Tratamiento de caracteres Unicode (RU005).
func TestCP_CRE_05_CaracteresUnicode(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	titulo, desc := "Revisión de mañana", "Verificar información de la exposición"
	esperarInsercion(mock, titulo, desc, usuarioA, 4)

	todo, err := CreateTodo(store, usuarioA, titulo, desc)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if todo.Title != titulo || todo.Description != desc {
		t.Fatalf("se alteraron los caracteres: %+v", todo)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-06 — Tratamiento de comillas y apóstrofes (RU005).
func TestCP_CRE_06_ComillasYApostrofes(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	titulo, desc := `Revisar "Go"`, "Libro de O'Reilly"
	esperarInsercion(mock, titulo, desc, usuarioA, 5)

	todo, err := CreateTodo(store, usuarioA, titulo, desc)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if todo.Title != titulo || todo.Description != desc {
		t.Fatalf("se alteraron las comillas o apóstrofes: %+v", todo)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-07 — Parametrización de contenido con apariencia de SQL (RU005, RU012).
func TestCP_CRE_07_ContenidoConAparienciaDeSQL(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	malicioso := "Prueba'); DELETE FROM todos; --"
	// El simulador exige el SQL exacto con "?": si el texto se concatenara,
	// la consulta sería distinta y la prueba fallaría. No se espera ningún DELETE.
	esperarInsercion(mock, malicioso, "", usuarioA, 6)

	todo, err := CreateTodo(store, usuarioA, malicioso, "")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if todo.Title != malicioso || todo.CreatedBy != usuarioA {
		t.Fatalf("el contenido o el propietario cambiaron: %+v", todo)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-08 — Creación de tareas con títulos iguales (RU005).
func TestCP_CRE_08_TitulosIguales(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarInsercion(mock, "Estudiar", "Matemáticas", usuarioA, 7)
	esperarInsercion(mock, "Estudiar", "Física", usuarioA, 8)

	t1, err1 := CreateTodo(store, usuarioA, "Estudiar", "Matemáticas")
	t2, err2 := CreateTodo(store, usuarioA, "Estudiar", "Física")

	if err1 != nil || err2 != nil {
		t.Fatalf("errores inesperados: %v / %v", err1, err2)
	}
	if t1.ID == t2.ID || t1.Description != "Matemáticas" || t2.Description != "Física" {
		t.Fatalf("las tareas no se conservaron por separado: %+v / %+v", t1, t2)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-09 — Asignación del propietario correspondiente (RU005, RU012).
func TestCP_CRE_09_PropietarioCorrecto(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	// El simulador solo acepta la inserción si el tercer parámetro es 101.
	esperarInsercion(mock, "Preparar exposición", "Revisar los temas asignados", usuarioA, 9)

	todo, err := CreateTodo(store, usuarioA, "Preparar exposición", "Revisar los temas asignados")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if todo.CreatedBy != usuarioA {
		t.Fatalf("propietario = %d, se esperaba %d", todo.CreatedBy, usuarioA)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-10 — Manejo de un error de escritura (RU005).
func TestCP_CRE_10_ErrorDeEscritura(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	errTecnico := errors.New("disco lleno")
	mock.ExpectExec(sqlInsertarTarea).
		WithArgs("Preparar exposición", "Revisar los temas asignados", usuarioA).
		WillReturnError(errTecnico)

	todo, err := CreateTodo(store, usuarioA, "Preparar exposición", "Revisar los temas asignados")

	if !errors.Is(err, errTecnico) {
		t.Fatalf("se esperaba el error de escritura, se obtuvo: %v", err)
	}
	if todo != (models.Todo{}) {
		t.Fatalf("se devolvió una tarea como si se hubiera creado: %+v", todo)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-11 — Independencia entre creaciones de usuarios diferentes (RU005, RU012).
func TestCP_CRE_11_UsuariosDiferentes(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	esperarInsercion(mock, "Tarea de A", "Datos de A", usuarioA, 10)
	esperarInsercion(mock, "Tarea de B", "Datos de B", usuarioB, 11)

	ta, errA := CreateTodo(store, usuarioA, "Tarea de A", "Datos de A")
	tb, errB := CreateTodo(store, usuarioB, "Tarea de B", "Datos de B")

	if errA != nil || errB != nil {
		t.Fatalf("errores inesperados: %v / %v", errA, errB)
	}
	if ta.CreatedBy != usuarioA || tb.CreatedBy != usuarioB {
		t.Fatalf("se intercambiaron los propietarios: A=%d B=%d", ta.CreatedBy, tb.CreatedBy)
	}
	verificarInteracciones(t, mock)
}

// CP-CRE-12 — Creación válida después de una operación fallida (RU005).
func TestCP_CRE_12_CreacionDespuesDeFallo(t *testing.T) {
	store, mock := nuevoStoreSimulado(t)
	mock.ExpectExec(sqlInsertarTarea).
		WithArgs("Primera", "Falla", usuarioA).
		WillReturnError(errors.New("error temporal"))
	esperarInsercion(mock, "Segunda", "Funciona", usuarioA, 12)

	_, err1 := CreateTodo(store, usuarioA, "Primera", "Falla")
	t2, err2 := CreateTodo(store, usuarioA, "Segunda", "Funciona")

	if err1 == nil {
		t.Fatal("la primera operación debía informar su error")
	}
	if err2 != nil {
		t.Fatalf("la segunda operación heredó el error anterior: %v", err2)
	}
	if t2.Title != "Segunda" || t2.Description != "Funciona" {
		t.Fatalf("la segunda tarea usó datos incorrectos: %+v", t2)
	}
	verificarInteracciones(t, mock)
}
