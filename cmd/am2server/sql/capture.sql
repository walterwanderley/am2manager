/* name: AddCapture :execresult */
/* http: POST /captures */
INSERT INTO capture(user_id, name, description, data)
VALUES(?,?,?,?);

/* name: RemoveCapture :execresult */
/* http: DELETE /captures/{id} */
DELETE FROM capture WHERE id = ?;

/* name: GetCapture :one */
/* http: GET /captures/{id} */
SELECT * FROM capture WHERE id = ?;

/* name: GetCaptureFile :one */
/* http: GET /captures/{id}/file */
UPDATE capture SET downloads = downloads + 1 WHERE id = ?
RETURNING data;

/* name: ListCaptures :many */
/* http: GET /captures */
SELECT id, name, description, downloads, created_at 
FROM capture
ORDER BY downloads, created_at DESC;
