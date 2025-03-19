package handlers

import (
	"api-test/database"
	"api-test/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateInteraction(c *gin.Context) {
	var interaction models.Interaction
	if err := c.ShouldBindJSON(&interaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()
	response, err := GetAIResponse(interaction.Input)
	duration := time.Since(start)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error contacting AI service"})
		return
	}

	interaction.Response = response

	_, err = database.DB.Exec(
		`INSERT INTO interactions (user_id, input, response, duration) VALUES (?, ?, ?, ?)`,
		interaction.UserID, interaction.Input, interaction.Response, int(duration.Seconds()),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, interaction)
}
