package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var dbURL string

func main() {
	viper.SetConfigFile("../../configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	dbURL = "http://localhost:" + viper.GetString("db.port")  // Или из env

	r := gin.Default()

	r.POST("/create", proxyToDB("/create"))
	r.GET("/list", proxyToDB("/list"))
	r.DELETE("/delete", proxyToDB("/delete"))
	r.PUT("/done", proxyToDB("/done"))

	port := viper.GetString("api.port")
	log.Fatal(r.Run(":" + port))
}

func proxyToDB(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := dbURL + path

		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		req.Header = c.Request.Header.Clone()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Set(k, v)
			}
		}

		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}