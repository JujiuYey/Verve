package handlers

import (
	"bufio"
	"encoding/json"
)

const (
	SSEStatus  = "status"
	SSESources = "sources"
	SSEAnswer  = "answer"
	SSEError   = "error"
)

func writeSSEEvent(w *bufio.Writer, eventType string, payload map[string]interface{}) error {
	payload["type"] = eventType
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := w.WriteString("data: "); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.WriteString("\n\n"); err != nil {
		return err
	}
	return w.Flush()
}
