package document

import (
	"database/sql"
	"fmt"
	"strconv"
)

// nullableString returns nil for an empty string so numeric DB columns receive NULL instead of "".
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type DocumentRepository interface {
	SetLatestObjId(tx *sql.Tx, document *Document) (int64, error)
	SaveMetadataWithObjId(tx *sql.Tx, objId *int64, document *Document) (int64, error)
	SaveMetadata(tx *sql.Tx, document *Document) (int64, error)
	InsertFilePath(tx *sql.Tx, document *Document) error
	SaveItems(tx *sql.Tx, objId int64, docId int64, items []Item) error
	// Read operations
	GetPurchasesByUser(userID int, year, month string) ([]Document, error)
	GetFilePathByObjId(objId int64, userID int) (string, string, error)
	// Transaction support
	BeginTx() (*sql.Tx, error)
}

type documentRepositoryImpl struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

// BeginTx starts a new database transaction.
func (r *documentRepositoryImpl) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// Resolve objID gapless by table if there is file(s) attachment.
func (r *documentRepositoryImpl) SetLatestObjId(tx *sql.Tx, document *Document) (int64, error) {

	var objID int64

	// Lock the counter row.
	err := tx.QueryRow(`
	SELECT obj_id FROM obj_id_counter WHERE name='document' FOR UPDATE
	`).Scan(&objID)

	if err != nil {
		return -1, fmt.Errorf("failed to lock obj_id_counter: %w", err)
	}

	_, err = tx.Exec(`
	UPDATE obj_id_counter SET obj_id = obj_id + 1 WHERE name = 'document'
	`)

	if err != nil {
		return -1, fmt.Errorf("failed to update latest obj ID: %w", err)
	}

	// Get the objID after updated.
	err = tx.QueryRow(`
	SELECT obj_id FROM obj_id_counter WHERE name='document'
		`).Scan(&objID)

	if err != nil {
		return -1, fmt.Errorf("failed to get latest obj ID: %w", err)
	}

	return objID, nil
}

func (r *documentRepositoryImpl) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, document *Document) (int64, error) {
	// tx, err := r.db.Begin()

	// if err != nil {
	// 	return -1, err
	// }

	// result, err := r.db.Exec(
	result, err := tx.Exec(
		`INSERT INTO obj_doc (guid, user_id, obj_id, name_or_title, file_name, description, file_size, extension, status, buy_from, buy_price, buy_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		document.GUID,
		document.UserId,
		objId,
		document.Title,
		nullableString(document.FileName),
		document.Description,
		document.FileSize,
		document.Extension,
		document.Status,
		document.PurchaseInfo.BuyFrom,
		nullableString(document.PurchaseInfo.BuyPrice),
		document.PurchaseInfo.BuyAt,
	)

	if err != nil {
		// tx.Rollback()
		return -1, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		// tx.Rollback()
		return -1, err
	}

	return id, nil
}

// Return int objID, if not -1.
func (r *documentRepositoryImpl) SaveMetadata(tx *sql.Tx, document *Document) (int64, error) {

	result, err := tx.Exec(
		`INSERT INTO obj_doc (guid, name_or_title, file_name, description, file_size, extension, status)
		VALUES (?,?,?,?,?,?,?)`,
		document.GUID, document.Title, nullableString(document.FileName), document.Description, document.FileSize, document.Extension, document.Status,
	)

	if err != nil {
		return -1, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

// SaveItems inserts all line items for a document into obj_item.
func (r *documentRepositoryImpl) SaveItems(tx *sql.Tx, objId int64, docId int64, items []Item) error {
	for _, item := range items {
		_, err := tx.Exec(
			`INSERT INTO obj_item (obj_id, doc_id, name, quantity, price, total) VALUES (?,?,?,?,?,?)`,
			objId,
			docId,
			item.Name,
			nullableString(item.Quantity),
			nullableString(item.UnitPrice),
			nullableString(item.SubTotal),
		)
		if err != nil {
			return fmt.Errorf("failed to save item %q: %w", item.Name, err)
		}
	}
	return nil
}

// purchaseRow holds raw scanned values from a single query row.
type purchaseRow struct {
	objId     int64
	title     string
	fileName  sql.NullString
	buyAt     sql.NullTime
	buyFrom   sql.NullString
	buyPrice  sql.NullString
	filePath  sql.NullString
	itemName  sql.NullString
	itemQty   sql.NullString
	itemPrice sql.NullString
	itemTotal sql.NullString
}

func (row *purchaseRow) toDocument() *Document {
	var buyAtPtr *Date
	if row.buyAt.Valid {
		d := Date{Time: row.buyAt.Time}
		buyAtPtr = &d
	}
	return &Document{
		Id:       strconv.FormatInt(row.objId, 10),
		Title:    row.title,
		FileName: row.fileName.String,
		FilePath: row.filePath.String,
		PurchaseInfo: PurchaseInfo{
			BuyAt:    buyAtPtr,
			BuyFrom:  row.buyFrom.String,
			BuyPrice: row.buyPrice.String,
		},
		Items: []Item{},
	}
}

func (row *purchaseRow) toItem() (Item, bool) {
	if !row.itemName.Valid || row.itemName.String == "" {
		return Item{}, false
	}
	return Item{
		Name:      row.itemName.String,
		Quantity:  row.itemQty.String,
		UnitPrice: row.itemPrice.String,
		SubTotal:  row.itemTotal.String,
	}, true
}

// buildPurchaseQuery constructs the SELECT query with optional year/month filters.
// TODO: check the status -1 or 1 for active.
func buildPurchaseQuery(userID int, year, month string) (string, []interface{}) {
	query := `
		SELECT
			d.obj_id, d.name_or_title, d.file_name, d.buy_at,
			d.buy_from, d.buy_price, p.file_path,
			i.name, i.quantity, i.price, i.total
		FROM obj_doc d
		LEFT JOIN obj_doc_path p ON p.doc_id = d.obj_id
		LEFT JOIN obj_item i ON i.doc_id = d.id
		WHERE d.user_id = ? AND d.status = 1`

	args := []interface{}{userID}
	if year != "" && year != "all" {
		query += " AND YEAR(d.buy_at) = ?"
		args = append(args, year)
	}
	if month != "" && month != "all" {
		query += " AND MONTH(d.buy_at) = ?"
		args = append(args, month)
	}
	return query + " ORDER BY d.buy_at DESC, d.obj_id", args
}

// GetPurchasesByUser returns all active purchases for a user, optionally filtered by year and month.
func (r *documentRepositoryImpl) GetPurchasesByUser(userID int, year, month string) ([]Document, error) {
	query, args := buildPurchaseQuery(userID, year, month)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query purchases: %w", err)
	}
	defer rows.Close()

	// docOrder preserves DESC ordering; docMap avoids duplicate doc entries from the item JOIN.
	docMap := make(map[int64]*Document)
	docOrder := []int64{}

	for rows.Next() {
		var row purchaseRow
		if err := rows.Scan(&row.objId, &row.title, &row.fileName, &row.buyAt,
			&row.buyFrom, &row.buyPrice, &row.filePath,
			&row.itemName, &row.itemQty, &row.itemPrice, &row.itemTotal); err != nil {
			return nil, fmt.Errorf("scan purchase row: %w", err)
		}

		if _, exists := docMap[row.objId]; !exists {
			docMap[row.objId] = row.toDocument()
			docOrder = append(docOrder, row.objId)
		}

		if item, ok := row.toItem(); ok {
			docMap[row.objId].Items = append(docMap[row.objId].Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	result := make([]Document, 0, len(docOrder))
	for _, id := range docOrder {
		result = append(result, *docMap[id])
	}
	return result, nil
}

// GetFilePathByObjId retrieves the file path and name for a document owned by the given user.
// TODO: check the status -1 or 1. Currently default is -1 for active.
func (r *documentRepositoryImpl) GetFilePathByObjId(objId int64, userID int) (string, string, error) {
	var filePath, fileName sql.NullString
	err := r.db.QueryRow(`
		SELECT p.file_path, d.file_name
		FROM obj_doc d
		LEFT JOIN obj_doc_path p ON p.doc_id = d.obj_id
		WHERE d.obj_id = ? AND d.user_id = ? AND d.status = 1
		LIMIT 1
	`, objId, userID).Scan(&filePath, &fileName)
	if err != nil {
		return "", "", fmt.Errorf("get file path: %w", err)
	}
	return filePath.String, fileName.String, nil
}

// Insert file path.
func (r *documentRepositoryImpl) InsertFilePath(tx *sql.Tx, document *Document) error {

	objId, err := strconv.ParseInt(document.Id, 10, 64)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO obj_doc_path (doc_id, file_path) VALUES (?,?)`, objId, document.FilePath,
	)
	if err != nil {
		return err
	}
	return nil
}
