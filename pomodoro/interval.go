// Реализация функционала таймеров / интервалов
// Типы интервалов:
//
//	Pomodoro - концентрация на задаче
//	short brek - коротокий перерыв
//	long break - длинный перерыв
package pomodoro

import (
	"context"
	"errors"
	"time"
)

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

// Интервал
type Interval struct {
	ID              int64
	StartTime       time.Time
	PlannedDuration time.Duration
	ActualDuration  time.Duration
	Category        string
	State           int
}

// Репозиторий интервалов
type Repository interface {
	// Создаёт новый интервал в репозитории
	Create(i Interval) (int64, error)

	// Обновить интервал
	Update(i Interval) error

	// Возвращает интервал по id
	ByID(id int64) (Interval, error)

	// Возвращает последний (текущий) интервал
	Last() (Interval, error)

	// Возвращает n последних интервалов типа "перерыв"
	Breaks(n int) ([]Interval, error)
}

// Ошибки
var (
	ErrNoIntervals        = errors.New("интервалы отсутствуют")
	ErrIntervalNotRunnins = errors.New("интервал не исполняется")
	ErrIntervalCompleted  = errors.New("интервал завершен или отменён")
	ErrInvalidState       = errors.New("неверное состояние интервала")
	ErrInvalidID          = errors.New("неверный индентификатор интервала")
)

// Конфигурация для создания нового интервала
type IntevalConfig struct {
	repo               Repository
	PomodoroDuration   time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
}

// Контруктор IntevalConfig
func NewConfig(repo Repository, pomodoro, shortBreak, longBreak time.Duration) *IntevalConfig {
	c := &IntevalConfig{
		repo:               repo,
		PomodoroDuration:   25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
	}

	if pomodoro > 0 {
		c.PomodoroDuration = pomodoro
	}
	if shortBreak > 0 {
		c.ShortBreakDuration = shortBreak
	}
	if longBreak > 0 {
		c.LongBreakDuration = longBreak
	}

	return c
}

// Возвращает следующую категорию для репозитория r
func nextCategory(r Repository) (string, error) {
	li, err := r.Last()
	// Интервалов ещё не было - нужно начинать работать (pomodoro)
	if err != nil && err == ErrNoIntervals {
		return CategoryPomodoro, nil
	}
	if err != nil {
		return "", err
	}

	// После перерыва возвращаемся к работе - pomodoro
	if li.Category == CategoryShortBreak || li.Category == CategoryLongBreak {
		return CategoryPomodoro, nil
	}

	// Если мы оказались здесь - то мы работаем сейчас, и следующий интервал - перерыв.
	// И нам нужно выяснить, какой тип перерыва будет следующим.

	// Возьмём 3 последних перерыва в слайс lastBreaks
	lastBreaks, err := r.Breaks(3)
	if err != nil {
		return "", err
	}

	// Если перерывов в целом было менее 3 - то следующий перерыв будет короткий
	if len(lastBreaks) < 3 {
		return CategoryShortBreak, nil
	}

	// Если среди 3-х последних перерывов был длинный - то следующий перерыв будет короткий
	for _, i := range lastBreaks {
		if i.Category == CategoryLongBreak {
			return CategoryShortBreak, nil
		}
	}

	// Все 3 последних перерыва были короткими - следующий будет длинный
	return CategoryLongBreak, nil
}

type Callback func(Interval)

func tick(ctx context.Context, id int64, config *IntevalConfig, start, periodic, end Callback) error {
	// Создаём тикер, в котором будет канал C, c сигналом каждую секунду,
	// в сигнале будет содержаться текущее время. Буфер канала - 1 элемент, если не успеем
	// вычиать из канала значение, оно потеряется без к-л побочных эффектов.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	i, err := config.repo.ByID(id)
	if err != nil {
		return err
	}
	// Время интервала закончится через колчисество секунд
	// i.PlannedDuration - i.ActualDuration
	// Мы можем перезапустить интервал после того, как он отработает
	// какое-то время и поставлен на паузу. Поэтому при старте (рестарте)
	// мы вычисляем время истечения с учетом возможного рестарта, когда в
	// ActualDuration уже накоплено какое-то количесто секунд.
	expire := time.After(i.PlannedDuration - i.ActualDuration)
	start(i)

	for {
		select {
		// Ждём и получаем сигнал из канала
		case <-ticker.C:
			// сюда попадаем каждую секунду

			// Получаем интервал из репозитория
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			// если интервал в состоянии StatePaused - не делаем ничего
			if i.State == StatePaused {
				return nil
			}

			// Увеличиваем продолжительность ActualDuration
			// на одну секунду (потому что мы здесь оказываемся каждую секунду)
			// Обновляем интервал в репозитории
			// Вызываем callback periodic
			i.ActualDuration += time.Second
			if err := config.repo.Update(i); err != nil {
				return err
			}
			periodic(i)
		case <-expire:
			// Таймер закончился
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}
			i.State = StateDone
			end(i)
			return config.repo.Update(i)
		case <-ctx.Done():
			// Получили сигнал из контекста - нужно прервать исполнение
			i, err := config.repo.ByID(id)
			if err != nil {
				return err
			}
			i.State = StateCancelled
			return config.repo.Update(i)
		}
	}
}
