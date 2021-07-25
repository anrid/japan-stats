package stats

type Population struct {
	Year        string
	Total       float64
	Prefectures []*Prefecture
}

type Disasters struct {
	Year                 string
	PersonsAffectedTotal float64
	Prefectures          []*Prefecture
}

type Prefecture struct {
	Stat       string
	Name       string
	NameJP     string
	Value      float64
	PctOfTotal float64
}
