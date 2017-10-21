package main

type ReviewStatistics struct {
    Data []struct {
        Data struct {
            CreatedAt            string `json:"created_at"`
            MeaningCorrect       int64  `json:"meaning_correct"`
            MeaningCurrentStreak int64  `json:"meaning_current_streak"`
            MeaningIncorrect     int64  `json:"meaning_incorrect"`
            MeaningMaxStreak     int64  `json:"meaning_max_streak"`
            PercentageCorrect    int64  `json:"percentage_correct"`
            ReadingCorrect       int64  `json:"reading_correct"`
            ReadingCurrentStreak int64  `json:"reading_current_streak"`
            ReadingIncorrect     int64  `json:"reading_incorrect"`
            ReadingMaxStreak     int64  `json:"reading_max_streak"`
            SubjectID            int64  `json:"subject_id"`
            SubjectType          string `json:"subject_type"`
        } `json:"data"`
        DataUpdatedAt string `json:"data_updated_at"`
        ID            int64  `json:"id"`
        Object        string `json:"object"`
        URL           string `json:"url"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         struct {
        Current     int64       `json:"current"`
        Last        int64       `json:"last"`
        NextURL     string      `json:"next_url"`
        PreviousURL interface{} `json:"previous_url"`
    } `json:"pages"`
    TotalCount int64  `json:"total_count"`
    URL        string `json:"url"`
}
