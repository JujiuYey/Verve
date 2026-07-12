import { CircleAlertIcon } from "lucide-react";

import type { WikiDocumentChangeRequest } from "@/api/learning";
import {
  Artifact,
  ArtifactContent,
  ArtifactDescription,
  ArtifactHeader,
  ArtifactTitle,
} from "@/components/ai-elements/artifact";
import {
  Confirmation,
  ConfirmationAccepted,
  ConfirmationAction,
  ConfirmationActions,
  ConfirmationRejected,
  ConfirmationRequest,
  ConfirmationTitle,
} from "@/components/ai-elements/confirmation";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export function CuratorArtifact({
  request,
  busy,
  onApply,
  onCancel,
  onRegenerate,
}: {
  request: WikiDocumentChangeRequest;
  busy: boolean;
  onApply: () => void;
  onCancel: () => void;
  onRegenerate: () => void;
}) {
  const approval =
    request.status === "proposed"
      ? { id: request.id }
      : { id: request.id, approved: request.status === "applied" };
  const state =
    request.status === "proposed"
      ? "approval-requested"
      : request.status === "applied"
        ? "output-available"
        : "output-denied";
  return (
    <div className="flex flex-col gap-3">
      <Artifact>
        <ArtifactHeader>
          <div>
            <ArtifactTitle>{request.change_summary}</ArtifactTitle>
            <ArtifactDescription>基于文档 v{request.base_version}</ArtifactDescription>
          </div>
        </ArtifactHeader>
        <ArtifactContent>
          <pre className="max-h-72 overflow-auto whitespace-pre-wrap font-mono text-xs leading-5">
            {request.proposed_diff}
          </pre>
        </ArtifactContent>
      </Artifact>
      {request.status === "conflict" || request.status === "failed" ? (
        <Alert>
          <CircleAlertIcon />
          <AlertTitle>
            {request.status === "conflict" ? "文档版本已经变化" : "修订建议处理失败"}
          </AlertTitle>
          <AlertDescription className="flex flex-col gap-2">
            <span>{request.error_message || "请基于当前文档重新生成建议。"}</span>
            <Button
              size="sm"
              variant="outline"
              className="self-start"
              onClick={onRegenerate}
              disabled={busy}
            >
              重新生成
            </Button>
          </AlertDescription>
        </Alert>
      ) : (
        <Confirmation approval={approval} state={state}>
          <ConfirmationTitle>
            <ConfirmationRequest>确认后会创建新的 Wiki 文档版本。</ConfirmationRequest>
            <ConfirmationAccepted>修订已应用，RAG 正在处理新版本。</ConfirmationAccepted>
            <ConfirmationRejected>这条修订建议已取消。</ConfirmationRejected>
          </ConfirmationTitle>
          <ConfirmationActions>
            <ConfirmationAction variant="outline" onClick={onCancel} disabled={busy}>
              取消
            </ConfirmationAction>
            <ConfirmationAction onClick={onApply} disabled={busy}>
              应用修订
            </ConfirmationAction>
          </ConfirmationActions>
        </Confirmation>
      )}
    </div>
  );
}
