
package main

import (

"bufio"

"fmt"

"os"

"strings"

)

type Config struct {

Port int

Blacklist map[string]bool

ListenAddr string

}

func LoadConfig(blacklistFile string) (*Config, error) {

config := &Config{

Port: 8080,

ListenAddr: "127.0.0.1:8080",

Blacklist: make(map[string]bool),

}

file, err := os.Open(blacklistFile)

if err != nil {

fmt.Printf("⚠️ Файл blacklist.txt не найден: %v\n", err)

fmt.Println("📝 Создаю файл с примерами доменов...")

return config, nil

}

defer file.Close()


scanner := bufio.NewScanner(file)

for scanner.Scan() {

domain := strings.TrimSpace(scanner.Text())

if domain != "" && !strings.HasPrefix(domain, "#") {

config.Blacklist[strings.ToLower(domain)] = true

}

}

  

if err := scanner.Err(); err != nil {

return nil, fmt.Errorf("ошибка при чтении файла: %v", err)

}

  

fmt.Printf("✅ Загружено %d доменов в черный список\n", len(config.Blacklist))

return config, nil

}

func (c *Config) IsBlocked(domain string) bool {

return c.Blacklist[strings.ToLower(domain)]

}

