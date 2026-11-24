package auth

// Plan de Pruebas 01 — Registro de usuarios y validación de credenciales.
// Casos CP-USR-07 a CP-USR-12: procesamiento de contraseñas con bcrypt real.

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

const (
	clavePrueba = "ClavePrueba!2026"
	otraClave   = "OtraClave!2026"
)

// hashDePrueba genera un hash válido de ClavePrueba!2026 para cada prueba.
func hashDePrueba(t *testing.T) string {
	t.Helper()
	h, err := HashPassword(clavePrueba)
	if err != nil {
		t.Fatalf("no se pudo preparar el hash de prueba: %v", err)
	}
	return h
}

// CP-USR-07 — Validación de una contraseña correcta (RU003, RU011).
func TestCP_USR_07_ContrasenaCorrecta(t *testing.T) {
	hash := hashDePrueba(t)

	if err := CheckPassword(hash, clavePrueba); err != nil {
		t.Fatalf("se esperaba que la contraseña coincidiera, se obtuvo error: %v", err)
	}
}

// CP-USR-08 — Rechazo de una contraseña incorrecta (RU003).
func TestCP_USR_08_ContrasenaIncorrecta(t *testing.T) {
	hash := hashDePrueba(t)

	err := CheckPassword(hash, otraClave)
	if err == nil {
		t.Fatal("se aceptó una contraseña incorrecta")
	}
	if !errors.Is(err, ErrCredencialesInvalidas) {
		t.Fatalf("se esperaba ErrCredencialesInvalidas, se obtuvo: %v", err)
	}
}

// CP-USR-09 — Manejo de un hash inválido o corrupto (RU003, RU011).
func TestCP_USR_09_HashInvalido(t *testing.T) {
	casos := map[string]string{
		"texto que no es hash": "esto-no-es-un-hash",
		"hash vacío":           "",
		"hash truncado":        "$2a$10$abc",
	}
	for nombre, hashInvalido := range casos {
		t.Run(nombre, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("el componente provocó un pánico: %v", r)
				}
			}()
			if err := CheckPassword(hashInvalido, clavePrueba); err == nil {
				t.Fatal("un hash inválido fue aceptado como contraseña válida")
			}
		})
	}
}

// CP-USR-10 — Generación de un hash verificable (RU011).
func TestCP_USR_10_HashVerificable(t *testing.T) {
	hash, err := HashPassword(clavePrueba)
	if err != nil {
		t.Fatalf("error al generar el hash: %v", err)
	}
	if hash == clavePrueba || strings.Contains(hash, clavePrueba) {
		t.Fatal("el hash contiene la contraseña en texto plano")
	}
	if _, err := bcrypt.Cost([]byte(hash)); err != nil {
		t.Fatalf("el resultado no es un hash bcrypt válido: %v", err)
	}
	if err := CheckPassword(hash, clavePrueba); err != nil {
		t.Fatalf("el hash no valida la contraseña original: %v", err)
	}
	if err := CheckPassword(hash, otraClave); err == nil {
		t.Fatal("el hash validó una contraseña distinta")
	}
}

// CP-USR-11 — Generación de hashes distintos para la misma contraseña (RU011).
func TestCP_USR_11_HashesDistintos(t *testing.T) {
	h1, err1 := HashPassword(clavePrueba)
	h2, err2 := HashPassword(clavePrueba)
	if err1 != nil || err2 != nil {
		t.Fatalf("error al generar hashes: %v / %v", err1, err2)
	}
	if h1 == h2 {
		t.Fatal("dos hashes de la misma contraseña son iguales: la sal no es aleatoria")
	}
	for i, h := range []string{h1, h2} {
		if err := CheckPassword(h, clavePrueba); err != nil {
			t.Fatalf("el hash %d no valida la contraseña original: %v", i+1, err)
		}
	}
}

// CP-USR-12 — Sensibilidad de la comparación a cambios en la contraseña (RU003, RU011).
func TestCP_USR_12_SensibleAMayusculasYEspacios(t *testing.T) {
	hash := hashDePrueba(t)

	variantes := map[string]string{
		"minúscula inicial": "clavePrueba!2026",
		"espacio final":     "ClavePrueba!2026 ",
	}
	for nombre, variante := range variantes {
		t.Run(nombre, func(t *testing.T) {
			if err := CheckPassword(hash, variante); err == nil {
				t.Fatalf("se aceptó la variante %q", variante)
			}
		})
	}
}
