# reflexia_integration

This package is designed to integrate with Reflexia, a tool for summarizing and analyzing software projects. It takes a GitHub repository URL or a local project path as input and processes the project to generate summaries of its packages.

The package relies on several external data sources, including environment variables, command line arguments, and a GitHub repository. Environment variables such as GH_USERNAME, GH_TOKEN, HELPER_URL, MODEL, API_TOKEN, EMBEDDINGS_AI_URL, EMBEDDINGS_AI_KEY, and EMBEDDINGS_DB_URL are used for configuration. Command line arguments allow users to specify the GitHub repository URL or local project path.

The package's functionality is centered around the InitPackageRunner function, which initializes a PackageRunnerService. This service is responsible for running and summarizing packages in the project. The function first loads environment variables and command line arguments to determine the input project. Then, it processes the working directory, either by cloning a GitHub repository or using a local project path. Next, it extracts project configuration and builds package files.

The PackageRunnerService then initializes a SummarizeService, which uses a language model to generate summaries of the packages, and an EmbeddingsService, which stores and retrieves embeddings of the packages. Finally, it creates a PackageRunnerService instance and returns it.

The processWorkingDirectory function handles the cloning of the GitHub repository or the use of the local project path. It also handles submodules and the depth of the clone.

The loadEnv function is responsible for loading environment variables and checking if they are empty.

The package's structure is as follows:

- reflexia_integration.go
- internal/reflexia_integration/reflexia_integration.go

The package imports several external libraries, including errors, flag, fmt, log, net/url, os, path/filepath, strings, github.com/JackBekket/reflexia/pkg, github.com/JackBekket/reflexia/pkg/package_runner, github.com/JackBekket/reflexia/pkg/project, github.com/JackBekket/reflexia/pkg/summarize, github.com/go-git/go-git/v5, github.com/go-git/go-git/v5/plumbing/transport/http, and github.com/tmc/langchaingo/llms.

The package's main function is to provide a way to integrate with Reflexia and process software projects to generate summaries of their packages. It does this by initializing a PackageRunnerService, which handles the entire process from loading the project to generating the summaries.

Edge cases for launching the application:

- If the GitHub repository URL or local project path is not provided, the application will use the current working directory as the input project.
- If the required environment variables are not set, the application will attempt to load them from a configuration file.
- If the required libraries are not installed, the application will attempt to install them automatically.

