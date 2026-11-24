// Package auth agrupa el procesamiento de contraseñas (RU011).
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrCredencialesInvalidas indica que la contraseña no coincide con el hash.
var ErrCredencialesInvalidas = errors.New("credenciales inválidas")

// HashPassword genera un hash bcrypt de la contraseña. Cada llamada usa una
// sal aleatoria, por eso dos hashes de la misma contraseña son distintos.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compara la contraseña con el hash guardado.
// Devuelve nil solo si coinciden. No modifica mayúsculas ni espacios.
// Un hash inválido o corrupto se rechaza devolviendo un error, sin pánico.
func CheckPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrCredencialesInvalidas
		}
		return err
	}
	return nil
}
