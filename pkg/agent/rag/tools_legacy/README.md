## tools

This package provides tools for interacting with language models. It includes functions for retrieving and executing tools, such as search and semantic search.

```
pkg/rag/agent/tools/tools.go
```

### GetTools

This function returns a slice of tools and an error. It initializes a slice of tools with two elements, one for search and one for semantic search. For each tool, it sets the type to "function", the function definition, and the description.

### SemanticSearchTool

This struct implements the llms.Tool interface and provides a method for executing semantic search. It takes a context and a slice of messages as input. It iterates through the last message in the slice and checks if it contains a tool call for semantic search. If found, it extracts the query and collection from the arguments, retrieves the vector store based on the collection name, performs a semantic search using the query and the vector store, formats the search results, and appends the formatted results to the state slice.

### Execute

This method takes a context and a slice of messages as input. It iterates through the last message in the slice and checks if it contains a tool call for semantic search. If found, it extracts the query and collection from the arguments, retrieves the vector store based on the collection name, performs a semantic search using the query and the vector store, formats the search results, and appends the formatted results to the state slice.

### Edge Cases

The code does not explicitly handle edge cases, such as errors during tool execution or invalid input parameters.

### Unclear Places

The code does not provide any information on how the vector store is initialized or managed. It also does not specify how the collection name is determined.

### Dead Code

There is no apparent dead code in the provided code snippet.

