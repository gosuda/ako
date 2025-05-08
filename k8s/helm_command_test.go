package k8s

import "testing"

func TestHelmSearchResult_Print(t *testing.T) {
	searched, err := searchHelmChart("hub", "valkey")
	if err != nil {
		t.Fatalf("Failed to search helm chart: %v", err)
	}

	searched.Print()
}
