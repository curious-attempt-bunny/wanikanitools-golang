package main

type Progress struct {
	Progresses  []*ProgressType        `json:"progresses"`
	StageLevels map[string]*StageLevel `json:"stage_levels"`
}

type ProgressType struct {
	Level          int    `json:"level"`
	Type           string `json:"type"`
	SrsLevelTotals []int  `json:"srs_level_totals"`
	GuruedTotal    int    `json:"gurued_total"`
	Max            int    `json:"max"`
}

type StageLevel struct {
	Level               int     `json:"level"`
	PercentageNextLevel float64 `json:"percentage_next_level"`
}
