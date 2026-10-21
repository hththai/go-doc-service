export const ISSUE_CODES = {
  duplicateFile: 1,
  missingFile: 2,
  metadataMismatch: 3,
  multiplePaths: 4,
} as const;

export type IssueCode = (typeof ISSUE_CODES)[keyof typeof ISSUE_CODES];

export type DocIssue = {
  objId: number;
  title: string;
  code: IssueCode;
  message: string;
  repairable: boolean;
  fileName: string;
  extraFiles?: string[];
};

export type IssueRef = {
  objId: number;
  code: IssueCode;
};

export type RepairOutcome = IssueRef & {
  message: string;
};

export type RepairResult = {
  repaired: RepairOutcome[];
  skipped: RepairOutcome[];
};

const BASE = `${import.meta.env.VITE_API_URL}/v1/auth/maintenance`;

export async function verifyDocuments(): Promise<DocIssue[]> {
  const res = await fetch(`${BASE}/verify`, { credentials: "include" });
  if (!res.ok) throw new Error("Failed to verify documents");
  const body: { issues: DocIssue[] } = await res.json();
  return body.issues ?? [];
}

export async function repairIssues(issues: IssueRef[]): Promise<RepairResult> {
  const res = await fetch(`${BASE}/repair`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ issues }),
  });
  if (!res.ok) throw new Error("Failed to repair documents");
  return res.json();
}
