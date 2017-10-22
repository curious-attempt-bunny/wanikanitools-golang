package main

import "encoding/json"
import "log"
import "fmt"

var apiKeyReviewStatisticPageCounts map[string]int = make(map[string]int)

type ReviewStatistics struct {
    Data []ReviewStatisticsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int    `json:"total_count"`
    URL        string `json:"url"`
}

type ReviewStatisticsData struct {
    Data struct {
        CreatedAt            string `json:"created_at"`
        MeaningCorrect       int    `json:"meaning_correct"`
        MeaningCurrentStreak int    `json:"meaning_current_streak"`
        MeaningIncorrect     int    `json:"meaning_incorrect"`
        MeaningMaxStreak     int    `json:"meaning_max_streak"`
        PercentageCorrect    int    `json:"percentage_correct"`
        ReadingCorrect       int    `json:"reading_correct"`
        ReadingCurrentStreak int    `json:"reading_current_streak"`
        ReadingIncorrect     int    `json:"reading_incorrect"`
        ReadingMaxStreak     int    `json:"reading_max_streak"`
        SubjectID            int    `json:"subject_id"`
        SubjectType          string `json:"subject_type"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    ID            int    `json:"id"`
    Object        string `json:"object"`
    URL           string `json:"url"`
}

func getReviewStatistics(apiKey string, chResult chan *ReviewStatistics) {
    ch := make(chan *ReviewStatistics)
    maxPages, isApiKeyPageCountPresent := apiKeyReviewStatisticPageCounts[apiKey]
    if !isApiKeyPageCountPresent {
        maxPages = 1
    }
    fmt.Printf("getReviewStatistics assuming maxPages = %d\n", maxPages)

    for page := 1; page <= maxPages; page++ {
        go getReviewStatisticsPage(apiKey, page, ch)
    }
    
    results := <-ch
    if (results.Pages.Last) > maxPages {
        apiKeyReviewStatisticPageCounts[apiKey] = results.Pages.Last
        for page := maxPages+1; page <= results.Pages.Last; page++ {
            go getReviewStatisticsPage(apiKey, page, ch)
        }
        maxPages = results.Pages.Last
    }

    for page := 2; page <= maxPages; page++ {
        resultsPage := <-ch
        results.Data = append(results.Data, resultsPage.Data...)
    }

    results.Pages.Current = 1

    chResult <- results
}


func getReviewStatisticsPage(apiKey string, page int, ch chan *ReviewStatistics) {
    body := getUrl(apiKey, fmt.Sprintf("https://wanikani.com/api/v2/review_statistics?page=%d",page))
    var results ReviewStatistics
    
    err := json.Unmarshal(body, &results)
    if err != nil {
        log.Fatal("error:", err, string(body))
        panic(err)        
    }

    ch <- &results
}