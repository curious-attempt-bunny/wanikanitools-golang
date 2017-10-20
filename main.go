package main

import (
	"log"
	"net/http"
	"os"
    "io/ioutil"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.GET("/api/v2/subjects", func(c *gin.Context) {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", "https://wanikani.com/api/v2/subjects", nil)
		req.Header.Add("Authorization", "Token token=" + os.Getenv("WANIKANI_V2_API_KEY"))
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		c.String(http.StatusOK, string(body))
	})

	router.Run(":" + port)
}
