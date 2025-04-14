// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"
	"time"
)

type Account struct {
	ID            int64
	Number        string
	Destination   string
	Description   string
	AssociationID int64
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
}

type Association struct {
	ID            int64
	Name          string
	Address       string
	Administrator string
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
}

type Building struct {
	ID              int64
	Name            string
	Address         string
	CadastralNumber string
	TotalArea       float64
	AssociationID   int64
	CreatedAt       sql.NullTime
	UpdatedAt       sql.NullTime
}

type Category struct {
	ID            int64
	Type          string
	Family        string
	Name          string
	AssociationID int64
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
}

type Expense struct {
	ID          int64
	Amount      float64
	Description string
	Destination string
	Date        time.Time
	Month       int64
	Year        int64
	CategoryID  int64
	AccountID   int64
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
}

type Owner struct {
	ID                   int64
	Name                 string
	NormalizedName       string
	IdentificationNumber string
	ContactPhone         string
	ContactEmail         string
	FirstDetectedAt      sql.NullTime
	CreatedAt            sql.NullTime
	UpdatedAt            sql.NullTime
}

type Ownership struct {
	ID                   int64
	UnitID               int64
	OwnerID              int64
	StartDate            sql.NullTime
	EndDate              interface{}
	IsActive             sql.NullBool
	RegistrationDocument string
	RegistrationDate     time.Time
	CreatedAt            sql.NullTime
	UpdatedAt            sql.NullTime
}

type RefreshToken struct {
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Login     string
	ExpiresAt time.Time
	RevokedAt sql.NullTime
}

type Unit struct {
	ID              int64
	CadastralNumber string
	BuildingID      int64
	UnitNumber      string
	Address         string
	Entrance        int64
	Area            float64
	Part            float64
	UnitType        string
	Floor           int64
	RoomCount       int64
	CreatedAt       sql.NullTime
	UpdatedAt       sql.NullTime
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	ToptSecret   string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UsersAssociation struct {
	ID            int64
	UserID        int64
	AssociationID int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
