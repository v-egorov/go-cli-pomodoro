// Реализация функционала таймеров / интервалов
// Типы интервалов:
//
//	Pomodoro - концентрация на задаче
//	short brek - коротокий перерыв
//	long break - длинный перерыв
package pomodoro

import "time"

// Категории (типы) интервалов
const (
	CategoryPomodoro   = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak  = "LongBreak"
)

// Состояния интервалов
const (
	StateNotStarted = iota
	StateRunning
	StatePaused
	StateDone
	StateCancelled
)

type Interval struct {
	ID              int64
	StartTime       time.Time
	PlannedDuration time.Duration
	ActualDuration  time.Duration
	Category        string
	State           int
}
