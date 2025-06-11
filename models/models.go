package models

import "time"

type CreateEntryRequest struct {
	ProjectID    string    `json:"projectId"`
	Date         time.Time `json:"date"`
	MinutesSpent int       `json:"minutesSpent"`
	Mood         string    `json:"mood"`
	Content      string    `json:"content"`
	Title        string    `json:"title"`
}

type GetUserInfoResponse struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Projects    []Project `json:"projects"`
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}
