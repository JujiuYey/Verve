package service

import (
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func collectText(iter *adk.AsyncIterator[*adk.AgentEvent]) (string, error) {
	var sb strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		output := event.Output.MessageOutput
		if output.Message != nil && output.Message.Role != schema.Tool {
			sb.WriteString(output.Message.Content)
		}
	}
	return sb.String(), nil
}

func truncateForAgentLog(text string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxRunes]) + "..."
}
