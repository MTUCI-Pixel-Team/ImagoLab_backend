package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func copyToGoFile(filename string) {
	// Добавляем расширение .txt, если его нет
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	// Открываем исходный файл
	srcFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Файл '%s' не найден: %v\n", filename, err)
		return
	}
	defer srcFile.Close()

	// Создаем имя файла с расширением .go
	goFilename := "migrations/actualVersion.go"

	// Создаем файл назначения
	dstFile, err := os.Create(goFilename)
	if err != nil {
		fmt.Printf("Ошибка создания файла '%s': %v\n", goFilename, err)
		return
	}
	defer dstFile.Close()

	// Копируем содержимое
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Printf("Ошибка копирования содержимого: %v\n", err)
		return
	}

	fmt.Printf("Файл успешно скопирован как '%s'.\n", goFilename)
}
