package service

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

type MarkdownChunk struct {
	HeadingPath string
	BlockType   string
	Content     string
	ContentHash string
	TokenCount  int
}

type MarkdownChunker struct {
	maxChars int
}

func NewMarkdownChunker(maxChars int) *MarkdownChunker {
	if maxChars <= 0 {
		maxChars = 1800
	}
	return &MarkdownChunker{maxChars: maxChars}
}

func (c *MarkdownChunker) Chunk(markdown string) []MarkdownChunk {
	lines := strings.Split(markdown, "\n")
	headings := make([]string, 0, 4)
	chunks := make([]MarkdownChunk, 0)
	var buf strings.Builder
	blockType := "paragraph"
	inFence := false
	orderedListRE := regexp.MustCompile(`^\d+\.\s+`)
	headingRE := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	flush := func() {
		content := strings.TrimSpace(buf.String())
		if content == "" {
			buf.Reset()
			return
		}
		path := strings.Join(headings, " > ")
		if path == "" {
			path = "Document"
		}
		chunks = append(chunks, MarkdownChunk{
			HeadingPath: path,
			BlockType:   blockType,
			Content:     content,
			ContentHash: hashText(path + "\n" + content),
			TokenCount:  estimateTokens(content),
		})
		buf.Reset()
		blockType = "paragraph"
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			blockType = "code"
			buf.WriteString(line + "\n")
			continue
		}
		if !inFence {
			if match := headingRE.FindStringSubmatch(trimmed); match != nil {
				flush()
				level := len(match[1])
				title := strings.TrimSpace(match[2])
				if level <= len(headings) {
					headings = headings[:level-1]
				}
				headings = append(headings, title)
				continue
			}
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || orderedListRE.MatchString(trimmed) {
				blockType = "list"
			}
			if buf.Len() > c.maxChars && trimmed == "" {
				flush()
				continue
			}
		}
		buf.WriteString(line + "\n")
	}
	flush()
	return chunks
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func estimateTokens(text string) int {
	runes := []rune(text)
	return (len(runes) + 3) / 4
}
