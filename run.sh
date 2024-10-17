# test - запуск тестов
# migrate - запуск миграций
# migrate -rollback - откат миграций
# По умолчанию сборка и запуск сервера
run_tests() {
    echo "Running tests..."
    go test ./... -v
    if [ $? -ne 0 ]; then
        echo "Tests failed."
        exit 1
    fi
}

run_migrate() {
    echo "Running migrations..."
    go run db/cmd/runMigrations.go
    if [ $? -ne 0 ]; then
        echo "Migration failed."
        exit 1
    fi
}

run_migrate_rollback() {
    echo "Rolling back migrations..."
    go run db/cmd/runMigrations.go -rollback
    if [ $? -ne 0 ]; then
        echo "Migration rollback failed."
        exit 1
    fi
}

build_and_run_server() {
    echo "Building server..."
    go build -o server main.go
    if [ $? -ne 0 ]; then
        echo "Build failed. Aborting server start."
        exit 1
    fi
    
    echo "Starting server..."
    sudo ./server
}

if [ "$1" = "test" ]; then
    run_tests
elif [ "$1" = "migrate" ]; then
    if [ "$2" = "-rollback" ]; then
        run_migrate_rollback
    else
        run_migrate
    fi
else
    build_and_run_server
fi