package services

import (
	"errors"
	"strings"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

// ErrTituloObligatorio: la tarea no tiene título (RU005).
var ErrTituloObligatorio = errors.New("el título es obligatorio")

// TodoRepository es lo que necesita el servicio de tareas.
// models.TodoStore lo cumple.
type TodoRepository interface {
	Create(userID int64, title, description string) (models.Todo, error)
	Update(userID, todoID int64, title, description string) error
}

// CreateTodo valida y crea una tarea del usuario (RU005, RU012).
// El título es obligatorio; la descripción es opcional.
// Si el título está vacío no se llama a la base de datos.
func CreateTodo(repo TodoRepository, userID int64, title, description string) (models.Todo, error) {
	if strings.TrimSpace(title) == "" {
		return models.Todo{}, ErrTituloObligatorio
	}
	return repo.Create(userID, title, description)
}

// UpdateTodo valida y edita el título y la descripción de una tarea del usuario (RU007).
func UpdateTodo(repo TodoRepository, userID, todoID int64, title, description string) error {
	if strings.TrimSpace(title) == "" {
		return ErrTituloObligatorio
	}
	return repo.Update(userID, todoID, title, description)
}
