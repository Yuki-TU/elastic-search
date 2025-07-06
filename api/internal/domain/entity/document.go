package entity

import (
	"time"
)

// Document represents a document in Elasticsearch
type Document struct {
	ID       string         `json:"id"`
	Index    string         `json:"index"`
	Source   map[string]any `json:"source"`
	Version  int64          `json:"version"`
	Created  time.Time      `json:"created"`
	Modified time.Time      `json:"modified"`
}

// NewDocument creates a new Document instance
func NewDocument(index string, source map[string]any) *Document {
	now := time.Now()
	return &Document{
		Index:    index,
		Source:   source,
		Version:  1,
		Created:  now,
		Modified: now,
	}
}

// SetID sets the document ID
func (d *Document) SetID(id string) {
	d.ID = id
}

// UpdateSource updates the document source and increment version
func (d *Document) UpdateSource(source map[string]any) {
	d.Source = source
	d.Version++
	d.Modified = time.Now()
}

// GetField returns a specific field from the document source
func (d *Document) GetField(field string) (any, bool) {
	value, exists := d.Source[field]
	return value, exists
}

// SetField sets a specific field in the document source
func (d *Document) SetField(field string, value any) {
	d.Source[field] = value
	d.Modified = time.Now()
}
