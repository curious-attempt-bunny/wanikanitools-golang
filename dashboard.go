package main

type Dashboard struct {
    LeechesTotal int `json:"leeches_total"`
    Levels       struct {
        Order     []string `json:"order"`
        Unstarted Level `json:"unstarted"`
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
    SrsLevelLeechTotal    int   `json:"srs_level_leech_total"`
    SrsLevelLeechTrends   []int `json:"srs_level_leech_trends"`
    SrsLevelLeechesTotals []int `json:"srs_level_leeches_totals"`
    SrsLevelTotals        []int `json:"srs_level_totals"`
    Total                 int   `json:"total"`
}