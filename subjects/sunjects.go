// Package subjects содержит каталог готовых предметов и диагностических тестов.
// Каждый предмет имеет базовую сложность и набор вопросов двух типов:
//   - вопросы на УРОВЕНЬ ЗНАНИЙ (currentLevel) — насколько студент уже знает предмет
//   - вопросы на СЛОЖНОСТЬ (difficulty) — насколько тяжёлым ощущается предмет
//
// По ответам система автоматически вычисляет difficulty (1-5) и currentLevel (0-100).
package subjects

// QuestionType — тип вопроса теста.
type QuestionType string

const (
	// TypeKnowledge — вопрос оценивает текущий уровень знаний студента.
	TypeKnowledge QuestionType = "knowledge"
	// TypeDifficulty — вопрос оценивает субъективную сложность предмета.
	TypeDifficulty QuestionType = "difficulty"
)

// Option — вариант ответа в вопросе.
type Option struct {
	Text  string `json:"text"`
	Score int    `json:"score"` // очки за выбор этого ответа (0-4)
}

// Question — один вопрос диагностического теста.
type Question struct {
	ID      int          `json:"id"`
	Text    string       `json:"text"`
	Type    QuestionType `json:"type"`
	Options []Option     `json:"options"`
}

// Subject — предмет с описанием и диагностическим тестом.
type Subject struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Emoji       string     `json:"emoji"`
	Description string     `json:"description"`
	BaseDiff    int        `json:"baseDifficulty"` // базовая сложность предмета 1-5
	Questions   []Question `json:"questions"`
}

// TestResult — результат диагностического теста.
type TestResult struct {
	Difficulty   int `json:"difficulty"`   // 1-5
	CurrentLevel int `json:"currentLevel"` // 0-100
}

// ComputeTestResult вычисляет difficulty и currentLevel по ответам пользователя.
// answers — map[questionID]выбранный индекс опции (0-3).
func ComputeTestResult(s Subject, answers map[int]int) TestResult {
	var knowledgeScore, knowledgeMax int
	var diffScore, diffMax int

	for _, q := range s.Questions {
		idx, ok := answers[q.ID]
		if !ok || idx < 0 || idx >= len(q.Options) {
			continue
		}
		score := q.Options[idx].Score
		switch q.Type {
		case TypeKnowledge:
			knowledgeScore += score
			knowledgeMax += 4
		case TypeDifficulty:
			diffScore += score
			diffMax += 4
		}
	}

	// Уровень знаний: 0-100% от набранных очков за knowledge-вопросы
	currentLevel := 5
	if knowledgeMax > 0 {
		currentLevel = (knowledgeScore * 100) / knowledgeMax
	}

	// Сложность: комбинация базовой сложности предмета + субъективных ответов
	difficulty := s.BaseDiff
	if diffMax > 0 {
		// субъективная сложность от 1 до 5 на основе ответов
		subjDiff := 1 + (diffScore*4)/diffMax
		// финальная сложность = 60% база + 40% субъективная
		difficulty = (s.BaseDiff*6 + subjDiff*4) / 10
		if difficulty < 1 {
			difficulty = 1
		}
		if difficulty > 5 {
			difficulty = 5
		}
	}

	return TestResult{
		Difficulty:   difficulty,
		CurrentLevel: currentLevel,
	}
}

// Catalog — список всех готовых предметов.
var Catalog = []Subject{
	{
		ID:          "math",
		Name:        "Математика",
		Emoji:       "∫",
		Description: "Алгебра, геометрия, анализ, теория вероятностей",
		BaseDiff:    4,
		Questions: []Question{
			{
				ID:   1,
				Text: "Как ты решаешь квадратные уравнения?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Никак — не помню формулу", Score: 0},
					{Text: "Только по дискриминанту, иногда ошибаюсь", Score: 1},
					{Text: "Уверенно, знаю несколько способов", Score: 2},
					{Text: "Легко, могу объяснить любым методом", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Что такое производная функции?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Не знаю этой темы", Score: 0},
					{Text: "Слышал(а), но объяснить затрудняюсь", Score: 1},
					{Text: "Знаю определение, могу считать базовые производные", Score: 2},
					{Text: "Свободно дифференцирую сложные функции", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Как ты воспринимаешь математические доказательства?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень тяжело, теряюсь сразу", Score: 4},
					{Text: "С трудом, требуется много времени", Score: 3},
					{Text: "Нормально, если разбирать медленно", Score: 2},
					{Text: "Легко, логика мне понятна", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Сколько времени тебе нужно чтобы разобрать новую тему?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Несколько дней повторений", Score: 4},
					{Text: "Целый день", Score: 3},
					{Text: "Пару часов", Score: 2},
					{Text: "Час или меньше", Score: 1},
				},
			},
			{
				ID:   5,
				Text: "Умеешь ли ты строить графики функций?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет", Score: 0},
					{Text: "Только простейшие (y = x, y = x²)", Score: 1},
					{Text: "Большинство стандартных функций", Score: 3},
					{Text: "Да, включая тригонометрические и логарифмы", Score: 4},
				},
			},
		},
	},
	{
		ID:          "physics",
		Name:        "Физика",
		Emoji:       "⚛",
		Description: "Механика, термодинамика, электричество, оптика",
		BaseDiff:    4,
		Questions: []Question{
			{
				ID:   1,
				Text: "Что описывает второй закон Ньютона?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Не знаю", Score: 0},
					{Text: "Что-то про силу и движение, точнее не помню", Score: 1},
					{Text: "F = ma — знаю и умею применять", Score: 3},
					{Text: "Знаю, умею решать задачи любой сложности", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Как ты решаешь задачи на электрические цепи?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Не умею", Score: 0},
					{Text: "Только простейшие (одно сопротивление)", Score: 1},
					{Text: "Последовательные и параллельные соединения", Score: 3},
					{Text: "Любые схемы, знаю законы Кирхгофа", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Как ты воспринимаешь физические формулы?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Нужно долго зубрить, не понимаю смысл", Score: 4},
					{Text: "Трудно, но понимаю когда объясняют", Score: 3},
					{Text: "Нормально, понимаю откуда они берутся", Score: 1},
					{Text: "Легко — вижу физический смысл", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Насколько сложна для тебя математическая часть задач?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень сложно — математика тянет вниз", Score: 4},
					{Text: "Сложно, часто ошибаюсь в расчётах", Score: 3},
					{Text: "Справляюсь с большинством задач", Score: 1},
					{Text: "Математика не проблема", Score: 0},
				},
			},
			{
				ID:   5,
				Text: "Что такое КПД и как его считать?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Не знаю", Score: 0},
					{Text: "Слышал(а), но не могу посчитать", Score: 1},
					{Text: "Знаю формулу, могу решить простую задачу", Score: 2},
					{Text: "Отлично знаю, применяю в разных темах", Score: 4},
				},
			},
		},
	},
	{
		ID:          "history",
		Name:        "История",
		Emoji:       "📜",
		Description: "Всемирная и отечественная история, даты, события, личности",
		BaseDiff:    2,
		Questions: []Question{
			{
				ID:   1,
				Text: "Знаешь ли ты основные даты XX века?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Почти нет", Score: 0},
					{Text: "Только самые главные (войны, революции)", Score: 1},
					{Text: "Знаю большинство ключевых событий", Score: 3},
					{Text: "Да, могу выстроить хронологию подробно", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Как ты запоминаешь исторических деятелей?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "С трудом, путаю имена и даты", Score: 0},
					{Text: "Помню только самых известных", Score: 1},
					{Text: "Помню многих и их роль в событиях", Score: 3},
					{Text: "Легко — вижу связи между личностями и эпохой", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Как тебе даётся запоминание большого объёма дат?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень тяжело, даты не держатся в голове", Score: 4},
					{Text: "Сложно, нужно много повторений", Score: 3},
					{Text: "Нормально, если использовать ассоциации", Score: 1},
					{Text: "Легко — у меня хорошая память на числа", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Умеешь ли ты анализировать исторические причины и следствия?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет, просто пытаюсь запомнить факты", Score: 0},
					{Text: "Немного понимаю, но с трудом", Score: 1},
					{Text: "Да, понимаю причинно-следственные связи", Score: 3},
					{Text: "Легко — вижу закономерности в истории", Score: 4},
				},
			},
		},
	},
	{
		ID:          "chemistry",
		Name:        "Химия",
		Emoji:       "⚗",
		Description: "Органическая и неорганическая химия, реакции, таблица Менделеева",
		BaseDiff:    4,
		Questions: []Question{
			{
				ID:   1,
				Text: "Умеешь ли ты расставлять коэффициенты в реакциях?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет", Score: 0},
					{Text: "Только простые реакции", Score: 1},
					{Text: "Большинство реакций, метод электронного баланса знаю", Score: 3},
					{Text: "Любые реакции, включая окислительно-восстановительные", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Как ты ориентируешься в таблице Менделеева?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "С трудом нахожу нужный элемент", Score: 0},
					{Text: "Знаю основные элементы и их свойства", Score: 1},
					{Text: "Понимаю периодические закономерности", Score: 3},
					{Text: "Свободно использую таблицу в задачах", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Насколько сложно тебе писать уравнения химических реакций?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень сложно — не понимаю логику", Score: 4},
					{Text: "Трудно, нужно смотреть в учебник", Score: 3},
					{Text: "Справляюсь, иногда ошибаюсь", Score: 2},
					{Text: "Легко — логика ясна", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Умеешь ли ты решать задачи на молярные расчёты?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет, не знаю что такое моль", Score: 0},
					{Text: "Знаю определение, но задачи не решаю", Score: 1},
					{Text: "Решаю стандартные задачи", Score: 3},
					{Text: "Решаю любые задачи, включая многоступенчатые", Score: 4},
				},
			},
		},
	},
	{
		ID:          "english",
		Name:        "Английский язык",
		Emoji:       "EN",
		Description: "Грамматика, лексика, чтение, аудирование, письмо",
		BaseDiff:    3,
		Questions: []Question{
			{
				ID:   1,
				Text: "Как ты воспринимаешь речь на слух (аудирование)?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Почти ничего не понимаю", Score: 0},
					{Text: "Понимаю медленную речь с усилием", Score: 1},
					{Text: "Понимаю бытовую речь и несложные видео", Score: 3},
					{Text: "Понимаю носителей языка без субтитров", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Как ты знаешь английскую грамматику?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Едва знаю базовые времена", Score: 0},
					{Text: "Present/Past/Future — знаю, остальное сложно", Score: 1},
					{Text: "Знаю большинство времён и конструкций", Score: 3},
					{Text: "Свободно применяю сложную грамматику", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Насколько сложно тебе говорить по-английски?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень сложно — теряюсь и молчу", Score: 4},
					{Text: "Сложно — говорю медленно с паузами", Score: 3},
					{Text: "Нормально — объяснюсь в простых ситуациях", Score: 1},
					{Text: "Легко — говорю уверенно", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Как быстро ты читаешь и понимаешь английские тексты?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Очень медленно, нужно переводить каждое слово", Score: 0},
					{Text: "Медленно, понимаю общий смысл", Score: 1},
					{Text: "Нормально, понимаю большую часть без словаря", Score: 3},
					{Text: "Быстро, читаю как на родном языке", Score: 4},
				},
			},
		},
	},
	{
		ID:          "biology",
		Name:        "Биология",
		Emoji:       "🧬",
		Description: "Клетка, генетика, анатомия, экология, эволюция",
		BaseDiff:    3,
		Questions: []Question{
			{
				ID:   1,
				Text: "Знаешь ли ты строение клетки?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет", Score: 0},
					{Text: "Помню несколько органелл", Score: 1},
					{Text: "Знаю основные органеллы и их функции", Score: 3},
					{Text: "Знаю детально, включая отличия прокариот и эукариот", Score: 4},
				},
			},
			{
				ID:   2,
				Text: "Как ты понимаешь законы Менделя?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Не знаю", Score: 0},
					{Text: "Слышал(а), но не понимаю", Score: 1},
					{Text: "Знаю, решаю простые задачи по генетике", Score: 3},
					{Text: "Знаю, решаю задачи любой сложности", Score: 4},
				},
			},
			{
				ID:   3,
				Text: "Сколько терминов тебе нужно запомнить и насколько это тяжело?",
				Type: TypeDifficulty,
				Options: []Option{
					{Text: "Очень много терминов, запоминать тяжело", Score: 4},
					{Text: "Многовато, иногда путаюсь", Score: 2},
					{Text: "Нормально, если разобрать смысл", Score: 1},
					{Text: "Легко — логика помогает запомнить", Score: 0},
				},
			},
			{
				ID:   4,
				Text: "Понимаешь ли ты процесс фотосинтеза?",
				Type: TypeKnowledge,
				Options: []Option{
					{Text: "Нет", Score: 0},
					{Text: "В общих чертах — свет + CO₂ → глюкоза", Score: 1},
					{Text: "Знаю световую и тёмную фазы", Score: 3},
					{Text: "Знаю детально, могу объяснить на молекулярном уровне", Score: 4},
				},
			},
		},
	},
}

// GetByID возвращает предмет по его ID, или false если не найден.
func GetByID(id string) (Subject, bool) {
	for _, s := range Catalog {
		if s.ID == id {
			return s, true
		}
	}
	return Subject{}, false
}
