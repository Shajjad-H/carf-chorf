package vehicle

import (
	"errors"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	dto "carf-chorf.com/app/pkg/dto_validations"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// findCachedRecord performs the pre-API lookup. VIN has priority; otherwise
// it looks up by plate and state. It returns nil, nil when nothing matches.
func findCachedRecord(req dto.LookupRequest) (*domaine.CacheRecord, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	query := db.Where("provider = ?", domaine.CarbpiProvider)
	switch {
	case req.Vin != "":
		query = query.Where("vin = ?", strings.ToUpper(req.Vin))
	case req.Plate != "" && req.State != "":
		query = query.
			Where("plate = ? AND state = ?", strings.ToUpper(req.Plate), strings.ToUpper(req.State)).
			Order("updated_at DESC")
	default:
		return nil, nil
	}

	return firstOrNil(query)
}

// firstOrNil runs a lookup query and converts gorm.ErrRecordNotFound into nil.
func firstOrNil(query *gorm.DB) (*domaine.CacheRecord, error) {
	var record domaine.CacheRecord
	if err := query.First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}

// FindRecordByToken looks up the public vehicle handle used by checkout.
// Tokens are canonical UUIDv4 strings, so malformed values skip the query.
func FindRecordByToken(token string) (*domaine.CacheRecord, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	return firstOrNil(db.Where("token = ?", token))
}

// saveRecord upserts a provider response by its (vin, provider) cache key.
func saveRecord(incoming *domaine.CacheRecord) (*domaine.CacheRecord, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return upsertRecord(db, incoming)
}

// upsertRecord inserts a new record or updates the matching VIN row. A rare
// concurrent insert can fail; the user can retry the lookup.
func upsertRecord(db *gorm.DB, incoming *domaine.CacheRecord) (*domaine.CacheRecord, error) {
	existing, err := firstOrNil(
		db.Where("vin = ? AND provider = ?", incoming.Vin, incoming.Provider),
	)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return updateRecord(db, existing, incoming)
	}

	if incoming.Token == "" {
		token, err := newToken()
		if err != nil {
			return nil, err
		}
		incoming.Token = token
	}

	if err := db.Create(incoming).Error; err != nil {
		return nil, err
	}
	return incoming, nil
}

// updateRecord applies the latest plate/state values and saves only when
// something changed. Vehicle data and the existing token are preserved.
func updateRecord(db *gorm.DB, existing, incoming *domaine.CacheRecord) (*domaine.CacheRecord, error) {
	changed := false

	if incoming.Plate != "" && existing.Plate != incoming.Plate {
		existing.Plate = incoming.Plate
		changed = true
	}
	if incoming.State != "" && existing.State != incoming.State {
		existing.State = incoming.State
		changed = true
	}
	if existing.Token == "" {
		token, err := newToken()
		if err != nil {
			return nil, err
		}
		existing.Token = token
		changed = true
	}

	if !changed {
		return existing, nil
	}
	if err := db.Save(existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

// newToken returns a random UUIDv4 string.
func newToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
