package document

const IndexNameLibraries = "libraries"

type Library struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

func GetDefaultLibraries() []Library {
	return libraryList
}
