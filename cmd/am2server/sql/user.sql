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