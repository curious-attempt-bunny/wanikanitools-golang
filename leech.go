package main

import "bufio"
import "fmt"

// import "io"
// import "io/ioutil"
import "log"
import "math"
import "os"
import "strings"

type Leech struct {
	Name               string   `json:"name"`
	Level              int      `json:"level"`
	PrimaryMeaning     string   `json:"primary_meaning"`
	Meanings           []string `json:"meanings"`
	PrimaryReading     string   `json:"primary_reading"`
	Readings           []string `json:"readings"`
	SrsStage           int      `json:"srs_stage"`
	SrsStageName       string   `json:"srs_stage_name"`
	SubjectID          int      `json:"subject_id"`
	SubjectType        string   `json:"subject_type"`
	WorstCurrentStreak int      `json:"worst_current_streak"`
	WorstIncorrect     int      `json:"worst_incorrect"`
	WorstScore         float64  `json:"worst_score"`
	WorstType          string   `json:"worst_type"`
	ReviewOrder        int      `json:"-"`
}

type LeechLesson struct {
	LeechesAvailable int               `json:"leeches_available"`
	TooRecentCount   int               `json:"too_recent_count"`
	TooSoonCount     int               `json:"too_soon_count"`
	TrainedCount     int               `json:"trained_count"`
	LeechLessonItems []LeechLessonItem `json:"leech_lesson_items"`
}

type LeechLessonItem struct {
	Name            string     `json:"name"`
	SubjectType     string     `json:"type"`
	TrainType       string     `json:"train_type"`
	CorrectAnswers  []string   `json:"correct_answers"`
	TryAgainAnswers []string   `json:"try_again_answers"`
	Leech           LeechTrain `json:"leech"`
}

type LeechTrain struct {
	Key            string `json:"key"`
	WorstIncorrect int    `json:"worst_incorrect"`
}

var similar map[string][]string

func init() {
	similar = make(map[string][]string)

	file, err := os.Open("static/similar.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		// fmt.Printf("%q\n", parts)
		similar[parts[0]] = parts[1:]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getLeeches(apiKey string) (LeechList, *ReviewStatistics, *Assignments, *ResourceError) {
	leeches := make(LeechList, 0)

	chReviewStatistics := make(chan *ReviewStatistics)
	go getReviewStatistics(apiKey, chReviewStatistics)

	chAssignments := make(chan *Assignments)
	go getAssignments(apiKey, chAssignments)

	chSummary := make(chan *Summary)
	go getSummary(apiKey, chSummary)

	summary := <-chSummary
	if len(summary.Error) > 0 {
		return nil, nil, nil, &ResourceError{Category: "summary", ErrorMessage: summary.Error}
	}

	subjectReviewOrder := make(map[int]int)
	for i := 0; i < len(summary.Data.ReviewsPerHour); i++ {
		reviewsPerHour := summary.Data.ReviewsPerHour[i]
		for j := 0; j < len(reviewsPerHour.SubjectIds); j++ {
			subjectReviewOrder[reviewsPerHour.SubjectIds[j]] = i
		}
	}

	assignments := <-chAssignments
	if len(assignments.Error) > 0 {
		return nil, nil, nil, &ResourceError{Category: "assignments", ErrorMessage: assignments.Error}
	}

	assignmentsDataMap := make(map[int]AssignmentsData)
	for i := 0; i < len(assignments.Data); i++ {
		assignmentsDataMap[assignments.Data[i].Data.SubjectID] = assignments.Data[i]
	}

	reviewStatistics := <-chReviewStatistics
	if len(reviewStatistics.Error) > 0 {
		return nil, nil, nil, &ResourceError{Category: "reviewStatistics", ErrorMessage: reviewStatistics.Error}
	}

	for i := 0; i < len(reviewStatistics.Data); i++ {
		reviewStatistic := reviewStatistics.Data[i]
		if reviewStatistic.Data.SubjectType == "radical" {
			continue
		}
		if reviewStatistic.Data.MeaningIncorrect+reviewStatistic.Data.MeaningCorrect == 0 {
			continue
		}
		if reviewStatistic.Data.MeaningCorrect < 4 {
			// has not yet made it to Guru (approximate)
			continue
		}

		meaningScore := float64(reviewStatistic.Data.MeaningIncorrect) / math.Pow(float64(reviewStatistic.Data.MeaningCurrentStreak), 1.5)
		readingScore := float64(reviewStatistic.Data.ReadingIncorrect) / math.Pow(float64(reviewStatistic.Data.ReadingCurrentStreak), 1.5)

		if meaningScore < 1.0 && readingScore < 1.0 {
			continue
		}

		assignment := assignmentsDataMap[reviewStatistic.Data.SubjectID]

		if len(assignment.Data.BurnedAt) > 0 {
			continue
		}

		subject, isSubjectCached := subjectsDataMap[reviewStatistic.Data.SubjectID]
		if !isSubjectCached {
			fmt.Printf("Cache miss for subject ID %d - reloading\n", reviewStatistic.Data.SubjectID)
			chSubjects := make(chan *Subjects)
			go getSubjects(apiKey, chSubjects)
			subjects := <-chSubjects
			if len(subjects.Error) > 0 {
				return nil, nil, nil, &ResourceError{Category: "subjects", ErrorMessage: subjects.Error}
			}

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

		leech.Level = subject.Data.Level

		leech.Meanings = make([]string, len(subject.Data.Meanings))
		for j := 0; j < len(subject.Data.Meanings); j++ {
			leech.Meanings[j] = subject.Data.Meanings[j].Meaning
			if subject.Data.Meanings[j].Primary {
				leech.PrimaryMeaning = subject.Data.Meanings[j].Meaning
			}
		}

		leech.Readings = make([]string, len(subject.Data.Readings))
		for j := 0; j < len(subject.Data.Readings); j++ {
			leech.Readings[j] = subject.Data.Readings[j].Reading
			if subject.Data.Readings[j].Primary {
				leech.PrimaryReading = subject.Data.Readings[j].Reading
			}
		}

		leech.SrsStage = assignment.Data.SrsStage
		leech.SrsStageName = assignment.Data.SrsStageName

		if meaningScore > readingScore {
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
		// fmt.Printf("%-v\n", leech)
	}

	return leeches, reviewStatistics, assignments, nil
}
