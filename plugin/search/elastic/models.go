package elastic

type document struct {
	EntityName string `json:"entityName"`
	Category   string `json:"category"`
	Value      string `json:"value"`
	Link       string `json:"link"`
}
type esSearchQuery struct {
	Source struct {
		Excludes []string `json:"excludes"`
	} `json:"_source"`
	Query struct {
		MultiMatch struct {
			Query     string `json:"query"`
			Fuzziness int    `json:"fuzziness"`
		} `json:"multi_match"`
	} `json:"query"`
	Highlight struct {
		Fields struct {
			Value struct {
				NumberOfFragments int `json:"number_of_fragments"`
			} `json:"value"`
		} `json:"fields"`
	} `json:"highlight"`
	Size int `json:"size"`
	From int `json:"from"`
}

type errResponse struct {
	Error struct {
		Reason string `json:"reason"`
		Type   string `json:"type"`
	} `json:"error"`
}

type deleteQuery struct {
	Query struct {
		Terms map[string][]string `json:"terms"`
	} `json:"query"`
}

type deleteResponse struct {
	Errors  []interface{} `json:"failures"`
	Deleted int           `json:"deleted"`
}
