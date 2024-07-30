/* name: AddUser :execresult */
/* http: POST /users */
INSERT INTO user(name, email, status, picture) 
VALUES(?,?,?,?);

/* name: GetUser :one */
/* http: GET /users/{id} */
SELECT * FROM user WHERE id = ?;

/* name: ContUsers :one */
/* http: GET /users/count */
SELECT count(*) FROM user;

/* name: GetUserByEmail :one */
/* http: GET /users */
SELECT * from user WHERE email = ?;

/* name: updateUserPicture :execresult */
UPDATE user SET picture = ?, updated_at = now()  WHERE id = ?;