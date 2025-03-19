package models

type TaskChatRequest struct {
	UserId        int    `json:"user_id"`
	TaskId        int    `json:"task_id"`
	Message       string `json:"message"`
	Level         string `json:"level"`
	Language      string `json:"language"`
	Task          string `json:"task"`
	TimeRemaining int    `json:"time_remaining"`
	TimeSpent     int    `json:"time_spent"`
}

type TaskChatResponse struct {
	Message string `json:"message"`
}
