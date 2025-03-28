// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	AddCharacterSoulcore(ctx context.Context, arg AddCharacterSoulcoreParams) error
	AddListCharacter(ctx context.Context, arg AddListCharacterParams) error
	AddSoulcoreToList(ctx context.Context, arg AddSoulcoreToListParams) error
	CreateAnonymousUser(ctx context.Context, id uuid.UUID) (User, error)
	CreateCharacter(ctx context.Context, arg CreateCharacterParams) (Character, error)
	CreateCharacterClaim(ctx context.Context, arg CreateCharacterClaimParams) (CharacterClaim, error)
	CreateCreature(ctx context.Context, name string) (Creature, error)
	CreateList(ctx context.Context, arg CreateListParams) (List, error)
	CreateSoulcoreSuggestions(ctx context.Context, arg CreateSoulcoreSuggestionsParams) error
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeactivateCharacterListMemberships(ctx context.Context, characterID uuid.UUID) error
	DeleteSoulcoreSuggestion(ctx context.Context, arg DeleteSoulcoreSuggestionParams) error
	GetCharacter(ctx context.Context, id uuid.UUID) (Character, error)
	GetCharacterByName(ctx context.Context, name string) (Character, error)
	GetCharacterClaim(ctx context.Context, arg GetCharacterClaimParams) (CharacterClaim, error)
	GetCharacterSoulcores(ctx context.Context, characterID uuid.UUID) ([]GetCharacterSoulcoresRow, error)
	GetCharacterSuggestions(ctx context.Context, characterID uuid.UUID) ([]GetCharacterSuggestionsRow, error)
	GetCharactersByUserID(ctx context.Context, userID uuid.UUID) ([]Character, error)
	GetClaimByID(ctx context.Context, id uuid.UUID) (GetClaimByIDRow, error)
	GetCreatures(ctx context.Context) ([]Creature, error)
	GetList(ctx context.Context, id uuid.UUID) (List, error)
	GetListByShareCode(ctx context.Context, shareCode uuid.UUID) (List, error)
	GetListMembers(ctx context.Context, listID uuid.UUID) ([]GetListMembersRow, error)
	GetListSoulcore(ctx context.Context, arg GetListSoulcoreParams) (GetListSoulcoreRow, error)
	GetListSoulcores(ctx context.Context, listID uuid.UUID) ([]GetListSoulcoresRow, error)
	GetListsByAuthorId(ctx context.Context, authorID uuid.UUID) ([]List, error)
	GetMembers(ctx context.Context, listID uuid.UUID) ([]ListsUser, error)
	GetPendingClaims(ctx context.Context) ([]GetPendingClaimsRow, error)
	GetPendingClaimsToCheck(ctx context.Context) ([]GetPendingClaimsToCheckRow, error)
	GetPendingSuggestionsForUser(ctx context.Context, userID uuid.UUID) ([]GetPendingSuggestionsForUserRow, error)
	GetUserByEmail(ctx context.Context, email pgtype.Text) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]GetUserCharactersRow, error)
	GetUserLists(ctx context.Context, authorID uuid.UUID) ([]GetUserListsRow, error)
	IsUserListMember(ctx context.Context, arg IsUserListMemberParams) (bool, error)
	MigrateAnonymousUser(ctx context.Context, arg MigrateAnonymousUserParams) (User, error)
	RemoveCharacterSoulcore(ctx context.Context, arg RemoveCharacterSoulcoreParams) error
	RemoveListSoulcore(ctx context.Context, arg RemoveListSoulcoreParams) error
	UpdateCharacterOwner(ctx context.Context, arg UpdateCharacterOwnerParams) (Character, error)
	UpdateClaimStatus(ctx context.Context, arg UpdateClaimStatusParams) (CharacterClaim, error)
	UpdateSoulcoreStatus(ctx context.Context, arg UpdateSoulcoreStatusParams) error
	VerifyEmail(ctx context.Context, arg VerifyEmailParams) error
}

var _ Querier = (*Queries)(nil)
