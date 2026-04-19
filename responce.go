package main

import (
    "html/template"
    "net/http"
)

var blockedTemplate = template.Must(template.New("blocked").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>🚫 Access Blocked</title>
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
        <h1>🚫 Access Blocked</h1>
        <p>This website is in the proxy server's blacklist.</p>
        <div class="reason">
            <strong>Reason:</strong><br>
            {{.Reason}}
        </div>
        <p class="footer">Powered by Go Proxy Filter</p>
    </div>
</body>
</html>
`))

// BlockedResponse отправляет красивую страницу блокировки
func BlockedResponse(w http.ResponseWriter, reason string) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusForbidden)
    
    data := struct {
        Reason string
    }{
        Reason: reason,
    }
    
    blockedTemplate.Execute(w, data)
}