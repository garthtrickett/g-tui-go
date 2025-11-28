package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// --- api client ---
type (
	errMsg struct{ err error }
)

type APIClient interface {
	FetchModels() ([]list.Item, error)
	GenerateContent(model, prompt string) (string, error)
	GenerateContentWithImage(model, prompt, imageBase64, mimeType string) (string, error)
}

type LiveAPIClient struct {
	apiKey string
}

func NewLiveAPIClient(apiKey string) *LiveAPIClient {
	return &LiveAPIClient{apiKey: apiKey}
}

func (c *LiveAPIClient) FetchModels() ([]list.Item, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", c.apiKey)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Models []struct {
			Name                       string   `json:"name"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	items := []list.Item{}
	for _, model := range response.Models {
		for _, method := range model.SupportedGenerationMethods {
			if method == "generateContent" {
				items = append(items, item{title: strings.TrimPrefix(model.Name, "models/")})
			}
		}
	}
	return items, nil
}

func (c *LiveAPIClient) GenerateContent(model, prompt string) (string, error) {
	logToFile("--- New API Call ---")
	logToFile(fmt.Sprintf("Model: %s", model))
	logToFile(fmt.Sprintf("Prompt: %s", prompt))

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, c.apiKey)
	logToFile(fmt.Sprintf("URL: %s", url))

	writeFileTool := &Tool{
		FunctionDeclarations: []*FunctionDeclaration{
			{
				Name:        "writeFile",
				Description: "Write a file to the local file system.",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]*Property{
						"name": {
							Type:        "string",
							Description: "The name of the file to write.",
						},
						"content": {
							Type:        "string",
							Description: "The content to write to the file.",
						},
					},
					Required: []string{"name", "content"},
				},
			},
		},
	}

	    reqBody := &apiRequest{

	        Contents: []*Content{

	            {

	                Parts: []Part{

	                    {Text: prompt},

	                },

	            },

	        },

	        Tools: []*Tool{writeFileTool},

	    }

	

	    for {

	        reqBytes, err := json.Marshal(reqBody)

	        if err != nil {

	            logToFile(fmt.Sprintf("ERROR marshalling request: %v", err))

	            return "", err

	        }

	        logToFile(fmt.Sprintf("Request Body: %s", string(reqBytes)))

	

	        res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))

	        if err != nil {

	            logToFile(fmt.Sprintf("ERROR making POST request: %v", err))

	            return "", err

	        }

	        defer res.Body.Close()

	

	        body, err := io.ReadAll(res.Body)

	        if err != nil {

	            logToFile(fmt.Sprintf("ERROR reading response body: %v", err))

	            return "", err

	        }

	        logToFile(fmt.Sprintf("Response Body: %s", string(body)))

	

	        var response apiResponse

	        if err := json.Unmarshal(body, &response); err != nil {

	            logToFile(fmt.Sprintf("ERROR unmarshalling response: %v", err))

	            return "", err

	        }

	

	        		if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {

	

	        			part := response.Candidates[0].Content.Parts[0]

	

	        			logToFile(fmt.Sprintf("API Response Part: %+v", part))

	

	        			logToFile(fmt.Sprintf("Part.FunctionCall is nil: %t", part.FunctionCall == nil))

	

	        			if part.FunctionCall != nil {

	

	        				logToFile("FunctionCall detected.")

	

	        				if part.FunctionCall.Name == "writeFile" {

	

	        					logToFile("writeFile function call detected.")

	                    					name, ok1 := part.FunctionCall.Args["name"].(string)

	                    					rawContent, ok2 := part.FunctionCall.Args["content"]

	                    					content := fmt.Sprintf("%v", rawContent)

	                    					logToFile(fmt.Sprintf("Type assertion for name: %v, content: %v", ok1, ok2))

	                    					if ok1 && ok2 {

	                    						logToFile(fmt.Sprintf("Attempting to write file '%s' with content '%s'", name, content))

	                    						_, err := writeFile(name, content)

	                        if err != nil {

	                            logToFile(fmt.Sprintf("ERROR writing file: %v", err))

	                            return "", err

	                        }

	

	                        reqBody.Contents = append(reqBody.Contents, &Content{

	                            Parts: []Part{

	                                {

	                                    FunctionResponse: &FunctionResponse{

	                                        Name: "writeFile",

	                                        Response: map[string]interface{}{

	                                            "success": err == nil,

	                                        },

	                                    },

	                                },

	                            },

	                            Role: "tool",

	                        })

	                        return "File written successfully.", nil

	                    }

	                }

	            } else {

	
				responseText := response.Candidates[0].Content.Parts[0].Text
				logToFile(fmt.Sprintf("Success. Response text: %s", responseText))
				return responseText, nil
			}
		}
		logToFile("No response from model.")
		return "No response from model.", nil
	}
}

// --- api call types ---

type apiRequest struct {
	Contents []*Content `json:"contents"`
	Tools    []*Tool    `json:"tools,omitempty"`
}

type apiResponse struct {
	Candidates []struct {
		Content *Content `json:"content"`
	} `json:"candidates"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type Part struct {
	Text          string         `json:"text,omitempty"`
	InlineData    *InlineData    `json:"inline_data,omitempty"`
	FunctionCall  *FunctionCall  `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"function_response,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type FunctionResponse struct {
	Name     string `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type Tool struct {
	FunctionDeclarations []*FunctionDeclaration `json:"function_declarations"`
}

type FunctionDeclaration struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  *Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]*Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func writeFile(name string, content string) (string, error) {
	err := os.WriteFile(name, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return "File written successfully.", nil
}

func (c *LiveAPIClient) GenerateContentWithImage(model, prompt, imageBase64, mimeType string) (string, error) {
	logToFile("--- New API Call With Image ---")
	logToFile(fmt.Sprintf("Model: %s", model))
	logToFile(fmt.Sprintf("Prompt: %s", prompt))
	logToFile(fmt.Sprintf("MimeType: %s", mimeType))

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, c.apiKey)
	logToFile(fmt.Sprintf("URL: %s", url))

	writeFileTool := &Tool{
		FunctionDeclarations: []*FunctionDeclaration{
			{
				Name:        "writeFile",
				Description: "Write a file to the local file system.",
				Parameters: &Parameters{
					Type: "object",
					Properties: map[string]*Property{
						"name": {
							Type:        "string",
							Description: "The name of the file to write.",
						},
						"content": {
							Type:        "string",
							Description: "The content to write to the file.",
						},
					},
					Required: []string{"name", "content"},
				},
			},
		},
	}

	reqBody := &apiRequest{
		Contents: []*Content{
			{
				Parts: []Part{
					{Text: prompt},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     imageBase64,
						},
					},
				},
			},
		},
		Tools: []*Tool{writeFileTool},
	}

	for {
		reqBytes, err := json.Marshal(reqBody)
		if err != nil {
			logToFile(fmt.Sprintf("ERROR marshalling request: %v", err))
			return "", err
		}
		logToFile(fmt.Sprintf("Request Body: %s", string(reqBytes)))

		res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
		if err != nil {
			logToFile(fmt.Sprintf("ERROR making POST request: %v", err))
			return "", err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			logToFile(fmt.Sprintf("ERROR reading response body: %v", err))
			return "", err
		}
		logToFile(fmt.Sprintf("Response Body: %s", string(body)))

		var response apiResponse
		if err := json.Unmarshal(body, &response); err != nil {
			logToFile(fmt.Sprintf("ERROR unmarshalling response: %v", err))
			return "", err
		}

		if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
			part := response.Candidates[0].Content.Parts[0]
			logToFile(fmt.Sprintf("API Response Part: %+v", part))
			logToFile(fmt.Sprintf("Part.FunctionCall is nil: %t", part.FunctionCall == nil))
			if part.FunctionCall != nil {
				logToFile("FunctionCall detected.")
				if part.FunctionCall.Name == "writeFile" {
					logToFile("writeFile function call detected.")
					name, ok1 := part.FunctionCall.Args["name"].(string)
					rawContent, ok2 := part.FunctionCall.Args["content"]
					content := fmt.Sprintf("%v", rawContent)
					logToFile(fmt.Sprintf("Type assertion for name: %v, content: %v", ok1, ok2))
					if ok1 && ok2 {
						logToFile(fmt.Sprintf("Attempting to write file '%s' with content '%s'", name, content))
						_, err := writeFile(name, content)
						if err != nil {
							logToFile(fmt.Sprintf("ERROR writing file: %v", err))
							return "", err
						}

						reqBody.Contents = append(reqBody.Contents, &Content{
							Parts: []Part{
								{
									FunctionResponse: &FunctionResponse{
										Name: "writeFile",
										Response: map[string]interface{}{
											"success": err == nil,
										},
									},
								},
							},
							Role: "tool",
						})
						return "File written successfully.", nil
					}
				}
			} else {
				responseText := response.Candidates[0].Content.Parts[0].Text
				logToFile(fmt.Sprintf("Success. Response text: %s", responseText))
				return responseText, nil
			}
		}
		logToFile("No response from model.")
		return "No response from model.", nil
	}
}
