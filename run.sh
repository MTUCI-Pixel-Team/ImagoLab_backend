# test - запуск тестов
# migrate - запуск миграций
# migrate -rollback - откат миграций
# dev - сборка и запуск сервера в режиме разработки
# По умолчанию сборка и запуск сервера
run_tests() {
    echo "Running tests..."
    go test ./... -v
    if [ $? -ne 0 ]; then
        echo "Tests failed."
        exit 1
    fi
}

load_dependencies() {
    echo "Downloading dependencies..."
    go mod download
    if [ $? -ne 0 ]; then
        echo "Failed to download dependencies."
        exit 1
    fi
}

pre_build() {
    ./run.sh test
    if [ $? -ne 0 ]; then
        echo "Tests failed. Aborting build."
        exit 1
    fi

    ./run.sh migrate
    if [ $? -ne 0 ]; then
        echo "Migration failed. Aborting build."
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

build_and_run_dev_server() {
    echo "Building server..."
    go build -o server main.go
    if [ $? -ne 0 ]; then
        echo "Build failed. Aborting server start."
        exit 1
    fi
    
    echo "Starting dev server..."
    sudo ./server
}

build_and_run_server() {
    pre_build
    load_dependencies

    echo "Building server..."
    go build -trimpath -ldflags "-s -w -X 'main.Production=true'" -o server main.go
    if [ $? -ne 0 ]; then
        echo "Build failed. Aborting server start."
        exit 1
    fi

    echo "Starting server..."
}

if [ "$1" = "test" ]; then
    run_tests
elif [ "$1" = "migrate" ]; then
    if [ "$2" = "-rollback" ]; then
        run_migrate_rollback
    else
        if [ "$2" = "" ]; then
            run_migrate
        else
            echo "Invalid argument."
            exit 1
        fi
    fi
else
    if [ "$1" == "dev" ]; then
        build_and_run_dev_server
    elif [ "$1" != "" ]; then
        echo "Invalid argument."
        exit 1
    else 
        build_and_run_server
    fi
    
fi