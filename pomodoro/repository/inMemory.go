package repository

import (
	"fmt"
	"log/slog"
	"sync"

	"vegorov.ru/go-cli/pomo/pomodoro"
)

// Репозиторий для работы с интервалами в памяти
type inMemoryRepo struct {
	// не именованное поле, методы будут доступны в структуре inMemoryRepo
	sync.RWMutex
	// Интервалы хранятся в слайсе intervals, доступ к - семафорим через RWMutex
	intervals []pomodoro.Interval
}

func NewInMemoryRepo() *inMemoryRepo {
	slog.Debug("Creating InMemoryRepo")
	return &inMemoryRepo{
		intervals: []pomodoro.Interval{},
	}
}

// Записывает интервал в репозиторий, возвращет ID в репозитории
func (r *inMemoryRepo) Create(i pomodoro.Interval) (int64, error) {
	// Слайсы не concurent-safe, поэтому для операций над слайсом
	// []inetravls нужен механизм RWMutex
	r.Lock()
	defer r.Unlock()

	// в ID по сути будет 1-based номер по порядку в слайсе
	i.ID = int64(len(r.intervals)) + 1

	r.intervals = append(r.intervals, i)

	return i.ID, nil
}

// Обновляет интервал в репозитории
func (r *inMemoryRepo) Update(i pomodoro.Interval) error {
	r.Lock()
	defer r.Unlock()

	if i.ID == 0 {
		return fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, i.ID)
	}

	// Заменяем в слайсе значение на новое - которое пришло в параметре i
	r.intervals[i.ID-1] = i
	return nil
}

func (r *inMemoryRepo) ByID(id int64) (pomodoro.Interval, error) {
	r.Lock()
	defer r.Unlock()

	i := pomodoro.Interval{}
	if id == 0 {
		return i, fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, id)
	}

	i = r.intervals[id-1]
	return i, nil
}

// Возвращает последний интервал из репозитория
func (r *inMemoryRepo) Last() (pomodoro.Interval, error) {
	r.Lock()
	defer r.Unlock()

	i := pomodoro.Interval{}
	if len(r.intervals) == 0 {
		return i, pomodoro.ErrNoIntervals
	}
	return r.intervals[len(r.intervals)-1], nil
}

// Возвращает n последних перерывов из репозитория
func (r *inMemoryRepo) Breaks(n int) ([]pomodoro.Interval, error) {
	r.Lock()
	defer r.Unlock()

	// Пустышка для накопления данных для возврата
	returnData := []pomodoro.Interval{}

	for k := len(r.intervals) - 1; k >= 0; k-- {
		if r.intervals[k].Category == pomodoro.CategoryPomodoro {
			// В категориях (типах) интервалов у нас есть только CategoryPomodoro (работа)
			// и перерывы - поэтому скипаем работу, а если не скипнули - то это перерыв,
			// а они-то нам и нужны.
			continue
		}
		// Накапливаем слайс с перерывами
		returnData = append(returnData, r.intervals[k])
		if len(returnData) == n {
			return returnData, nil
		}
	}
	return returnData, nil
}
