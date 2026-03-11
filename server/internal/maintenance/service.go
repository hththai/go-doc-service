package maintenance

// I. Verify obj_doc id equals to obj_id_counter

// II. Verify total files is matched with obj_doc_path.
// 1. Find number of files in a directory file/0/
// 2. In database, obj_doc_path fine the number of records.
// 3. Verify if those number is matched.
// 4. If matched, GOOD.

// III. REPAIR if not.
// 0. Report the list of obj_doc that need to be repaired with obj_id.
// 1. Flag the file as "REPAIR" in database.
// 	[ ] 1.1 Schema: Need column as system_flag true false (false by default) for obj_doc.
// !!!2. Move the file to a temporary folder.

// If number of files less than obj_doc_path:
// 1. Flag the missing record by obj_id and report.

// If number of files more than obj_doc_path:
// 1. Check if there is duplicated number files. ex: 8.pdf 8.png 8.txt.
// 2. In obj_doc_path, find the address of obj_id 8, file_path: ./filedata/0/0/8.pdf
// 3. Verify with obj_doc: extension, file_size
