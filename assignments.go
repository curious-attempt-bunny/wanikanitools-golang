package main

import "encoding/json"
import "fmt"

var apiKeyAssignmentsPageCounts map[string]int = make(map[string]int)

type Assignments struct {
    Data []AssignmentsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int  `json:"total_count"`
    URL        string `json:"url"`
    Error      string `json:"-"`
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

func getAssignments(apiKey string, chResult chan *Assignments) {
    ch := make(chan *Assignments)
    maxPages, isApiKeyPageCountPresent := apiKeyAssignmentsPageCounts[apiKey]
    if !isApiKeyPageCountPresent {
        maxPages = 1
    }
    fmt.Printf("getAssignments assuming maxPages = %d\n", maxPages)

    for page := 1; page <= maxPages; page++ {
        go getAssignmentsPage(apiKey, page, ch)
    }
    
    results := <-ch
    if len(results.Error) > 0 {
        chResult <- results
        return
    }

    if (results.Pages.Last > maxPages) {
        apiKeyAssignmentsPageCounts[apiKey] = results.Pages.Last

        for page := maxPages+1; page <= results.Pages.Last; page++ {
            go getAssignmentsPage(apiKey, page, ch)
        }
        maxPages = results.Pages.Last
    }

    for page := 2; page <= maxPages; page++ {
        resultsPage := <-ch
        if len(resultsPage.Error) > 0 {
            results.Error = resultsPage.Error
            chResult <- results
            return
        }

        results.Data = append(results.Data, resultsPage.Data...)
    }

    results.Pages.Current = 1

    chResult <- results
}


func getAssignmentsPage(apiKey string, page int, ch chan *Assignments) {
    body, err := getUrl(apiKey, fmt.Sprintf("https://wanikani.com/api/v2/assignments?page=%d",page))
    if err != nil {
        ch <- &Assignments{Error: err.Error()}
        return
    }

    var results Assignments
    
    err = json.Unmarshal(body, &results)
    if err != nil {
        ch <- &Assignments{Error: err.Error()}
        return
    }

    ch <- &results
}
