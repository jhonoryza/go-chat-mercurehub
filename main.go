package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// message structure
type ChatMessage struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// payload structure to mercure
type MercureUpdate struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system env")
	}

	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Nuxt dev server
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/chat", func(c *gin.Context) {
		var msg ChatMessage
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// message serialization to JSON
		data, err := json.Marshal(msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode message"})
			return
		}

		// publish to Mercure
		err = publishToMercure("admin-chat", string(data))
		if err != nil {
			errMsg := fmt.Sprintf("failed, %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": msg,
		})
	})

	r.Run(":8080")
}

// publish to Mercure hub
func publishToMercure(topic string, data string) error {
	hubURL := os.Getenv("MERCURE_URL")         // ex: "http://localhost:3000/.well-known/mercure"
	jwtToken := os.Getenv("MERCURE_JWT_TOKEN") // JWT for publish

	form := make(map[string][]string)
	form["topic"] = []string{topic}
	form["data"] = []string{data}

	req, err := http.NewRequest("POST", hubURL, bytes.NewBufferString(encodeForm(form)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mercure publish failed: %s", resp.Status)
	}

	return nil
}

// helper encode form
func encodeForm(data map[string][]string) string {
	result := ""
	for key, values := range data {
		for _, value := range values {
			if result != "" {
				result += "&"
			}
			result += key + "=" + value
		}
	}
	return result
}
