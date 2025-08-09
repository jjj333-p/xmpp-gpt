package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Question represents a question-answer pair in the database
type Question struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuestionDB wraps the database connection pool for question operations
type QuestionDB struct {
	pool *pgxpool.Pool
}

// Init creates a connection pool to the database
func (qdb *QuestionDB) Init(pgUrl string) {
	// create connection pool
	dbpool, err := pgxpool.New(context.Background(), pgUrl)
	if err != nil {
		panic(err)
	}

	// store pool in struct
	qdb.pool = dbpool
	fmt.Println("Connected to database")
}

// Close closes the database connection pool
func (qdb *QuestionDB) Close() {
	qdb.pool.Close()
}

func (qdb *QuestionDB) CreateWithID(ctx context.Context, id, userID, question string) error {
	query := `
		INSERT INTO questions (id, user_id, question)
		VALUES ($1, $2, $3)`

	_, err := qdb.pool.Exec(ctx, query, id, userID, question)
	if err != nil {
		return fmt.Errorf("failed to create question with ID: %w", err)
	}

	return nil
}

// GetByID retrieves a question by its ID
func (qdb *QuestionDB) GetByID(ctx context.Context, id string) (*Question, error) {
	q := &Question{}
	query := `
		SELECT id, user_id, question, answer, created_at, updated_at
		FROM questions
		WHERE id = $1`

	err := qdb.pool.QueryRow(ctx, query, id).Scan(
		&q.ID, &q.UserID, &q.Question, &q.Answer, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get question by ID: %w", err)
	}

	return q, nil
}

// GetByUserIDWithLimit retrieves questions for a user with pagination
func (qdb *QuestionDB) GetByUserIDWithLimit(ctx context.Context, userID string, limit, offset int) ([]Question, error) {
	query := `
		SELECT id, user_id, question, answer, created_at, updated_at
		FROM questions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := qdb.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions with limit: %w", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		q := Question{}
		err := rows.Scan(&q.ID, &q.UserID, &q.Question, &q.Answer, &q.CreatedAt, &q.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

// GetByUserIDWithLimitAnswered retrieves answered questions for a user with pagination
func (qdb *QuestionDB) GetByUserIDWithLimitAnswered(ctx context.Context, userID string, limit, offset int) ([]Question, error) {
	query := `
		SELECT id, user_id, question, answer, created_at, updated_at
		FROM questions
		WHERE user_id = $1 AND answer != ''
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := qdb.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get answered questions with limit: %w", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		q := Question{}
		err := rows.Scan(&q.ID, &q.UserID, &q.Question, &q.Answer, &q.CreatedAt, &q.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

// Update modifies an existing question's answer
func (qdb *QuestionDB) Update(ctx context.Context, id, newAnswer string) error {
	query := `
		UPDATE questions
		SET answer = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := qdb.pool.Exec(ctx, query, id, newAnswer)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no question found with ID: %s", id)
	}

	return nil
}

// Delete removes a question from the database
func (qdb *QuestionDB) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM questions WHERE id = $1`

	result, err := qdb.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no question found with ID: %s", id)
	}

	return nil
}

// DeleteByUserID removes all questions for a specific user
func (qdb *QuestionDB) DeleteByUserID(ctx context.Context, userID string) (int64, error) {
	query := `DELETE FROM questions WHERE user_id = $1`

	result, err := qdb.pool.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete questions for user: %w", err)
	}

	return result.RowsAffected(), nil
}

// CountByUserID returns the total number of questions for a user
func (qdb *QuestionDB) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM questions WHERE user_id = $1`

	err := qdb.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count questions: %w", err)
	}

	return count, nil
}
