// Package services contiene las reglas de negocio: registro de usuarios,
// inicio de sesión y validación de tareas. No conoce HTTP ni SQL.
package services

import (
	"errors"
	"strings"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/auth"
	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

var (
	// ErrDatosObligatorios: falta nombre, correo o contraseña.
	ErrDatosObligatorios = errors.New("nombre, correo y contraseña son obligatorios")
	// ErrLoginInvalido: correo o contraseña incorrectos (mismo mensaje para ambos casos).
	ErrLoginInvalido = errors.New("correo o contraseña incorrectos")
)

// UserRepository es lo que necesita el servicio de usuarios para guardar y buscar cuentas.
// models.UserStore lo cumple; en las pruebas se usa un doble.
type UserRepository interface {
	CreateUser(username, email, passwordHash string) error
	GetUserByEmail(email string) (models.User, error)
}

// Register registra una cuenta (RU001, RU002, RU011).
// La contraseña se guarda solo como hash bcrypt.
func Register(repo UserRepository, username, email, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(email) == "" || password == "" {
		return ErrDatosObligatorios
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	return repo.CreateUser(username, email, hash)
}

// Login valida correo y contraseña (RU003, RU004) y devuelve el usuario.
func Login(repo UserRepository, email, password string) (models.User, error) {
	u, err := repo.GetUserByEmail(email)
	if errors.Is(err, models.ErrUsuarioNoEncontrado) {
		return models.User{}, ErrLoginInvalido
	}
	if err != nil {
		return models.User{}, err
	}
	if err := auth.CheckPassword(u.Password, password); err != nil {
		return models.User{}, ErrLoginInvalido
	}
	return u, nil
}
