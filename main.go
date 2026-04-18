package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🌐 HTTP Прокси-Сервер с Фильтрацией")
	fmt.Println(strings.Repeat("=", 60))

	// Загружаем конфигурацию и черный список
	config, err := LoadConfig("blacklist.txt")
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Создаем фильтр
	filter := NewFilter(config)

	// Создаем и запускаем сервер
	server := NewServer(config, filter)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Ошибка при запуске сервера:", err)
		}
	}()

	// Ждем сигнала прерывания (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\n🛑 Останавливаем сервер...")
	server.Stop()
	fmt.Println("✅ Сервер остановлен")
}
