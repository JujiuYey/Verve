import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, useNavigate } from "@tanstack/react-router";
import {
  CircleAlertIcon,
  FilePenIcon,
  GraduationCapIcon,
  MessageSquareTextIcon,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  sessionKeys,
  useCompleteSession,
  useCreateSession,
  useSessionDetail,
  useSubmitTurn,
  type LearningAgentType,
  type TimelineItem,
  type WikiDocumentChangeRequest,
} from "@/api/learning";
import { ragWikiKeys, useDocumentIndexStatus, useRetryDocumentIndex } from "@/api/rag/wiki";
import {
  documentApi,
  useApplyWikiChangeRequest,
  useCancelWikiChangeRequest,
} from "@/api/wiki/document";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { AgentComposer } from "./_components/agent-composer";
import { AgentTimeline } from "./_components/agent-timeline";
import { curatorRegenerationInput, filterTimelineByAgent, mergeTimeline } from "./_lib/timeline";

const routeApi = getRouteApi("/_layout/learn/feynman-practice/$documentId");

const agentRoutes = {
  listener: "/learn/feynman-practice/$documentId/listener",
  teacher: "/learn/feynman-practice/$documentId/teacher",
  curator: "/learn/feynman-practice/$documentId/curator",
} as const;

type AgentDrafts = Record<LearningAgentType, string>;

const emptyAgentDrafts = (): AgentDrafts => ({ listener: "", teacher: "", curator: "" });

export function FeynmanWorkbenchPage({ agentType }: { agentType: LearningAgentType }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { documentId } = routeApi.useParams();
  const { sessionId: searchSessionId } = routeApi.useSearch();
  const requestedSessionId = searchSessionId?.trim() || "";
  const routeIdentity = `${documentId}:${requestedSessionId || "new"}`;

  const createSession = useCreateSession();
  const creatingRouteRef = useRef("");
  const currentRouteIdentityRef = useRef(routeIdentity);
  const previousRouteIdentityRef = useRef(routeIdentity);
  const currentRequestIdentityRef = useRef("");
  const draftsRef = useRef<AgentDrafts>(emptyAgentDrafts());

  const [createAttempt, setCreateAttempt] = useState(0);
  const [createdSession, setCreatedSession] = useState<{
    routeIdentity: string;
    id: string;
  } | null>(null);
  const [creatingIdentity, setCreatingIdentity] = useState("");
  const [createFailureIdentity, setCreateFailureIdentity] = useState("");
  const [drafts, setDrafts] = useState<AgentDrafts>(emptyAgentDrafts);
  const [timeline, setTimeline] = useState<TimelineItem[]>([]);
  const [completedSummary, setCompletedSummary] = useState("");
  const [submittingRequestId, setSubmittingRequestId] = useState("");
  const [completingIdentity, setCompletingIdentity] = useState("");
  const [busyChangeRequestId, setBusyChangeRequestId] = useState("");

  const createdSessionId = createdSession?.routeIdentity === routeIdentity ? createdSession.id : "";
  const sessionId = requestedSessionId || createdSessionId;
  const isCreating = creatingIdentity === routeIdentity;
  const createFailed = createFailureIdentity === routeIdentity;
  const requestIdentity = `${documentId}:${sessionId}`;
  currentRouteIdentityRef.current = routeIdentity;
  currentRequestIdentityRef.current = requestIdentity;
  draftsRef.current = drafts;

  const { data: document, isLoading: isDocumentLoading } = useQuery({
    queryKey: ["wiki-document", documentId],
    queryFn: () => documentApi.findOne(documentId),
    enabled: !!documentId,
  });
  const sessionDetailQuery = useSessionDetail(sessionId);
  const sessionDetail = sessionDetailQuery.data;
  const submitTurn = useSubmitTurn(sessionId);
  const completeSession = useCompleteSession(sessionId);
  const applyChangeRequest = useApplyWikiChangeRequest();
  const cancelChangeRequest = useCancelWikiChangeRequest();
  const indexStatusQuery = useDocumentIndexStatus(documentId);
  const retryIndex = useRetryDocumentIndex();

  useEffect(() => {
    if (previousRouteIdentityRef.current === routeIdentity) return;
    previousRouteIdentityRef.current = routeIdentity;
    setCreatedSession(null);
    setCreatingIdentity("");
    setCreateFailureIdentity("");
    setDrafts(emptyAgentDrafts());
    setTimeline([]);
    setCompletedSummary("");
    setSubmittingRequestId("");
    setCompletingIdentity("");
    setBusyChangeRequestId("");
    creatingRouteRef.current = "";
  }, [routeIdentity]);

  useEffect(() => {
    if (
      requestedSessionId ||
      !documentId ||
      createdSessionId ||
      creatingRouteRef.current === routeIdentity
    ) {
      return;
    }

    creatingRouteRef.current = routeIdentity;
    setCreatingIdentity(routeIdentity);
    setCreateFailureIdentity("");
    createSession
      .mutateAsync({ document_id: documentId })
      .then((result) => {
        if (currentRouteIdentityRef.current === routeIdentity) {
          setCreatedSession({ routeIdentity, id: result.session_id });
          void navigate({
            to: agentRoutes[agentType],
            params: { documentId },
            search: { sessionId: result.session_id },
            replace: true,
          });
        }
      })
      .catch(() => {
        if (currentRouteIdentityRef.current === routeIdentity) {
          setCreateFailureIdentity(routeIdentity);
        }
      })
      .finally(() => {
        if (currentRouteIdentityRef.current === routeIdentity) {
          setCreatingIdentity("");
        }
      });
  }, [
    createAttempt,
    createSession,
    createdSessionId,
    agentType,
    documentId,
    navigate,
    requestedSessionId,
    routeIdentity,
  ]);

  useEffect(() => {
    if (
      !sessionDetail ||
      sessionDetail.session.id !== sessionId ||
      sessionDetail.session.document_id !== documentId
    ) {
      return;
    }

    setTimeline((current) => mergeTimeline(current, sessionDetail.timeline || []));
    setCompletedSummary(
      sessionDetail.session.status === "completed"
        ? sessionDetail.session.summary || "你的解释记录已经保存。"
        : "",
    );
  }, [documentId, sessionDetail, sessionId]);

  const submitAgentTurn = async (
    agentType: LearningAgentType,
    contentOverride?: string,
    replacesChangeRequestId?: string,
  ) => {
    const submittedAnswer = contentOverride ?? drafts[agentType];
    const content = submittedAnswer.trim();
    const submittedIdentity = requestIdentity;
    if (!sessionReady || !content || submittingRequestId) return;
    const requestId = crypto.randomUUID();
    const createdAt = new Date().toISOString();
    const localItem = makeLocalTimelineItem(sessionId, requestId, agentType, content, createdAt);
    setSubmittingRequestId(requestId);
    setTimeline((current) => [...current, localItem]);
    try {
      const item = await submitTurn.mutateAsync({
        request_id: requestId,
        agent_type: agentType,
        content,
        replaces_change_request_id: replacesChangeRequestId,
      });
      if (currentRequestIdentityRef.current !== submittedIdentity) return;
      setTimeline((current) => upsertTimelineItem(current, item));
      if (contentOverride === undefined && draftsRef.current[agentType] === submittedAnswer) {
        setDrafts((current) => ({ ...current, [agentType]: "" }));
      }
      void queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
    } catch (error) {
      if (currentRequestIdentityRef.current === submittedIdentity) {
        setTimeline((current) => current.filter((item) => item.turn.request_id !== requestId));
        toast.error(error instanceof Error ? error.message : "处理学习轮次失败，请重试");
      }
    } finally {
      setSubmittingRequestId((current) => (current === requestId ? "" : current));
    }
  };

  const finishPractice = async () => {
    const submittedIdentity = requestIdentity;
    if (
      !sessionReady ||
      listenerTurnCount === 0 ||
      drafts.listener.trim() ||
      completingIdentity === submittedIdentity
    ) {
      return;
    }

    setCompletingIdentity(submittedIdentity);
    try {
      const result = await completeSession.mutateAsync();
      if (currentRequestIdentityRef.current !== submittedIdentity) return;
      setCompletedSummary(result.summary || "你的解释记录已经保存。");
      void queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
      toast.success("本次费曼练习已结束");
    } catch (error) {
      if (currentRequestIdentityRef.current === submittedIdentity) {
        toast.error(error instanceof Error ? error.message : "结束练习失败");
      }
    } finally {
      if (currentRequestIdentityRef.current === submittedIdentity) {
        setCompletingIdentity("");
      }
    }
  };

  const applyCuratorChange = async (item: TimelineItem) => {
    if (item.artifact?.type !== "wiki_change_request") return;
    const id = item.artifact.data.id;
    setBusyChangeRequestId(id);
    try {
      const result = await applyChangeRequest.mutateAsync(id);
      setTimeline((current) => replaceChangeRequest(current, result.change_request));
      void refreshDocumentState();
      toast.success("Wiki 修订已应用");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "应用修订失败");
      void sessionDetailQuery.refetch();
    } finally {
      setBusyChangeRequestId("");
    }
  };

  const cancelCuratorChange = async (item: TimelineItem) => {
    if (item.artifact?.type !== "wiki_change_request") return;
    const request = item.artifact.data;
    setBusyChangeRequestId(request.id);
    try {
      await cancelChangeRequest.mutateAsync(request.id);
      setTimeline((current) =>
        replaceChangeRequest(current, {
          ...request,
          status: "cancelled",
          updated_at: new Date().toISOString(),
        }),
      );
      toast.success("修订建议已取消");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "取消修订失败");
    } finally {
      setBusyChangeRequestId("");
    }
  };

  const regenerateCuratorChange = (item: TimelineItem) => {
    const input = curatorRegenerationInput(item);
    void submitAgentTurn("curator", input.content, input.replaces_change_request_id);
  };

  const refreshDocumentState = () => {
    void queryClient.invalidateQueries({ queryKey: ["wiki-document", documentId] });
    void queryClient.invalidateQueries({ queryKey: ["wiki-document-content", documentId] });
    void queryClient.invalidateQueries({ queryKey: ragWikiKeys.documentStatus(documentId) });
    void queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
  };

  const retryCreateSession = () => {
    creatingRouteRef.current = "";
    setCreateFailureIdentity("");
    setCreateAttempt((attempt) => attempt + 1);
  };

  const startNewSession = () => {
    navigate({
      to: agentRoutes[agentType],
      params: { documentId },
      search: {},
      replace: true,
    });
  };

  if (createFailed && !sessionId) {
    return (
      <SessionProblem
        title="无法开始这次练习"
        description="文档可能不存在，或者当前账号没有访问权限。"
        primaryLabel="重试"
        onPrimary={retryCreateSession}
        onBack={() => navigate({ to: "/wiki" })}
      />
    );
  }

  if (
    isDocumentLoading ||
    isCreating ||
    !sessionId ||
    sessionDetailQuery.isLoading ||
    sessionDetailQuery.isPending
  ) {
    return <WorkbenchSkeleton />;
  }

  if (sessionDetailQuery.isError || !sessionDetail) {
    return (
      <SessionProblem
        title="无法读取练习会话"
        description="会话可能已被删除，或者当前账号没有访问权限。"
        primaryLabel="重试读取"
        onPrimary={() => void sessionDetailQuery.refetch()}
        secondaryLabel={requestedSessionId ? "新建练习" : undefined}
        onSecondary={requestedSessionId ? startNewSession : undefined}
        onBack={() => navigate({ to: "/wiki" })}
      />
    );
  }

  if (sessionDetail.session.document_id !== documentId) {
    return (
      <SessionProblem
        title="会话与文章不匹配"
        description="这个会话属于另一篇文章，不能在当前页面继续。"
        primaryLabel="为本文新建练习"
        onPrimary={startNewSession}
        onBack={() => navigate({ to: "/wiki" })}
      />
    );
  }

  const sessionReady = sessionDetail.session.id === sessionId;
  const isSubmitting = Boolean(submittingRequestId);
  const isCompleting = completingIdentity === requestIdentity;
  const isCompleted = Boolean(completedSummary || sessionDetail.session.status === "completed");
  const itemsByAgent: Record<LearningAgentType, TimelineItem[]> = {
    listener: filterTimelineByAgent(timeline, "listener"),
    teacher: filterTimelineByAgent(timeline, "teacher"),
    curator: filterTimelineByAgent(timeline, "curator"),
  };
  const listenerTurnCount = itemsByAgent.listener.filter(
    (item) => item.artifact?.type === "explanation_review",
  ).length;
  const title = document?.filename?.replace(/\.md$/i, "") || "整篇费曼练习";

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-hidden p-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold">{title}</h1>
          <p className="mt-1 text-sm text-muted-foreground">整篇文章是一轮完整的费曼练习</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={() => navigate({ to: "/wiki" })}>
            返回 Wiki
          </Button>
          <Button variant="secondary" onClick={() => navigate({ to: "/learn/feynman" })}>
            学习 Agent
          </Button>
        </div>
      </div>

      <Tabs
        value={agentType}
        onValueChange={(value) => {
          const nextAgentType = value as LearningAgentType;
          void navigate({
            to: agentRoutes[nextAgentType],
            params: { documentId },
            search: { sessionId },
          });
        }}
        className="min-h-0 flex-1 overflow-hidden"
      >
        <TabsList className="h-auto max-w-full flex-wrap justify-start">
          <TabsTrigger value="listener">
            <MessageSquareTextIcon />
            开始讲解
            {itemsByAgent.listener.length > 0 ? ` (${itemsByAgent.listener.length})` : ""}
          </TabsTrigger>
          <TabsTrigger value="teacher">
            <GraduationCapIcon />
            教学补充{itemsByAgent.teacher.length > 0 ? ` (${itemsByAgent.teacher.length})` : ""}
          </TabsTrigger>
          <TabsTrigger value="curator">
            <FilePenIcon />
            修订文档{itemsByAgent.curator.length > 0 ? ` (${itemsByAgent.curator.length})` : ""}
          </TabsTrigger>
        </TabsList>
        <TabsContent value={agentType} className="flex min-h-0 overflow-hidden">
          <section className="flex min-h-0 flex-1 flex-col overflow-hidden bg-background">
            {agentType === "teacher" &&
            (indexStatusQuery.data?.status === "pending" ||
              indexStatusQuery.data?.status === "running") ? (
              <Alert className="mt-4">
                <CircleAlertIcon />
                <AlertTitle>当前文档版本正在建立索引</AlertTitle>
                <AlertDescription>教学补充暂时不会使用旧版本证据。</AlertDescription>
              </Alert>
            ) : agentType === "teacher" && indexStatusQuery.data?.status === "failed" ? (
              <Alert className="mt-4">
                <CircleAlertIcon />
                <AlertTitle>当前文档版本索引失败</AlertTitle>
                <AlertDescription className="flex items-center justify-between gap-3">
                  <span>{indexStatusQuery.data.error_message || "请重试索引。"}</span>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() =>
                      retryIndex.mutate(documentId, { onSuccess: refreshDocumentState })
                    }
                    disabled={retryIndex.isPending}
                  >
                    重试
                  </Button>
                </AlertDescription>
              </Alert>
            ) : null}
            {isCompleted ? (
              <Alert className="mt-4">
                <AlertTitle>本次练习已结束</AlertTitle>
                <AlertDescription>{completedSummary || "你的解释记录已经保存。"}</AlertDescription>
              </Alert>
            ) : null}
            <AgentTimeline
              agentType={agentType}
              items={itemsByAgent[agentType]}
              busyChangeRequestId={busyChangeRequestId}
              onApply={(item) => void applyCuratorChange(item)}
              onCancel={(item) => void cancelCuratorChange(item)}
              onRegenerate={regenerateCuratorChange}
            />
            <AgentComposer
              agentType={agentType}
              value={drafts[agentType]}
              disabled={!sessionReady || isCompleting || isCompleted}
              isSubmitting={isSubmitting}
              canComplete={agentType === "listener" && !isCompleted && listenerTurnCount > 0}
              isCompleting={isCompleting}
              onChange={(value) => setDrafts((current) => ({ ...current, [agentType]: value }))}
              onSubmit={() => void submitAgentTurn(agentType)}
              onComplete={() => void finishPractice()}
            />
          </section>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function makeLocalTimelineItem(
  sessionId: string,
  requestId: string,
  agentType: LearningAgentType,
  content: string,
  createdAt: string,
): TimelineItem {
  const turnId = `local-${requestId}`;
  return {
    turn: {
      id: turnId,
      session_id: sessionId,
      request_id: requestId,
      agent_type: agentType,
      status: "processing",
      started_at: createdAt,
      created_at: createdAt,
      updated_at: createdAt,
    },
    user_message: {
      id: `${turnId}-user`,
      session_id: sessionId,
      turn_id: turnId,
      role: "user",
      content,
      created_at: createdAt,
    },
  };
}

function upsertTimelineItem(current: TimelineItem[], item: TimelineItem) {
  const index = current.findIndex(
    (entry) => entry.turn.id === item.turn.id || entry.turn.request_id === item.turn.request_id,
  );
  if (index < 0) return [...current, item];
  const next = [...current];
  next[index] = item;
  return next;
}

function replaceChangeRequest(
  current: TimelineItem[],
  request: WikiDocumentChangeRequest,
): TimelineItem[] {
  return current.map((item) =>
    item.artifact?.type === "wiki_change_request" && item.artifact.data.id === request.id
      ? { ...item, artifact: { type: "wiki_change_request", data: request } }
      : item,
  );
}

function SessionProblem({
  title,
  description,
  primaryLabel,
  secondaryLabel,
  onPrimary,
  onSecondary,
  onBack,
}: {
  title: string;
  description: string;
  primaryLabel: string;
  secondaryLabel?: string;
  onPrimary: () => void;
  onSecondary?: () => void;
  onBack: () => void;
}) {
  return (
    <Empty className="h-full border-0 p-6">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <CircleAlertIcon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <div className="flex flex-wrap justify-center gap-2">
          <Button variant="outline" onClick={onBack}>
            返回 Wiki
          </Button>
          {secondaryLabel && onSecondary ? (
            <Button variant="secondary" onClick={onSecondary}>
              {secondaryLabel}
            </Button>
          ) : null}
          <Button onClick={onPrimary}>{primaryLabel}</Button>
        </div>
      </EmptyContent>
    </Empty>
  );
}

function WorkbenchSkeleton() {
  return (
    <div className="flex h-full flex-col gap-4 p-6">
      <div className="flex flex-col gap-2">
        <Skeleton className="h-7 w-64" />
        <Skeleton className="h-4 w-48" />
      </div>
      <Skeleton className="h-9 w-56" />
      <div className="grid min-h-0 flex-1 gap-4 lg:grid-cols-[240px_minmax(0,1fr)]">
        <Skeleton className="h-full" />
        <Skeleton className="h-full" />
      </div>
    </div>
  );
}
