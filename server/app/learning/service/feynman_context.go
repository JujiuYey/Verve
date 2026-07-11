package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	rag_db "verve/app/rag/models/db"
	rag_payload "verve/app/rag/models/payload"
	wiki_db "verve/app/wiki/models/db"
)

const (
	FullDocumentCharacterLimit = 24000
	FeynmanContextModeFull     = "full"
	FeynmanContextModeRAG      = "rag"

	feynmanSearchLimit        = 6
	feynmanNeighborRadius     = 1
	minimumFeynmanRAGEvidence = 2
)

type FeynmanEvidence struct {
	ChunkID     string `json:"chunk_id"`
	ChunkIndex  int    `json:"chunk_index"`
	HeadingPath string `json:"heading_path"`
	Content     string `json:"content"`
}

type FeynmanDocumentContext struct {
	DocumentID                 string            `json:"document_id"`
	Title                      string            `json:"title"`
	Outline                    []string          `json:"outline"`
	Evidence                   []FeynmanEvidence `json:"evidence"`
	FullText                   string            `json:"full_text"`
	Mode                       string            `json:"mode"`
	ContextSufficient          bool              `json:"context_sufficient"`
	ContextInsufficiencyReason string            `json:"context_insufficiency_reason,omitempty"`
}

type FeynmanDocumentSource interface {
	LoadDocument(ctx context.Context, documentID string) (*wiki_db.Document, string, error)
	SearchDocument(ctx context.Context, documentID, query string, limit int) ([]rag_payload.SearchResult, error)
	FindNeighbors(ctx context.Context, documentID string, indexes []int, radius int) ([]*rag_db.WikiChunk, error)
}

type FeynmanContextBuilder struct {
	source FeynmanDocumentSource
}

func NewFeynmanContextBuilder(source FeynmanDocumentSource) *FeynmanContextBuilder {
	return &FeynmanContextBuilder{source: source}
}

func (b *FeynmanContextBuilder) Build(ctx context.Context, documentID, learnerExplanation string) (*FeynmanDocumentContext, error) {
	if b == nil || b.source == nil {
		return nil, fmt.Errorf("feynman document source is required")
	}
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}
	doc, markdown, err := b.source.LoadDocument(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("load Feynman document %q: %w", documentID, err)
	}
	if doc == nil {
		return nil, fmt.Errorf("load Feynman document %q: document is nil", documentID)
	}
	if strings.TrimSpace(markdown) == "" {
		return nil, fmt.Errorf("Feynman document %q markdown is empty", documentID)
	}

	result := &FeynmanDocumentContext{
		DocumentID: doc.ID,
		Title:      doc.Filename,
		Outline:    markdownHeadingPaths(markdown),
		Evidence:   []FeynmanEvidence{},
	}
	if utf8.RuneCountInString(markdown) <= FullDocumentCharacterLimit {
		result.Mode = FeynmanContextModeFull
		result.FullText = markdown
		result.ContextSufficient = true
		return result, nil
	}

	result.Mode = FeynmanContextModeRAG
	query := strings.TrimSpace(learnerExplanation)
	if query == "" {
		result.ContextInsufficiencyReason = "learner explanation is empty, so relevant document evidence cannot be retrieved"
		return result, nil
	}
	hits, err := b.source.SearchDocument(ctx, documentID, query, feynmanSearchLimit)
	if err != nil {
		return nil, fmt.Errorf("search Feynman document %q: %w", documentID, err)
	}
	indexes := make([]int, 0, len(hits))
	for _, hit := range hits {
		if hit.DocumentID == documentID {
			indexes = append(indexes, hit.ChunkIndex)
		}
	}
	neighbors, err := b.source.FindNeighbors(ctx, documentID, indexes, feynmanNeighborRadius)
	if err != nil {
		return nil, fmt.Errorf("find Feynman evidence neighbors for %q: %w", documentID, err)
	}
	result.Evidence = mergeFeynmanEvidence(documentID, hits, neighbors)
	result.ContextSufficient = len(result.Evidence) >= minimumFeynmanRAGEvidence
	if !result.ContextSufficient {
		result.ContextInsufficiencyReason = fmt.Sprintf("retrieval returned %d unique evidence chunk(s); at least %d are required", len(result.Evidence), minimumFeynmanRAGEvidence)
	}
	return result, nil
}

func markdownHeadingPaths(markdown string) []string {
	var stack [6]string
	paths := make([]string, 0)
	appendHeading := func(level int, title string) {
		stack[level-1] = title
		for i := level; i < len(stack); i++ {
			stack[i] = ""
		}
		parts := make([]string, 0, level)
		for i := 0; i < level; i++ {
			if stack[i] != "" {
				parts = append(parts, stack[i])
			}
		}
		paths = append(paths, strings.Join(parts, " > "))
	}

	lines := strings.Split(markdown, "\n")
	var fence byte
	var fenceLength int
	for i, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		marker, markerLength := markdownFenceMarker(line)
		if fence != 0 {
			if marker == fence && markerLength >= fenceLength {
				fence = 0
				fenceLength = 0
			}
			continue
		}
		if marker != 0 {
			fence = marker
			fenceLength = markerLength
			continue
		}
		if level := markdownSetextLevel(line); level != 0 && i > 0 {
			title := strings.TrimSpace(lines[i-1])
			if title != "" {
				appendHeading(level, title)
			}
			continue
		}

		level := 0
		for level < len(line) && level < len(stack) && line[level] == '#' {
			level++
		}
		if level == 0 || level >= len(line) || (line[level] != ' ' && line[level] != '\t') {
			continue
		}
		title := strings.TrimSpace(line[level:])
		title = strings.TrimSpace(strings.TrimRight(title, "#"))
		if title == "" {
			continue
		}
		appendHeading(level, title)
	}
	return paths
}

func markdownFenceMarker(line string) (byte, int) {
	if len(line) < 3 || (line[0] != '`' && line[0] != '~') {
		return 0, 0
	}
	length := 1
	for length < len(line) && line[length] == line[0] {
		length++
	}
	if length < 3 {
		return 0, 0
	}
	return line[0], length
}

func markdownSetextLevel(line string) int {
	if line == "" {
		return 0
	}
	marker := line[0]
	if marker != '=' && marker != '-' {
		return 0
	}
	for i := 1; i < len(line); i++ {
		if line[i] != marker {
			return 0
		}
	}
	if marker == '=' {
		return 1
	}
	return 2
}

func mergeFeynmanEvidence(documentID string, hits []rag_payload.SearchResult, neighbors []*rag_db.WikiChunk) []FeynmanEvidence {
	byIndex := make(map[int]FeynmanEvidence, len(hits)+len(neighbors))
	add := func(item FeynmanEvidence) {
		if strings.TrimSpace(item.Content) == "" {
			return
		}
		byIndex[item.ChunkIndex] = item
	}
	for _, hit := range hits {
		if hit.DocumentID == documentID {
			add(FeynmanEvidence{ChunkID: hit.ChunkID, ChunkIndex: hit.ChunkIndex, HeadingPath: hit.HeadingPath, Content: hit.Content})
		}
	}
	for _, chunk := range neighbors {
		if chunk != nil && chunk.DocumentID == documentID {
			add(FeynmanEvidence{ChunkID: chunk.ID, ChunkIndex: chunk.ChunkIndex, HeadingPath: chunk.HeadingPath, Content: chunk.Content})
		}
	}
	evidence := make([]FeynmanEvidence, 0, len(byIndex))
	for _, item := range byIndex {
		evidence = append(evidence, item)
	}
	sort.SliceStable(evidence, func(i, j int) bool {
		if evidence[i].ChunkIndex == evidence[j].ChunkIndex {
			return evidence[i].ChunkID < evidence[j].ChunkID
		}
		return evidence[i].ChunkIndex < evidence[j].ChunkIndex
	})
	return evidence
}
