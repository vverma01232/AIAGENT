package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Scrapper function to scarp data using URL
func ScrapeData(url string) (string, error) {
	scrapperUrl := os.Getenv("SCRAPPERURI")
	requestBody := map[string]interface{}{"url": url}

	// Marshal request body to JSON
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send POST request to scrapper URI
	req, err := http.NewRequest("POST", scrapperUrl+"/scrape", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("error occurred while making the scrape request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error occurred while generating the response: %w", err)
	}
	defer resp.Body.Close()

	// Read and unmarshal the response
	var scrapeResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"` // Scraped data
			} `json:"message"`
		} `json:"choices"`
	}

	scrapeResBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error occurred while reading the response body: %w", err)
	}

	err = json.Unmarshal(scrapeResBody, &scrapeResponse)
	if err != nil {
		return "", fmt.Errorf("error occurred while unmarshaling the response body: %w", err)
	}

	// Check if any scraped content is available
	if len(scrapeResponse.Choices) > 0 {
		return scrapeResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content found in the scraper response")
}
