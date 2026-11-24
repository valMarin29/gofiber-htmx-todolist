# Gestor de tareas — gofiber-htmx-todolist

Trabajo final de **Ingeniería de Software III** — Institución Universitaria EAM.
**Pipeline de ejecución de pruebas unitarias (Plan de Pruebas).**

Integrantes: Valentina Marín, Juan David Rodríguez y Diego Andrés Osorio.
Docente: Deibi Octavio Riascos Botero.

---

## 1. De qué se trata

Es un gestor de tareas web con usuarios. Cada persona se registra, inicia sesión y
maneja **solo sus propias tareas**: crear, ver, editar, marcar como completada o
pendiente y eliminar.

En este repositorio están:

- **La aplicación**: Go + [Fiber](https://gofiber.io) + [HTMX](https://htmx.org) + SQLite.
- **Las 60 pruebas unitarias** de nuestro Plan de Pruebas (casos `CP-USR`, `CP-CRE`,
  `CP-CON`, `CP-ACT` y `CP-ELI`).
- **El pipeline CI/CD** en GitHub Actions, con los mismos seis pasos del taller anterior.

---

## 2. Cómo funciona el pipeline

Archivo: [`.github/workflows/pipeline.yml`](.github/workflows/pipeline.yml)

```
push / pull request
        │
        ▼
 4a. Lint (gofmt + go vet) ──► 4b. Build (go build) ──► 5. Pruebas unitarias ──► 6a. Imagen Docker ──► 6b. Despliegue
                                                          (60 casos + cobertura ≥ 70 %)
```

| Paso del taller | En GitHub Actions | Qué hace |
|---|---|---|
| 1. Trigger | `on:` | Corre en cada push a `main` o `develop`, en cada pull request y con el botón **Run workflow**. |
| 2. Infraestructura | `runs-on: ubuntu-latest` | Cada job corre en una máquina Ubuntu limpia de GitHub. |
| 3. Checkout | `actions/checkout` | Descarga el código del commit que inició el pipeline. |
| 4. Compilación | jobs `lint` y `build` | Revisa el formato (`gofmt`), busca errores (`go vet`) y compila el ejecutable. |
| 5. Pruebas unitarias | job `unit-tests` | Corre las pruebas del Plan de Pruebas y exige cobertura mínima de 70 %. |
| 6. Despliegue | jobs `docker-image` y `deploy` | Crea la imagen Docker, arranca el contenedor y verifica que responde. En `main` publica la imagen en GitHub Container Registry. |

**Si una prueba falla, el pipeline se detiene en el paso 5**: el job queda en rojo ❌ y
los jobs de imagen y despliegue **no se ejecutan** (quedan como *skipped*).

### Parámetros del pipeline

Están al inicio de `pipeline.yml`, en la sección `env:`:

| Parámetro | Valor | Para qué sirve |
|---|---|---|
| `COVERAGE_MIN` | `70` | Cobertura mínima exigida. Si es menor, el pipeline falla. |
| `GOTESTSUM_VERSION` | `v1.13.0` | Versión de gotestsum (genera el reporte JUnit). |
| `COVERAGE_PKGS` | `./internal/auth/...,./internal/models/...,./internal/services/...` | Paquetes que se miden en la cobertura. |

**No hace falta configurar secretos ni llaves.** El despliegue usa el token automático
de GitHub (`GITHUB_TOKEN`).

---

## 3. Guía para ejecutar el pipeline (para el docente)

### Opción A — En GitHub, sin instalar nada (recomendada)

1. Entrar al repositorio y hacer clic en **Fork** (arriba a la derecha) → **Create fork**.
2. En el fork, abrir la pestaña **Actions**. GitHub pide activar los workflows en los forks:
   clic en **"I understand my workflows, go ahead and enable them"**.
3. En el menú izquierdo elegir **Pipeline CI/CD** → botón **Run workflow** → rama `main` →
   **Run workflow**.
4. Abrir la ejecución que aparece. Se ven los jobs en orden. Al terminar, todos quedan en verde ✅.
5. Al final de la página está el **resumen**: pruebas aprobadas, casos del plan (60 / 60)
   y cobertura.

#### Hacer fallar una prueba a propósito

1. En el fork, abrir un archivo de pruebas, por ejemplo
   `internal/models/todo_delete_test.go`.
2. Clic en el lápiz ✏️ (**Edit this file**).
3. Cambiar un valor esperado. Por ejemplo, en `TestCP_ELI_01_EliminarTareaPendiente` cambiar
   `esperarEliminacion(mock, 11, usuarioA, 1)` por `esperarEliminacion(mock, 99, usuarioA, 1)`.
4. **Commit changes** directamente en `main`.
5. El pipeline arranca solo. El job **5. Pruebas unitarias** queda en rojo ❌, el resumen
   muestra qué caso falló y los jobs **6a** y **6b** no se ejecutan.

### Opción B — En el computador (para correr las pruebas localmente)

**Requisitos:**

- [Git](https://git-scm.com/downloads)
- [Go 1.27 o superior](https://go.dev/dl/) (verificar con `go version`)
- Opcional: [Docker](https://docs.docker.com/get-docker/) para probar la imagen

**Pasos:**

```bash
git clone https://github.com/valMarin29/gofiber-htmx-todolist.git
cd gofiber-htmx-todolist

# Descargar dependencias
go mod download

# Correr las pruebas unitarias (muestra cada caso)
go test ./... -v

# Correr las pruebas con cobertura (igual que el pipeline)
go test ./... -coverprofile=coverage.out -coverpkg=./internal/auth/...,./internal/models/...,./internal/services/...
go tool cover -func=coverage.out | tail -1

# Correr solo los casos de un módulo, por ejemplo eliminación
go test ./... -v -run 'TestCP_ELI'
```

**Levantar la aplicación:**

```bash
go run ./cmd/server
```

Abrir <http://localhost:8080>, crear una cuenta y empezar a usarla.
La base de datos se guarda en el archivo `todos.db`.

Variables opcionales: `PORT` (puerto, por defecto `8080`), `DB_PATH` (archivo SQLite,
por defecto `todos.db`) y `VIEWS_DIR` (plantillas, por defecto `./views`).

**Con Docker:**

```bash
CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/server
docker build -t gestor-tareas .
docker run -p 8080:8080 gestor-tareas
```

---

## 4. Pruebas unitarias del Plan de Pruebas

Las pruebas usan el paquete estándar `testing` de Go. Ninguna usa la base de datos real:

- El **acceso a datos** se prueba con [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock),
  un simulador SQL que exige el **texto exacto** de cada consulta. Así se comprueba que los
  datos van como parámetros (`?`) y nunca concatenados en el SQL, y que cada consulta
  filtra por el dueño de la tarea (`created_by`).
- Las **contraseñas** se prueban con la implementación real de bcrypt.
- El **registro** se prueba con un doble del repositorio que guarda las llamadas recibidas.

Cada función de prueba se llama igual que su caso del plan (`TestCP_USR_01_...`), así es
fácil ubicarla.

| Documento del plan | Casos | Archivo de pruebas |
|---|---|---|
| 01 Registro y credenciales | CP-USR-01 a 03 | `internal/services/users_test.go` |
| | CP-USR-04 a 06 | `internal/models/user_test.go` |
| | CP-USR-07 a 12 | `internal/auth/password_test.go` |
| 02 Creación de tareas | CP-CRE-01 a 12 | `internal/services/todos_create_test.go` |
| 03 Consulta de tareas | CP-CON-01 a 12 | `internal/models/todo_list_test.go` |
| 04 Actualización y estado | CP-ACT-01 a 14 | `internal/services/todos_update_test.go` |
| 05 Eliminación de tareas | CP-ELI-01 a 10 | `internal/models/todo_delete_test.go` |
| **Total** | **60 casos** | |

El documento 06 (caja negra) se ejecuta de forma manual sobre la aplicación, como indica el plan.

Además hay 5 **pruebas complementarias** (`complementarias_test.go`) que no hacen parte de
los 60 casos. Cubren el inicio de sesión y la detección del correo duplicado en SQLite.

### Decisiones tomadas al contrastar el plan con el código

- **Título obligatorio**: lo valida la capa de servicios (`internal/services`). Un título
  vacío o solo con espacios se rechaza sin llamar a la base de datos.
- **Orden del listado**: el esquema no tiene fecha de creación, así que la más reciente es
  la de `id` mayor (`ORDER BY id DESC`).
- **Cambio de estado**: recibe un valor explícito (completada / pendiente); no alterna.
- **Tarea inexistente o ajena**: las dos devuelven el mismo resultado ("tarea no
  encontrada"), para no revelar datos de otra cuenta.
- **CP-ELI-08**: aplica, porque la eliminación revisa las filas afectadas. Si no se puede
  leer ese número, se informa que el resultado no se pudo confirmar.

---

## 5. Estructura del proyecto

```
gofiber-htmx-todolist/
├── .github/workflows/pipeline.yml   pipeline CI/CD
├── cmd/server/main.go               punto de entrada
├── internal/
│   ├── auth/                        hash y validación de contraseñas (bcrypt)
│   ├── db/                          conexión SQLite y creación de tablas
│   ├── handlers/                    rutas HTTP (Fiber + HTMX)
│   ├── models/                      acceso a datos de usuarios y tareas
│   └── services/                    reglas de negocio (registro, login, validaciones)
├── views/                           plantillas HTML
├── Dockerfile
└── go.mod
```

## 6. Requisitos del sistema cubiertos

| Requisito | Descripción | Dónde se prueba |
|---|---|---|
| RU001–RU003, RU011 | Registro, correo único, inicio de sesión, contraseñas con bcrypt | CP-USR |
| RU005 | Crear tareas (título obligatorio, descripción opcional) | CP-CRE |
| RU006 | Consultar tareas | CP-CON |
| RU007, RU008 | Editar tareas y cambiar su estado | CP-ACT |
| RU009 | Eliminar tareas | CP-ELI |
| RU012 | Cada usuario solo accede a sus tareas | CP-CRE, CP-CON, CP-ACT, CP-ELI |
