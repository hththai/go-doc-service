**src/api/upload.test.ts — 9 tests**

1. getPurchases: GET request with credentials, absolute fileUrl conversion, missing fileUrl stays undefined, non-ok response throws
2. uploadFile: POST request, missing title throws, non-ok throws, buyAt date format conversion

**src/components/Summary/Summary.test.tsx — 10 tests**

1. Loading and error states
2. Renders all purchases, shows correct grand total
3. Year and month filter options populated from real data
4. Month filtering hides non-matching purchases
5. Year change resets month filter to "all"
6. Year filtering hides purchases from other years
7. Empty array shows "no purchases found"

**src/components/Summary/PurchaseRow.test.tsx — 12 tests**

1. Title, store, price, item count rendered correctly
2. Filename renders as a button when fileUrl present, plain text otherwise
3. onPreview called with correct args on file click, onSelect not triggered by file click
4. onSelect called when row is clicked
5. Item sub-table rendered when items exist, absent when empty
6. Unit price and subtotal shown per item