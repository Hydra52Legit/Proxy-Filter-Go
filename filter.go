package main

import (
	"fmt"
	"net/http"
	"strings"
)

// Filter проверяет запрос и решает, должен ли он быть заблокирован
type Filter struct {
	config *Config
}

// NewFilter создает новый фильтр
func NewFilter(config *Config) *Filter {
	return &Filter{config: config}
}

// CheckRequest проверяет, должен ли запрос быть заблокирован
func (f *Filter) CheckRequest(req *http.Request) (blocked bool, reason string) {
	// Получаем хост из заголовка Host или URL
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// Убираем порт если есть (например: example.com:443 -> example.com)
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Проверяем основной домен
	if f.config.IsBlocked(host) {
		return true, fmt.Sprintf("домен '%s' заблокирован", host)
	}

	// Проверяем поддомены (например: ads.google.com, если google.com в черном списке)
	parts := strings.Split(host, ".")
	for i := 1; i < len(parts); i++ {
		domain := strings.Join(parts[i:], ".")
		if f.config.IsBlocked(domain) {
			return true, fmt.Sprintf("поддомен '%s' (родитель: '%s') заблокирован", host, domain)
		}
	}

	return false, ""
}

// BlockedResponse возвращает HTTP ответ для заблокированного запроса
func BlockedResponse(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusForbidden) // 403 Forbidden
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🚫 Доступ заблокирован</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.3);
            text-align: center;
            max-width: 500px;
        }
        h1 {
            color: #e74c3c;
            font-size: 32px;
            margin: 0 0 20px 0;
        }
        p {
            color: #555;
            font-size: 16px;
            line-height: 1.6;
            margin: 10px 0;
        }
        .reason {
            background: #f8f9fa;
            border-left: 4px solid #e74c3c;
            padding: 15px;
            margin: 20px 0;
            text-align: left;
            border-radius: 4px;
            font-family: monospace;
            color: #333;
        }
        .footer {
            font-size: 12px;
            color: #999;
            margin-top: 30px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚫 Доступ заблокирован</h1>
        <p>Этот сайт содержится в черном списке прокси-сервера.</p>
        <div class="reason">
            <strong>Причина:</strong><br>
            %s
        </div>
        <p class="footer">Powered by Go Proxy Filter</p>
    </div>
</body>
</html>
`, reason)

	fmt.Fprint(w, html)
}
