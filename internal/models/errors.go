// Package models contiene el acceso a datos (SQLite) de usuarios y tareas.
// Cada store recibe un *sql.DB, así las pruebas pueden usar un simulador SQL
// (go-sqlmock) en lugar de la base de datos real.
package models

import "errors"

var (
	// ErrUsuarioNoEncontrado: no existe una cuenta con ese correo.
	ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")
	// ErrCorreoDuplicado: el correo ya está registrado (restricción UNIQUE).
	ErrCorreoDuplicado = errors.New("el correo ya está registrado")
	// ErrTareaNoEncontrada: no hay una tarea con ese id que pertenezca al usuario.
	// Se usa igual para tareas inexistentes y ajenas, para no revelar datos de otra cuenta.
	ErrTareaNoEncontrada = errors.New("tarea no encontrada")
	// ErrResultadoNoConfirmado: la escritura se ejecutó pero no se pudo confirmar
	// cuántas filas afectó.
	ErrResultadoNoConfirmado = errors.New("no se pudo confirmar el resultado de la operación")
)
