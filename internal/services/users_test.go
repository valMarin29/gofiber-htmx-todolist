package services

// Plan de Pruebas 01 — Registro de usuarios y validación de credenciales.
// Casos CP-USR-01 a CP-USR-03: registro con un doble del repositorio de usuarios.

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

// repoUsuariosFalso es un doble de UserRepository que registra las llamadas.
type repoUsuariosFalso struct {
	errCrear     error
	inserciones  []insercionUsuario
	otrasLlamada int // cualquier llamada distinta de CreateUser
}

type insercionUsuario struct {
	username, email, hash string
}

func (r *repoUsuariosFalso) CreateUser(username, email, hash string) error {
	r.inserciones = append(r.inserciones, insercionUsuario{username, email, hash})
	return r.errCrear
}

func (r *repoUsuariosFalso) GetUserByEmail(string) (models.User, error) {
	r.otrasLlamada++
	return models.User{}, models.ErrUsuarioNoEncontrado
}

const (
	nombrePrueba = "usuario_prueba"
	correoPrueba = "usuario.a@example.com"
	clavePrueba  = "ClavePrueba!2026"
)

// CP-USR-01 — Registro exitoso de una cuenta (RU001, RU011).
func TestCP_USR_01_RegistroExitoso(t *testing.T) {
	repo := &repoUsuariosFalso{}

	err := Register(repo, nombrePrueba, correoPrueba, clavePrueba)

	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}
	if len(repo.inserciones) != 1 {
		t.Fatalf("se esperaba una única inserción, hubo %d", len(repo.inserciones))
	}
	ins := repo.inserciones[0]
	if ins.username != nombrePrueba || ins.email != correoPrueba {
		t.Fatalf("datos enviados a persistencia incorrectos: %+v", ins)
	}
	if ins.hash == clavePrueba || strings.Contains(ins.hash, clavePrueba) {
		t.Fatal("la contraseña se envió a persistencia en texto plano")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ins.hash), []byte(clavePrueba)); err != nil {
		t.Fatalf("lo enviado a persistencia no es un hash bcrypt válido de la contraseña: %v", err)
	}
}

// CP-USR-02 — Rechazo de un correo ya registrado (RU002).
func TestCP_USR_02_CorreoDuplicado(t *testing.T) {
	// El duplicado se detecta por el error de unicidad que devuelve la inserción.
	repo := &repoUsuariosFalso{errCrear: models.ErrCorreoDuplicado}

	err := Register(repo, "otro_usuario", correoPrueba, clavePrueba)

	if !errors.Is(err, models.ErrCorreoDuplicado) {
		t.Fatalf("se esperaba ErrCorreoDuplicado, se obtuvo: %v", err)
	}
	if len(repo.inserciones) != 1 {
		t.Fatalf("se esperaba un solo intento de inserción, hubo %d", len(repo.inserciones))
	}
	if repo.otrasLlamada != 0 {
		t.Fatal("se hicieron llamadas adicionales sobre la cuenta existente")
	}
}

// CP-USR-03 — Manejo de un error técnico durante el registro (RU001).
func TestCP_USR_03_ErrorTecnicoEnRegistro(t *testing.T) {
	errTecnico := errors.New("disco lleno")
	repo := &repoUsuariosFalso{errCrear: errTecnico}

	err := Register(repo, nombrePrueba, correoPrueba, clavePrueba)

	if err == nil {
		t.Fatal("se informó éxito aunque la escritura falló")
	}
	if !errors.Is(err, errTecnico) {
		t.Fatalf("se esperaba el error técnico original, se obtuvo: %v", err)
	}
	if errors.Is(err, models.ErrCorreoDuplicado) {
		t.Fatal("el fallo técnico se clasificó como correo duplicado")
	}
}
