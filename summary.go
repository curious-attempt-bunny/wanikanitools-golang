package main

import "encoding/json"
import "log"

type Summary struct {
    Data struct {
        LessonSubjectIds []int       `json:"lesson_subject_ids"`
        ReviewSubjectIds []int `json:"review_subject_ids"`
        ReviewsPerHour   []struct {
            AvailableAt string  `json:"available_at"`
            SubjectIds  []int `json:"subject_ids"`
        } `json:"reviews_per_hour"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    URL           string `json:"url"`
}

func getSummary(apiKey string, chResult chan *Summary) {
    body := getUrl(apiKey, "https://wanikani.com/api/v2/summary")
    var results Summary
    
    err := json.Unmarshal(body, &results)
    if err != nil {
        log.Fatal("error:", err, string(body))
        panic(err)
    }

    chResult <- &results
}