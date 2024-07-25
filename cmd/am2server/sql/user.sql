/* name: AddUser :execresult */
/* http: POST /users */
INSERT INTO user(login, email, pass, status) 
VALUES(?,?,?,?);

/* name: RemoveUser :execresult */
/* http: DELETE /users/{id} */
DELETE FROM user WHERE id = ?;

/* name: GetUser :one */
/* http: GET /users/{id} */
SELECT * FROM user WHERE id = ?;

/* name: ContUsers :one */
/* http: GET /users/count */
SELECT count(*) FROM user;

/* name: SetUserPassword :execresult */
/* http: PATCH /users/{id}/pass */
UPDATE user SET pass = ? WHERE id = ?;

/* name: ValidateUserEmail :execresult */
/* http: GET /users/{id}/token/{pass} */
UPDATE user SET status = 'VALID' WHERE id = ? AND pass = ?;