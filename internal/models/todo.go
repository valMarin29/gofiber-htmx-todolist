package models

import (
	"database/sql"
	"fmt"
)

// Todo representa una tarea de un usuario.
type Todo struct {
	ID          int64
	Title       string
	Description string
	Completed   bool
	CreatedBy   int64
}

// Consultas SQL de tareas. Todas filtran por created_by para que cada usuario
// solo vea y modifique sus propias tareas (RU012).
const (
	QueryInsertTodo = "INSERT INTO todos (title, description, created_by) VALUES (?, ?, ?)"
	// El esquema no tiene fecha de creación: la más reciente es la de id mayor.
	QueryListTodos     = "SELECT id, title, description, completed, created_by FROM todos WHERE created_by = ? ORDER BY id DESC"
	QueryUpdateTodo    = "UPDATE todos SET title = ?, description = ? WHERE id = ? AND created_by = ?"
	QuerySetCompleted  = "UPDATE todos SET completed = ? WHERE id = ? AND created_by = ?"
	QueryDeleteTodo    = "DELETE FROM todos WHERE id = ? AND created_by = ?"
	QueryGetTodoByUser = "SELECT id, title, description, completed, created_by FROM todos WHERE id = ? AND created_by = ?"
)

// TodoStore accede a la tabla todos.
type TodoStore struct {
	DB *sql.DB
}

// NewTodoStore crea un store de tareas.
func NewTodoStore(db *sql.DB) *TodoStore {
	return &TodoStore{DB: db}
}

// Create inserta una tarea y devuelve la tarea creada con su id.
func (s *TodoStore) Create(userID int64, title, description string) (Todo, error) {
	res, err := s.DB.Exec(QueryInsertTodo, title, description, userID)
	if err != nil {
		return Todo{}, fmt.Errorf("crear tarea: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Todo{}, fmt.Errorf("crear tarea: %w", err)
	}
	return Todo{ID: id, Title: title, Description: description, CreatedBy: userID}, nil
}

// ListByUser devuelve las tareas del usuario, de la más reciente a la más antigua.
// Si no tiene tareas devuelve un listado vacío (no nil) y sin error.
// Si hay un error devuelve nil, para no presentar un listado parcial como completo.
func (s *TodoStore) ListByUser(userID int64) ([]Todo, error) {
	rows, err := s.DB.Query(QueryListTodos, userID)
	if err != nil {
		return nil, fmt.Errorf("consultar tareas: %w", err)
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedBy); err != nil {
			return nil, fmt.Errorf("leer tarea: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer tareas: %w", err)
	}
	return todos, nil
}

// GetByUser devuelve una tarea del usuario (usada por la interfaz para editar).
func (s *TodoStore) GetByUser(userID, todoID int64) (Todo, error) {
	var t Todo
	err := s.DB.QueryRow(QueryGetTodoByUser, todoID, userID).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedBy)
	if err == sql.ErrNoRows {
		return Todo{}, ErrTareaNoEncontrada
	}
	if err != nil {
		return Todo{}, fmt.Errorf("consultar tarea: %w", err)
	}
	return t, nil
}

// Update cambia título y descripción de una tarea del usuario.
func (s *TodoStore) Update(userID, todoID int64, title, description string) error {
	res, err := s.DB.Exec(QueryUpdateTodo, title, description, todoID, userID)
	return confirmarUnaFila(res, err, "actualizar tarea")
}

// SetCompleted cambia el estado de una tarea del usuario al valor indicado.
func (s *TodoStore) SetCompleted(userID, todoID int64, completed bool) error {
	res, err := s.DB.Exec(QuerySetCompleted, completed, todoID, userID)
	return confirmarUnaFila(res, err, "cambiar estado")
}

// Delete elimina una tarea del usuario.
func (s *TodoStore) Delete(userID, todoID int64) error {
	res, err := s.DB.Exec(QueryDeleteTodo, todoID, userID)
	return confirmarUnaFila(res, err, "eliminar tarea")
}

// confirmarUnaFila revisa el resultado de un UPDATE o DELETE:
// error técnico -> error; 0 filas -> ErrTareaNoEncontrada;
// no se pudo leer filas afectadas -> ErrResultadoNoConfirmado; 1 fila -> nil.
func confirmarUnaFila(res sql.Result, err error, op string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, ErrResultadoNoConfirmado, err)
	}
	if n == 0 {
		return ErrTareaNoEncontrada
	}
	return nil
}
