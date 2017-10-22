package main

import (
	"log"
	"os"
    "fmt"
    "net/http"
    "math"
    "sort"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgin/v1"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()

	if len(os.Getenv("NEW_RELIC_APP_NAME")) > 0 && len(os.Getenv("NEW_RELIC_LICENSE_KEY")) > 0 {
		config := newrelic.NewConfig(os.Getenv("NEW_RELIC_APP_NAME"), os.Getenv("NEW_RELIC_LICENSE_KEY"))
		app, err := newrelic.NewApplication(config)
		if (err != nil) {
			panic(err)
		}
		router.Use(nrgin.Middleware(app))
	}

	router.Use(CORSMiddleware())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	withApiKey := router.Group("/")
	withApiKey.Use(ApiKeyAuth())
	{
		withApiKey.GET("/api/v2/subjects", func(c *gin.Context) {
			apiKey := c.MustGet("apiKey").(string)
			ch := make(chan *Subjects)
			go getSubjects(apiKey, ch)
			subjects := <-ch
			
			fmt.Printf("%-v\n", subjectsDataMap[19])
			fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
			fmt.Printf("data has length %d\n", len(subjects.Data))
			c.JSON(200, subjects)
		})

		withApiKey.GET("/srs/status", func(c *gin.Context) {
			apiKey := c.MustGet("apiKey").(string)

			chReviewStatistics := make(chan *ReviewStatistics)
			go getReviewStatistics(apiKey, chReviewStatistics)

			chAssignments := make(chan *Assignments)
			go getAssignments(apiKey, chAssignments)

			chSummary := make(chan *Summary)
			go getSummary(apiKey, chSummary)
			
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
			reviewStatistics := <-chReviewStatistics

			dashboard := Dashboard{}
			dashboard.Levels.Order = []string{ "apprentice", "guru", "master", "enlightened", "burned" }

			leeches := make(LeechList, 0)
			
			srsLevelTotals := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	        leechTotals := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	        
	        for i := 0; i<len(assignments.Data); i++ {
	        	srsLevelTotals[assignments.Data[i].Data.SrsStage] += 1
	        }

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

				subject, isSubjectCached := subjectsDataMap[reviewStatistic.Data.SubjectID]
				if !isSubjectCached {
					fmt.Printf("Cache miss for subject ID %d - reloading\n", reviewStatistic.Data.SubjectID)
					chSubjects := make(chan *Subjects)
					go getSubjects(apiKey, chSubjects)
					<-chSubjects
					subject, isSubjectCached = subjectsDataMap[reviewStatistic.Data.SubjectID]
					if !isSubjectCached {
						fmt.Printf("Double cache miss for subject ID %d - skipping\n", reviewStatistic.Data.SubjectID)
						continue
					}
				}

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
				leechTotals[leech.SrsStage] += 1
				// fmt.Printf("%-v\n", leech)
			}

			sort.Sort(leeches)
			retainedLeeches := 10
			if retainedLeeches > len(leeches) {
				retainedLeeches = len(leeches)
			}

			dashboard.ReviewOrder = leeches[0:retainedLeeches]
			dashboard.LeechesTotal = len(leeches)
			dashboard.SrsLevelTotals = srsLevelTotals
	        dashboard.SrsLevelLeechesTotals = leechTotals

			dashboard.Levels.Apprentice. SrsLevelTotals = srsLevelTotals[1:5]
			dashboard.Levels.Guru.       SrsLevelTotals = srsLevelTotals[5:7]
			dashboard.Levels.Master.     SrsLevelTotals = srsLevelTotals[7:8]
			dashboard.Levels.Enlightened.SrsLevelTotals = srsLevelTotals[8:9]
			dashboard.Levels.Burned.     SrsLevelTotals = srsLevelTotals[9:10]

			dashboard.Levels.Apprentice. Total = srsLevelTotals[1] + srsLevelTotals[2] + srsLevelTotals[3] + srsLevelTotals[4]
			dashboard.Levels.Guru.       Total = srsLevelTotals[5] + srsLevelTotals[6]
			dashboard.Levels.Master.     Total = srsLevelTotals[7]
			dashboard.Levels.Enlightened.Total = srsLevelTotals[8]
			dashboard.Levels.Burned.     Total = srsLevelTotals[9]

			dashboard.Levels.Apprentice. SrsLevelLeechesTotals = leechTotals[1:5]
			dashboard.Levels.Guru.       SrsLevelLeechesTotals = leechTotals[5:7]
			dashboard.Levels.Master.     SrsLevelLeechesTotals = leechTotals[7:8]
			dashboard.Levels.Enlightened.SrsLevelLeechesTotals = leechTotals[8:9]
			dashboard.Levels.Burned.     SrsLevelLeechesTotals = leechTotals[9:10]


			dashboard.Levels.Apprentice. LeechesTotal = leechTotals[1] + leechTotals[2] + leechTotals[3] + leechTotals[4]
			dashboard.Levels.Guru.       LeechesTotal = leechTotals[5] + leechTotals[6]
			dashboard.Levels.Master.     LeechesTotal = leechTotals[7]
			dashboard.Levels.Enlightened.LeechesTotal = leechTotals[8]
			dashboard.Levels.Burned.     LeechesTotal = leechTotals[9]

			c.JSON(200, dashboard)
		})
	}

	router.Run(":" + port)
}

type LeechList []Leech

func (p LeechList) Len() int { return len(p) }
func (p LeechList) Less(i, j int) bool { return p[i].ReviewOrder < p[j].ReviewOrder }
func (p LeechList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
