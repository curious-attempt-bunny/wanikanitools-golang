package main

import (
	"log"
	"os"
    "fmt"
    "net/http"
    "math"
    "sort"

	"github.com/gin-gonic/gin"
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

		chSummary := make(chan *Summary)
		go getSummary(chSummary)
		
		summary := <-chSummary
		subjectReviewOrder := make(map[int]int)
		for i := 0; i<len(summary.Data.ReviewsPerHour); i++ {
			reviewsPerHour := summary.Data.ReviewsPerHour[i]
			for j := 0; j<len(reviewsPerHour.SubjectIds); j++ {
				subjectReviewOrder[reviewsPerHour.SubjectIds[j]] = i
			}
		}

		assignments := <-chAssignments
		assignmentsDataMap := make(map[int]AssignmentsData)
		for i := 0; i<len(assignments.Data); i++ {
	        assignmentsDataMap[assignments.Data[i].Data.SubjectID] = assignments.Data[i]
	    }
		<-chSubjects
		reviewStatistics := <-chReviewStatistics

		dashboard := Dashboard{}
		dashboard.Levels.Order = []string{ "apprentice", "guru", "master", "enlightened", "burned" }

		leeches := make(LeechList, 0)
		
		for i := 0; i<len(reviewStatistics.Data); i++ {
			reviewStatistic := reviewStatistics.Data[i]
			if reviewStatistic.Data.SubjectType == "radical" {
				continue
			}
			if (reviewStatistic.Data.MeaningIncorrect + reviewStatistic.Data.MeaningCorrect == 0) {
				continue
			}
			if (reviewStatistic.Data.MeaningCorrect < 4) {
				// has not yet made it to Guru (approximate)
				continue;
			}

            meaningScore := float64(reviewStatistic.Data.MeaningIncorrect) / math.Pow(float64(reviewStatistic.Data.MeaningCurrentStreak), 1.5)
            readingScore := float64(reviewStatistic.Data.ReadingIncorrect) / math.Pow(float64(reviewStatistic.Data.ReadingCurrentStreak), 1.5)
            
            if (meaningScore < 1.0 && readingScore < 1.0) {
            	continue;
            }

			assignment := assignmentsDataMap[reviewStatistic.Data.SubjectID]

			if (len(assignment.Data.BurnedAt) > 0) {
				continue;
			}

			subject := subjectsDataMap[reviewStatistic.Data.SubjectID]

			leech := Leech{}

			if len(subject.Data.Character) > 0 {
				leech.Name = subject.Data.Character 
			} else {
				leech.Name = subject.Data.Characters
			}

			for j := 0; j<len(subject.Data.Meanings); j++ {
				if (subject.Data.Meanings[j].Primary) {
					leech.PrimaryMeaning = subject.Data.Meanings[j].Meaning
					break
				}
			}

			for j := 0; j<len(subject.Data.Readings); j++ {
				if (subject.Data.Readings[j].Primary) {
					leech.PrimaryReading = subject.Data.Readings[j].Reading
					break
				}
			}

			leech.SrsStage = assignment.Data.SrsStage			
			leech.SrsStageName = assignment.Data.SrsStageName

			if (meaningScore > readingScore) {
				leech.WorstType = "meaning"
				leech.WorstScore = meaningScore
				leech.WorstCurrentStreak = reviewStatistic.Data.MeaningCurrentStreak
				leech.WorstIncorrect = reviewStatistic.Data.MeaningIncorrect
			} else {
				leech.WorstType = "reading"
				leech.WorstScore = readingScore
				leech.WorstCurrentStreak = reviewStatistic.Data.ReadingCurrentStreak
				leech.WorstIncorrect = reviewStatistic.Data.ReadingIncorrect
			}

    		leech.SubjectID = subject.ID
			leech.SubjectType = subject.Object

			var isComingUpForReview bool
			leech.ReviewOrder, isComingUpForReview = subjectReviewOrder[leech.SubjectID]
			if !isComingUpForReview {
				leech.ReviewOrder = 1000
			}
			leeches = append(leeches, leech)
			fmt.Printf("%-v\n", leech)
		}

		sort.Sort(leeches)
		retainedLeeches := 10
		if retainedLeeches > len(leeches) {
			retainedLeeches = len(leeches)
		}

		dashboard.ReviewOrder = leeches[0:retainedLeeches]

		c.JSON(200, dashboard)
	})

	router.Run(":" + port)
}

type LeechList []Leech

func (p LeechList) Len() int { return len(p) }
func (p LeechList) Less(i, j int) bool { return p[i].ReviewOrder < p[j].ReviewOrder }
func (p LeechList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
