package parser

import (
	"encoding/json"
	"fmt"
)

// Document within the database
type Document struct {
	// OldID of the document
	OldID string `json:"_id"`
	// Name of the document
	Name string `json:"name"`
	// NewID of the document
	NewID string
}

// ParseDocument from JSON bytes
func ParseDocument(in []byte) (Document, error) {
	var doc Document

	err := json.Unmarshal(in, &doc)
	if err != nil {
		return doc, fmt.Errorf("error unmarshalling database file: %w", err)
	}

	doc.NewID = NewRandomID()

	return doc, nil
}
