# Imagen del Gestor de tareas.
# Usa el ejecutable que ya compiló el pipeline (bin/api), igual que en el taller:
# se compila una sola vez y lo que se prueba es lo que se despliega.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY bin/api /app/api
COPY views /app/views

ENV PORT=8080 \
    VIEWS_DIR=/app/views \
    DB_PATH=/tmp/todos.db

EXPOSE 8080
ENTRYPOINT ["/app/api"]
