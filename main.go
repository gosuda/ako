package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
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
		Model: ollama.ModelGemma3p12B,
	})
	if err != nil {
		panic(err)
	}

	fmt.Print("Enter your package description: ")
	var userInput string = "로그를 MQ, Rest API로 전송해야해. 필요한 인터페이스 패키지와 kafka, nats, http client에 대한 패키지가 필요해."
	//fmt.Scanln(&userInput)
	prompt := prompts.GetGoPackageGenerationPrompt(userInput)

	resp, err := cli.Generate(ctx, prompt)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", resp)

	// Extract suggested technologies
	suggestedTechnologies := prompts.ExtractSuggestedTechnologies(resp)
	for _, tech := range suggestedTechnologies {
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

		for _, lib := range sr.Hits {
			fmt.Printf("Library ID: %s, Name: %s, URL: %s, Tags: %v, Description: %s\n", lib.ID, lib.Name, lib.URL, lib.Tags, lib.Description)
		}
	}
}
