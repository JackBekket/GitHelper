package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	embd "github.com/JackBekket/hellper/lib/embeddings"
	embeddings "github.com/JackBekket/hellper/lib/embeddings"
	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v65/github"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores"
)

var ClientArray []*github.Client
var AI string
var API_TOKEN string
var DB string
var NS string

// Define your API endpoint for handling webhook requests.
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var client *github.Client

	/*
	_ = godotenv.Load()
	_id := os.Getenv("APP_ID")
	app_id := _id
	*/


	// Extract webhook event type from the header
	eventType := github.WebHookType(r)
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}



	// Parse event based on eventType
	switch eventType {
	case "installation_repositories":
		event := new(github.InstallationRepositoriesEvent)
		err := json.Unmarshal(requestBody, event)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling installation_repositories event: %v", err), http.StatusBadRequest)
			return
		}
		if len(event.RepositoriesAdded) > 0 {
			repoOwner := event.Sender.GetLogin()
			repoName := *event.RepositoriesAdded[0].Name
			fmt.Printf("App installed for repository: %s/%s\n", repoOwner, repoName)
		}

	case "issues":
		event := new(github.IssuesEvent)
		err := json.Unmarshal(requestBody, event)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling issues event: %v", err), http.StatusBadRequest)
			return
		}

		if event.GetAction() == "opened" {
			repoOwner := event.GetRepo().GetOwner().GetLogin()
			repoName := event.GetRepo().GetName()
			issueID := event.GetIssue().GetNumber()
			issueTitle := event.GetIssue().GetTitle()
			issueBody := event.GetIssue().GetBody()

			fmt.Printf("New issue opened: %s/%s Issue: %d Title: %s\n", repoOwner, repoName, issueID, issueTitle)
			fmt.Println("Issue body is: ", issueBody)

			// Respond to the issue
			response, err := generateResponse(issueBody, repoName)
			if err != nil {
				log.Println("Can't generate response")
				log.Println(err)
				response = "Can't generate response bleep-bloop"
			}
			fmt.Println("response generated")
			//createClients()
			respond(client, repoOwner, repoName, int64(issueID), response)
		}

	default:
		http.Error(w, "Unknown event type received", http.StatusBadRequest)
	}
}

func generateResponse(prompt string, namespace string) (string, error) {
	collection, err := getCollection(AI, API_TOKEN, DB, namespace) // getting all docs from (whole collection) for namespace (repo_name)
	if err != nil {
		log.Println(err)
	}
	/* opts := vectorstores.WithFilters(map[string]string{
		"type": "doc",
	}) */

		fmt.Println("namespace is: ", namespace)

		//🤕🤕🤕
		searchResults, err := embeddings.SemanticSearch(prompt, 2, collection)
		if err != nil {
			return "", err
		}
	
		contextBuilder := strings.Builder{}
		for _, doc := range searchResults {
			contextBuilder.WriteString(doc.PageContent)
			contextBuilder.WriteString("\n")
		}
		contexts := contextBuilder.String()
	
		fmt.Sprintf("Context: %s\n\nQuestion: %s", contexts, prompt)


	response, err := rag(prompt, AI, API_TOKEN, 1, collection)
	if err != nil {
		return "", err
	}
	return response, nil
}

func respond(client *github.Client, owner string, repo string, id int64, response string) {
	ctx := context.Background()
	// Craft a reply message from the response from the 3rd service.
	replyMessage := fmt.Sprintf("Here's the response from our 3rd service:\n%s", response)

	// Create a new comment on the issue using the GitHub API.

	//client := github.NewClient(nil)
	a, b, err := client.Issues.CreateComment(ctx, owner, repo, int(id), &github.IssueComment{
		// Configure the comment with the issue's ID and other necessary details.
		Body: &replyMessage,
	})

	fmt.Println("Var #1: ", a)
	fmt.Println("Var #2: ", b)
	if err != nil {
		fmt.Println("Error creating comment on issue: ", err)
	} else {
		fmt.Println("Comment successfully created!")
	}

}



// get only last installment, bug, will be deprecated
func createClient(key_path string, app_id int) *github.Client {
	//const gitHost = "https://git.api.com"


	/*
	privatePem, err := os.ReadFile(key_path)
	if err != nil {
		log.Fatalf("failed to read pem: %v", err)
	}


	itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, int64(app_id), privatePem)
	if err != nil {
		log.Fatalf("failed to create app transport: %v\n", err)
	}
	*/

	tr := http.DefaultTransport
	pk_name := os.Getenv("PRIVATE_KEY_NAME")

	itr, err := ghinstallation.NewAppsTransportKeyFromFile(tr, int64(app_id), pk_name)
    if err != nil {
        log.Fatal(err)
    }




	//itr.BaseURL = gitHost

	//create git client with app transport
	client := github.NewClient(
		&http.Client{
			Transport: itr,
			Timeout:   time.Second * 30,
		},
	)
	//)

	if client == nil {
		log.Fatalf("failed to create git client for app: %v\n", err)
	}

	installations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list installations: %v\n", err)
	}

	//capture our installationId for our app
	//we need this for the access token
	var installID int64
	for _, val := range installations {
		installID = val.GetID()
	}

	token, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		installID,
		&github.InstallationTokenOptions{})
	if err != nil {
		log.Fatalf("failed to create installation token: %v\n", err)
	}

	apiClient := github.NewClient(nil).WithAuthToken(
		token.GetToken(),
	)
	if apiClient == nil {
		log.Fatalf("failed to create new git client with token: %v\n", err)
	}

	//log.Println("gh client: ", apiClient)

	return apiClient
}


func createClients(app_id int) []*github.Client  {


	var result []*github.Client
	tr := http.DefaultTransport
	pk_name := os.Getenv("PRIVATE_KEY_NAME")

	itr, err := ghinstallation.NewAppsTransportKeyFromFile(tr, int64(app_id), pk_name)
    if err != nil {
        log.Fatal(err)
    }

	//create git client with app transport
	client := github.NewClient(
		&http.Client{
			Transport: itr,
			Timeout:   time.Second * 30,
		},
	)
	//)

	if client == nil {
		log.Fatalf("failed to create git client for app: %v\n", err)
	}

	installations, _, err := client.Apps.ListInstallations(context.Background(), &github.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list installations: %v\n", err)
	}

	// log installations
	println("installations: ", installations)


	//capture our installationId for our app
	//we need this for the access token
	var installID int64
	for _, val := range installations {
		installID = val.GetID()


		//TODO: input any other check here before creating a client


		token, _, err := client.Apps.CreateInstallationToken(
			context.Background(),
			installID,
			&github.InstallationTokenOptions{})
		if err != nil {
			log.Fatalf("failed to create installation token: %v\n", err)
		}

		apiClient := github.NewClient(nil).WithAuthToken(
			token.GetToken(),
		)
		if apiClient == nil {
			log.Fatalf("failed to create new git client with token: %v\n", err)
		}

		fmt.Println("API CLIENT: ",apiClient)

		//append
		result = append(result, apiClient)
	}

	



	log.Println("clients created: ", result)

	//eturn apiClient
	return result
}



func main() {
	fmt.Println("main process started")

	// creating github client from private key
	_ = godotenv.Load()
	_id := os.Getenv("APP_ID")
	//wh_secret := os.Getenv("WEBHOOK_SECRET")
	//pk_path := os.Getenv("PRIVATE_KEY_PATH")
	app_id, err := strconv.Atoi(_id)
	if err != nil {
		// ... handle error
		panic(err)
	}
	ClientArray = createClients(app_id)

	//Test getting vectorstore from .env
	// In production name should be replaced by event value
	ai := os.Getenv("AI_ENDPOINT")
	apit := os.Getenv("API_TOKEN")
	db_link := os.Getenv("DB_URL")
	//namesp := os.Getenv("REPO_NAME")

	AI = ai
	API_TOKEN = apit
	DB = db_link
	//NS = namesp

	// ... (Set up your webhook endpoint and start the server)
	http.HandleFunc("/webhook", handleWebhook)
	//log.Fatal(http.ListenAndServe(":8086", nil))
	log.Fatal(http.ListenAndServe(":8186", nil))
}

func getCollection(ai_url string, api_token string, db_link string, namespace string) (vectorstores.VectorStore, error) {
	store, err := embd.GetVectorStoreWithOptions(ai_url, api_token, db_link, namespace) // ai, api, db, namespace
	if err != nil {
		return nil, err
	}
	return store, nil
}

func rag(question string, ai_url string, api_token string, numOfResults int, store vectorstores.VectorStore) (result string, err error) {
	//base_url := os.Getenv("AI_BASEURL")
	base_url := ai_url

	// Create an embeddings client using the specified API and embedding model
	llm, err := openai.New(
		openai.WithBaseURL(base_url),
		openai.WithAPIVersion("v1"),
		openai.WithToken(api_token),
		openai.WithModel("tiger-gemma-9b-v1-i1"),
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	if err != nil {
		return "", err
	}

	//🤕🤕🤕
	searchResults, err := embeddings.SemanticSearch(question, numOfResults, store)
	if err != nil {
		return "", err
	}

	contextBuilder := strings.Builder{}
	for _, doc := range searchResults {
		contextBuilder.WriteString(doc.PageContent)
		contextBuilder.WriteString("\n")
	}
	contexts := contextBuilder.String()

	fullPrompt := fmt.Sprintf("Context: %s\n\nQuestion: %s", contexts, question)

	result, err = chains.Run(
		context.Background(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, numOfResults),
		),
		fullPrompt,
		chains.WithMaxTokens(8192),
	)
	if err != nil {
		return "", err
	}

	fmt.Println("====final answer====\n", result)

	return result, nil
}
