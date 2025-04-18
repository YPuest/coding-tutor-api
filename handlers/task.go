package handlers

import (
	"api-test/database"
	"api-test/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GenerateTask(c *gin.Context) {
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("%+v\n", req)

	prompt := fmt.Sprintf(`
Goal:
Erstelle eine klar formulierte, praxisnahe Programmieraufgabe für Studierende mit abwechslungsreichem Einstieg. 
Nutze unterschiedliche Szenarien oder Anwendungsbeispiele (z. B. Spiel, Alltag, Wissenschaft, Web, Textverarbeitung).
Die Aufgabe soll sich nicht wie eine Standardformulierung anfühlen.

Return Format:
- Präzise Aufgabenstellung (max. 150 Wörter).
- Realistische Zeiteinschätzung zur Bearbeitung (minimale und maximale Zeit in Minuten).
- Gib keine Code-Fences an.
- Exaktes JSON-Format (zwingend im JSON-Format, keine illegalen Zeichen, keinerlei zusätzlichen Text!):
{
  "task": "<Aufgabenbeschreibung>",
  "time_estimation_minutes": <geschätzte Zeit als Zahl>
}

Instructions:
- Verwende für den Einstieg kreative Kontexte, damit sich Aufgaben unterschiedlich anfühlen.
- Stelle sicher, dass die Aufgabe nur mit Standardbibliotheken lösbar ist.
- Stelle sicher, dass die Aufgabe in einer einzigen Datei lösbar ist.
- Wann immer möglich, soll ein bestimmter, dem Schwierigkeitsgrad entsprechender Algorithmus abgefragt werden.
- Gib realistische und nicht überzogene Zeitschätzungen an. Die Zeitschätzung darf auf keinen Fall 0 sein!

Context Dump:
- Programmiersprache: "%s";
- Schwierigkeitsgrad: "%s";
- Zusätzliche Anmerkungen: "%s"
`, req.Language, req.Level, req.Comment)

	response, err := GetAIResponse(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Kontaktieren der KI"})
		return
	}

	jsonString, err := CleanAndExtractJSON(response)
	if err != nil {
		log.Printf("Fehler beim Extrahieren von JSON: %v\nOriginal: %s\n", err, response)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	var taskResponse models.TaskResponse
	if err := json.Unmarshal([]byte(jsonString), &taskResponse); err != nil {
		log.Printf("json.Unmarshal-Fehler: %v\nBereinigtes JSON: %s\n", err, jsonString)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	log.Printf("%+v\n", taskResponse)

	c.JSON(http.StatusOK, taskResponse)
}

func SaveTask(c *gin.Context) {
	var req models.TaskSaveRequest

	log.Printf("%+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := database.DB.Exec(`
		INSERT INTO tasks (user_id, description, language, level, time_estimated)
		VALUES (?, ?, ?, ?, ?)
	`, req.UserID, req.Description, strings.ToLower(req.Language), req.Level, req.TimeEstimation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der Aufgabe"})
		return
	}

	taskID, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Task-ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"message": "Aufgabe erfolgreich gespeichert",
	})
}

func EvaluateTask(c *gin.Context) {
	var req models.TaskEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("%+v\n", req)

	useAI := ""
	if req.UseAI {
		useAI = "ja"
	} else {
		useAI = "nein"
	}

	prompt := fmt.Sprintf(`
Goal:
Bewerte die eingereichte Lösung zu folgender Aufgabe.

Return Format:
- Eine kurze Bewertung hinsichtlich Codequalität, Lesbarkeit und Effizienz. Bitte beachte, ob KI genutzt wurde.
- Bewerte explizit, ob die Lösung einer Musterlösung nahe kommt.
- Formuliere die Bewertung motivierend und konstruktiv, auch wenn es Schwächen gibt.
- Gib am Ende einen kurzen Verbesserungsvorschlag ("Tipp") – maximal ein Satz.
- Vergib eine Schulnote von 1,0 (sehr gut) bis 6,0 (ungenügend). Schritte von 0,1 sind möglich.
- Begründe in der Bewertung, weshalb du diese Note vergeben hast.
- Vergleich zwischen geschätzter Zeit und benötigter Zeit (realistisch, zu schnell, zu langsam).
- Generiere eine mögliche und gültige Lösung, die als Code ausführbar ist.
- Gib keine Code-Fences an.
- Exaktes JSON-Format (zwingend im JSON-Format, keine illegalen Zeichen, keinerlei zusätzlichen Text!):
{
  "rating": "<Bewertung mit Hinweis und Verbesserungsvorschlag>",
  "mark": "<Schulnote: x,y>",
  "time_comparison": <Vergleich der Zeiten>,
  "solution": <generierte Lösung>
}

Warnings:
- Gib objektive und realistische Bewertungen.
- Beachte den Schwierigkeitsgrad (super-easy bis super-hard).
- Wenn die eingereichte Lösung einer Musterlösung entspricht muss die Note mindestens eine 1,3 sein.

Context Dump:
- Aufgabe: "%s";
- Eingereichter Code: "%s";
- Level: "%s";
- Sprache: "%s";
- KI-Nutzung: "%s";
- Zeitangabe: %d Sekunden;
- Tatsächlich benötigte Zeit: %d Sekunden
`, req.Task, req.Code, req.Level, req.Language, useAI, req.TimeEstimation, req.TimeSpent)

	response, err := GetAIResponse(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei KI-Anfrage"})
		return
	}

	log.Println(response)

	jsonString, err := CleanAndExtractJSON(response)
	if err != nil {
		log.Printf("Fehler beim Extrahieren von JSON: %v\nOriginal: %s\n", err, response)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	var evalResponse models.TaskEvaluation
	if err := json.Unmarshal([]byte(jsonString), &evalResponse); err != nil {
		log.Printf("json.Unmarshal-Fehler: %v\nBereinigtes JSON: %s\n", err, jsonString)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Parsen der KI-Antwort"})
		return
	}

	markStr := strings.Replace(evalResponse.Mark, ",", ".", 1)
	mark, err := strconv.ParseFloat(markStr, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim konvertieren der Note"})
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO solutions (task_id, code, rating, mark, ai_usage, time_spent)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		req.TaskID,
		req.Code,
		evalResponse.Rating,
		mark,
		req.UseAI,
		req.TimeSpent,
	)

	log.Printf("%+v\n", evalResponse)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der User-Nachricht"})
		return
	}

	c.JSON(http.StatusOK, evalResponse)
}
