# Test Report — Report Component

**File:** `src/components/Report/Report.test.tsx`
**Runner:** Vitest v3.2.4
**Last run:** 2026-02-27
**Result:** 7 passed, 0 failed

---

## Setup

| Item | Detail |
|------|--------|
| Test framework | Vitest |
| Render helper | @testing-library/react |
| Query wrapper | `QueryClientProvider` (retry disabled) |
| `recharts` mock | `ResponsiveContainer` replaced with a plain `<div>` — jsdom has no layout engine, so the real component returns zero dimensions and cannot render the chart |
| API mock | `@/api/upload.getPurchases` mocked via `vi.mock` |

---

## Test Cases

### 1. Loading state
**ID:** RT-01
**Description:** Shows a loading indicator while the fetch is pending.
**Input:** `getPurchases` returns a promise that never resolves.
**Expected:** Element matching `/loading/i` is in the document.
**Result:** PASS

---

### 2. Error state
**ID:** RT-02
**Description:** Shows an error message when the fetch rejects.
**Input:** `getPurchases` rejects with `Error("Network error")`.
**Expected:** Element matching `/failed to load data/i` is in the document.
**Result:** PASS

---

### 3. Empty state
**ID:** RT-03
**Description:** Shows an empty-state message when the server returns no purchases.
**Input:** `getPurchases` resolves with `[]`.
**Expected:** Element matching `/no expense data/i` is in the document.
**Result:** PASS

---

### 4. Grand total — all years
**ID:** RT-04
**Description:** Displays the correct sum across all purchases when no year filter is active.
**Input:** Three purchases — `$87.45` (Jan 2025), `$134.90` (Feb 2025), `$50.00` (Jun 2024).
**Expected:** `$272.35` is visible.
**Calculation:** 87.45 + 134.90 + 50.00 = 272.35
**Result:** PASS

---

### 5. Year filter options
**ID:** RT-05
**Description:** The year dropdown is populated from the dates present in the data.
**Input:** Purchases spanning 2024 and 2025.
**Expected:** `<option>` elements for both `2024` and `2025` exist.
**Result:** PASS

---

### 6. Year filter updates total
**ID:** RT-06
**Description:** Selecting a specific year recalculates the displayed grand total to include only matching purchases.
**Input:** Same three purchases as RT-04; user selects year `2025`.
**Expected:** Total updates to `$222.35` (87.45 + 134.90).
**Result:** PASS

---

### 7. Selected year with no data
**ID:** RT-07
**Description:** When the active year filter matches no purchases, the empty-state message does not appear (the year option itself is not in the list, so the chart simply shows no data points without showing the empty message).
**Input:** One purchase dated 2024; user selects year `2025` via `fireEvent`.
**Expected:** `/no expense data/i` is **not** in the document.
**Result:** PASS

---

## Coverage Notes

| Area | Covered |
|------|---------|
| Loading / error / empty states | Yes |
| Grand total calculation | Yes |
| Year filter — option population | Yes |
| Year filter — total recalculation | Yes |
| Chart rendering (SVG/canvas) | No — mocked at `ResponsiveContainer` level; chart internals not asserted |
| Custom tooltip (`ChartTooltip`) | No — tooltip only activates on pointer hover, not testable in jsdom |
| `buildChartData` grouping logic | Indirectly via total assertions |

---

## Known Limitations

- **Recharts `ResponsiveContainer`** is mocked because jsdom does not implement CSS layout. The chart SVG is rendered but not size-constrained; pixel-level assertions are not possible.
- **Tooltip** visibility requires a real pointer event and DOM dimensions; unit testing it is out of scope. Consider an E2E test (e.g. Playwright) if tooltip content needs to be verified.
