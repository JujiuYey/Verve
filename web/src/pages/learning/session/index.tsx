import { useNavigate, useParams } from "@tanstack/react-router";
import { CircleAlertIcon, LoaderCircleIcon } from "lucide-react";
import { useEffect } from "react";

import { useSessionDetail } from "@/api/learning";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export function SessionPage() {
  const { sessionId } = useParams({ from: "/_layout/learn/session/$sessionId" });
  const navigate = useNavigate();
  const { data: detail, isError } = useSessionDetail(sessionId);

  useEffect(() => {
    if (!detail?.session.document_id) return;
    navigate({
      to: "/learn/feynman-practice/$documentId",
      params: { documentId: detail.session.document_id },
      search: { sessionId },
      replace: true,
    });
  }, [detail, navigate, sessionId]);

  return (
    <Empty className="h-full border-0 p-6">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          {isError ? <CircleAlertIcon /> : <LoaderCircleIcon className="animate-spin" />}
        </EmptyMedia>
        <EmptyTitle>{isError ? "会话无法打开" : "正在打开文章练习"}</EmptyTitle>
        <EmptyDescription>
          {isError
            ? "这个会话可能已被删除，或者当前账号没有访问权限。"
            : "学习会话现在直接绑定整篇 Wiki 文章。"}
        </EmptyDescription>
      </EmptyHeader>
      {isError ? (
        <EmptyContent>
          <Button variant="outline" onClick={() => navigate({ to: "/learn/feynman" })}>
            返回学习 Agent
          </Button>
        </EmptyContent>
      ) : null}
    </Empty>
  );
}
