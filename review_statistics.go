package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/url"

type ReviewStatistics struct {
    Data []ReviewStatisticsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int    `json:"total_count"`
    URL        string `json:"url"`
    Error      string `json:"-"`
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
    var results *ReviewStatistics

    apiEntity := "review_statistics"
    uri := "https://api.wanikani.com/v2/"+apiEntity
    cacheFile := fmt.Sprintf("%s/%s_%s.json", GetCacheDir(), apiKey, apiEntity)
    raw, err := ioutil.ReadFile(cacheFile)
    if (err != nil) {
        // cache miss
        results, err = getReviewStatisticsPage(apiKey, uri)
        if err != nil {
            chResult <- &ReviewStatistics{Error: err.Error()}
            return
        }    
    } else {
        // cache hit
        err = json.Unmarshal(raw, &results)
        if err != nil {
            chResult <- &ReviewStatistics{Error: err.Error()}
            return
        }
        v := url.Values{}
        v.Set("updated_after", results.DataUpdatedAt)
        results.Pages.NextURL = uri+"?"+v.Encode()
    }

    itemDataMap := make(map[int]ReviewStatisticsData)
    for _, item := range results.Data {
        itemDataMap[item.Data.SubjectID] = item
    }

    lastResult := results
    for len(lastResult.Pages.NextURL) > 0 {
        lastResult, err = getReviewStatisticsPage(apiKey, lastResult.Pages.NextURL)
        if err != nil {
            chResult <- &ReviewStatistics{Error: err.Error()}
            return
        }

        for _, item := range lastResult.Data {
            itemDataMap[item.Data.SubjectID] = item
        }
    }

    results.Data = make([]ReviewStatisticsData, 0, len(itemDataMap))
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

func getReviewStatisticsPage(apiKey string, pageUrl string) (*ReviewStatistics, error) {
    body, err := getUrl(apiKey, pageUrl)
    if err != nil {
        return nil, err
    }

    var results ReviewStatistics
    
    err = json.Unmarshal(body, &results)
    if err != nil {
        return nil, err
    }

    return &results, nil
}
