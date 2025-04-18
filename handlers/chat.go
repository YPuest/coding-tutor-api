package handlers

import (
	"api-test/database"
	"api-test/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1]) // entfernt Anführungszeichen
}

func TaskSendChat(c *gin.Context) {
	var req models.TaskChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Printf("TaskSendChat: ShouldBindJSON error: %v", err)
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO interactions (user_id, task_id, role, content, time_remaining, time_spent)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		req.UserId,
		req.TaskId,
		"user",
		req.Message,
		req.TimeRemaining,
		req.TimeSpent,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der User-Nachricht"})
		log.Printf("TaskSendChat: Insertion failed: %v", err)
		return
	}

	rows, err := database.DB.Query(`
		SELECT role, content FROM (
			SELECT id, role, content FROM interactions
			WHERE task_id = ?
			ORDER BY id DESC
			LIMIT 10
		) ORDER BY id ASC
	`, req.TaskId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden des Chatverlaufs"})
		log.Printf("TaskSendChat: Query failed: %v", err)
		return
	}
	defer rows.Close()

	var historyJSON = "[\n"
	first := true
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		if !first {
			historyJSON += ",\n"
		}
		first = false
		historyJSON += fmt.Sprintf(`  { "role": "%s", "content": "%s" }`, escapeJSON(role), escapeJSON(content))
	}
	historyJSON += "\n]"

	prompt := fmt.Sprintf(`
Goal:
Beantworte die Frage, entsprechend dem Level.

Return Format:
- Exaktes JSON-Format (zwingend im JSON-Format, keine illegalen Zeichen, keinerlei zusätzlichen Text!):
{
  "message": "<Antwort>"
}

Warnings:
- Stelle sicher, dass deine Antwort dem Level der Aufgabe entspricht.
- Stelle sicher, dass deine Antwort zur Aufgabenstellung passt.
- Stelle sicher, dass deine Antwort nicht die Lösung enthält, das darf nur ignoriert werden wenn EXPLIZIT nach der Lösung gefragt wird.
- Sollte die Nachricht nicht zum Thema Programmieren passen, antworte bitte mit "Diese Nachricht passt nicht zum Thema. Ich kann nur themenbezogene Nachrichten beantworten.". Sei hierbei aber nicht zu streng!

Context Dump:
- Vorherige Konversation: %s
- Aktuelle Nachricht: "%s"
- Schwierigkeitsgrad: "%s"
- Aufgabe: "%s"
`, historyJSON, req.Message, req.Level, req.Task)

	response, err := GetAIResponse(prompt)
	log.Printf("TaskSendChat: Prompt: %v", prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Kontaktieren der KI"})
		log.Printf("TaskSendChat: GetAIResponse error: %v", err)
		return
	}

	jsonString, err := CleanAndExtractJSON(response)
	if err != nil {
		log.Printf("Fehler beim Extrahieren von JSON: %v\nOriginal: %s\n", err, response)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	var taskChatResponse models.TaskChatResponse
	if err := json.Unmarshal([]byte(jsonString), &taskChatResponse); err != nil {
		log.Printf("json.Unmarshal-Fehler: %v\nBereinigtes JSON: %s\n", err, jsonString)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO interactions (user_id, task_id, role, content)
		VALUES (?, ?, ?, ?)
	`,
		req.UserId,
		req.TaskId,
		"assistant",
		taskChatResponse.Message,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der KI-Antwort"})
		return
	}

	log.Printf("KI-Antwort: %s\n", response)

	c.JSON(http.StatusOK, taskChatResponse)
}
