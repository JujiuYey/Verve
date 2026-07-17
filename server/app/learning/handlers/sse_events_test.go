package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteSSEEventWritesFlatTypedPayload(t *testing.T) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	if err := writeSSEEvent(writer, SSEStatus, map[string]interface{}{"phase": "generating"}); err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(buffer.String())
	if !strings.HasPrefix(line, "data: ") {
		t.Fatalf("frame = %q", line)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["type"] != SSEStatus || payload["phase"] != "generating" {
		t.Fatalf("payload = %#v", payload)
	}
}
