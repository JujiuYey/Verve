import { describe, expect, it } from "vitest";

import { parseKnowledgeQAEvent } from "./knowledge-qa";

describe("parseKnowledgeQAEvent", () => {
  it("parses the structured event vocabulary", () => {
    expect(parseKnowledgeQAEvent('{"type":"status","phase":"generating"}')).toEqual({
      type: "status",
      phase: "generating",
    });
    expect(
      parseKnowledgeQAEvent('{"type":"answer","knowledgeAnswer":"回答","learningAdvice":"建议"}'),
    ).toEqual({ type: "answer", knowledgeAnswer: "回答", learningAdvice: "建议" });
    expect(
      parseKnowledgeQAEvent(
        '{"type":"sources","sources":[{"documentId":"doc-1","documentTitle":"database.md","folderPath":"数据库","headingPath":"事务","score":0.91}]}',
      ),
    ).toMatchObject({ type: "sources", sources: [{ documentId: "doc-1", score: 0.91 }] });
  });

  it("rejects legacy and malformed events", () => {
    expect(() => parseKnowledgeQAEvent('{"type":"tool_call"}')).toThrow("知识问答事件格式错误");
    expect(() => parseKnowledgeQAEvent('{"type":"answer","knowledgeAnswer":"回答"}')).toThrow(
      "知识问答事件格式错误",
    );
    expect(() =>
      parseKnowledgeQAEvent('{"type":"sources","sources":[{"documentId":"doc-1"}]}'),
    ).toThrow("知识问答事件格式错误");
  });
});
