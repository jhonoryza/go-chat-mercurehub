package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

// message structure
type ChatMessage struct {
	UserID  string `json:"user_id" binding:"required,max=50"`
	Message string `json:"message" binding:"required"`
	IsRead  bool   `json:"is_read"`
	Channel string `json:"channel" binding:"required,max=100"`
}

// payload structure to mercure
type MercureUpdate struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system env")
	}

	// connect to postgres
	db, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("❌ Cannot connect to database:", err)
	}
	defer db.Close()

	// test connection
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Cannot ping database:", err)
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

		// Simpan ke DB
		query := `
        INSERT INTO messages (channel, user_id, message, is_read)
        VALUES ($1, $2, $3, $4)`
		_, err := db.Exec(query, msg.Channel, msg.UserID, msg.Message, msg.IsRead)
		if err != nil {
			log.Println("DB insert error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
			return
		}

		// message serialization to JSON
		data, err := json.Marshal(msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode message"})
			return
		}

		// publish to Mercure
		err = publishToMercure(msg.Channel, string(data))
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

	r.GET("/messages", func(c *gin.Context) {
		beforeID := c.Query("before") // optional

		query := `
        SELECT id, channel, user_id, message, is_read, created_at
        FROM messages
        WHERE channel = $1
    `
		args := []any{"admin-chat"}

		if beforeID != "" {
			query += " AND id < $2"
			args = append(args, beforeID)
		}

		query += " ORDER BY id DESC LIMIT 20"

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Println("DB query error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
			return
		}
		defer rows.Close()

		messages := []ChatMessage{}
		for rows.Next() {
			var m ChatMessage
			var id int64
			var createdAt time.Time
			rows.Scan(&id, &m.Channel, &m.UserID, &m.Message, &m.IsRead, &createdAt)
			messages = append(messages, m)
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
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
