import { useQuery } from "@tanstack/react-query";
import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { BookOpenIcon, CircleAlertIcon, MessageSquareTextIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  useCompleteSession,
  useCreateSession,
  useReviewExplanation,
  useSessionDetail,
  type LearningExplanationReview,
} from "@/api/learning";
import { documentApi } from "@/api/wiki/document";
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

import { PracticePanel } from "./_components/practice-panel";
import { SourcePanel } from "./_components/source-panel";

const routeApi = getRouteApi("/_layout/learn/feynman-practice/$documentId");

export function FeynmanWorkbenchPage() {
  const navigate = useNavigate();
  const { documentId } = routeApi.useParams();
  const createSession = useCreateSession();
  const creatingDocumentRef = useRef("");
  const currentDocumentRef = useRef(documentId);
  const previousDocumentRef = useRef(documentId);
  const [createAttempt, setCreateAttempt] = useState(0);
  const [sessionId, setSessionId] = useState("");
  const [activeView, setActiveView] = useState("source");
  const [answer, setAnswer] = useState("");
  const [turns, setTurns] = useState<LearningExplanationReview[]>([]);
  const [completedSummary, setCompletedSummary] = useState("");

  currentDocumentRef.current = documentId;

  const { data: document, isLoading: isDocumentLoading } = useQuery({
    queryKey: ["wiki-document", documentId],
    queryFn: () => documentApi.findOne(documentId),
    enabled: !!documentId,
  });
  const { data: sessionDetail, isError: isSessionDetailError } = useSessionDetail(sessionId);
  const reviewExplanation = useReviewExplanation(sessionId);
  const completeSession = useCompleteSession(sessionId);

  useEffect(() => {
    if (previousDocumentRef.current === documentId) return;
    previousDocumentRef.current = documentId;
    setSessionId("");
    setActiveView("source");
    setAnswer("");
    setTurns([]);
    setCompletedSummary("");
    creatingDocumentRef.current = "";
  }, [documentId]);

  useEffect(() => {
    if (!documentId || sessionId || creatingDocumentRef.current === documentId) return;
    creatingDocumentRef.current = documentId;
    createSession
      .mutateAsync({ document_id: documentId })
      .then((result) => {
        if (currentDocumentRef.current === documentId) {
          setSessionId(result.session_id);
        }
      })
      .catch(() => {
        // Keep the document marked until the user explicitly retries.
      });
  }, [createAttempt, createSession, documentId, sessionId]);

  useEffect(() => {
    if (!sessionDetail) return;
    setTurns(sessionDetail.reviews || []);
    if (sessionDetail.session.status === "completed") {
      setCompletedSummary(sessionDetail.session.summary || "");
    }
  }, [sessionDetail]);

  const submitExplanation = async () => {
    const explanation = answer.trim();
    if (!sessionId || !explanation) return;

    try {
      const review = await reviewExplanation.mutateAsync({ explanation });
      const createdAt = new Date().toISOString();
      setTurns((current) => [
        ...current,
        {
          ...review,
          id: `local-${createdAt}-${current.length}`,
          session_id: sessionId,
          document_id: documentId,
          user_id: "",
          explanation,
          created_at: createdAt,
        },
      ]);
      setAnswer("");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "审阅解释失败，请重试");
    }
  };

  const finishPractice = async () => {
    if (!sessionId || turns.length === 0) return;
    try {
      const result = await completeSession.mutateAsync();
      setCompletedSummary(result.summary || "你的解释记录已经保存。");
      toast.success("本次费曼练习已结束");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "结束练习失败");
    }
  };

  const retryCreateSession = () => {
    createSession.reset();
    creatingDocumentRef.current = "";
    setCreateAttempt((attempt) => attempt + 1);
  };

  if (createSession.isError && !sessionId) {
    return (
      <Empty className="h-full border-0 p-6">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <CircleAlertIcon />
          </EmptyMedia>
          <EmptyTitle>无法开始这次练习</EmptyTitle>
          <EmptyDescription>文档可能不存在，或者当前账号没有访问权限。</EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <div className="flex flex-wrap justify-center gap-2">
            <Button variant="outline" onClick={() => navigate({ to: "/wiki" })}>
              返回 Wiki
            </Button>
            <Button onClick={retryCreateSession}>重试</Button>
          </div>
        </EmptyContent>
      </Empty>
    );
  }

  if (isDocumentLoading || (!sessionId && createSession.isPending)) {
    return <WorkbenchSkeleton />;
  }

  const title = document?.filename?.replace(/\.md$/i, "") || "整篇费曼练习";
  const isCompleted = Boolean(completedSummary || sessionDetail?.session.status === "completed");

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

      {isSessionDetailError ? (
        <Alert>
          <CircleAlertIcon />
          <AlertTitle>历史记录读取失败</AlertTitle>
          <AlertDescription>
            <p>你仍然可以继续解释，新提交会保存在当前会话中。</p>
          </AlertDescription>
        </Alert>
      ) : null}

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
        <TabsContent value="source" className="min-h-0 overflow-hidden">
          <SourcePanel documentId={documentId} />
        </TabsContent>
        <TabsContent value="explain" className="min-h-0 overflow-hidden">
          <PracticePanel
            answer={answer}
            turns={turns}
            disabled={!sessionId || reviewExplanation.isPending || isCompleted}
            isSubmitting={reviewExplanation.isPending}
            isCompleting={completeSession.isPending}
            isCompleted={isCompleted}
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
