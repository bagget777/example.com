// Package calc реализует математическую модель планировщика подготовки к экзамену.
//
// Модель построена на трёх понятиях математического анализа:
//
//  1. Экстремум — функция эффективности E(x) от количества часов занятий
//     в день x имеет один глобальный максимум (колоколообразная кривая).
//     Эта точка и есть "оптимальное число часов в день".
//
//  2. Производная — функция уровня знаний K(t) растёт со временем нелинейно
//     (логистическая кривая): сначала быстро, затем рост замедляется
//     (закон убывающей отдачи). Производная K'(t) показывает скорость
//     усвоения материала в каждый конкретный день.
//
//  3. Интеграл — накопленные знания K(t) являются интегралом (суммой)
//     мгновенной скорости усвоения K'(t) по всем прошедшим дням:
//     K(t) = K(0) + ∫₀ᵗ K'(τ) dτ.
//     На практике интеграл считается численно (метод сумм / Эйлера) —
//     то есть как накопительная сумма дневных приростов знаний.
package calc

import "math"

// Difficulty — сложность предмета по 5-балльной шкале.
type Difficulty int

const (
	DifficultyVeryEasy Difficulty = 1
	DifficultyEasy     Difficulty = 2
	DifficultyMedium   Difficulty = 3
	DifficultyHard     Difficulty = 4
	DifficultyVeryHard Difficulty = 5
)

// Input — входные данные пользователя.
type Input struct {
	DaysLeft     int     // сколько дней осталось до экзамена
	FreeHours    float64 // сколько часов в день пользователь МОЖЕТ заниматься (максимум)
	Difficulty   Difficulty
	CurrentLevel float64 // текущий уровень знаний, 0..100 (%)
}

// EfficiencyPoint — одна точка графика эффективности E(x).
type EfficiencyPoint struct {
	Hours      float64 `json:"hours"`
	Efficiency float64 `json:"efficiency"`
}

// DayPoint — одна точка на графике роста знаний (день -> уровень знаний).
type DayPoint struct {
	Day        int     `json:"day"`
	Knowledge  float64 `json:"knowledge"`  // K(t), накопленный уровень знаний, %
	GrowthRate float64 `json:"growthRate"` // K'(t), прирост знаний за этот день (производная)
}

// Result — полный результат расчёта, отдаваемый на фронтенд в виде JSON.
type Result struct {
	OptimalHours    float64           `json:"optimalHours"`    // x_opt — точка экстремума E(x)
	PredictedResult float64           `json:"predictedResult"` // K(T) — итоговый уровень знаний к экзамену
	BurnoutRisk     string            `json:"burnoutRisk"`     // "низкий" | "средний" | "высокий"
	BurnoutMessage  string            `json:"burnoutMessage"`
	EfficiencyCurve []EfficiencyPoint `json:"efficiencyCurve"` // график E(x), x от 0.5 до 12 часов
	KnowledgeCurve  []DayPoint        `json:"knowledgeCurve"`  // график K(t) и K'(t) по дням
	TotalKnowledge  float64           `json:"totalKnowledge"`  // интеграл — сумма всех приростов K'(t)
	Input           Input             `json:"input"`
}

// efficiency считает эффективность занятий E(x) при x часах в день.
//
// Это колоколообразная (гауссова) функция с единственным максимумом —
// классический пример функции, исследуемой на экстремум через производную:
// E'(x) = 0 в точке x = xOpt (это и есть вершина параболы/купола,
// о которой говорится в схеме из текста задания).
//
// Чем сложнее предмет и чем выше текущий уровень знаний, тем меньше
// оптимальное число часов в день (выше риск переутомления при той же нагрузке).
func efficiency(x float64, difficulty Difficulty, level float64) (eff float64, xOpt float64) {
	xOpt = 7.0 - 0.6*float64(difficulty) - 0.015*level
	xOpt = clamp(xOpt, 2.0, 9.0)

	const width = 5.5 // "ширина купола" — насколько резко падает эффективность при отклонении от оптимума
	eff = 100 * math.Exp(-((x-xOpt)*(x-xOpt))/(2*width))
	if eff < 0 {
		eff = 0
	}
	return eff, xOpt
}

// burnoutRisk оценивает риск выгорания исходя из того, насколько сильно
// желаемая нагрузка (freeHours) превышает оптимум xOpt, найденный как экстремум E(x).
func burnoutRisk(freeHours, xOpt float64) (level string, message string) {
	diff := freeHours - xOpt
	switch {
	case diff <= 1.0:
		return "низкий", "Нагрузка близка к оптимальной — мозг будет успевать усваивать материал."
	case diff <= 3.0:
		return "средний", "Вы планируете заниматься заметно дольше оптимума — возможна усталость к концу дня."
	default:
		return "высокий", "Запланированная нагрузка сильно превышает оптимум — высокий риск выгорания и падения эффективности."
	}
}

// knowledgeGrowthRate возвращает коэффициент скорости роста знаний (параметр
// логистической модели), зависящий от эффективности занятий и сложности предмета.
func knowledgeGrowthRate(effPercent float64, difficulty Difficulty) float64 {
	effFactor := effPercent / 100.0
	denom := float64(difficulty) / 3.0
	if denom < 0.6 {
		denom = 0.6
	}
	return 0.18 * effFactor / denom
}

// Compute выполняет полный расчёт: находит экстремум E(x), строит кривую
// эффективности, моделирует рост знаний K(t) по дням (логистическая кривая)
// и численно интегрирует производную K'(t), чтобы получить накопленные знания.
func Compute(in Input) Result {
	if in.DaysLeft < 1 {
		in.DaysLeft = 1
	}
	if in.CurrentLevel < 0 {
		in.CurrentLevel = 0
	}
	if in.CurrentLevel > 99 {
		in.CurrentLevel = 99
	}
	if in.FreeHours <= 0 {
		in.FreeHours = 1
	}

	// 1. Находим экстремум функции эффективности E(x).
	_, xOpt := efficiency(in.FreeHours, in.Difficulty, in.CurrentLevel)

	// Оптимальное число часов не может физически превышать доступное пользователю время.
	optimalHours := math.Min(xOpt, in.FreeHours)
	if optimalHours < 0.5 {
		optimalHours = 0.5
	}

	// 2. Строим график E(x) для визуализации купола (x от 0.5 до 12 часов).
	curve := make([]EfficiencyPoint, 0, 24)
	for h := 0.5; h <= 12.0; h += 0.5 {
		e, _ := efficiency(h, in.Difficulty, in.CurrentLevel)
		curve = append(curve, EfficiencyPoint{
			Hours:      round1(h),
			Efficiency: round1(e),
		})
	}

	// 3. Моделируем рост знаний K(t): логистическая кривая, движимая
	//    производной K'(t), которая убывает по мере приближения к 100%
	//    (закон убывающей отдачи — каждый следующий час приносит меньше пользы).
	effAtOptimal, _ := efficiency(optimalHours, in.Difficulty, in.CurrentLevel)
	growthRate := knowledgeGrowthRate(effAtOptimal, in.Difficulty)

	k0 := in.CurrentLevel
	if k0 < 1 {
		k0 = 1
	}
	A := (100.0 - k0) / k0

	knowledgePoints := make([]DayPoint, 0, in.DaysLeft+1)
	var totalGrowth float64 // интеграл K'(t) — накопленная сумма приростов
	kPrev := k0
	for day := 0; day <= in.DaysLeft; day++ {
		kt := 100.0 / (1 + A*math.Exp(-growthRate*float64(day)))
		growth := 0.0
		if day > 0 {
			growth = kt - kPrev // дискретная производная K'(t) ≈ ΔK
			totalGrowth += growth
		}
		knowledgePoints = append(knowledgePoints, DayPoint{
			Day:        day,
			Knowledge:  round1(kt),
			GrowthRate: round2(growth),
		})
		kPrev = kt
	}

	finalLevel := knowledgePoints[len(knowledgePoints)-1].Knowledge

	risk, riskMsg := burnoutRisk(in.FreeHours, xOpt)

	return Result{
		OptimalHours:    round1(optimalHours),
		PredictedResult: finalLevel,
		BurnoutRisk:     risk,
		BurnoutMessage:  riskMsg,
		EfficiencyCurve: curve,
		KnowledgeCurve:  knowledgePoints,
		TotalKnowledge:  round1(totalGrowth),
		Input:           in,
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
