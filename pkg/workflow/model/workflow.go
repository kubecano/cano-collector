package model

// AlertInput - uproszczony typ wejściowy do matching (do rozbudowy)
type AlertInput struct {
	AlertName   string
	Status      string
	Severity    string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}
