package entity

import (
	"time"
)

// Document は Elasticsearch のドキュメントを表す
type Document struct {
	ID       string         `json:"id"`
	Index    string         `json:"index"`
	Source   map[string]any `json:"source"`
	Version  int64          `json:"version"`
	Created  time.Time      `json:"created"`
	Modified time.Time      `json:"modified"`
}

// NewDocument は新しい Document インスタンスを作成する
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

// SetID はドキュメント ID を設定する
func (d *Document) SetID(id string) {
	d.ID = id
}

// UpdateSource はドキュメントのソースを更新し、バージョンをインクリメントする
func (d *Document) UpdateSource(source map[string]any) {
	d.Source = source
	d.Version++
	d.Modified = time.Now()
}

// GetField はドキュメントソースから特定のフィールドを取得する
func (d *Document) GetField(field string) (any, bool) {
	value, exists := d.Source[field]
	return value, exists
}

// SetField はドキュメントソースに特定のフィールドを設定する
func (d *Document) SetField(field string, value any) {
	d.Source[field] = value
	d.Modified = time.Now()
}
