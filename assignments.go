package main

type Assignments struct {
    Data []struct {
        Data struct {
            AvailableAt  string      `json:"available_at"`
            BurnedAt     interface{} `json:"burned_at"`
            Level        int64       `json:"level"`
            Passed       bool        `json:"passed"`
            PassedAt     string      `json:"passed_at"`
            Resurrected  bool        `json:"resurrected"`
            SrsStage     int64       `json:"srs_stage"`
            SrsStageName string      `json:"srs_stage_name"`
            StartedAt    string      `json:"started_at"`
            SubjectID    int64       `json:"subject_id"`
            SubjectType  string      `json:"subject_type"`
            UnlockedAt   string      `json:"unlocked_at"`
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
