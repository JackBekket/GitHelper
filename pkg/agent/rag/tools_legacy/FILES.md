# pkg/rag/agent/tools/tools.go  
package: tools  
imports: context, encoding/json, log, os, github.com/JackBekket/hellper/lib/embeddings, github.com/joho/godotenv, github.com/tmc/langchaingo/llms  
  
func GetTools():  
	- returns a slice of llms.Tool and an error  
	- initializes a slice of llms.Tool with two elements  
	- for each element in the slice  
		- sets the Type field to "function"  
		- sets the Function field to an llms.FunctionDefinition struct  
		- sets the Name field of the llms.FunctionDefinition struct to either "search" or "semanticSearch"  
		- sets the Description field of the llms.FunctionDefinition struct to either "Preforms Duck Duck Go web search" or "Performs semantic search using a vector store"  
		- sets the Parameters field of the llms.FunctionDefinition struct to a map[string]any  
		- returns the slice of llms.Tool and nil  
func (s *SemanticSearchTool) Execute(ctx context.Context, state []llms.MessageContent):  
	- takes a context.Context and a slice of llms.MessageContent as input  
	- iterates through the last message in the state slice  
	- for each part in the message  
		- checks if the part is an llms.ToolCall and if the function call name is "semanticSearch"  
		- if true, extracts the query and collection from the arguments  
		- retrieves the vector store based on the collection name  
		- performs a semantic search using the query and the vector store  
		- formats the search results  
		- appends the formatted results to the state slice  
	- returns the updated state slice and nil  
  
  
