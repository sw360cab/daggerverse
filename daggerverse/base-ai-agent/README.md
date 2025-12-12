# Basice AI Agent

This is a basic example of an AI Agent run via Dagger.

In a local environment, it leverages Docker Model Runner (DMR).

## Setting up Docker Model Runner

- Pull a model from Docker Hub, via `docker model pull` command
- Setup env variables in a `.env` file to configure DMR in Dagger (see `sample.env`)

```sh
OPENAI_BASE_URL=http://model-runner.docker.internal/engines/v1/
OPENAI_MODEL=index.docker.io/ai/gpt-oss-safeguard:20B
```

### Context caveats

When running a dagger func backed up by DMR LLM, it may happen that the call ends up in a crash.
This is often due to `context_size` overflow. Basically the LLM used in Dagger is demanding more
_Input Tokens_ when running in its loop than the ones available from the LLM model in use.
That is a protection mechansim to avoid a prompt to consume all the HW resources of the machine running the LLM.

In the case of DMR, this values is shipped with the model itself and cannot be changed directly.
As today, the only method is by running an explicit compose file, configureing the model (see [docker-compose.yaml](./docker-compose.yaml)) with an higer `context_size` value.

In the example here, the LLM model used locally, had a default context size of `4096` which was not able to handle the request in the prompt.

## See Also

- [LLM Providers | Dagger](https://docs.dagger.io/reference/configuration/llm/#ollama)
- [LLM Integration | Dagger](https://docs.dagger.io/features/llm/)
- [Add an AI agent to an Existing Project | Dagger](https://docs.dagger.io/getting-started/quickstarts/agent-in-project/#create-an-agentic-function)
- [Introduction - Container Use](https://container-use.com/introduction)
- [Building AI agents made easy with Goose and Docker | Docker](https://www.docker.com/blog/building-ai-agents-with-goose-and-docker/)
