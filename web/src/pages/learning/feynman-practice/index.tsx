import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { BookOpenIcon, CircleAlertIcon, MessageSquareTextIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  sessionKeys,
  useCompleteSession,
  useCreateSession,
  useReviewExplanation,
  useSessionDetail,
  type LearningExplanationReview,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
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

import { PracticePanel } from "./_components/practice-panel";
import { SourcePanel } from "./_components/source-panel";
import { mergeReviewTurns } from "./_lib/review-turns";

const routeApi = getRouteApi("/_layout/learn/feynman-practice/$documentId");

export function FeynmanWorkbenchPage() {
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
  const answerRef = useRef("");

  const [createAttempt, setCreateAttempt] = useState(0);
  const [createdSession, setCreatedSession] = useState<{
    routeIdentity: string;
    id: string;
  } | null>(null);
  const [creatingIdentity, setCreatingIdentity] = useState("");
  const [createFailureIdentity, setCreateFailureIdentity] = useState("");
  const [activeView, setActiveView] = useState("source");
  const [answer, setAnswer] = useState("");
  const [turns, setTurns] = useState<LearningExplanationReview[]>([]);
  const [completedSummary, setCompletedSummary] = useState("");
  const [reviewingIdentity, setReviewingIdentity] = useState("");
  const [completingIdentity, setCompletingIdentity] = useState("");

  const createdSessionId = createdSession?.routeIdentity === routeIdentity ? createdSession.id : "";
  const sessionId = requestedSessionId || createdSessionId;
  const isCreating = creatingIdentity === routeIdentity;
  const createFailed = createFailureIdentity === routeIdentity;
  const requestIdentity = `${documentId}:${sessionId}`;
  currentRouteIdentityRef.current = routeIdentity;
  currentRequestIdentityRef.current = requestIdentity;
  answerRef.current = answer;

  const { data: document, isLoading: isDocumentLoading } = useQuery({
    queryKey: ["wiki-document", documentId],
    queryFn: () => documentApi.findOne(documentId),
    enabled: !!documentId,
  });
  const sessionDetailQuery = useSessionDetail(sessionId);
  const sessionDetail = sessionDetailQuery.data;
  const reviewExplanation = useReviewExplanation(sessionId);
  const completeSession = useCompleteSession(sessionId);

  useEffect(() => {
    if (previousRouteIdentityRef.current === routeIdentity) return;
    previousRouteIdentityRef.current = routeIdentity;
    setCreatedSession(null);
    setCreatingIdentity("");
    setCreateFailureIdentity("");
    setActiveView("source");
    setAnswer("");
    setTurns([]);
    setCompletedSummary("");
    setReviewingIdentity("");
    setCompletingIdentity("");
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
    documentId,
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

    setTurns((current) => mergeReviewTurns(current, sessionDetail.reviews || []));
    setCompletedSummary(
      sessionDetail.session.status === "completed"
        ? sessionDetail.session.summary || "你的解释记录已经保存。"
        : "",
    );
  }, [documentId, sessionDetail, sessionId]);

  const submitExplanation = async () => {
    const submittedAnswer = answer;
    const explanation = submittedAnswer.trim();
    const submittedIdentity = requestIdentity;
    if (!sessionReady || !explanation || reviewingIdentity === submittedIdentity) return;

    setReviewingIdentity(submittedIdentity);
    try {
      const review = await reviewExplanation.mutateAsync({
        request_id: crypto.randomUUID(),
        explanation,
      });
      if (currentRequestIdentityRef.current !== submittedIdentity) return;

      const createdAt = new Date().toISOString();
      setTurns((current) => [
        ...current,
        {
          ...review,
          id: `local-${createdAt}-${current.length}`,
          turn_id: "",
          session_id: sessionId,
          document_id: documentId,
          user_id: "",
          explanation,
          created_at: createdAt,
        },
      ]);
      if (answerRef.current === submittedAnswer) {
        setAnswer("");
      }
      void queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
    } catch (error) {
      if (currentRequestIdentityRef.current === submittedIdentity) {
        toast.error(error instanceof Error ? error.message : "审阅解释失败，请重试");
      }
    } finally {
      if (currentRequestIdentityRef.current === submittedIdentity) {
        setReviewingIdentity("");
      }
    }
  };

  const finishPractice = async () => {
    const submittedIdentity = requestIdentity;
    if (
      !sessionReady ||
      turns.length === 0 ||
      answer.trim() ||
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

  const retryCreateSession = () => {
    creatingRouteRef.current = "";
    setCreateFailureIdentity("");
    setCreateAttempt((attempt) => attempt + 1);
  };

  const startNewSession = () => {
    navigate({
      to: "/learn/feynman-practice/$documentId",
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
  const isReviewing = reviewingIdentity === requestIdentity;
  const isCompleting = completingIdentity === requestIdentity;
  const isCompleted = Boolean(completedSummary || sessionDetail.session.status === "completed");
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
        value={activeView}
        onValueChange={setActiveView}
        className="min-h-0 flex-1 overflow-hidden"
      >
        <TabsList>
          <TabsTrigger value="source">
            <BookOpenIcon />
            阅读文章
          </TabsTrigger>
          <TabsTrigger value="explain">
            <MessageSquareTextIcon />
            开始讲解{turns.length > 0 ? ` (${turns.length})` : ""}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="source" className="flex min-h-0 overflow-hidden">
          <SourcePanel documentId={documentId} />
        </TabsContent>
        <TabsContent value="explain" className="flex min-h-0 overflow-hidden">
          <PracticePanel
            answer={answer}
            turns={turns}
            disabled={!sessionReady || isReviewing || isCompleting || isCompleted}
            isSubmitting={isReviewing}
            isCompleting={isCompleting}
            isCompleted={isCompleted}
            hasDraft={answer.trim().length > 0}
            completedSummary={completedSummary}
            onAnswerChange={setAnswer}
            onSubmit={() => void submitExplanation()}
            onComplete={() => void finishPractice()}
          />
        </TabsContent>
      </Tabs>
    </div>
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
