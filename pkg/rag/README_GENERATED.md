**rag_test**
================

**Overview**
------------

The `rag_test` package is a test suite for the RAG (Retrieval-Augmented Generation) model. It contains three test functions: `Test_RagWithFilteres`, `Test_RagReflexia`, and `Test_RefinedQA_RAG`. These tests aim to verify the functionality of the RAG model.

**Test Functions**
-----------------

### Test_RagWithFilteres
This test function calls the RAG model with optional filters and tests the response.

### Test_RagReflexia
This test function calls the Refined QA method of RAG and tests the response.

### Test_RefinedQA_RAG
This test function calls the Refined QA method of RAG and tests the response.

**getCollection Function**
-------------------------

This function is used to retrieve a vector store from the `embeddings` package.

**Configuration and Launch Options**
--------------------------------

* Environment variables:
	+ `AI_ENDPOINT`
	+ `API_TOKEN`
	+ `DB_URL`
* Command-line arguments:
	* None
* Files and paths:
	* `getCollection` function: `embeddings` package

**Launch Options**
-----------------

* The package can be launched using the `go test` command.
* The package can be launched using the `go run rag_test.go` command.

**Edge Cases**
--------------

* The package can be launched with the `AI_ENDPOINT`, `API_TOKEN`, and `DB_URL` environment variables set.
* The package can be launched with the `embeddings` package path specified.

**Notes**
--------

* The package does not contain any dead code or unclear places.
* The package does not contain any TODO comments.

