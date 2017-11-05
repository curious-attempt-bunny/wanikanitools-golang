package main

type Progress struct {
    Progresses []*ProgressType `json:"progresses"`
}

type ProgressType struct {
    Level           int     `json:"level"`
    Type            string  `json:"type"`
    SrsLevelTotals  []int   `json:"srs_level_totals"`
    GuruedTotal     int     `json:"gurued_total"`
    Max             int     `json:"max"`
}