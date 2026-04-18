package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Server представляет наш прокси-сервер
type Server struct {
	config       *Config
	filter       *Filter
	requestCount int64 // счётчик запросов
	blockedCount int64 // счётчик заблокированных запросов
	httpServer   *http.Server
	mux          *sync.Mutex
}

// NewServer создает новый прокси-сервер
func NewServer(config *Config, filter *Filter) *Server {
	return &Server{
		config: config,
		filter: filter,
		mux:    &sync.Mutex{},
	}
}

// Start запускает сервер
func (s *Server) Start() error {
	// Создаем HTTP обработчик
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)
	mux.HandleFunc("/stats", s.handleStats)

	s.httpServer = &http.Server{
		Addr:           s.config.ListenAddr,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	fmt.Printf("\n🚀 Прокси-сервер запущен на http://%s\n", s.config.ListenAddr)
	fmt.Printf("📋 Статистика доступна на http://%s/stats\n\n", s.config.ListenAddr)

	return s.httpServer.ListenAndServe()
}

// handleRequest обработчик для всех HTTP запросов
// ✨ Здесь работает горутина для каждого запроса!
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Увеличиваем счётчик запросов (atomically, безопасно для многопоточности)
	atomic.AddInt64(&s.requestCount, 1)
	requestNum := atomic.LoadInt64(&s.requestCount)

	// Получаем хост и путь для логирования
	method := r.Method
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	path := r.RequestURI

	fmt.Printf("[%d] 📨 Новый запрос: %s %s%s\n", requestNum, method, host, path)

	// Проверяем, заблокирован ли домен
	blocked, reason := s.filter.CheckRequest(r)
	if blocked {
		atomic.AddInt64(&s.blockedCount, 1)
		fmt.Printf("[%d] 🚫 ЗАБЛОКИРОВАН: %s\n", requestNum, reason)
		BlockedResponse(w, reason)
		return
	}

	// Если это запрос к нашему прокси напрямую
	if strings.HasPrefix(r.RequestURI, "/stats") || strings.HasPrefix(r.RequestURI, "/") && r.URL.Host == "" {
		return
	}

	fmt.Printf("[%d] ✅ Пропускаю: %s\n", requestNum, host)

	// Проксируем запрос (используем встроенный httputil.ReverseProxy)
	s.proxyRequest(w, r, requestNum)
}

// proxyRequest проксирует запрос на реальный сервер
func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request, requestNum int64) {
	// Если нет схемы в URL, добавляем http
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}

	// Если нет хоста в URL, берем из заголовка Host
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}

	// Создаём обратный прокси
	director := func(req *http.Request) {
		// Удаляем hop-by-hop заголовки (они не должны передаваться)
		req.RequestURI = ""
		// Указываем оригинальный IP
		if clientIP := getClientIP(r); clientIP != "" {
			req.Header.Set("X-Forwarded-For", clientIP)
		}
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("User-Agent", "Go-Proxy/1.0")
	}

	proxy := &httputil.ReverseProxy{Director: director}

	// Перехватываем ошибки проксирования
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("[%d] ❌ Ошибка при проксировании: %v\n", requestNum, err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Error: %v", err)
	}

	// Логируем успешное проксирование
	fmt.Printf("[%d] 🔄 Проксирую на: %s\n", requestNum, r.URL.String())

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)

	fmt.Printf("[%d] ✔️  Запрос завершён\n", requestNum)
}

// handleStats показывает статистику прокси
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	totalRequests := atomic.LoadInt64(&s.requestCount)
	blockedRequests := atomic.LoadInt64(&s.blockedCount)

	blockedPercent := float64(0)
	if totalRequests > 0 {
		blockedPercent = (float64(blockedRequests) / float64(totalRequests)) * 100
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>📊 Статистика Прокси</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0% , #764ba2 100%);
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.3);
            padding: 30px;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-top: 0;
        }
        .stat {
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 15px 0;
            border-radius: 4px;
        }
        .stat-label {
            color: #666;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .stat-value {
            color: #333;
            font-size: 32px;
            font-weight: bold;
            margin-top: 8px;
        }
        .stat-percent {
            color: #e74c3c;
            font-size: 24px;
            font-weight: bold;
        }
        .footer {
            text-align: center;
            color: #999;
            font-size: 12px;
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
        }
        .blacklist-info {
            background: #e8f4f8;
            border-left: 4px solid #3498db;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 Статистика Прокси-Сервера</h1>
        
        <div class="stat">
            <div class="stat-label">Всего запросов</div>
            <div class="stat-value">%d</div>
        </div>
        
        <div class="stat">
            <div class="stat-label">Заблокировано</div>
            <div class="stat-value">%d <span class="stat-percent">(%.1f%%)</span></div>
        </div>
        
        <div class="stat">
            <div class="stat-label">Пропущено</div>
            <div class="stat-value">%d</div>
        </div>

        <div class="blacklist-info">
            <strong>📋 Размер черного списка:</strong> %d доменов
        </div>

        <div class="footer">
            <p>Прокси-сервер работает на http://%s</p>
            <p>Обновите страницу для получения свежих данных</p>
        </div>
    </div>
    <script>
        // Автоматическое обновление каждые 5 секунд
        setTimeout(() => location.reload(), 5000);
    </script>
</body>
</html>
`, totalRequests, blockedRequests, blockedPercent, totalRequests-blockedRequests, len(s.config.Blacklist), s.config.ListenAddr)

	fmt.Fprint(w, html)
}

// getClientIP извлекает IP клиента из запроса
func getClientIP(r *http.Request) string {
	// Сначала проверяем X-Forwarded-For (если запрос прошел через другой прокси)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	// Затем X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// И наконец, RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Stop останавливает сервер
func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}
