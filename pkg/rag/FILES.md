# pkg/rag/rag_test.go  
**Package Name:** rag_test  
**Imports:**  
- `fmt`  
- `log`  
- `os`  
- `testing`  
- `github.com/JackBekket/GitHelper/pkg/rag` (RAG)  
- `github.com/JackBekket/hellper/lib/embeddings` (embeddings)  
- `github.com/joho/godotenv` (godotenv)  
- `github.com/tmc/langchaingo/vectorstores` (vectorstores)  
  
**External Data and Input Sources:**  
  
* Environment variables:  
	+ `AI_ENDPOINT`  
	+ `API_TOKEN`  
	+ `DB_URL`  
* `getCollection` function:  
	+ `ai_url` (string)  
	+ `api_token` (string)  
	+ `db_link` (string)  
	+ `namespace` (string)  
  
**TODO Comments:**  
  
* None found in the provided code.  
  
**Summary:**  
  
### Overview  
The `rag_test` package contains three test functions: `Test_RagWithFilteres`, `Test_RagReflexia`, and `Test_RefinedQA_RAG`. These tests aim to test the functionality of the RAG (Retrieval-Augmented Generation) model.  
  
### Test Functions  
#### Test_RagWithFilteres  
This test function calls the RAG model with optional filters and tests the response.  
  
#### Test_RagReflexia  
This test function calls the Refined QA method of RAG and tests the response.  
  
#### Test_RefinedQA_RAG  
This test function calls the Refined QA method of RAG and tests the response.  
  
### getCollection Function  
This function is used to retrieve a vector store from the `embeddings` package.  
  
**  
  
