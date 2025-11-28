package main

import (
	"errors"
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// MockAPIClient is a mock implementation of the APIClient interface for testing.
type MockAPIClient struct {
	FetchModelsFunc     func() ([]list.Item, error)
	GenerateContentFunc func(model, prompt string) (string, error)
}

func (m *MockAPIClient) FetchModels() ([]list.Item, error) {
	if m.FetchModelsFunc != nil {
		return m.FetchModelsFunc()
	}
	return []list.Item{
		item{title: "gemini-pro"},
		item{title: "gemini-pro-vision"},
	}, nil
}

func (m *MockAPIClient) GenerateContent(model, prompt string) (string, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(model, prompt)
	}
	return "mocked response", nil
}

func TestInitialModel(t *testing.T) {
	// Unset the API key to test the default case
	os.Unsetenv("GEMINI_API_KEY")

	// Test case 1: No API key
	t.Run("No API Key", func(t *testing.T) {
		model := initialModel()

		if model.state != showChat {
			t.Errorf("Expected state to be showChat, but got %v", model.state)
		}

		if model.selectedModel != "gemini-3-pro-preview" {
			t.Errorf("Expected selectedModel to be \"gemini-3-pro-preview\", but got %s", model.selectedModel)
		}
	})

	// Set the API key to test the pre-selection case
	os.Setenv("GEMINI_API_KEY", "test-api-key")

	// Test case 2: API key is set
	t.Run("API Key is set", func(t *testing.T) {
		model := initialModel()

		if model.state != showChat {
			t.Errorf("Expected state to be showChat, but got %v", model.state)
		}

		if model.selectedModel != "gemini-3-pro-preview" {
			t.Errorf("Expected selectedModel to be \"gemini-3-pro-preview\", but got %s", model.selectedModel)
		}
	})

	// Clean up the environment variable
	os.Unsetenv("GEMINI_API_KEY")
}

func TestFetchModels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := &MockAPIClient{}
		cmd := fetchModels(client)
		msg := cmd()

		fetchedMsg, ok := msg.(fetchedModelsMsg)
		if !ok {
			t.Fatalf("Expected fetchedModelsMsg, but got %T", msg)
		}

		if len(fetchedMsg) != 2 {
			t.Errorf("Expected 2 models, but got %d", len(fetchedMsg))
		}
	})

	t.Run("Error", func(t *testing.T) {
		client := &MockAPIClient{
			FetchModelsFunc: func() ([]list.Item, error) {
				return nil, errors.New("API error")
			},
		}
		cmd := fetchModels(client)
		msg := cmd()

		errMsg, ok := msg.(errMsg)
		if !ok {
			t.Fatalf("Expected errMsg, but got %T", msg)
		}

		if errMsg.err.Error() != "API error" {
			t.Errorf("Expected error 'API error', but got '%s'", errMsg.err.Error())
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("fetchedModelsMsg", func(t *testing.T) {
		m := initialModel()
		models := []list.Item{item{title: "test-model"}}
		msg := fetchedModelsMsg(models)

		newModel, _ := m.Update(msg)
		updatedModel := newModel.(model)

		if len(updatedModel.list.Items()) != 1 {
			t.Errorf("Expected list to have 1 item, but got %d", len(updatedModel.list.Items()))
		}
	})

	t.Run("apiResponseMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showChat
		msg := apiResponseMsg("test response")

		newModel, _ := m.Update(msg)
		updatedModel := newModel.(model)

		if len(updatedModel.messages) != 1 {
			t.Fatalf("Expected 1 message, but got %d", len(updatedModel.messages))
		}

		if updatedModel.messages[0] != "Gemini: test response" {
			t.Errorf("Expected message to be 'Gemini: test response', but got '%s'", updatedModel.messages[0])
		}
	})

		t.Run("KeyEnter in showChat", func(t *testing.T) {
			m := initialModel()
			m.textInput.SetValue("Hello Gemini")
			client := &MockAPIClient{
				GenerateContentFunc: func(model, prompt string) (string, error) {
					return "response from model", nil
				},
			}
			m.client = client
			msg := tea.KeyMsg{Type: tea.KeyEnter}

			newModel, _ := m.Update(msg)
			updatedModel := newModel.(model)

			if !updatedModel.loading {
				t.Errorf("Expected loading to be true after sending message")
			}
			if len(updatedModel.messages) != 1 {
				t.Errorf("Expected 1 message after sending, got %d", len(updatedModel.messages))
			}
			if updatedModel.messages[0] != "You: Hello Gemini" {
				t.Errorf("Expected message to be 'You: Hello Gemini', got '%s'", updatedModel.messages[0])
			}

			// Simulate API response
			apiMsg := apiResponseMsg("response from model")

			finalModelRaw, _ := updatedModel.Update(apiMsg)
			finalModel := finalModelRaw.(model)

			if finalModel.loading {
				t.Errorf("Expected loading to be false after API response")
			}
			if len(finalModel.messages) != 2 {
				t.Errorf("Expected 2 messages after API response, got %d", len(finalModel.messages))
			}
			if finalModel.messages[1] != "Gemini: response from model" {
				t.Errorf("Expected response message, got %s", finalModel.messages[1])
			}
		})
	}
