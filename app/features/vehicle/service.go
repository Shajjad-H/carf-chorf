package vehicle

import (
	"context"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/carbpi_service"
	dto "carf-chorf.com/app/pkg/dto_validations"
)

// lookup returns a cached record when the VIN or plate/state matches. On a
// miss it queries carbpi and stores the response. Records are keyed by VIN;
// plate/state is an alternate lookup path into the same cache.
func lookup(ctx context.Context, req dto.LookupRequest) (*domaine.CacheRecord, error) {
	cached, err := findCachedRecord(req)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	data, err := fetchFromCarbpi(ctx, req)
	if err != nil {
		return nil, err
	}

	incoming := buildRecord(data, req)

	// The initial VIN lookup already proved there is no record for this VIN,
	// so create it directly. createRecord handles a concurrent insert race.
	if req.Vin != "" {
		return createRecord(incoming)
	}

	// A plate lookup only learns the VIN from the API response. If a VIN-only
	// record already exists, fill in the plate/state from this request instead
	// of creating a duplicate.
	return mergeOrCreateRecord(incoming)
}

// fetchFromCarbpi chooses the documented carbpi endpoint for the request.
func fetchFromCarbpi(ctx context.Context, req dto.LookupRequest) (*carbpi_service.Record, error) {
	if req.Vin != "" {
		return carbpi_service.CheckCarfaxRecords(ctx, req.Vin)
	}
	return carbpi_service.CheckCarfaxRecordsByPlate(ctx, req.Plate, req.State)
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
