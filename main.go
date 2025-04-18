package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/snowmerak/hasu/internal/ollama"
	"github.com/snowmerak/hasu/internal/prompts"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
	fmt.Printf("Suggested Technologies: %+v", suggestedTechnologies)
}
