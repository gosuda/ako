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

var libraryList = []Library{
	{
		ID:          "1",
		Name:        "rs/zerolog",
		URL:         "https://github.com/rs/zerolog",
		Tags:        []string{"logger", "golang", "zero allocation", "performance"},
		Description: "rs/zerolog is a fast and simple logger for Go. It is designed to be zero allocation and high performance, making it ideal for high-throughput applications.",
	},
}
