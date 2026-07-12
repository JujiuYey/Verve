package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/adk"
	"github.com/sergi/go-diff/diffmatchpatch"

	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
	"verve/infrastructure/llm"
	"verve/infrastructure/llm/prompts"
)

const CuratorDocumentCharacterLimit = 60000

type CuratorResult struct {
	ChangeSummary   string `json:"change_summary"`
	ProposedContent string `json:"proposed_content"`
}

type CuratorRequest struct {
	UserID                  string
	DocumentID              string
	TurnID                  string
	RequestID               string
	Instruction             string
	ReplacesChangeRequestID *string
}

type changeRequestWriter interface {
	CreateProposal(context.Context, *wiki_db.DocumentChangeRequest) error
	FindChangeRequest(context.Context, string) (*wiki_db.DocumentChangeRequest, error)
}

type CuratorService struct {
	source FeynmanDocumentSource
	writer changeRequestWriter
	run    agentTextRunner
}

func NewCuratorService(source FeynmanDocumentSource, writer changeRequestWriter) *CuratorService {
	return newCuratorService(source, writer, runWikiCurator)
}

func newCuratorService(source FeynmanDocumentSource, writer changeRequestWriter, run agentTextRunner) *CuratorService {
	return &CuratorService{source: source, writer: writer, run: run}
}

func (s *CuratorService) Propose(ctx context.Context, request CuratorRequest) (*wiki_db.DocumentChangeRequest, error) {
	instruction := strings.TrimSpace(request.Instruction)
	if instruction == "" {
		return nil, errors.New("instruction is required")
	}
	document, content, err := s.source.LoadDocument(ctx, request.UserID, request.DocumentID)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(content) > CuratorDocumentCharacterLimit {
		return nil, fmt.Errorf("document exceeds Curator limit of %d code points", CuratorDocumentCharacterLimit)
	}
	if request.ReplacesChangeRequestID != nil {
		if s.writer == nil {
			return nil, wiki_repo.ErrChangeRequestForbidden
		}
		existing, err := s.writer.FindChangeRequest(ctx, *request.ReplacesChangeRequestID)
		if err != nil {
			return nil, err
		}
		if existing.DocumentID != document.ID || existing.RequestedBy != request.UserID || existing.SourceType != "learning_turn" || (existing.Status != wiki_db.ChangeRequestStatusConflict && existing.Status != wiki_db.ChangeRequestStatusFailed) {
			return nil, wiki_repo.ErrChangeRequestForbidden
		}
	}
	text, err := s.run(ctx, prompts.WikiCuratorQueryPrompt(prompts.WikiCuratorQueryInput{
		DocumentTitle: document.Filename, Content: content, Instruction: instruction,
	}))
	if err != nil {
		return nil, fmt.Errorf("run WikiCurator: %w", err)
	}
	result, err := parseCuratorResult(text)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	proposal := &wiki_db.DocumentChangeRequest{
		DocumentID: document.ID, RequestedBy: request.UserID, SourceType: "learning_turn", SourceID: request.TurnID,
		RequestID: request.RequestID, ReplacesChangeRequestID: request.ReplacesChangeRequestID,
		BaseVersion: document.CurrentVersion, Instruction: instruction, ChangeSummary: result.ChangeSummary,
		ProposedContent: result.ProposedContent, ProposedDiff: buildUnifiedDiff(content, result.ProposedContent),
		Status: wiki_db.ChangeRequestStatusProposed, CreatedAt: now, UpdatedAt: now,
	}
	if s.writer != nil {
		if err := s.writer.CreateProposal(ctx, proposal); err != nil {
			return nil, err
		}
	}
	return proposal, nil
}

func runWikiCurator(ctx context.Context, query string) (string, error) {
	agent, err := llm.NewWikiCuratorAgent(ctx)
	if err != nil {
		return "", err
	}
	return collectText(adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent}).Query(ctx, query))
}

func parseCuratorResult(text string) (*CuratorResult, error) {
	for _, candidate := range jsonObjectCandidates(strings.TrimSpace(text)) {
		var result CuratorResult
		if json.Unmarshal([]byte(candidate), &result) != nil {
			continue
		}
		result.ChangeSummary = strings.TrimSpace(result.ChangeSummary)
		if result.ChangeSummary != "" && strings.TrimSpace(result.ProposedContent) != "" {
			return &result, nil
		}
	}
	return nil, errors.New("WikiCurator output is invalid")
}

func buildUnifiedDiff(current, proposed string) string {
	dmp := diffmatchpatch.New()
	chars1, chars2, lines := dmp.DiffLinesToChars(current, proposed)
	diffs := dmp.DiffMain(chars1, chars2, false)
	diffs = dmp.DiffCharsToLines(diffs, lines)
	var sb strings.Builder
	sb.WriteString("--- current.md\n+++ proposed.md\n")
	sb.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", lineCount(current), lineCount(proposed)))
	for _, diff := range diffs {
		prefix := " "
		if diff.Type == diffmatchpatch.DiffDelete {
			prefix = "-"
		}
		if diff.Type == diffmatchpatch.DiffInsert {
			prefix = "+"
		}
		for _, line := range strings.SplitAfter(diff.Text, "\n") {
			if line == "" {
				continue
			}
			sb.WriteString(prefix)
			sb.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				sb.WriteByte('\n')
			}
		}
	}
	return sb.String()
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}
	count := strings.Count(value, "\n")
	if !strings.HasSuffix(value, "\n") {
		count++
	}
	return count
}
