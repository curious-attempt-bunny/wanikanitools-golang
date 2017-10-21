package main

import (
	"log"
	"os"
    "fmt"
    "net/http"

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
		ch := make(chan *Subjects)
		go getSubjects(ch)
		subjects := <-ch
		
		fmt.Printf("%-v\n", subjectsDataMap[19])
		fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
		fmt.Printf("data has length %d\n", len(subjects.Data))
		c.JSON(200, subjects)
	})

	router.GET("/srs/status", func(c *gin.Context) {
		chSubjects := make(chan *Subjects)
		go getSubjects(chSubjects)

		chReviewStatistics := make(chan *ReviewStatistics)
		go getReviewStatistics(chReviewStatistics)

		chAssignments := make(chan *Assignments)
		go getAssignments(chAssignments)
		
		<-chSubjects
		reviewStatistics := <-chReviewStatistics
		<-chAssignments

		// get assignments
		// get review statistics

		// iterate review statistics
		// - exclude burned_at assignments
		// - exclude not-yet-passed assignments
		// - calculate scores
		// - calculate worst score
		// - remove worst score < 1.0
		// - determine readings and meanings

		// if subjects == nil {
		// 	subjects = getSubjects()
		// 	for i := 0; i<len(subjects.Data); i++ {
		// 		subjectsDataMap[subjects.Data[i].ID] = subjects.Data[i]
		// 	}
		// }

		// fmt.Printf("%-v\n", subjectsDataMap[19])
		// fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
		// fmt.Printf("data has length %d\n", len(subjects.Data))
		c.JSON(200, reviewStatistics)
	})

	router.Run(":" + port)
}
