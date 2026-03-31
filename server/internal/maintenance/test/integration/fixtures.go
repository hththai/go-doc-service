package integration

const (
	errorInsertMsg       = "\033[31mfailed to insert test data: %v\033[0m"
	errorDropAllTableMsg = "\033[31mfailed to drop table: %s\033[0m"
	errorFailToInsertMsg = "failed to insert test data: %v"

	dropObjDocTable   = "DROP TABLE IF EXISTS obj_doc"
	createObjDocTable = "CREATE TABLE obj_doc (obj_id BIGINT PRIMARY KEY)"
	insertObjDocData  = "INSERT INTO obj_doc (obj_id) VALUES (1), (2), (3)"

	dropObjIdCounterTable   = "DROP TABLE IF EXISTS obj_id_counter"
	createObjIdCounterTable = "CREATE TABLE obj_id_counter (name VARCHAR(255) PRIMARY KEY, obj_id BIGINT)"
	insertObjIdCounterData  = "INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 456)"

	dropObjVerificationTable   = "DROP TABLE IF EXISTS obj_doc_verification"
	createObjVerficiationTable = `CREATE TABLE obj_doc_verification (
								id BIGINT AUTO_INCREMENT PRIMARY KEY,
								obj_id BIGINT NOT NULL,
								has_issue TINYINT(1) NOT NULL DEFAULT 0,
								issue_code VARCHAR(50) NULL,
								issue_message VARCHAR(255) NULL,
								verified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
								FOREIGN KEY (obj_id) REFERENCES obj_doc(obj_id)
								)`

	createObjDocDetailTable = `CREATE TABLE obj_doc (
								obj_id BIGINT PRIMARY KEY,
								status INT NOT NULL,
								name_or_title VARCHAR(100) NOT NULL,
								description VARCHAR(1000),
								created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
								modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
								buy_from VARCHAR(1000),
								buy_price REAL,
								sold_price REAL,
								buy_at TIMESTAMP NULL,
								sold_at TIMESTAMP NULL,
								file_size REAL,
								extension VARCHAR(10)
							)
							`
	insertObjDocDetailData = `INSERT INTO obj_doc (
								obj_id,
								status,
								name_or_title,
								description,
								created_at,
								modified_at,
								buy_from,
								buy_price,
								sold_price,
								buy_at,
								sold_at,
								file_size,
								extension
							) VALUES
							(1, 1, 'Invoice Jan', 'January invoice document', NOW(), NOW(), 'Amazon', 19.99, NULL, '2024-01-10 10:00:00', NULL, 2048, 'pdf'),
							(2, 1, 'Receipt Laptop', 'Laptop purchase receipt', NOW(), NOW(), 'BestBuy', 1299.00, NULL, '2024-02-15 14:30:00', NULL, 512, 'jpg'),
							(3, 0, 'Contract Draft', 'Draft version of contract', NOW(), NOW(), NULL, NULL, NULL, NULL, NULL, 4096, 'docx'),
							(4, 1, 'Warranty Card', 'Warranty information', NOW(), NOW(), 'Apple', 0.00, NULL, '2024-03-01 09:00:00', NULL, 1024, 'png');
							`
)
