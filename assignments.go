package main

type Assignments struct {
    Data []struct {
        Data struct {
            AvailableAt  string      `json:"available_at"`
            BurnedAt     interface{} `json:"burned_at"`
            Level        int       `json:"level"`
            Passed       bool        `json:"passed"`
            PassedAt     string      `json:"passed_at"`
            Resurrected  bool        `json:"resurrected"`
            SrsStage     int       `json:"srs_stage"`
            SrsStageName string      `json:"srs_stage_name"`
            StartedAt    string      `json:"started_at"`
            SubjectID    int       `json:"subject_id"`
            SubjectType  string      `json:"subject_type"`
            UnlockedAt   string      `json:"unlocked_at"`
        } `json:"data"`
        DataUpdatedAt string `json:"data_updated_at"`
        ID            int  `json:"id"`
        Object        string `json:"object"`
        URL           string `json:"url"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int  `json:"total_count"`
    URL        string `json:"url"`
}
