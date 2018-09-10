package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/url"

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
    var results *Assignments

    apiEntity := "assignments"
    uri := "https://api.wanikani.com/v2/"+apiEntity
    cacheFile := fmt.Sprintf("%s/%s_%s.json", GetCacheDir(), apiKey, apiEntity)
    raw, err := ioutil.ReadFile(cacheFile)
    if (err != nil) {
        // cache miss
        results, err = getAssignmentsPage(apiKey, uri)
        if err != nil {
            chResult <- &Assignments{Error: err.Error()}
            return
        }    
    } else {
        // cache hit
        err = json.Unmarshal(raw, &results)
        if err != nil {
            chResult <- &Assignments{Error: err.Error()}
            return
        }
        v := url.Values{}
        v.Set("updated_after", results.DataUpdatedAt)
        results.Pages.NextURL = uri+"?"+v.Encode()
    }

    itemDataMap := make(map[int]AssignmentsData)
    for _, item := range results.Data {
        itemDataMap[item.Data.SubjectID] = item
    }

    lastResult := results
    for len(lastResult.Pages.NextURL) > 0 {
        lastResult, err = getAssignmentsPage(apiKey, lastResult.Pages.NextURL)
        if err != nil {
            chResult <- &Assignments{Error: err.Error()}
            return
        }

        for _, item := range lastResult.Data {
            itemDataMap[item.Data.SubjectID] = item
        }
    }

    results.Data = make([]AssignmentsData, 0, len(itemDataMap))
    for _, item := range itemDataMap {
        results.Data = append(results.Data, item)
    }
    
    chResult <- results

    raw, err = json.MarshalIndent(results, "", "  ")
    if (err != nil) {
        fmt.Printf("Error marshalling %s cache: %s\n", apiEntity, err.Error())
        return
    }

    err = ioutil.WriteFile(cacheFile, raw, 0644)
    if (err != nil) {
        fmt.Printf("Error writing %s cache: %s\n", apiEntity, err.Error())
        return
    }
}

func getAssignmentsPage(apiKey string, pageUrl string) (*Assignments, error) {
    body, err := getUrl(apiKey, pageUrl)
    if err != nil {
        return nil, err
    }

    var results Assignments
    
    err = json.Unmarshal(body, &results)
    if err != nil {
        return nil, err
    }

    return &results, nil
}
