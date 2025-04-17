package models

type TaskRequest struct {
	Language string `json:"language"`
	Level    string `json:"level"`
	Comment  string `json:"comment"`
}

type TaskResponse struct {
	Task           string `json:"task"`
	TimeEstimation int    `json:"time_estimation_minutes"`
}

type TaskSaveRequest struct {
	UserID         int    `json:"user_id"`
	Description    string `json:"description"`
	Language       string `json:"language"`
	Level          string `json:"level"`
	TimeEstimation int    `json:"time_estimated"`
}

type TaskEvaluationRequest struct {
	UserID         int    `json:"user_id"`
	TaskID         int    `json:"task_id"`
	Code           string `json:"code"`
	Level          string `json:"level"`
	Language       string `json:"language"`
	Task           string `json:"task"`
	UseAI          bool   `json:"use_ai"`
	TimeEstimation int    `json:"time_estimation"`
	TimeSpent      int    `json:"time_spent"`
}

type TaskEvaluation struct {
	Rating         string `json:"rating"`
	Mark           string `json:"mark"`
	TimeComparison string `json:"time_comparison"`
	Solution       string `json:"solution"`
}
