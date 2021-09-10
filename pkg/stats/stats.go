package stats

type Population struct {
	Country     string
	Year        string
	Total       float64
	Prefectures Prefectures
}

type Prefectures []*Prefecture

func (prs Prefectures) Find(name string) *Prefecture {
	for _, p := range prs {
		if p.Name == name {
			return p
		}
	}
	return nil
}

type Disasters struct {
	Country              string
	Year                 string
	PersonsAffectedTotal float64
	Prefectures          Prefectures
}

type Prefecture struct {
	Name      string
	Statistic string
	Value     float64
}

type DisastersInPrefecture struct {
	Prefecture       string
	AffectedTotal    float64
	AffectedAvg      float64
	YearsTotal       int
	PopulationTotal  float64
	PopulationPct    float64
	OfAllAffectedPct float64
}
