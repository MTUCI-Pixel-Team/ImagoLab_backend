# test - запуск тестов
# migrate - запуск миграций
# migrate -rollback - откат миграций
# migrate -rollback -version - замена файла models.go на указанную версию
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
    version=$1

    if [ -z "$version" ]; then
        echo "Please specify a version number."
        exit 1
    fi

    echo "Rolling back migrations to version $version..."
    go run db/cmd/runMigrations.go -rollback -version "$version"
    if [ $? -ne 0 ]; then
        echo "Migration rollback failed."
        exit 1
    fi
    run_migrate
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
    if [ "$2" = "-rollback" ] && [ "$3" = "-version" ] && [ -n "$4" ]; then
        run_migrate_rollback "$4"
    else
        if [ "$2" = "" ]; then
            run_migrate
        else
            echo "Invalid argument."
            exit 1
        fi
    fi
else
    build_and_run_server
fi