# EditTable Component

A CRUD-capable purchases table. Mirrors the read-only `Summary` component but adds create, edit, and delete operations via modals.

---

## File Structure

```
src/components/EditTable/
├── EditTable.tsx           # Main container
├── EditTableRow.tsx        # Table row with Edit / Delete actions
├── PurchaseFormModal.tsx   # Create & Edit dialog form
├── DeleteConfirmModal.tsx  # Delete confirmation dialog
├── EditTable.test.tsx      # 18 integration tests
└── EditTableRow.test.tsx   # 12 unit tests
```

---

## Usage

```tsx
import EditTable from "@/components/EditTable/EditTable";

<EditTable />
```

Drop it into any authenticated route the same way `<Summary />` is used.

---

## Features

| Feature | Detail |
|---------|--------|
| Year / Month filters | Derived from fetched data; month resets when year changes |
| Grand total | Sum of filtered purchases, shown in filter bar and table footer |
| New Purchase | Opens blank form modal |
| Edit Purchase | Opens pre-filled form modal |
| Delete Purchase | Opens confirmation dialog before calling API |
| Inline items sub-table | Shown on desktop (≥ 640 px) below each row |
| Auto-calculated subtotal | Updates per item as Qty × Unit Price is entered |
| "Use items total" shortcut | One-click fills Total Price from the sum of item subtotals |

---

## Data Flow

```
useQuery(getPurchases)
    │
    ▼
EditTable (filters, totals, modal state)
    │
    ├── EditTableRow × N  ──► onEdit  ──► PurchaseFormModal
    │                    └──► onDelete ──► DeleteConfirmModal
    │
    ├── useMutation(createPurchase)  ← form submit (create mode)
    ├── useMutation(updatePurchase)  ← form submit (edit mode)
    └── useMutation(deletePurchase)  ← confirm delete
            │
            └── invalidateQueries(["purchases"]) on success
```

---

## Sub-components

### `EditTableRow`

| Prop | Type | Description |
|------|------|-------------|
| `purchase` | `Purchase` | Row data |
| `onEdit` | `(p: Purchase) => void` | Called when Edit button is clicked |
| `onDelete` | `(p: Purchase) => void` | Called when Delete button is clicked |

### `PurchaseFormModal`

| Prop | Type | Description |
|------|------|-------------|
| `mode` | `"create" \| "edit"` | Controls heading and which mutation fires |
| `purchase` | `Purchase?` | Pre-fills the form in edit mode |
| `onSave` | `(data: PurchaseFormData) => void` | Called on valid form submit |
| `onClose` | `() => void` | Called on Cancel, backdrop click, or Escape |
| `isSaving` | `boolean` | Disables Save button while mutation is pending |

**Exported type:**

```ts
export type PurchaseFormData = {
  title: string;
  buyFrom: string;
  buyAt: string;       // YYYY-MM-DD
  buyPrice: string;
  items: Array<{
    itemName: string;
    itemQty: string;
    unitPrice: string;
    subTotal: string;  // auto-calculated
  }>;
};
```

### `DeleteConfirmModal`

| Prop | Type | Description |
|------|------|-------------|
| `purchase` | `Purchase` | Used to display the purchase title in the prompt |
| `onConfirm` | `() => void` | Called when Delete is confirmed |
| `onCancel` | `() => void` | Called on Cancel, backdrop click, or Escape |
| `isDeleting` | `boolean` | Disables both buttons while mutation is pending |

---

## API Functions (`src/api/upload.tsx`)

```ts
createPurchase(data: PurchasePayload): Promise<Purchase>
// POST /v1/auth/purchases

updatePurchase(id: string, data: PurchasePayload): Promise<Purchase>
// PATCH /v1/auth/purchases/:id

deletePurchase(id: string): Promise<void>
// DELETE /v1/auth/purchases/:id
```

`PurchasePayload` is structurally identical to `PurchaseFormData` — TypeScript accepts either.

---

## Shared Dependencies

| Import | Source |
|--------|--------|
| `Purchase`, `Item`, `formatDate`, `formatCurrency`, `MONTH_NAMES` | `../Summary/mockData` |
| `getPurchases` | `../../api/upload` |
| Icons (`Pencil`, `Trash2`, `Plus`, `X`, `AlertTriangle`) | `lucide-react` |
| `useQuery`, `useMutation`, `useQueryClient` | `@tanstack/react-query` |

---

## Tests

Run with:

```bash
npx vitest run src/components/EditTable
```

**EditTable.test.tsx** (18 tests)
- Loading / error / empty states
- Year and month filter behaviour, month reset on year change
- Opens New / Edit / Delete modals
- Pre-fills edit form from purchase data
- Calls `createPurchase`, `updatePurchase`, `deletePurchase` on submit / confirm
- Closes form modal on Cancel

**EditTableRow.test.tsx** (12 tests)
- Renders title, store, price, item count
- `onEdit` / `onDelete` callback isolation
- Inline items sub-table present / absent based on items array
