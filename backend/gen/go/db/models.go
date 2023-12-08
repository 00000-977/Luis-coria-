// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package db_queries

import (
	"github.com/jackc/pgx/v5/pgtype"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

type NeosyncApiAccount struct {
	ID             pgtype.UUID
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
	AccountType    int16
	AccountSlug    string
	TemporalConfig *pg_models.TemporalConfig
}

type NeosyncApiAccountApiKey struct {
	ID          pgtype.UUID
	AccountID   pgtype.UUID
	KeyValue    string
	CreatedByID pgtype.UUID
	UpdatedByID pgtype.UUID
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
	ExpiresAt   pgtype.Timestamp
	KeyName     string
	UserID      pgtype.UUID
}

type NeosyncApiAccountInvite struct {
	ID           pgtype.UUID
	AccountID    pgtype.UUID
	SenderUserID pgtype.UUID
	Email        string
	Token        string
	Accepted     pgtype.Bool
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
	ExpiresAt    pgtype.Timestamp
}

type NeosyncApiAccountUserAssociation struct {
	ID        pgtype.UUID
	AccountID pgtype.UUID
	UserID    pgtype.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type NeosyncApiConnection struct {
	ID               pgtype.UUID
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
	Name             string
	AccountID        pgtype.UUID
	ConnectionConfig *pg_models.ConnectionConfig
	CreatedByID      pgtype.UUID
	UpdatedByID      pgtype.UUID
}

type NeosyncApiJob struct {
	ID                pgtype.UUID
	CreatedAt         pgtype.Timestamp
	UpdatedAt         pgtype.Timestamp
	Name              string
	AccountID         pgtype.UUID
	Status            int16
	ConnectionOptions *pg_models.JobSourceOptions
	Mappings          []*pg_models.JobMapping
	CronSchedule      pgtype.Text
	CreatedByID       pgtype.UUID
	UpdatedByID       pgtype.UUID
}

type NeosyncApiJobDestinationConnectionAssociation struct {
	ID           pgtype.UUID
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
	JobID        pgtype.UUID
	ConnectionID pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

type NeosyncApiTransformer struct {
	ID                pgtype.UUID
	CreatedAt         pgtype.Timestamp
	UpdatedAt         pgtype.Timestamp
	Name              string
	Description       string
	Type              string
	Source            string
	AccountID         pgtype.UUID
	TransformerConfig *pg_models.TransformerConfigs
	CreatedByID       pgtype.UUID
	UpdatedByID       pgtype.UUID
}

type NeosyncApiUser struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	UserType  int16
}

type NeosyncApiUserIdentityProviderAssociation struct {
	ID              pgtype.UUID
	UserID          pgtype.UUID
	Auth0ProviderID string
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}
