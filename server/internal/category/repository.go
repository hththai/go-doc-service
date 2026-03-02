package category

import (
	"2_Go/internal/obj"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	CreateCategory(userID int, name, color string) (*Category, error)
	GetCategoriesByUser(userID int) ([]Category, error)
	UpdateCategory(guid string, userID int, name, color string) error
	SoftDeleteCategory(guid string, userID int) error

	// Document-category link management (tx-aware for use inside purchase transactions).
	SaveDocCategories(tx *sql.Tx, docId int64, categoryGUIDs []string, userID int) error
	DeleteDocCategories(tx *sql.Tx, docId int64) error
	GetCategoriesByDocIds(docIds []int64) (map[int64][]Category, error)
}

type categoryRepositoryImpl struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepositoryImpl{db: db}
}

func (r *categoryRepositoryImpl) CreateCategory(userID int, name, color string) (*Category, error) {
	guid := uuid.New().String()
	result, err := r.db.Exec(
		`INSERT INTO obj_category (guid, user_id, name, color, status) VALUES (?,?,?,?,?)`,
		guid, userID, name, nullableString(color), obj.StatusActive,
	)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Category{GUID: guid, Id: int(id), UserId: userID, Name: name, Color: color, Status: obj.StatusActive}, nil
}

func (r *categoryRepositoryImpl) GetCategoriesByUser(userID int) ([]Category, error) {
	rows, err := r.db.Query(
		`SELECT id, guid, name, color, status, created_at, modified_at
		 FROM obj_category WHERE user_id=? AND status=1 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		var color sql.NullString
		if err := rows.Scan(&c.Id, &c.GUID, &c.Name, &color, &c.Status, &c.CreatedAt, &c.ModifiedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		c.Color = color.String
		c.UserId = userID
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *categoryRepositoryImpl) UpdateCategory(guid string, userID int, name, color string) error {
	result, err := r.db.Exec(
		`UPDATE obj_category SET name=?, color=? WHERE guid=? AND user_id=? AND status=1`,
		name, nullableString(color), guid, userID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *categoryRepositoryImpl) SoftDeleteCategory(guid string, userID int) error {
	result, err := r.db.Exec(
		`UPDATE obj_category SET status=-1 WHERE guid=? AND user_id=? AND status=1`,
		guid, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SaveDocCategories resolves each GUID to its integer id (scoped to userID) then inserts into obj_doc_category.
func (r *categoryRepositoryImpl) SaveDocCategories(tx *sql.Tx, docId int64, categoryGUIDs []string, userID int) error {
	if len(categoryGUIDs) == 0 {
		return nil
	}
	for _, guid := range categoryGUIDs {
		var catId int
		err := tx.QueryRow(
			`SELECT id FROM obj_category WHERE guid=? AND user_id=? AND status=1`,
			guid, userID,
		).Scan(&catId)
		if err != nil {
			return fmt.Errorf("resolve category guid %s: %w", guid, err)
		}
		if _, err := tx.Exec(
			`INSERT IGNORE INTO obj_doc_category (doc_id, category_id) VALUES (?,?)`,
			docId, catId,
		); err != nil {
			return fmt.Errorf("link category: %w", err)
		}
	}
	return nil
}

// DeleteDocCategories removes all category links for a document (call before re-linking on update).
func (r *categoryRepositoryImpl) DeleteDocCategories(tx *sql.Tx, docId int64) error {
	_, err := tx.Exec(`DELETE FROM obj_doc_category WHERE doc_id=?`, docId)
	return err
}

// GetCategoriesByDocIds fetches active categories for a batch of doc IDs in a single query.
// Returns a map of docId -> []Category.
func (r *categoryRepositoryImpl) GetCategoriesByDocIds(docIds []int64) (map[int64][]Category, error) {
	if len(docIds) == 0 {
		return map[int64][]Category{}, nil
	}

	placeholders := strings.Repeat("?,", len(docIds))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]interface{}, len(docIds))
	for i, id := range docIds {
		args[i] = id
	}

	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT dc.doc_id, c.id, c.guid, c.user_id, c.name, c.color, c.status, c.created_at, c.modified_at
		FROM obj_category c
		INNER JOIN obj_doc_category dc ON dc.category_id = c.id
		WHERE dc.doc_id IN (%s) AND c.status=1
		ORDER BY c.name`, placeholders), args...)
	if err != nil {
		return nil, fmt.Errorf("get categories by doc ids: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]Category)
	for rows.Next() {
		var docId int64
		var c Category
		var color sql.NullString
		if err := rows.Scan(&docId, &c.Id, &c.GUID, &c.UserId, &c.Name, &color, &c.Status, &c.CreatedAt, &c.ModifiedAt); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		c.Color = color.String
		result[docId] = append(result[docId], c)
	}
	return result, rows.Err()
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
