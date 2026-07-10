package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "verve/app/learning/models/db"
	wiki_db "verve/app/wiki/models/db"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

type ObjectiveGenerationService struct {
	db *database.DatabaseService
}

func NewObjectiveGenerationService(db *database.DatabaseService) *ObjectiveGenerationService {
	return &ObjectiveGenerationService{db: db}
}

type GeneratedObjective struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type ObjectiveGenerationResult struct {
	Objectives []GeneratedObjective `json:"objectives"`
}

func (s *ObjectiveGenerationService) GenerateFromMarkdown(ctx context.Context, userID string, doc *wiki_db.Document, folder *wiki_db.Folder, markdown string) ([]*learning_db.LearningObjective, error) {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil, errors.New("markdown 为空")
	}

	agent, err := llm.NewObjectiveGeneratorAgent(ctx)
	if err != nil {
		log.Printf("❌ Objective Generator 初始化失败: document_id=%s err=%v", doc.ID, err)
		return nil, err
	}

	query := buildObjectiveGenerationQuery(doc, folder, markdown)
	log.Printf("📚 Objective Generator 开始生成: document_id=%s filename=%q markdown_chars=%d", doc.ID, doc.Filename, len(markdown))

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		log.Printf("❌ Objective Generator 调用失败: document_id=%s err=%v", doc.ID, err)
		return nil, err
	}
	log.Printf("📚 Objective Generator 原始输出: document_id=%s output_chars=%d output_preview=%q", doc.ID, len(text), truncateForAgentLog(text, 1200))

	result, err := parseObjectiveGenerationOutput(text)
	if err != nil {
		log.Printf("❌ Objective Generator JSON 解析失败: document_id=%s err=%v raw=%q", doc.ID, err, truncateForAgentLog(text, 2000))
		return nil, err
	}
	if len(result.Objectives) == 0 {
		return nil, errors.New("未生成学习小节")
	}

	_, total, err := s.db.Objectives.CountByFolder(ctx, doc.FolderID)
	if err != nil {
		return nil, err
	}

	stageTitle := strings.TrimSpace(doc.Filename)
	folderID := doc.FolderID
	docID := doc.ID
	folderPath := strings.TrimSpace(folder.Name)
	objectives := make([]*learning_db.LearningObjective, 0, len(result.Objectives))
	for i, item := range result.Objectives {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		detail := strings.TrimSpace(item.Detail)
		obj := &learning_db.LearningObjective{
			UserID:           userID,
			StageTitle:       &stageTitle,
			Title:            title,
			SourceDocumentID: &docID,
			SourceFolderID:   &folderID,
			SourceFolderPath: &folderPath,
			OrderIndex:       total + i + 1,
			Status:           "pending",
			MasteryLevel:     "none",
		}
		if detail != "" {
			obj.Detail = &detail
		}
		objectives = append(objectives, obj)
	}
	if len(objectives) == 0 {
		return nil, errors.New("未生成有效学习小节")
	}

	if err := s.db.Objectives.BulkCreate(ctx, objectives); err != nil {
		return nil, err
	}
	return objectives, nil
}

func buildObjectiveGenerationQuery(doc *wiki_db.Document, folder *wiki_db.Folder, markdown string) string {
	var sb strings.Builder
	sb.WriteString("文件夹:")
	sb.WriteString(folder.Name)
	sb.WriteString("\n文档:")
	sb.WriteString(doc.Filename)
	sb.WriteString("\n\nMarkdown 学习资料:\n")
	sb.WriteString(truncateMarkdown(markdown, 24000))
	return sb.String()
}

func truncateMarkdown(markdown string, limit int) string {
	markdown = strings.TrimSpace(markdown)
	if len(markdown) <= limit {
		return markdown
	}
	return markdown[:limit] + "\n\n...(资料过长,已截断)"
}

func parseObjectiveGenerationOutput(text string) (*ObjectiveGenerationResult, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var lastErr error
	for _, candidate := range jsonObjectCandidates(text) {
		var out ObjectiveGenerationResult
		if err := json.Unmarshal([]byte(candidate), &out); err != nil {
			repaired := escapeLikelyUnescapedStringQuotes(candidate)
			if repaired == candidate {
				lastErr = err
				continue
			}
			if repairErr := json.Unmarshal([]byte(repaired), &out); repairErr != nil {
				lastErr = err
				continue
			}
		}
		normalized := make([]GeneratedObjective, 0, len(out.Objectives))
		for _, obj := range out.Objectives {
			obj.Title = strings.TrimSpace(obj.Title)
			obj.Detail = strings.TrimSpace(obj.Detail)
			if obj.Title == "" {
				continue
			}
			normalized = append(normalized, obj)
		}
		if len(normalized) == 0 {
			lastErr = errors.New("JSON 对象没有有效 objectives")
			continue
		}
		out.Objectives = normalized
		return &out, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未找到 JSON 对象")
}

func escapeLikelyUnescapedStringQuotes(input string) string {
	var sb strings.Builder
	sb.Grow(len(input))

	inString := false
	escaped := false
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if !inString {
			sb.WriteByte(ch)
			if ch == '"' {
				inString = true
			}
			continue
		}

		if escaped {
			sb.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			sb.WriteByte(ch)
			escaped = true
			continue
		}
		if ch == '"' {
			if isLikelyStringTerminator(input, i+1) {
				sb.WriteByte(ch)
				inString = false
			} else {
				sb.WriteString(`\"`)
			}
			continue
		}
		sb.WriteByte(ch)
	}

	return sb.String()
}

func isLikelyStringTerminator(input string, start int) bool {
	for i := start; i < len(input); i++ {
		switch input[i] {
		case ' ', '\n', '\r', '\t':
			continue
		case ':', ',', '}', ']':
			return true
		default:
			return false
		}
	}
	return true
}
