// Package handlers содержит HTTP-обработчики приложения.
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"examplan/internal/calc"
	"examplan/internal/db"
)

// App хранит зависимости, нужные обработчикам (БД, шаблоны).
type App struct {
	DB        *db.DB
	Templates *template.Template
}

// New создаёт новый App, парсит HTML-шаблоны.
func New(database *db.DB, templatesGlob string) (*App, error) {
	tmpl, err := template.ParseGlob(templatesGlob)
	if err != nil {
		return nil, err
	}
	return &App{DB: database, Templates: tmpl}, nil
}

// Index отдаёт главную страницу.
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := a.Templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Printf("ошибка рендера index.html: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// calcRequest — тело запроса POST /api/calculate
type calcRequest struct {
	Subject      string  `json:"subject"`
	DaysLeft     int     `json:"daysLeft"`
	FreeHours    float64 `json:"freeHours"`
	Difficulty   int     `json:"difficulty"`
	CurrentLevel float64 `json:"currentLevel"`
	Save         bool    `json:"save"`
}

// calcResponse — тело ответа POST /api/calculate
type calcResponse struct {
	Result calc.Result `json:"result"`
	PlanID int64       `json:"planId,omitempty"`
}

// Calculate принимает входные параметры пользователя, считает план подготовки
// и (опционально) сохраняет его в историю.
func (a *App) Calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req calcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "не удалось разобрать запрос: "+err.Error())
		return
	}

	if req.Subject == "" {
		req.Subject = "Предмет"
	}
	if req.Difficulty < 1 || req.Difficulty > 5 {
		writeJSONError(w, http.StatusBadRequest, "сложность должна быть от 1 до 5")
		return
	}
	if req.DaysLeft < 1 {
		writeJSONError(w, http.StatusBadRequest, "количество дней должно быть не меньше 1")
		return
	}
	if req.FreeHours <= 0 || req.FreeHours > 16 {
		writeJSONError(w, http.StatusBadRequest, "часы в день должны быть от 0 до 16")
		return
	}
	if req.CurrentLevel < 0 || req.CurrentLevel > 100 {
		writeJSONError(w, http.StatusBadRequest, "текущий уровень должен быть от 0 до 100")
		return
	}

	in := calc.Input{
		DaysLeft:     req.DaysLeft,
		FreeHours:    req.FreeHours,
		Difficulty:   calc.Difficulty(req.Difficulty),
		CurrentLevel: req.CurrentLevel,
	}
	result := calc.Compute(in)

	resp := calcResponse{Result: result}

	if req.Save {
		id, err := a.DB.SavePlan(req.Subject, in, result)
		if err != nil {
			log.Printf("ошибка сохранения плана: %v", err)
			// не блокируем ответ пользователю из-за ошибки сохранения
		} else {
			resp.PlanID = id
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// History отдаёт последние сохранённые расчёты.
func (a *App) History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	records, err := a.DB.ListPlans(30)
	if err != nil {
		log.Printf("ошибка чтения истории: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "не удалось загрузить историю")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

// DeleteHistory удаляет запись истории по ID (?id=).
func (a *App) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "некорректный id")
		return
	}
	if err := a.DB.DeletePlan(id); err != nil {
		log.Printf("ошибка удаления записи: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "не удалось удалить запись")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ошибка кодирования JSON-ответа: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
