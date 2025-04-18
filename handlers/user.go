package handlers

import (
	"api-test/database"
	"api-test/models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Hashen des Passworts"})
		return
	}

	_, err = database.DB.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		user.Username, string(hashed),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User erfolgreich registriert"})
}

func LoginUser(c *gin.Context) {
	var creds models.Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE username = ?", creds.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültige Anmeldedaten"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültige Anmeldedaten"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login erfolgreich",
		"user_id": user.ID,
	})
}

func GetUserStats(c *gin.Context) {
	userID := c.Query("user_id")

	var stats models.Stats

	err := database.DB.Get(&stats, `
		SELECT 
			COALESCE(AVG(solutions.mark), 0) AS avg_mark,
			(COUNT(CASE WHEN solutions.ai_usage > 0 THEN 1 END) * 100.0 / 
			 COUNT(CASE WHEN solutions.task_id IS NOT NULL THEN 1 END)) AS ai_usage_rate,
			COUNT(tasks.id) AS total_tasks,
			SUM(CASE WHEN solutions.task_id IS NOT NULL THEN 1 ELSE 0 END) AS completed_tasks,
			(
				SELECT '{' || GROUP_CONCAT('"' || language || '": ' || task_count) || '}'
			 	FROM 
			 	    (
			 	    	SELECT language, COUNT(*) AS task_count
				   		FROM tasks 
				   		WHERE user_id = ? 
				   		GROUP BY language
					)
			) AS language_usage
		FROM tasks
			LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.user_id = ?`, userID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Statistiken"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetUserStatsFull(c *gin.Context) {
	userID := c.Query("user_id")

	var stats models.StatsFull

	err := database.DB.QueryRow(`
       	SELECT 
			COALESCE(AVG(solutions.mark), 0) AS avg_mark,
			(COUNT(CASE WHEN solutions.ai_usage > 0 THEN 1 END) * 100.0 / 
			 COUNT(CASE WHEN solutions.task_id IS NOT NULL THEN 1 END)) AS ai_usage_rate,
			COUNT(tasks.id) AS total_tasks,
			SUM(CASE WHEN solutions.task_id IS NOT NULL THEN 1 ELSE 0 END) AS completed_tasks
		FROM tasks
			LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.user_id = ?`, userID).Scan(
		&stats.AvgMark, &stats.AIUsageRate, &stats.TotalTasks, &stats.CompletedTasks,
	)

	if err != nil {
		log.Printf("DB Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Statistiken"})
		return
	}

	rows, err := database.DB.Query(`
        SELECT language, COUNT(*) 
        FROM tasks
        WHERE user_id = ? 
        GROUP BY language`, userID)

	if err != nil {
		log.Printf("DB Error (Language Distribution): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Sprachenverteilung"})
		return
	}
	defer rows.Close()

	stats.LanguageDistribution = make(map[string]int)
	for rows.Next() {
		var lang string
		var count int
		rows.Scan(&lang, &count)
		stats.LanguageDistribution[lang] = count
	}

	rows, err = database.DB.Query(`
        SELECT language, 
               SUM(CASE WHEN solutions.mark IS NOT NULL THEN 1 ELSE 0 END) as completed, 
               SUM(CASE WHEN solutions.mark IS NULL THEN 1 ELSE 0 END) as not_completed
        FROM tasks 
        LEFT JOIN solutions ON tasks.id = solutions.task_id
        WHERE tasks.user_id = ?
        GROUP BY language`, userID)

	if err != nil {
		log.Printf("DB Error (Task Status Chart): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen des Aufgabenstatus"})
		return
	}
	defer rows.Close()

	stats.TaskStatusChart = make(map[string]map[string]int)
	for rows.Next() {
		var lang string
		var completed, notCompleted int
		rows.Scan(&lang, &completed, &notCompleted)
		stats.TaskStatusChart[lang] = map[string]int{"completed": completed, "not_completed": notCompleted}
	}

	rows, err = database.DB.Query(`
        SELECT language, 
               SUM(CASE WHEN solutions.ai_usage = 1 THEN 1 ELSE 0 END) as with_ai, 
               SUM(CASE WHEN solutions.ai_usage = 0 THEN 1 ELSE 0 END) as without_ai
        FROM tasks 
        LEFT JOIN solutions ON tasks.id = solutions.task_id
        WHERE tasks.user_id = ?
        GROUP BY language`, userID)

	if err != nil {
		log.Printf("DB Error (AI Usage Chart): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der KI-Nutzung"})
		return
	}
	defer rows.Close()

	stats.AIUsageChart = make(map[string]map[string]int)
	for rows.Next() {
		var lang string
		var withAI, withoutAI int
		rows.Scan(&lang, &withAI, &withoutAI)
		stats.AIUsageChart[lang] = map[string]int{"with_ai": withAI, "without_ai": withoutAI}
	}

	c.JSON(http.StatusOK, stats)
}

func GetUserStatsLanguage(c *gin.Context) {
	userID := c.Query("user_id")
	language := c.Query("language")

	var stats models.StatsLanguage
	stats.TaskLevels = make(map[string]int)

	err := database.DB.QueryRow(`
		SELECT 
			(
				SELECT COUNT(*) 
				FROM tasks 
				WHERE tasks.language = ? 
				  AND tasks.user_id = ?
			) AS total_tasks,
			COUNT(solutions.id) AS completed_tasks,
			COALESCE(AVG(solutions.mark), 0) AS avg_mark
		FROM tasks
			LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.language = ? 
			AND tasks.user_id = ?
			AND solutions.task_id IS NOT NULL`,
		language, userID, language, userID).Scan(&stats.TotalTasks, &stats.CompletedTasks, &stats.AvgMark)

	if err != nil {
		log.Printf("DB Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Sprachstatistiken"})
		return
	}

	err = database.DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN solutions.ai_usage > 0 THEN 1 END) AS ai_with_usage,
			COUNT(CASE WHEN solutions.ai_usage = 0 THEN 1 END) AS ai_without_usage
		FROM tasks
			LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.language = ? 
			AND tasks.user_id = ?
			AND solutions.task_id IS NOT NULL;`,
		language, userID).Scan(&stats.AIWithUsage, &stats.AIWithoutUsage)

	if err != nil {
		log.Printf("DB Error (AI Usage): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der KI-Nutzung"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT level, COUNT(*) 
		FROM tasks 
		WHERE language = ? AND user_id = ? 
		GROUP BY level`,
		language, userID)

	if err != nil {
		log.Printf("DB Error (Task Levels): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Aufgabenlevel"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			log.Printf("DB Scan Error: %v", err)
			continue
		}
		stats.TaskLevels[level] = count
	}

	log.Printf("%+v\n", stats)

	c.JSON(http.StatusOK, stats)
}

func GetUserTasks(c *gin.Context) {
	userID := c.Query("user_id")

	log.Printf("GetUserTasks(%s)", userID)

	var tasks models.Tasks

	err := database.DB.Select(&tasks, `
		SELECT
		    tasks.id, tasks.description, tasks.language, solutions.mark,
    		tasks.level, COALESCE(solutions.ai_usage, 0) as ai_usage, 
    		COALESCE(solutions.time_spent, 0) as time_spent,
    		tasks.time_estimated, solutions.rating
		FROM tasks
        	LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.user_id = ?`, userID)

	if err != nil {
		log.Printf("DB Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Aufgaben"})
		return
	}

	log.Printf("Fetched Tasks: %+v", tasks)

	c.JSON(http.StatusOK, tasks)
}

func GetSingleTask(c *gin.Context) {
	taskID := c.Param("task_id")

	var task models.Task

	err := database.DB.Get(&task, `
		SELECT
			tasks.id, tasks.description, tasks.language, tasks.level,
			COALESCE(solutions.mark, NULL) as mark,
			COALESCE(solutions.rating, 'Keine Bewertung') as rating, 
			COALESCE(solutions.time_spent, 0) as time_spent, 
			tasks.time_estimated,
			COALESCE(solutions.ai_usage, 0) as ai_usage, 
			COALESCE(solutions.code, '') as code
		FROM tasks
		LEFT JOIN solutions ON tasks.id = solutions.task_id
		WHERE tasks.id = ?`, taskID)

	if err != nil {
		log.Printf("DB Error (task fetch): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Aufgabe"})
		return
	}
	var interactions []models.TaskInteraction
	err = database.DB.Select(&interactions, `
		SELECT * FROM interactions
		WHERE task_id = ?
		ORDER BY id`, taskID)

	if err != nil {
		log.Printf("DB Error (interaction fetch): %v", err)
		interactions = []models.TaskInteraction{}
	}

	task.Interactions = interactions

	c.JSON(http.StatusOK, task)
}

func ChangeUsername(c *gin.Context) {
	var req models.ChangeUsername

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
		return
	}

	_, err := database.DB.Exec("UPDATE users SET username = ? WHERE id = ?", req.Username, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Ändern des Nutzernamens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nutzername erfolgreich geändert"})
}

func ChangePassword(c *gin.Context) {
	var req models.ChangePassword

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
		return
	}

	var hashedPassword string
	err := database.DB.Get(&hashedPassword, "SELECT password FROM users WHERE id = ?", req.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Benutzer nicht gefunden"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Altes Passwort ist falsch"})
		return
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Hashen des neuen Passworts"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedNewPassword), req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Ändern des Passworts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passwort erfolgreich geändert"})
}

func DeleteAccount(c *gin.Context) {
	var req models.DeleteAccount

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
		return
	}

	tx, err := database.DB.Beginx()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Starten der Transaktion"})
		return
	}

	_, err = tx.Exec("DELETE FROM interactions WHERE task_id IN (SELECT id FROM tasks WHERE user_id = ?)", req.UserID)
	if err != nil {
		log.Printf("Error deleting interactions: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen der Interaktionen"})
		return
	}

	_, err = tx.Exec("DELETE FROM solutions WHERE task_id IN (SELECT id FROM tasks WHERE user_id = ?)", req.UserID)
	if err != nil {
		log.Printf("Error deleting solutions: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen der Lösungen"})
		return
	}

	_, err = tx.Exec("DELETE FROM tasks WHERE user_id = ?", req.UserID)
	if err != nil {
		log.Printf("Error deleting tasks: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen der Aufgaben"})
		return
	}

	_, err = tx.Exec("DELETE FROM users WHERE id = ?", req.UserID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Benutzers"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Transaction commit error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abschließen der Transaktion"})
		return
	}

	log.Printf("Account with user_id=%s successfully deleted.", req.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Konto erfolgreich gelöscht"})
}
