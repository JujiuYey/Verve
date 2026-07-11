const OPEN_TAG = "<ACTION>";
const CLOSE_TAG = "</ACTION>";

export interface ActionStreamFilter {
  push(chunk: string): string;
  flush(): string;
}

export function createActionStreamFilter(): ActionStreamFilter {
  let insideAction = false;
  let pending = "";

  return {
    push(chunk) {
      let input = pending + chunk;
      let output = "";
      pending = "";

      while (input) {
        const tag = insideAction ? CLOSE_TAG : OPEN_TAG;
        const tagStart = input.indexOf("<");

        if (tagStart === -1) {
          if (!insideAction) output += input;
          break;
        }

        if (tagStart > 0) {
          if (!insideAction) output += input.slice(0, tagStart);
          input = input.slice(tagStart);
        }

        if (tag.startsWith(input)) {
          pending = input;
          break;
        }

        if (input.startsWith(tag)) {
          input = input.slice(tag.length);
          insideAction = !insideAction;
          continue;
        }

        if (!insideAction) output += "<";
        input = input.slice(1);
      }

      return output;
    },

    flush() {
      const output = insideAction ? "" : pending;
      insideAction = false;
      pending = "";
      return output;
    },
  };
}
