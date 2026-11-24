// Package handlers conecta las rutas HTTP (Fiber + HTMX) con los servicios y modelos.
package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/valMarin29/gofiber-htmx-todolist/internal/models"
	"github.com/valMarin29/gofiber-htmx-todolist/internal/services"
)

const claveUsuario = "user_id"

// App agrupa las dependencias de los handlers.
type App struct {
	Users    *models.UserStore
	Todos    *models.TodoStore
	Sessions *session.Store
}

// Routes registra todas las rutas de la aplicación.
func (a *App) Routes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/todos") })
	app.Get("/register", a.registerPage)
	app.Post("/register", a.register)
	app.Get("/login", a.loginPage)
	app.Post("/login", a.login)
	app.Post("/logout", a.logout)

	todos := app.Group("/todos", a.requireLogin)
	todos.Get("/", a.listTodos)
	todos.Post("/", a.createTodo)
	todos.Put("/:id", a.updateTodo)
	todos.Patch("/:id/status", a.setStatus)
	todos.Delete("/:id", a.deleteTodo)
}

// ---- autenticación ----

func (a *App) registerPage(c *fiber.Ctx) error {
	return c.Render("register", fiber.Map{}, "layouts/main")
}

func (a *App) register(c *fiber.Ctx) error {
	err := services.Register(a.Users, c.FormValue("username"), c.FormValue("email"), c.FormValue("password"))
	if err != nil {
		msg := "No se pudo crear la cuenta."
		if errors.Is(err, models.ErrCorreoDuplicado) || errors.Is(err, services.ErrDatosObligatorios) {
			msg = err.Error()
		}
		return c.Status(fiber.StatusBadRequest).Render("register", fiber.Map{"Error": msg}, "layouts/main")
	}
	return c.Render("login", fiber.Map{"Info": "Cuenta creada. Ya puedes iniciar sesión."}, "layouts/main")
}

func (a *App) loginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{}, "layouts/main")
}

func (a *App) login(c *fiber.Ctx) error {
	u, err := services.Login(a.Users, c.FormValue("email"), c.FormValue("password"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Render("login", fiber.Map{"Error": services.ErrLoginInvalido.Error()}, "layouts/main")
	}
	sess, err := a.Sessions.Get(c)
	if err != nil {
		return err
	}
	if err := sess.Regenerate(); err != nil {
		return err
	}
	sess.Set(claveUsuario, u.ID)
	sess.Set("username", u.Username)
	if err := sess.Save(); err != nil {
		return err
	}
	return c.Redirect("/todos")
}

func (a *App) logout(c *fiber.Ctx) error {
	sess, err := a.Sessions.Get(c)
	if err == nil {
		_ = sess.Destroy()
	}
	return c.Redirect("/login")
}

// requireLogin deja pasar solo a usuarios con sesión (RU010).
func (a *App) requireLogin(c *fiber.Ctx) error {
	sess, err := a.Sessions.Get(c)
	if err != nil {
		return err
	}
	id, ok := sess.Get(claveUsuario).(int64)
	if !ok || id == 0 {
		if c.Get("HX-Request") == "true" {
			c.Set("HX-Redirect", "/login")
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Redirect("/login")
	}
	c.Locals(claveUsuario, id)
	c.Locals("username", sess.Get("username"))
	return c.Next()
}

// El id del usuario siempre sale de la sesión, nunca del formulario (RU012).
func usuarioActual(c *fiber.Ctx) int64 {
	id, _ := c.Locals(claveUsuario).(int64)
	return id
}

func idTarea(c *fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}

// ---- tareas ----

func (a *App) listTodos(c *fiber.Ctx) error {
	todos, err := a.Todos.ListByUser(usuarioActual(c))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No se pudieron cargar las tareas")
	}
	return c.Render("todos", fiber.Map{"Todos": todos, "Username": c.Locals("username")}, "layouts/main")
}

func (a *App) createTodo(c *fiber.Ctx) error {
	t, err := services.CreateTodo(a.Todos, usuarioActual(c), c.FormValue("title"), c.FormValue("description"))
	if errors.Is(err, services.ErrTituloObligatorio) {
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No se pudo crear la tarea")
	}
	return c.Render("partials/todo", t)
}

func (a *App) updateTodo(c *fiber.Ctx) error {
	id, err := idTarea(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	userID := usuarioActual(c)
	err = services.UpdateTodo(a.Todos, userID, id, c.FormValue("title"), c.FormValue("description"))
	switch {
	case errors.Is(err, services.ErrTituloObligatorio):
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	case errors.Is(err, models.ErrTareaNoEncontrada):
		return fiber.ErrNotFound
	case err != nil:
		return fiber.NewError(fiber.StatusInternalServerError, "No se pudo actualizar la tarea")
	}
	t, err := a.Todos.GetByUser(userID, id)
	if err != nil {
		return fiber.ErrNotFound
	}
	return c.Render("partials/todo", t)
}

func (a *App) setStatus(c *fiber.Ctx) error {
	id, err := idTarea(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	userID := usuarioActual(c)
	completed := c.FormValue("completed") == "true"
	err = a.Todos.SetCompleted(userID, id, completed)
	if errors.Is(err, models.ErrTareaNoEncontrada) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No se pudo cambiar el estado")
	}
	t, err := a.Todos.GetByUser(userID, id)
	if err != nil {
		return fiber.ErrNotFound
	}
	return c.Render("partials/todo", t)
}

func (a *App) deleteTodo(c *fiber.Ctx) error {
	id, err := idTarea(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	err = a.Todos.Delete(usuarioActual(c), id)
	if errors.Is(err, models.ErrTareaNoEncontrada) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "No se pudo eliminar la tarea")
	}
	// HTMX reemplaza el elemento por una respuesta vacía: la tarea desaparece sin recargar.
	return c.SendString("")
}
