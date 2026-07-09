import type { ChatStatus } from "ai";
import { BotIcon, PlayIcon, SearchIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { ragWikiApi, type WikiKnowledgeSearchResult } from "@/api/rag/wiki";
import type { WikiAgentInstance } from "@/api/wiki/agent-instance";
import {
  Conversation,
  ConversationContent,
  ConversationEmptyState,
  ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import { Message, MessageContent, MessageResponse } from "@/components/ai-elements/message";
import {
  PromptInput,
  PromptInputBody,
  PromptInputFooter,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputTools,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

interface KnowledgeSearchPanelProps {
  agentInstance?: WikiAgentInstance | null;
  agentLoading?: boolean;
  onStartLearning?: () => void;
  rootFolderId?: string;
  scopeLabel?: string;
}

function formatScore(score: number) {
  return `${Math.round(score * 100)}%`;
}

export function KnowledgeSearchPanel({
  agentInstance,
  agentLoading,
  onStartLearning,
  rootFolderId,
  scopeLabel,
}: KnowledgeSearchPanelProps) {
  const [results, setResults] = useState<WikiKnowledgeSearchResult[]>([]);
  const [lastQuery, setLastQuery] = useState("");
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(false);
  const status: ChatStatus = loading ? "submitted" : "ready";

  const handleSearch = async (message: PromptInputMessage) => {
    if (!rootFolderId) {
      toast.error("请先进入一个知识库目录");
      return;
    }
    const nextQuery = message.text.trim();
    if (!nextQuery) return;

    setLoading(true);
    setSearched(true);
    setLastQuery(nextQuery);
    try {
      const nextResults = await ragWikiApi.search({
        root_folder_id: rootFolderId,
        query: nextQuery,
        limit: 6,
      });
      setResults(nextResults);
    } catch (error) {
      setResults([]);
      toast.error(error instanceof Error ? error.message : "知识检索失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex h-full min-h-[460px] flex-col overflow-hidden rounded-lg border bg-background">
      <div className="flex items-center gap-2 border-b px-4 py-3">
        <BotIcon className="size-4 text-muted-foreground" />
        <div className="text-sm font-medium">{agentInstance?.name || "知识检索"}</div>
        <Badge variant="outline">{rootFolderId ? scopeLabel || "当前知识库" : "未选择知识库"}</Badge>
        {agentLoading ? <Badge variant="secondary">创建 Agent</Badge> : null}
        {agentInstance ? <Badge variant="secondary">{agentInstance.agent_key}</Badge> : null}
        {loading ? <Badge variant="secondary">检索中</Badge> : null}
        {agentInstance && onStartLearning ? (
          <Button className="ml-auto" size="sm" variant="outline" onClick={onStartLearning}>
            <PlayIcon className="size-4" />
            继续学习
          </Button>
        ) : null}
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <Conversation className="min-h-full">
          <ConversationContent className="min-h-full gap-5 p-4">
            {!searched ? (
              <ConversationEmptyState
                description={
                  rootFolderId
                    ? "输入概念、问题或关键词，当前 Wiki Agent 会检索这个知识库里已经解析的文档片段。"
                    : "进入一个知识库目录后可检索已解析内容。"
                }
                icon={<SearchIcon className="size-8" />}
                title={agentInstance?.name || "检索知识库上下文"}
              />
            ) : (
              <>
                <Message from="user">
                  <MessageContent>
                    <div className="whitespace-pre-wrap text-sm leading-6">{lastQuery}</div>
                  </MessageContent>
                </Message>

                <Message from="assistant">
                  <MessageContent className="w-full">
                    {loading ? (
                      <span className="text-sm text-muted-foreground">正在检索相关片段...</span>
                    ) : results.length === 0 ? (
                      <span className="text-sm text-muted-foreground">没有找到相关片段</span>
                    ) : (
                      <div className="space-y-3">
                        {results.map((result) => (
                          <div key={result.chunk_id} className="rounded-md border bg-muted/20 p-3">
                            <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
                              <span className="font-medium text-foreground">
                                {result.document_title}
                              </span>
                              <span>匹配度 {formatScore(result.score)}</span>
                            </div>
                            <div className="mt-1 text-xs text-muted-foreground">
                              {result.folder_path}
                              {result.heading_path ? ` / ${result.heading_path}` : ""}
                            </div>
                            <MessageResponse className="mt-2 line-clamp-4 max-w-none text-sm leading-6">
                              {result.content}
                            </MessageResponse>
                          </div>
                        ))}
                      </div>
                    )}
                  </MessageContent>
                </Message>
              </>
            )}
          </ConversationContent>
          <ConversationScrollButton />
        </Conversation>
      </ScrollArea>

      <div className="border-t p-4">
        <PromptInput onSubmit={handleSearch}>
          <PromptInputBody>
            <PromptInputTextarea
              className="min-h-11"
              disabled={!rootFolderId || loading}
              placeholder="搜索概念、问题或关键词..."
            />
          </PromptInputBody>
          <PromptInputFooter>
            <PromptInputTools>
              <Badge variant="outline">search_wiki_knowledge</Badge>
            </PromptInputTools>
            <PromptInputSubmit disabled={!rootFolderId || loading} status={status} />
          </PromptInputFooter>
        </PromptInput>
      </div>
    </div>
  );
}
