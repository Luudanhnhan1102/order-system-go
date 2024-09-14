package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const mailtrapAPIURL = "https://send.api.mailtrap.io/api/send"

var apiToken = os.Getenv("MAILTRAP_API_TOKEN")

type EmailData struct {
	To      []EmailAddress `json:"to"`
	Subject string         `json:"subject"`
	Text    string         `json:"text"`
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func SendEmail(emailData EmailData) error {
	if apiToken == "" {
		fmt.Println("Mailtrap API token not defined. Skipping email send.")
		return nil
	}

	payload := struct {
		From    EmailAddress   `json:"from"`
		To      []EmailAddress `json:"to"`
		Subject string         `json:"subject"`
		Text    string         `json:"text"`
	}{
		From:    EmailAddress{Email: "mailtrap@demomailtrap.com", Name: "Cursor Experiment Project"},
		To:      emailData.To,
		Subject: emailData.Subject,
		Text:    emailData.Text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling email data: %v", err)
	}

	req, err := http.NewRequest("POST", mailtrapAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Token", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
