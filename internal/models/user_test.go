package models

// Plan de Pruebas 01 — Registro de usuarios y validación de credenciales.
// Casos CP-USR-04 a CP-USR-06: recuperación de cuentas por correo (simulador SQL).

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const sqlUsuarioPorCorreo = "SELECT id, username, email, password FROM users WHERE email = ?"

var columnasUsuario = []string{"id", "username", "email", "password"}

// CP-USR-04 — Recuperación de un usuario existente por correo (RU003).
func TestCP_USR_04_RecuperarUsuarioExistente(t *testing.T) {
	db, mock := nuevoSimulador(t)
	hash := "$2a$10$7EqJtq98hPqEX7fNZaFWoO5c5w3S5E5F5yF5yF5yF5yF5yF5yF5yO"
	mock.ExpectQuery(sqlUsuarioPorCorreo).
		WithArgs("usuario.a@example.com").
		WillReturnRows(sqlmock.NewRows(columnasUsuario).
			AddRow(int64(101), "usuario_prueba", "usuario.a@example.com", hash))

	u, err := NewUserStore(db).GetUserByEmail("usuario.a@example.com")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	esperado := User{ID: 101, Username: "usuario_prueba", Email: "usuario.a@example.com", Password: hash}
	if u != esperado {
		t.Fatalf("usuario recuperado = %+v, se esperaba %+v", u, esperado)
	}
	verificarInteracciones(t, mock)
}

// CP-USR-05 — Consulta de un correo inexistente (RU003).
func TestCP_USR_05_CorreoInexistente(t *testing.T) {
	db, mock := nuevoSimulador(t)
	mock.ExpectQuery(sqlUsuarioPorCorreo).
		WithArgs("inexistente@example.com").
		WillReturnRows(sqlmock.NewRows(columnasUsuario)) // cero filas

	u, err := NewUserStore(db).GetUserByEmail("inexistente@example.com")

	if !errors.Is(err, ErrUsuarioNoEncontrado) {
		t.Fatalf("se esperaba ErrUsuarioNoEncontrado, se obtuvo: %v", err)
	}
	if u != (User{}) {
		t.Fatalf("se devolvió una cuenta con datos aunque no existe: %+v", u)
	}
	verificarInteracciones(t, mock)
}

// CP-USR-06 — Manejo de un error técnico al recuperar una cuenta (RU003).
func TestCP_USR_06_ErrorTecnicoAlConsultar(t *testing.T) {
	db, mock := nuevoSimulador(t)
	errTecnico := errors.New("base de datos bloqueada")
	mock.ExpectQuery(sqlUsuarioPorCorreo).
		WithArgs("usuario.a@example.com").
		WillReturnError(errTecnico)

	u, err := NewUserStore(db).GetUserByEmail("usuario.a@example.com")

	if err == nil {
		t.Fatal("el fallo técnico se interpretó como consulta exitosa")
	}
	if !errors.Is(err, errTecnico) {
		t.Fatalf("se esperaba conservar el error técnico, se obtuvo: %v", err)
	}
	if errors.Is(err, ErrUsuarioNoEncontrado) || errors.Is(err, sql.ErrNoRows) {
		t.Fatal("el fallo técnico se confundió con un usuario inexistente")
	}
	if u != (User{}) {
		t.Fatalf("se devolvió una cuenta junto con el error: %+v", u)
	}
	verificarInteracciones(t, mock)
}
