package service

import (
	"strings"
	"testing"
)

func TestMarkdownChunkerPreservesHeadingPath(t *testing.T) {
	chunker := NewMarkdownChunker(1200)
	chunks := chunker.Chunk("# Go\n\n## Channel\n\nChannel is a typed pipe.\n\n### Close\n\nClosing broadcasts completion.")
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].HeadingPath != "Go > Channel" {
		t.Fatalf("heading path = %q", chunks[0].HeadingPath)
	}
	if chunks[1].HeadingPath != "Go > Channel > Close" {
		t.Fatalf("heading path = %q", chunks[1].HeadingPath)
	}
}

func TestMarkdownChunkerKeepsCodeFenceTogether(t *testing.T) {
	chunker := NewMarkdownChunker(80)
	chunks := chunker.Chunk("## Example\n\n```go\nfunc main() {\n println(\"hi\")\n}\n```\n\nAfter text.")
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	if !strings.Contains(chunks[0].Content, "func main") {
		t.Fatalf("code fence was split away: %q", chunks[0].Content)
	}
}

func TestMarkdownChunkerEmptyMarkdown(t *testing.T) {
	chunker := NewMarkdownChunker(80)
	chunks := chunker.Chunk(" \n\n\t")
	if len(chunks) != 0 {
		t.Fatalf("expected no chunks, got %d", len(chunks))
	}
}
