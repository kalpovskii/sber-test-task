package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var dbURL string

func initConfig() {
	viper.SetEnvPrefix("CHECKLIST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) 
	viper.AutomaticEnv()
}

func main() {
	initConfig()

	apiPort := viper.GetString("API_PORT")
	dbURL = viper.GetString("DB_SERVICE_URL") 

	if apiPort == "" || dbURL == "" {
		log.Fatal("api.port or db.service_url is not configured")
	}

	log.Printf("API started on :%s", apiPort)
	log.Printf("Proxying to %s", dbURL)

	r := gin.Default()

	r.POST("/create", proxyToDB("/create"))
	r.GET("/list", proxyToDB("/list"))
	r.DELETE("/delete", proxyToDB("/delete"))
	r.PUT("/done", proxyToDB("/done"))

	log.Fatal(r.Run(":" + apiPort))
}

func proxyToDB(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := http.NewRequest(c.Request.Method, dbURL+path, c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		req.Header = c.Request.Header.Clone()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}