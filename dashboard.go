package main

type Dashboard struct {
    LeechesTotal int `json:"leeches_total"`
    Levels       struct {
        Order     []string `json:"order"`
        Apprentice Level `json:"apprentice"`
        Guru Level `json:"guru"`
        Master Level `json:"master"`
        Enlightened Level `json:"enlightened"`
        Burned Level `json:"burned"`
    } `json:"levels"`
    ReviewOrder []Leech `json:"review_order"`
    SrsLevelLeechesTotals []int `json:"srs_level_leeches_totals"`
    SrsLevelTotals        []int `json:"srs_level_totals"`
}

type Level struct {
    LeechesTotal          int   `json:"leeches_total"`
    SrsLevelLeechesTotals []int `json:"srs_level_leeches_totals"`
    SrsLevelTotals        []int `json:"srs_level_totals"`
    Total                 int   `json:"total"`
}