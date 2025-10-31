package response

import (
	"time"

	"github.com/zuxt268/sales/internal/model"
)

type Log struct {
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Logs struct {
	Logs []*Log `json:"logs"`
	Paginate
}

func GetLog(l *model.Log) *Log {
	return &Log{
		Name:      l.Name,
		Category:  l.Category,
		Message:   l.Message,
		CreatedAt: l.CreatedAt,
	}
}

func GetLogs(ls []*model.Log, total int64) *Logs {
	respLogs := make([]*Log, len(ls))
	for i, l := range ls {
		respLogs[i] = GetLog(l)
	}
	return &Logs{
		Logs: respLogs,
		Paginate: Paginate{
			Total: total,
			Count: len(ls),
		},
	}
}
