package main

import "encoding/json"
import "log"
import "fmt"

type Assignments struct {
    Data []AssignmentsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int  `json:"total_count"`
    URL        string `json:"url"`
}

type AssignmentsData struct {
    Data struct {
        AvailableAt  string      `json:"available_at"`
        BurnedAt     string      `json:"burned_at"`
        Level        int         `json:"level"`
        Passed       bool        `json:"passed"`
        PassedAt     string      `json:"passed_at"`
        Resurrected  bool        `json:"resurrected"`
        SrsStage     int         `json:"srs_stage"`
        SrsStageName string      `json:"srs_stage_name"`
        StartedAt    string      `json:"started_at"`
        SubjectID    int         `json:"subject_id"`
        SubjectType  string      `json:"subject_type"`
        UnlockedAt   string      `json:"unlocked_at"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    ID            int  `json:"id"`
    Object        string `json:"object"`
    URL           string `json:"url"`
}

func getAssignments(chResult chan *Assignments) {
    ch := make(chan *Assignments)
    maxPages := 1
    for page := 1; page <= maxPages; page++ {
        go getAssignmentsPage(page, ch)
    }
    
    results := <-ch
    if (results.Pages.Last > maxPages) {
        for page := maxPages+1; page <= results.Pages.Last; page++ {
            go getAssignmentsPage(page, ch)
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


func getAssignmentsPage(page int, ch chan *Assignments) {
    body := getUrl(fmt.Sprintf("https://wanikani.com/api/v2/assignments?page=%d",page))
    var results Assignments
    
    err := json.Unmarshal(body, &results)
    if err != nil {
        log.Fatal("error:", err, string(body))
    }

    ch <- &results
}