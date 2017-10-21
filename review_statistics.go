package main

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
