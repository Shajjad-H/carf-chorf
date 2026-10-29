package vehicle

import (
	"context"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/carbpi_service"
	dto "carf-chorf.com/app/pkg/dto_validations"
)

// Resolve returns a cached record when the VIN or plate/state matches and
// queries carbpi on a miss. It is shared by the public lookup and the dashboard
// report purchase flow.
func Resolve(ctx context.Context, req dto.LookupRequest) (*domaine.CacheRecord, error) {
	return lookup(ctx, req)
}

// lookup returns a cached record when the VIN or plate/state matches. On a
// miss it queries carbpi and stores the response. Records are keyed by VIN;
// plate/state is an alternate lookup path into the same cache.
func lookup(ctx context.Context, req dto.LookupRequest) (*domaine.CacheRecord, error) {
	// search by vin or plate and state
	cached, err := findCachedRecord(req)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	// did not found in db cache

	incoming, err := fetchFromCarbpi(ctx, req)
	if err != nil {
		return nil, err
	}

	return saveRecord(incoming)
}

// fetchFromCarbpi chooses the documented carbpi endpoint for the request and
// converts the provider response into the record shape we cache.
func fetchFromCarbpi(ctx context.Context, req dto.LookupRequest) (*domaine.CacheRecord, error) {
	var data *carbpi_service.Record
	var err error

	if req.Vin != "" {
		data, err = carbpi_service.CheckCarfaxRecords(ctx, req.Vin)
	} else {
		data, err = carbpi_service.CheckCarfaxRecordsByPlate(ctx, req.Plate, req.State)
	}
	if err != nil {
		return nil, err
	}

	return buildRecord(data, req), nil
}

// buildRecord converts a carbpi response into the record shape we cache. The
// request plate/state are fallbacks for providers that omit them.
func buildRecord(data *carbpi_service.Record, req dto.LookupRequest) *domaine.CacheRecord {
	plate := data.Plate
	if plate == "" {
		plate = req.Plate
	}
	state := data.State
	if state == "" {
		state = req.State
	}

	return &domaine.CacheRecord{
		Vin:           data.Vin,
		Provider:      domaine.CarbpiProvider,
		Plate:         plate,
		State:         state,
		Vehicle:       data.Vehicle,
		CarfaxRecords: data.CarfaxRecords,
	}
}
