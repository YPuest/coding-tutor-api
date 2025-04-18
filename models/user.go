package models

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Stats struct {
	AvgMark        float64 `db:"avg_mark" json:"avg_mark"`
	AIUsageRate    float64 `db:"ai_usage_rate" json:"ai_usage_rate"`
	TotalTasks     int     `db:"total_tasks" json:"total_tasks"`
	CompletedTasks int     `db:"completed_tasks" json:"completed_tasks"`
	LanguageUsage  string  `db:"language_usage" json:"language_usage"`
}

type StatsFull struct {
	AvgMark              *float64                  `db:"avg_mark" json:"avg_mark"`
	AIUsageRate          *float64                  `db:"ai_usage_rate" json:"ai_usage_rate"`
	TotalTasks           int                       `db:"total_tasks" json:"total_tasks"`
	CompletedTasks       int                       `db:"completed_tasks" json:"completed_tasks"`
	LanguageDistribution map[string]int            `json:"language_distribution"`
	TaskStatusChart      map[string]map[string]int `json:"task_status_chart"`
	AIUsageChart         map[string]map[string]int `json:"ai_usage_chart"`
}

type StatsLanguage struct {
	TotalTasks     int            `json:"total_tasks"`
	CompletedTasks int            `json:"completed_tasks"`
	AIWithUsage    int            `json:"ai_with_usage"`
	AIWithoutUsage int            `json:"ai_without_usage"`
	TaskLevels     map[string]int `json:"task_levels"`
	AvgMark        float64        `json:"avg_mark"`
}

type Tasks []struct {
	ID            int      `db:"id" json:"id"`
	Description   string   `db:"description" json:"description"`
	Language      string   `db:"language" json:"language"`
	Mark          *float64 `db:"mark" json:"mark"`
	Level         string   `db:"level" json:"level"`
	AIUsage       int      `db:"ai_usage" json:"ai_usage"`
	TimeSpent     *int     `db:"time_spent" json:"time_spent"`
	TimeEstimated int      `db:"time_estimated" json:"time_estimated"`
	Rating        *string  `db:"rating" json:"rating"`
}

type Task struct {
	ID            int               `json:"id" db:"id"`
	Description   string            `json:"description" db:"description"`
	Level         string            `json:"level" db:"level"`
	Language      string            `json:"language" db:"language"`
	Mark          *float64          `json:"mark" db:"mark"`
	Rating        *string           `json:"rating" db:"rating"`
	TimeSpent     *int              `json:"time_spent" db:"time_spent"`
	TimeEstimated int               `json:"time_estimated" db:"time_estimated"`
	AIUsage       int               `json:"ai_usage" db:"ai_usage"`
	Code          *string           `json:"code" db:"code"`
	Interactions  []TaskInteraction `json:"interactions"`
}

type TaskInteraction struct {
	ID            int    `json:"id" db:"id"`
	UserID        int    `json:"user_id" db:"user_id"`
	TaskID        int    `json:"task_id" db:"task_id"`
	Role          string `json:"role" db:"role"`
	Content       string `json:"content" db:"content"`
	TimeRemaining *int   `json:"time_remaining" db:"time_remaining"`
	TimeSpent     *int   `json:"time_spent" db:"time_spent"`
	CategoryID    *int   `json:"category_id" db:"category_id"`
}

type ChangeUsername struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

type ChangePassword struct {
	UserID      int    `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type DeleteAccount struct {
	UserID int `json:"user_id"`
}
