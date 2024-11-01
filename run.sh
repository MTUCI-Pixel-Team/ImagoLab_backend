# test - запуск тестов
# migrate - запуск миграций
# migrate -rollback - откат миграций
# dev - сборка и запуск сервера в режиме разработки
# По умолчанию сборка и запуск сервера
load_env() {
    local env_file_path=${1:-.env} 
    echo "Loading environment from $env_file_path..."

    if [ -f "$env_file_path" ]; then
        
        export $(cat "$env_file_path" | xargs)
        if [ -z "$RUNWARE_API_KEY" ]; then
            echo "Error loading environment: RUNWARE_API_KEY not set."
            exit 1
        else
            echo "Environment loaded successfully."
        fi
    else
        echo "$env_file_path file not found. Please make sure the path is correct."
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

run_tests() {
    echo "Running tests..."
    sudo go test ./... -v
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

pre_build() {
    load_env "$env_path"
    load_dependencies
    run_tests
    run_migrate
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

build_server() {
    echo "Pre-building..."
    export PRODUCTION=true
    if [ "$1" != "" ]; then
        env_path=$1
    else
        env_path=".env"
    fi

    pre_build "$env_path"
    
    echo "Building server..."
    go build -trimpath -ldflags "-s -w" -o server main.go
    if [ $? -ne 0 ]; then
        echo "Build failed. Aborting server start."
        exit 1
    fi
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
    if [ "$1" = "dev" ]; then
        build_and_run_dev_server
    elif [ "$1" == "env" ]; then
        if [ "$2" != "" ]; then
            env_path=$2
        else
            env_path=".env"
        fi
        build_server $env_path
        echo "Starting server..."
        sudo ./server
    elif [ "$1" != "" ]; then
        echo "Invalid argument."
        exit 1
    else 
        build_server
        echo "Starting server..."
        sudo ./server
    fi
    
fi