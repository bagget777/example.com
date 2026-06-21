// Command examplan запускает веб-сервер планировщика подготовки к экзамену.
package main

import (
	"log"
	"net/http"
	"os"

	"examplan/internal/db"
	"examplan/internal/handlers"
)

func main() {
	dbPath := getEnv("EXAMPLAN_DB", "examplan.db")

	// Render и большинство PaaS-платформ передают порт через переменную PORT
	// и ожидают, что сервис слушает именно его. EXAMPLAN_ADDR остаётся как
	// запасной вариант для локального запуска и ручной настройки.
	addr := getEnv("EXAMPLAN_ADDR", ":8080")
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("не удалось открыть базу данных: %v", err)
	}
	defer database.Close()

	app, err := handlers.New(database, "templates/*.html")
	if err != nil {
		log.Fatalf("не удалось загрузить шаблоны: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Index)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/calculate", app.Calculate)
	mux.HandleFunc("/api/history", app.History)
	mux.HandleFunc("/api/history/delete", app.DeleteHistory)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Сервер запущен: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
