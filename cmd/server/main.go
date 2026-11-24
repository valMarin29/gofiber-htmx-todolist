// Punto de entrada del Gestor de Tareas (gofiber-htmx-todolist).
package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/db"
	"github.com/valMarin29/gofiber-htmx-todolist/internal/handlers"
	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
)

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}

func main() {
	dbPath := env("DB_PATH", "todos.db")
	port := env("PORT", "8080")
	views := env("VIEWS_DIR", "./views")

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("no se pudo abrir la base de datos: %v", err)
	}
	defer conn.Close()

	app := fiber.New(fiber.Config{Views: html.New(views, ".html")})
	app.Use(logger.New())

	h := &handlers.App{
		Users: models.NewUserStore(conn),
		Todos: models.NewTodoStore(conn),
		Sessions: session.New(session.Config{
			Expiration:     2 * time.Hour,
			CookieHTTPOnly: true,
			CookieSameSite: "Lax",
		}),
	}
	h.Routes(app)

	log.Printf("Gestor de tareas escuchando en http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
