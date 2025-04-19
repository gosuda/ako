package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/snowmerak/hasu/internal/document"
	"github.com/snowmerak/hasu/internal/meilisearch"
	"github.com/snowmerak/hasu/internal/ollama"
	"github.com/snowmerak/hasu/internal/prompts"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	_ = ctx

	mc := meilisearch.New(meilisearch.NewConfig().WithApiKey(meilisearch.MasterKey))
	if err := meilisearch.Insert[document.Library](mc, document.IndexNameLibraries, document.GetDefaultLibraries()...); err != nil {
		panic(err)
	}

	cli, err := ollama.New(ollama.Config{
		Model: ollama.ModelGemma3p27B,
	})
	if err != nil {
		panic(err)
	}

	fmt.Print("Enter your package description: ")
	var userInput string
	fmt.Scanln(&userInput)
	prompt := prompts.GetGoPackageGenerationPrompt(userInput)

	resp, err := cli.Generate(ctx, prompt)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", resp)

	// Extract suggested technologies
	suggestedTechnologies := prompts.ExtractSuggestedTechnologies(resp)
	for _, tech := range suggestedTechnologies {
		tech = strings.ReplaceAll(tech, "\"", "")
		tech = strings.ReplaceAll(tech, "'", "")
		tech = strings.ReplaceAll(tech, "`", "")
		tech = strings.ReplaceAll(tech, "*", "")
		log.Printf("Searching for libraries related to technology: %s", tech)

		sr, err := meilisearch.Search[document.Library](mc, document.IndexNameLibraries, tech, meilisearch.SearchOption{
			Limit:  5,
			Offset: 0,
		})
		if err != nil {
			log.Printf("Error searching technologies: %v", err)
			continue
		}

		if len(sr.Hits) == 0 {
			log.Printf("No libraries found for technology: %s", tech)
			continue
		}

		fmt.Printf("Libraries related to technology '%s':\n", tech)
		for _, lib := range sr.Hits {
			fmt.Printf("Library ID: %s, Name: %s, URL: %s, Tags: %v, Description: %s\n", lib.ID, lib.Name, lib.URL, lib.Tags, lib.Description)
		}
	}
}
