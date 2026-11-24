package services

// Pruebas complementarias: NO forman parte de los 60 casos del Plan de Pruebas.
// Cubren el inicio de sesión (RU004 se verifica en caja negra) y la validación
// de datos obligatorios del registro.

import (
	"errors"
	"testing"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/auth"
	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

// repoLogin devuelve siempre el mismo usuario o error.
type repoLogin struct {
	user models.User
	err  error
}

func (r repoLogin) CreateUser(string, string, string) error    { return nil }
func (r repoLogin) GetUserByEmail(string) (models.User, error) { return r.user, r.err }

func TestComplementaria_Login(t *testing.T) {
	hash, err := auth.HashPassword(clavePrueba)
	if err != nil {
		t.Fatal(err)
	}
	cuenta := models.User{ID: usuarioA, Username: nombrePrueba, Email: correoPrueba, Password: hash}

	t.Run("credenciales correctas", func(t *testing.T) {
		u, err := Login(repoLogin{user: cuenta}, correoPrueba, clavePrueba)
		if err != nil || u.ID != usuarioA {
			t.Fatalf("resultado inesperado: %+v, %v", u, err)
		}
	})
	t.Run("contraseña incorrecta", func(t *testing.T) {
		if _, err := Login(repoLogin{user: cuenta}, correoPrueba, "OtraClave!2026"); !errors.Is(err, ErrLoginInvalido) {
			t.Fatalf("se esperaba ErrLoginInvalido, se obtuvo: %v", err)
		}
	})
	t.Run("correo inexistente", func(t *testing.T) {
		if _, err := Login(repoLogin{err: models.ErrUsuarioNoEncontrado}, "inexistente@example.com", clavePrueba); !errors.Is(err, ErrLoginInvalido) {
			t.Fatalf("se esperaba ErrLoginInvalido, se obtuvo: %v", err)
		}
	})
	t.Run("error técnico", func(t *testing.T) {
		errTecnico := errors.New("fallo")
		if _, err := Login(repoLogin{err: errTecnico}, correoPrueba, clavePrueba); !errors.Is(err, errTecnico) {
			t.Fatalf("se esperaba el error técnico, se obtuvo: %v", err)
		}
	})
}

func TestComplementaria_RegistroSinDatosObligatorios(t *testing.T) {
	repo := &repoUsuariosFalso{}
	if err := Register(repo, "", correoPrueba, clavePrueba); !errors.Is(err, ErrDatosObligatorios) {
		t.Fatalf("se esperaba ErrDatosObligatorios, se obtuvo: %v", err)
	}
	if len(repo.inserciones) != 0 {
		t.Fatal("se intentó guardar un registro incompleto")
	}
}
