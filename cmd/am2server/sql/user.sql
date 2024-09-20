/* name: addUser :execresult */
INSERT INTO user(name, email, status, picture) 
VALUES(?,?,?,?);

/* name: getUser :one */
/* http: GET /users/{id} */
SELECT * FROM user WHERE id = ?;

/* name: ContUsers :one */
/* http: GET /users/count */
SELECT count(*) FROM user;

/* name: GetUserByEmail :one */
/* http: GET /users */
SELECT * from user WHERE email = ?;

/* name: updateUserName :execresult */
/* http: PATCH /users/{id}/name */
UPDATE user SET name = ?, updated_at = date()  WHERE id = ?;

/* name: updateUserPicture :execresult */
UPDATE user SET picture = ?, updated_at = date()  WHERE id = ?;

/* name: addFavoriteCapture :execresult */
/* http: POST /users/{user_id}/captures/{capture_id} */
INSERT INTO user_favorite(user_id, capture_id) VALUES(?,?);

/* name: removeFavoriteCapture :execresult */
/* http: DELETE /users/{user_id}/captures/{capture_id} */
DELETE FROM user_favorite WHERE user_id = ? AND capture_id = ?;

/* name: getFavoriteCapture :one */
/* http: GET /users/{user_id}/captures/{capture_id} */
SELECT *
FROM user_favorite
WHERE user_id = ? AND capture_id = ?;

/* name: listFavoriteCaptures :many */
/* http: GET /users/{user_id}/captures */
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at, c.demo_link
FROM user_favorite uf, capture c
WHERE uf.capture_id = c.id AND uf.user_id = ?
ORDER BY uf.created_at DESC
LIMIT ? OFFSET ?;

/* name: ListAllFavoriteCaptures :many */
/* http: GET /favorites */
SELECT c.id, c.name, c.description, c.downloads, c.has_cab, c.type, c.created_at, c.demo_link, uf.user_id
FROM user_favorite uf, capture c
WHERE uf.capture_id = c.id
ORDER BY uf.created_at DESC
LIMIT ? OFFSET ?;
