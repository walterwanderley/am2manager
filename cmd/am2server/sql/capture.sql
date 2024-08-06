/* name: addCapture :execresult */
INSERT INTO capture(user_id, name, description, type, has_cab, data, am2_hash, data_hash, demo_link)
VALUES(?,?,?,?,?,?,?,?,?);

/* name: updateCapture :execresult */
/* http: PUT /captures/{id} */
UPDATE capture SET name = ?, 
description = ?, type = ?, has_cab = ?, demo_link = ?
WHERE id = ?;

/* name: GetCapture :one */
/* http: GET /captures/{id} */
SELECT * FROM capture WHERE id = ?;

/* name: GetCaptureFile :one */
/* http: GET /captures/{id}/file */
UPDATE capture SET downloads = downloads + 1 WHERE id = ?
RETURNING data, name;

/* name: SearchCaptures :many */
/* http: GET /captures */
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at, c.demo_link, AVG(r.rate) rate, uf.user_id fav
FROM capture c LEFT OUTER JOIN review r ON (c.id = r.capture_id)
LEFT OUTER JOIN user_favorite uf ON (c.id = uf.capture_id)
WHERE c.description LIKE '%'||sqlc.arg('arg')||'%' OR c.name LIKE '%'||sqlc.arg('arg')||'%' 
OR c.data_hash = sqlc.arg('arg') OR c.am2_hash = sqlc.arg('arg')
AND (uf.user_id = sqlc.arg('user') OR uf.user_id IS NULL)
GROUP BY c.id
ORDER BY rate DESC, c.downloads DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

/* name: totalSearchCaptures :one */
SELECT count(*)
FROM capture c
WHERE c.description LIKE '%'||sqlc.arg('arg')||'%' OR c.name LIKE '%'||sqlc.arg('arg')||'%' 
OR c.data_hash = sqlc.arg('arg') OR c.am2_hash = sqlc.arg('arg');

/* name: mostRecentCaptures :many */
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at 
FROM capture c
ORDER BY c.created_at DESC
LIMIT 5;

/* name: mostDownloadedCaptures :many */
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at 
FROM capture c
ORDER BY c.downloads DESC
LIMIT 5;

/* name: protectedTrainer :one */
SELECT * FROM protected_am2 WHERE am2_hash = ? LIMIT 1;

/* name: listReviewsByCapture :many */
SELECT * FROM review
WHERE capture_id = ?;

/* name: rateByCapture :one */
SELECT AVG(rate) FROM review WHERE capture_id = ?;