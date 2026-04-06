package model

import "github.com/google/uuid"

type Task struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Type     string    `json:"type"`
	Problem  string    `gorm:"unique" json:"problem"`
	Solution string    `json:"solution"`
	Level    string    `json:"level"`
	Boxed    string    `json:"boxed"`
	Likes    uint      `json:"likes"`
	Dislikes uint      `json:"dislikes"`
}

func NewTask(taskType, problem, solution, boxed, level string) *Task {
	return &Task{
		ID:       uuid.New(),
		Type:     taskType,
		Solution: solution,
		Problem:  problem,
		Boxed:    boxed,
		Level:    level,
	}
}
