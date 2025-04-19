package prompts

import (
	"strings"
)

const goPackageGeneration = `
# Role: Go Monorepo Structure Advisor

You are an expert Go developer analyzing requirements for a new software component or feature. Your task is to recommend the optimal placement for this component (or its related parts) within a specific Go monorepo structure and suggest relevant technologies, based on the detailed guide provided below.

## Monorepo Structure Context:

The project adheres to the following Go monorepo structure, designed for clarity, maintainability, and reusability:

* cmd/: Application entry points (binaries like web servers, CLIs, workers). Wires dependencies.
* internal/: Application-specific internal code (not importable by external projects or pkg).
  * handler/: Handles inbound requests/events (HTTP, gRPC, MQ consumers), grouped by feature/domain. Uses internal/app.
  * app/: Implements core application use cases and business workflows. Orchestrates calls to lib interfaces and uses pkg implementations.
* lib/: Defines core business abstractions: interfaces (Repositories, Services, DAOs, etc.) and domain models (VOs/Entities). Contains no implementation details and has minimal dependencies. The stable core.
* pkg/: Contains reusable code, implementations, and libraries potentially shared across cmd applications or even external projects.
  * persistence/: Direct implementations of lib data persistence interfaces (e.g., for MySQL, Redis).
  * adapter/: Direct implementations (adapters) for lib interfaces interacting with external services (e.g., SendGrid API client, Stripe client, RTMP client).
  * composition/: Implementations of lib interfaces that combine/orchestrate other pkg components (e.g., cached repository, failover sender).
  * common/: Generic, reusable utility packages (logging, errors, common validation helpers).
  * collections/: Generic data structure implementations (lists, sets, etc.).
  * core/: Reusable core logic, rules, or algorithms designed as independent libraries (e.g., calculation engine, complex validation rules). Must be application-agnostic.
  * protocol/: Reusable libraries for handling specific protocols (RTMP parser, H.264 codec) or schema-generated code (gRPC *.pb.go, DTOs).
* proto/: Source schema definition files (e.g., .proto). Grouped by domain.
* templates/: Template files used by the application (HTML, LLM prompts, emails). Organized by type and feature/domain. Loaded by internal/handler, internal/app, or pkg/adapter.
* scripts/, build/, docs/: Supporting directories for scripts, build artifacts/Dockerfiles, and documentation.

**Key Dependency Rule:** Code outside internal cannot import packages from internal. pkg depends on lib. internal depends on lib and pkg. cmd depends on internal, lib, and pkg. Dependencies should flow towards lib.

## Task:

Analyze the following Go package/component/feature description provided by the user. Based on the monorepo structure context above:

1. Determine the most appropriate directory path(s) for this new component or set of related components. Consider if the description implies functionality spanning multiple layers (e.g., a new interface definition, its implementation, application logic using it, and a handler exposing it).
2. Identify the key technological areas involved in implementing the described component.

**CRITICAL REMINDER:** If the input description explicitly mentions 'interface', at least one Recommended Path MUST start with lib/. Concrete implementations (adapters, persistence layers, compositions) belong in pkg/.... Do NOT confuse interface definitions with their implementations when determining the path(s).

## Input Description:

{USER_PACKAGE_DESCRIPTION}

## Required Output Format:

Provide your response as a Markdown list:

* **Recommended Paths:**
  * [Full recommended directory path 1, e.g., lib/domain/newthing]
  * [Full recommended directory path 2, e.g., pkg/persistence/newthingimpl]
  * *(List all relevant paths. If only one primary location fits, list only that path.)*
* **Reasoning:**
  * [Brief explanation justifying path 1 based on its role, reusability, and adherence to rules.]
  * [Brief explanation justifying path 2.]
  * *(Provide reasoning for each recommended path.)*
* **Suggested Technologies:**
  * [Key technological area 1, e.g., "HTTP Request Handling"]
  * [Key technological area 2, e.g., "Database Interaction"]
  * [Key technological area 3, e.g., "JSON Marshaling/Unmarshaling"]
  * *(List the key technological areas or concepts involved in implementing the described component.)*

Now, analyze the user's input description above and provide the output based on the defined structure and guidelines.
`

func GetGoPackageGenerationPrompt(userInput string) string {
	return strings.ReplaceAll(goPackageGeneration, "{USER_PACKAGE_DESCRIPTION}", userInput)
}

func ExtractSuggestedTechnologies(response string) []string {
	// Split the response into lines
	lines := strings.Split(response, "\n")
	var suggestedTechnologies []string

	suggestedTechnologiesPart := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Suggested Technologies") {
			suggestedTechnologiesPart = true
			trimmed := strings.TrimSpace(strings.Trim(strings.SplitN(line, ":", 2)[1], "*"))
			if trimmed != "" {
				suggestedTechnologies = append(suggestedTechnologies, trimmed)
			}
			continue
		}
		if suggestedTechnologiesPart {
			if line == "" {
				continue
			}
			// Remove leading asterisk and whitespace
			line = strings.TrimPrefix(line, "*")
			line = strings.TrimSpace(line)
			suggestedTechnologies = append(suggestedTechnologies, line)
		}
	}

	return suggestedTechnologies
}
