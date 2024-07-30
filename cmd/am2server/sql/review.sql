/* name: addReview :execresult */
/* http: POST /reviews */
INSERT INTO review(user_id, capture_id, rate, comment) 
VALUES(?,?,?,?);

/* name: removeReview :execresult */
/* http: DELETE /reviews/{id} */
DELETE FROM review WHERE id = ?;

/* name: getReview :one */
SELECT * FROM review WHERE id = ?;

/* name: ListReviewsByUser :many */
/* http: GET /users/{user_id}/reviews */
SELECT * FROM review
WHERE user_id = ?;

/* name: ListReviewsByCapture :many */
/* http: GET /captures/{capture_id}/reviews */
SELECT * FROM review
WHERE capture_id = ?;
