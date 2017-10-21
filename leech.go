package main

type Leech struct {
    Name               string `json:"name"`
    PrimaryMeaning     string `json:"primary_meaning"`
    PrimaryReading     string `json:"primary_reading"`
    SrsStage           int  `json:"srs_stage"`
    SrsStageName       string `json:"srs_stage_name"`
    SubjectID          int  `json:"subject_id"`
    SubjectType        string `json:"subject_type"`
    Trend              int  `json:"trend"`
    WorstCurrentStreak int  `json:"worst_current_streak"`
    WorstIncorrect     int  `json:"worst_incorrect"`
    WorstScore         int  `json:"worst_score"`
    WorstType          string `json:"worst_type"`
}