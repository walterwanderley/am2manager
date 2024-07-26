/* name: addCapture :execresult */
INSERT INTO capture(user_id, name, description, type, has_cab, data, am2_hash, data_hash)
VALUES(?,?,?,?,?,?,?,?);

/* name: RemoveCapture :execresult */
/* http: DELETE /captures/{id} */
DELETE FROM capture WHERE id = ?;

/* name: GetCapture :one */
/* http: GET /captures/{id} */
SELECT * FROM capture WHERE id = ?;

/* name: GetCaptureFile :one */
/* http: GET /captures/{id}/file */
UPDATE capture SET downloads = downloads + 1 WHERE id = ?
RETURNING data, name;

/* name: SearchCaptures :many */
/* http: GET /captures */
SELECT c.id, c.name, c.description, c.downloads, count(f.capture_id) AS fav, c.has_cab, c.type, c.created_at 
FROM capture c LEFT OUTER JOIN user_favorite f ON c.id = f.capture_id
WHERE c.description LIKE '%'||sqlc.arg('arg')||'%' OR c.name LIKE '%'||sqlc.arg('arg')||'%' 
GROUP BY f.capture_id
ORDER BY c.downloads, c.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
