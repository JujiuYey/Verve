import { describe, expect, it } from "vitest";

import { createActionStreamFilter } from "./action-stream-filter";

describe("createActionStreamFilter", () => {
  it("emits ordinary text immediately", () => {
    const filter = createActionStreamFilter();

    expect(filter.push("先看整体，再看细节。 ")).toBe("先看整体，再看细节。 ");
    expect(filter.push("继续解释。")).toBe("继续解释。");
  });

  it("holds a split opening tag and strips the action block", () => {
    const filter = createActionStreamFilter();

    expect(filter.push("可以开始练习了。<ACT")).toBe("可以开始练习了。");
    expect(filter.push('ION>{"type":"navigate_to_practice"}')).toBe("");
    expect(filter.push("</ACTION>")).toBe("");
  });

  it("strips JSON and a closing tag split across chunks", () => {
    const filter = createActionStreamFilter();

    expect(filter.push('正文<ACTION>{"document_id":')).toBe("正文");
    expect(filter.push('"doc-1"}</ACT')).toBe("");
    expect(filter.push("ION>尾声")).toBe("尾声");
  });

  it("strips multiple action blocks while preserving surrounding text", () => {
    const filter = createActionStreamFilter();

    expect(
      filter.push(
        '甲<ACTION>{"document_id":"doc-1"}</ACTION>乙' +
          '<ACTION>{"document_id":"doc-2"}</ACTION>丙',
      ),
    ).toBe("甲乙丙");
  });

  it("preserves less-than text that is not an action tag", () => {
    const filter = createActionStreamFilter();

    expect(filter.push("a < b，<AN 也不是标签")).toBe("a < b，<AN 也不是标签");
  });

  it("flushes a safe partial opening prefix at the end", () => {
    const filter = createActionStreamFilter();

    expect(filter.push("正文<ACT")).toBe("正文");
    expect(filter.flush()).toBe("<ACT");
    expect(filter.flush()).toBe("");
  });

  it("discards an unfinished action block at the end", () => {
    const filter = createActionStreamFilter();

    expect(filter.push('正文<ACTION>{"document_id":"doc-1"}')).toBe("正文");
    expect(filter.flush()).toBe("");
  });
});
