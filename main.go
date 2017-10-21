package main

import (
	"log"
	"net/http"
	"os"
    "io/ioutil"
    "encoding/json"
    "fmt"

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
		body := getUrl("https://wanikani.com/api/v2/subjects")
		var subjects Subjects
		err := json.Unmarshal(body, &subjects)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Printf("%d subjects pages in total", subjects.Pages.Last)
		c.JSON(200, subjects)
	})

	router.Run(":" + port)
}

func getUrl(url string) []byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Token token=" + os.Getenv("WANIKANI_V2_API_KEY"))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}