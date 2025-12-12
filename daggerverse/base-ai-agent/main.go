package main

import "dagger/base-ai-agent/internal/dagger"

type BaseAiAgent struct{}

// Write a Go program
func (m *BaseAiAgent) GoProgram(
	// The programming assignment, e.g. "write me a curl clone"
	assignment string,
) *dagger.Container {
	environment := dag.Env().
		WithStringInput("assignment", assignment, "the assignment to complete").
		WithContainerInput(
			"builder",
			dag.Container().From("golang").WithWorkdir("/app"),
			"a container to use for building Go code").
		WithContainerOutput("completed", "the completed assignment in the Golang container")

	work := dag.LLM().
		WithEnv(environment).
		WithPrompt(`
		You are an expert Go programmer with an assignment to create a Go program
		Create files in the default directory in $builder
		Always build the code to make sure it is valid
		Do not stop until your assignment is completed and the code builds
		Your assignment is: $assignment
	`)

	return work.
		Env().
		Output("completed").
		AsContainer().
		Terminal()
}

func (m *BaseAiAgent) GoAdvancedProgram(
	assignment string,
) *dagger.Directory {
	environment := dag.Env().
		WithStringInput("assignment", assignment, "the assignment to complete").
		WithContainerInput(
			"builder",
			dag.Container().From("golang").WithWorkdir("/app"),
			"a container to use for building Go code").
		WithContainerOutput("completed", "the completed assignment in the Golang container")

	prompt := dag.CurrentModule().
		Source().
		File("prompts/create_go_app_all.md")

	work := dag.LLM().
		WithEnv(environment).
		WithPromptFile(prompt)

	return work.
		Env().
		Output("completed").
		AsContainer().Directory("/app")
	// dagger call go-advanced-program --assignment "..." \
	//  export --path=out/go-advanced
}
