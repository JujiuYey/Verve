package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/eino/schema"

	learning_db "verve/app/learning/models/db"
	rag_payload "verve/app/rag/models/payload"
	"verve/infrastructure/llm"
	"verve/infrastructure/llm/prompts"
)

const (
	KnowledgeQAQuestionCharacterLimit = 2000
	KnowledgeQAHistoryMessageLimit    = 12
	KnowledgeQAHistoryCharacterBudget = 24000
	KnowledgeQARetrievalQueryBudget   = 3000
	KnowledgeQARetrievalLimit         = 8
	KnowledgeQAMemoryLimit            = 20
	MinimumKnowledgeQAEvidenceScore   = 0.65
)

var (
	ErrKnowledgeQAQuestionRequired = errors.New("question is required")
	ErrKnowledgeQAQuestionTooLong  = errors.New("question is too long")
)

type KnowledgeQAMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KnowledgeQARequest struct {
	Message string               `json:"message"`
	History []KnowledgeQAMessage `json:"history"`
}

type KnowledgeQASource struct {
	DocumentID    string  `json:"documentId"`
	DocumentTitle string  `json:"documentTitle"`
	FolderPath    string  `json:"folderPath"`
	HeadingPath   string  `json:"headingPath"`
	Score         float64 `json:"score"`
}

type KnowledgeQAAnswer struct {
	KnowledgeAnswer string `json:"knowledgeAnswer"`
	LearningAdvice  string `json:"learningAdvice"`
}

type PreparedKnowledgeQA struct {
	Sources         []KnowledgeQASource
	Prompt          string
	MemoryStatus    string
	ImmediateAnswer *KnowledgeQAAnswer
}

type globalKnowledgeRetriever interface {
	SearchAll(ctx context.Context, query string, limit int) ([]rag_payload.SearchResult, error)
}

type knowledgeQAMemoryReader interface {
	FindDocumentItemsBatch(ctx context.Context, documentIDs []string, limit int) ([]*learning_db.LearningMemoryItem, error)
}

type knowledgeQAModelRunner func(ctx context.Context, prompt string) (string, error)

type KnowledgeQAService struct {
	retriever globalKnowledgeRetriever
	memories  knowledgeQAMemoryReader
	run       knowledgeQAModelRunner
}

func NewKnowledgeQAService(retriever globalKnowledgeRetriever, memories knowledgeQAMemoryReader, resolver llm.AgentModelResolver) *KnowledgeQAService {
	return newKnowledgeQAService(retriever, memories, makeKnowledgeQAModelRunner(resolver))
}

func newKnowledgeQAService(retriever globalKnowledgeRetriever, memories knowledgeQAMemoryReader, run knowledgeQAModelRunner) *KnowledgeQAService {
	return &KnowledgeQAService{retriever: retriever, memories: memories, run: run}
}

func makeKnowledgeQAModelRunner(resolver llm.AgentModelResolver) knowledgeQAModelRunner {
	return func(ctx context.Context, query string) (string, error) {
		chatModel, err := llm.NewStructuredChatModel(ctx, resolver, llm.AgentKeyKnowledgeQA, llm.SceneKeyDefault)
		if err != nil {
			return "", err
		}
		message, err := chatModel.Generate(ctx, []*schema.Message{
			schema.SystemMessage(prompts.KnowledgeQAPrompt(prompts.Input{})),
			schema.UserMessage(query),
		})
		if err != nil {
			return "", err
		}
		if message == nil {
			return "", errors.New("knowledge QA model returned no message")
		}
		return message.Content, nil
	}
}

func (s *KnowledgeQAService) Prepare(ctx context.Context, request KnowledgeQARequest) (*PreparedKnowledgeQA, error) {
	question := strings.TrimSpace(request.Message)
	if question == "" {
		return nil, ErrKnowledgeQAQuestionRequired
	}
	if utf8.RuneCountInString(question) > KnowledgeQAQuestionCharacterLimit {
		return nil, ErrKnowledgeQAQuestionTooLong
	}
	if s == nil || s.retriever == nil {
		return nil, errors.New("knowledge retriever is not configured")
	}

	history := normalizeKnowledgeQAHistory(request.History)
	hits, err := s.retriever.SearchAll(ctx, buildKnowledgeQARetrievalQuery(question, history), KnowledgeQARetrievalLimit)
	if err != nil {
		return nil, fmt.Errorf("search Wiki knowledge: %w", err)
	}
	evidence := reliableKnowledgeQAEvidence(hits)
	if len(evidence) == 0 {
		return &PreparedKnowledgeQA{
			Sources: []KnowledgeQASource{}, MemoryStatus: "none",
			ImmediateAnswer: &KnowledgeQAAnswer{
				KnowledgeAnswer: "当前 Wiki 中没有检索到足够可靠的相关资料，暂时无法基于现有内容回答这个问题。",
				LearningAdvice:  "暂无相关学习记录",
			},
		}, nil
	}

	memoryStatus := "none"
	memoryItems := make([]*learning_db.LearningMemoryItem, 0)
	if s.memories == nil {
		memoryStatus = "unavailable"
	} else {
		memoryItems, err = s.memories.FindDocumentItemsBatch(ctx, evidenceDocumentIDs(evidence), KnowledgeQAMemoryLimit)
		if err != nil {
			memoryStatus = "unavailable"
			memoryItems = []*learning_db.LearningMemoryItem{}
		} else if hasKnowledgeQAMemory(memoryItems) {
			memoryStatus = "available"
		}
	}

	return &PreparedKnowledgeQA{
		Sources:      knowledgeQASources(evidence),
		MemoryStatus: memoryStatus,
		Prompt: prompts.KnowledgeQAQueryPrompt(prompts.KnowledgeQAQueryInput{
			Question: question, History: promptKnowledgeQAHistory(history),
			Evidence: promptKnowledgeQAEvidence(evidence), MemoryStatus: memoryStatus,
			MemoryItems: promptKnowledgeQAMemory(memoryItems),
		}),
	}, nil
}

func (s *KnowledgeQAService) Generate(ctx context.Context, prepared *PreparedKnowledgeQA) (*KnowledgeQAAnswer, error) {
	if prepared == nil {
		return nil, errors.New("prepared knowledge QA is required")
	}
	if prepared.ImmediateAnswer != nil {
		answer := *prepared.ImmediateAnswer
		return &answer, nil
	}
	if s == nil || s.run == nil {
		return nil, errors.New("knowledge QA model is not configured")
	}
	raw, err := s.run(ctx, prepared.Prompt)
	if err != nil {
		return nil, fmt.Errorf("generate knowledge QA answer: %w", err)
	}
	answer, err := parseKnowledgeQAAnswer(raw)
	if err != nil {
		return nil, err
	}
	switch prepared.MemoryStatus {
	case "none":
		answer.LearningAdvice = "暂无相关学习记录"
	case "unavailable":
		answer.LearningAdvice = "学习记录暂不可用"
	}
	return answer, nil
}

func parseKnowledgeQAAnswer(text string) (*KnowledgeQAAnswer, error) {
	for _, candidate := range jsonObjectCandidates(strings.TrimSpace(text)) {
		var payload struct {
			KnowledgeAnswer string `json:"knowledge_answer"`
			LearningAdvice  string `json:"learning_advice"`
		}
		if err := json.Unmarshal([]byte(candidate), &payload); err != nil {
			continue
		}
		payload.KnowledgeAnswer = strings.TrimSpace(payload.KnowledgeAnswer)
		payload.LearningAdvice = strings.TrimSpace(payload.LearningAdvice)
		if payload.KnowledgeAnswer == "" || payload.LearningAdvice == "" {
			continue
		}
		return &KnowledgeQAAnswer{KnowledgeAnswer: payload.KnowledgeAnswer, LearningAdvice: payload.LearningAdvice}, nil
	}
	return nil, errors.New("knowledge QA output is invalid")
}

func normalizeKnowledgeQAHistory(history []KnowledgeQAMessage) []KnowledgeQAMessage {
	valid := make([]KnowledgeQAMessage, 0, len(history))
	for _, item := range history {
		role := strings.TrimSpace(item.Role)
		content := strings.TrimSpace(item.Content)
		if (role != "user" && role != "assistant") || content == "" {
			continue
		}
		valid = append(valid, KnowledgeQAMessage{Role: role, Content: content})
	}
	if len(valid) > KnowledgeQAHistoryMessageLimit {
		valid = valid[len(valid)-KnowledgeQAHistoryMessageLimit:]
	}

	result := make([]KnowledgeQAMessage, 0, len(valid))
	remaining := KnowledgeQAHistoryCharacterBudget
	for i := len(valid) - 1; i >= 0 && remaining > 0; i-- {
		content := truncateRunes(valid[i].Content, remaining)
		if content == "" {
			break
		}
		result = append(result, KnowledgeQAMessage{Role: valid[i].Role, Content: content})
		remaining -= utf8.RuneCountInString(content)
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func buildKnowledgeQARetrievalQuery(question string, history []KnowledgeQAMessage) string {
	question = truncateRunes(strings.TrimSpace(question), KnowledgeQARetrievalQueryBudget)
	remaining := KnowledgeQARetrievalQueryBudget - utf8.RuneCountInString(question)
	previous := make([]string, 0, 2)
	for i := len(history) - 1; i >= 0 && len(previous) < 2; i-- {
		if history[i].Role != "user" || remaining <= 1 {
			continue
		}
		content := truncateRunes(history[i].Content, remaining-1)
		if content == "" {
			continue
		}
		previous = append(previous, content)
		remaining -= utf8.RuneCountInString(content) + 1
	}
	for left, right := 0, len(previous)-1; left < right; left, right = left+1, right-1 {
		previous[left], previous[right] = previous[right], previous[left]
	}
	return strings.Join(append(previous, question), "\n")
}

func reliableKnowledgeQAEvidence(hits []rag_payload.SearchResult) []rag_payload.SearchResult {
	result := make([]rag_payload.SearchResult, 0, len(hits))
	for _, hit := range hits {
		if hit.Score >= MinimumKnowledgeQAEvidenceScore && strings.TrimSpace(hit.Content) != "" {
			result = append(result, hit)
		}
	}
	return result
}

func evidenceDocumentIDs(evidence []rag_payload.SearchResult) []string {
	result := make([]string, 0, len(evidence))
	seen := make(map[string]struct{}, len(evidence))
	for _, item := range evidence {
		if _, exists := seen[item.DocumentID]; item.DocumentID == "" || exists {
			continue
		}
		seen[item.DocumentID] = struct{}{}
		result = append(result, item.DocumentID)
	}
	return result
}

func knowledgeQASources(evidence []rag_payload.SearchResult) []KnowledgeQASource {
	result := make([]KnowledgeQASource, 0, len(evidence))
	indexes := make(map[string]int, len(evidence))
	for _, item := range evidence {
		key := item.DocumentID + "\x00" + item.HeadingPath
		if index, exists := indexes[key]; exists {
			if item.Score > result[index].Score {
				result[index].Score = item.Score
			}
			continue
		}
		indexes[key] = len(result)
		result = append(result, KnowledgeQASource{
			DocumentID: item.DocumentID, DocumentTitle: item.DocumentTitle,
			FolderPath: item.FolderPath, HeadingPath: item.HeadingPath, Score: item.Score,
		})
	}
	return result
}

func hasKnowledgeQAMemory(items []*learning_db.LearningMemoryItem) bool {
	for _, item := range items {
		if item != nil && strings.TrimSpace(item.Statement) != "" {
			return true
		}
	}
	return false
}

func promptKnowledgeQAHistory(history []KnowledgeQAMessage) []prompts.KnowledgeQAHistoryMessage {
	result := make([]prompts.KnowledgeQAHistoryMessage, 0, len(history))
	for _, item := range history {
		result = append(result, prompts.KnowledgeQAHistoryMessage{Role: item.Role, Content: item.Content})
	}
	return result
}

func promptKnowledgeQAEvidence(evidence []rag_payload.SearchResult) []prompts.KnowledgeQAEvidence {
	result := make([]prompts.KnowledgeQAEvidence, 0, len(evidence))
	for _, item := range evidence {
		result = append(result, prompts.KnowledgeQAEvidence{
			DocumentID: item.DocumentID, DocumentTitle: item.DocumentTitle,
			FolderPath: item.FolderPath, HeadingPath: item.HeadingPath, Score: item.Score, Content: item.Content,
		})
	}
	return result
}

func promptKnowledgeQAMemory(items []*learning_db.LearningMemoryItem) []prompts.KnowledgeQAMemoryItem {
	result := make([]prompts.KnowledgeQAMemoryItem, 0, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.Statement) == "" {
			continue
		}
		documentID := ""
		if item.DocumentID != nil {
			documentID = strings.TrimSpace(*item.DocumentID)
		}
		result = append(result, prompts.KnowledgeQAMemoryItem{
			DocumentID: documentID, Kind: strings.TrimSpace(item.Kind),
			Statement: strings.TrimSpace(item.Statement), Confidence: strings.TrimSpace(item.Confidence),
		})
	}
	return result
}
