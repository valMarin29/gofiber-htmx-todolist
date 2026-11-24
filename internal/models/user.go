package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// User representa una cuenta registrada.
type User struct {
	ID       int64
	Username string
	Email    string
	Password string // hash bcrypt, nunca texto plano
}

// Consultas SQL de usuarios. Siempre con parámetros (?), nunca concatenadas.
const (
	QueryInsertUser     = "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	QueryGetUserByEmail = "SELECT id, username, email, password FROM users WHERE email = ?"
)

// UserStore accede a la tabla users.
type UserStore struct {
	DB *sql.DB
}

// NewUserStore crea un store de usuarios.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{DB: db}
}

// CreateUser inserta un usuario. passwordHash debe venir ya cifrado con bcrypt.
// Si el correo ya existe devuelve ErrCorreoDuplicado.
func (s *UserStore) CreateUser(username, email, passwordHash string) error {
	_, err := s.DB.Exec(QueryInsertUser, username, email, passwordHash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCorreoDuplicado
		}
		return fmt.Errorf("crear usuario: %w", err)
	}
	return nil
}

// GetUserByEmail busca un usuario por su correo.
// Si no existe devuelve ErrUsuarioNoEncontrado y un User vacío.
func (s *UserStore) GetUserByEmail(email string) (User, error) {
	var u User
	err := s.DB.QueryRow(QueryGetUserByEmail, email).Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUsuarioNoEncontrado
		}
		return User{}, fmt.Errorf("consultar usuario: %w", err)
	}
	return u, nil
}
