package main

type Progress struct {
    Vocabulary ProgressType `json:"vocabulary"`
    Radical    ProgressType `json:"radical"`
    Kanji      ProgressType `json:"kanji"`
}

type ProgressType struct {
    Level           int     `json:"level"`
    SrsLevelTotals  []int   `json:"srs_level_totals"`
    GuruedTotal     int     `json:"gurued_total"`
    Max             int     `json:"max"`
}