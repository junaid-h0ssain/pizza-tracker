package main

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"os"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Config holds all application configuration
type Config struct {
	Port             string
	DBPath           string
	SessionSecretKey string
}

func loadConfig() Config {
	return Config{
		Port:             getEnv("PORT", "8080"),
		DBPath:           getEnv("DATABASE_URL", "./data/orders.db"),
		SessionSecretKey: getEnv("SESSION_SECRET_KEY", "pizza-order-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadTemplates(router *gin.Engine) error {
	functions := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	}

	tmpl, err := template.New("").Funcs(functions).ParseGlob("tmp/*.tmpl")
	if err != nil {
		return err
	}

	router.SetHTMLTemplate(tmpl)
	return nil
}


func setupSessionStore(db *gorm.DB, secretKey []byte) sessions.Store {
	store := gormsessions.NewStore(db, true, secretKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   true,
		SameSite: 3, // Strict
	})
	slog.Info("Session store configured", "backend", "gorm-sqlite", "lifetime", "24h")
	return store
}

// Session helper functions

func GetSession(c *gin.Context) sessions.Session {
	return sessions.Default(c)
}

func SetSessionValue(c *gin.Context, key string, value interface{}) error {
	session := GetSession(c)
	session.Set(key, value)
	return session.Save()
}

func GetSessionString(c *gin.Context, key string) string {
	session := GetSession(c)
	val := session.Get(key)
	if val == nil {
		return ""
	}
	str, _ := val.(string)
	return str
}

func GetSessionValue(c *gin.Context, key string) interface{} {
	return GetSession(c).Get(key)
}

func ClearSession(c *gin.Context) error {
	session := GetSession(c)
	session.Clear()
	return session.Save()
}

func DeleteSessionKey(c *gin.Context, key string) error {
	session := GetSession(c)
	session.Delete(key)
	return session.Save()
}