// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: characters.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createCharacter = `-- name: CreateCharacter :one
INSERT INTO characters (user_id, name, world)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, world, created_at, updated_at
`

type CreateCharacterParams struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	World  string    `json:"world"`
}

func (q *Queries) CreateCharacter(ctx context.Context, arg CreateCharacterParams) (Character, error) {
	row := q.db.QueryRow(ctx, createCharacter, arg.UserID, arg.Name, arg.World)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.World,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCharacter = `-- name: GetCharacter :one
SELECT id, user_id, name, world, created_at, updated_at FROM characters
WHERE id = $1
`

func (q *Queries) GetCharacter(ctx context.Context, id uuid.UUID) (Character, error) {
	row := q.db.QueryRow(ctx, getCharacter, id)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.World,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCharactersByUserID = `-- name: GetCharactersByUserID :many
SELECT id, user_id, name, world, created_at, updated_at FROM characters
WHERE user_id = $1
`

func (q *Queries) GetCharactersByUserID(ctx context.Context, userID uuid.UUID) ([]Character, error) {
	rows, err := q.db.Query(ctx, getCharactersByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Character{}
	for rows.Next() {
		var i Character
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Name,
			&i.World,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserCharacters = `-- name: GetUserCharacters :many
SELECT id, name, world
FROM characters
WHERE user_id = $1
ORDER BY name
`

type GetUserCharactersRow struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	World string    `json:"world"`
}

func (q *Queries) GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]GetUserCharactersRow, error) {
	rows, err := q.db.Query(ctx, getUserCharacters, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetUserCharactersRow{}
	for rows.Next() {
		var i GetUserCharactersRow
		if err := rows.Scan(&i.ID, &i.Name, &i.World); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
