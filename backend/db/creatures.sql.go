// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: creatures.sql

package db

import (
	"context"
)

const createCreature = `-- name: CreateCreature :one
INSERT INTO creatures (name)
VALUES ($1)
RETURNING id, name
`

func (q *Queries) CreateCreature(ctx context.Context, name string) (Creature, error) {
	row := q.db.QueryRow(ctx, createCreature, name)
	var i Creature
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getCreatures = `-- name: GetCreatures :many
SELECT id, name
FROM creatures
ORDER BY name
`

func (q *Queries) GetCreatures(ctx context.Context) ([]Creature, error) {
	rows, err := q.db.Query(ctx, getCreatures)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Creature{}
	for rows.Next() {
		var i Creature
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
