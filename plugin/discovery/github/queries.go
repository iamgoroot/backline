package github

type commitQuery struct {
	Repository struct {
		Object struct {
			CommitResourcePath string `graphql:"commitResourcePath"`
		} `graphql:"object(expression: $expression)"`
	} `graphql:"repository(name: $repo, owner: $owner) "`
}

type contentQuery struct {
	Repository struct {
		Object struct {
			CommitResourcePath string `graphql:"commitResourcePath"`
			Tree               struct {
				Entries []struct {
					Extension string `graphql:"extension"`
					Type      string `graphql:"type"`
					Path      string `graphql:"path"`
					Object    struct {
						Blob struct {
							Text string `graphql:"text"`
						} `graphql:"... on Blob"`
					} `graphql:"object"`
				} `graphql:"entries"`
			} `graphql:"... on Tree"`
		} `graphql:"object(expression: $expression)"`
	} `graphql:"repository(name: $repo, owner: $owner) "`
}

type fileContentQuery struct {
	Repository struct {
		Object struct {
			CommitResourcePath string `graphql:"commitResourcePath"`
			Blob               struct {
				Text     string `graphql:"text"`
				IsBinary bool   `graphql:"isBinary"`
			} `graphql:"... on Blob"`
		} `graphql:"object(expression: $expression)"`
	} `graphql:"repository(name: $repo, owner: $owner) "`
}
