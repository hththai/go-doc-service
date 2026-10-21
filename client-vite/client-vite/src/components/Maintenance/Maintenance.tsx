import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import {
  ISSUE_CODES,
  repairIssues,
  verifyDocuments,
  type DocIssue,
  type IssueCode,
  type IssueRef,
  type RepairResult,
} from "../../api/maintenance";
import RepairConfirmModal from "./RepairConfirmModal";

const ISSUE_LABELS: Record<IssueCode, string> = {
  [ISSUE_CODES.duplicateFile]: "Duplicate file",
  [ISSUE_CODES.missingFile]: "Missing file",
  [ISSUE_CODES.metadataMismatch]: "Wrong file details",
  [ISSUE_CODES.multiplePaths]: "Multiple file paths",
};

const buttonClass =
  "px-4 py-1.5 text-sm text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors";

const issueKey = (i: IssueRef) => `${i.objId}:${i.code}`;

export default function Maintenance() {
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [confirming, setConfirming] = useState(false);
  const [repairResult, setRepairResult] = useState<RepairResult | null>(null);

  const verifyMutation = useMutation({
    mutationFn: verifyDocuments,
    onSuccess: () => setSelected(new Set()),
  });

  const repairMutation = useMutation({
    mutationFn: repairIssues,
    onSuccess: (result) => {
      setRepairResult(result);
      setConfirming(false);
      verifyMutation.mutate();
    },
    onError: () => setConfirming(false),
  });

  const issues: DocIssue[] | undefined = verifyMutation.data;
  const repairable = issues?.filter((i) => i.repairable) ?? [];
  const allSelected =
    repairable.length > 0 && repairable.every((i) => selected.has(issueKey(i)));

  function toggle(issue: DocIssue) {
    setSelected((prev) => {
      const next = new Set(prev);
      const key = issueKey(issue);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }

  function toggleAll() {
    setSelected(allSelected ? new Set() : new Set(repairable.map(issueKey)));
  }

  function handleVerify() {
    setRepairResult(null);
    verifyMutation.mutate();
  }

  function handleRepair() {
    const refs = repairable
      .filter((i) => selected.has(issueKey(i)))
      .map(({ objId, code }) => ({ objId, code }));
    repairMutation.mutate(refs);
  }

  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Document Maintenance
      </div>

      <div className="border border-gray-200 rounded-b-md overflow-hidden">
        <div className="bg-gray-50 px-4 py-3 flex flex-wrap items-center gap-4 border-b border-gray-200">
          <p className="text-sm text-gray-600">
            Check your documents for duplicate, missing or mismatched files.
          </p>
          <div className="ml-auto flex gap-3">
            <button
              type="button"
              onClick={handleVerify}
              disabled={verifyMutation.isPending}
              className={buttonClass}
            >
              {verifyMutation.isPending ? "Verifying…" : "Verify my documents"}
            </button>
            {issues && repairable.length > 0 && (
              <button
                type="button"
                onClick={() => setConfirming(true)}
                disabled={selected.size === 0 || repairMutation.isPending}
                className={buttonClass}
              >
                Repair selected ({selected.size})
              </button>
            )}
          </div>
        </div>

        <div className="px-4 py-4">
          {verifyMutation.isError && (
            <p className="text-sm text-red-600">
              Could not verify documents. Please try again.
            </p>
          )}
          {repairMutation.isError && (
            <p className="text-sm text-red-600">
              Could not repair documents. Please try again.
            </p>
          )}

          {repairResult && <RepairSummary result={repairResult} />}

          {!issues && !verifyMutation.isPending && !verifyMutation.isError && (
            <p className="text-sm text-gray-500">
              Click "Verify my documents" to start.
            </p>
          )}

          {issues?.length === 0 && (
            <p className="text-sm text-green-700">No issues found.</p>
          )}

          {issues && issues.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-500 border-b border-gray-200">
                    <th className="py-2 pr-3 w-8">
                      <input
                        type="checkbox"
                        aria-label="Select all repairable issues"
                        checked={allSelected}
                        disabled={repairable.length === 0}
                        onChange={toggleAll}
                      />
                    </th>
                    <th className="py-2 pr-3">Document</th>
                    <th className="py-2 pr-3">Issue</th>
                    <th className="py-2 pr-3">Files</th>
                    <th className="py-2">Details</th>
                  </tr>
                </thead>
                <tbody>
                  {issues.map((issue) => (
                    <tr
                      key={issueKey(issue)}
                      className="border-b border-gray-100 align-top"
                    >
                      <td className="py-2 pr-3">
                        <input
                          type="checkbox"
                          aria-label={`Select ${ISSUE_LABELS[issue.code]} for ${issue.title || issue.objId}`}
                          checked={selected.has(issueKey(issue))}
                          disabled={!issue.repairable}
                          onChange={() => toggle(issue)}
                        />
                      </td>
                      <td className="py-2 pr-3">
                        <span className="font-medium text-gray-900">
                          {issue.title || "Untitled"}
                        </span>
                        <span className="block text-xs text-gray-400">
                          #{issue.objId}
                        </span>
                      </td>
                      <td className="py-2 pr-3 text-gray-700">
                        {ISSUE_LABELS[issue.code] ?? `Issue ${issue.code}`}
                      </td>
                      <td className="py-2 pr-3 text-gray-700">
                        {issue.fileName}
                        {issue.extraFiles?.map((f) => (
                          <span
                            key={f}
                            className="block text-xs text-amber-700"
                          >
                            extra: {f}
                          </span>
                        ))}
                      </td>
                      <td className="py-2 text-gray-500">
                        {issue.message}
                        {!issue.repairable && (
                          <span className="block text-xs text-gray-400">
                            Cannot be repaired automatically.
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {confirming && (
        <RepairConfirmModal
          count={selected.size}
          onConfirm={handleRepair}
          onCancel={() => setConfirming(false)}
          isRepairing={repairMutation.isPending}
        />
      )}
    </div>
  );
}

function RepairSummary({ result }: Readonly<{ result: RepairResult }>) {
  return (
    <div className="mb-4 rounded-md border border-gray-200 bg-gray-50 p-3 text-sm">
      <p className="font-medium text-gray-900">
        Repaired {result.repaired.length}, skipped {result.skipped.length}
      </p>
      <ul className="mt-1 space-y-0.5">
        {result.repaired.map((o) => (
          <li key={`r-${issueKey(o)}`} className="text-green-700">
            #{o.objId}: {o.message}
          </li>
        ))}
        {result.skipped.map((o) => (
          <li key={`s-${issueKey(o)}`} className="text-amber-700">
            #{o.objId} skipped: {o.message}
          </li>
        ))}
      </ul>
    </div>
  );
}
