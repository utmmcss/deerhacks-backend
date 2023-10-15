package models

type DiscordError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
