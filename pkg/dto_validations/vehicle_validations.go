package dto_validations

import (
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// LookupRequest contains either a VIN or a plate/state pair. When both are
// submitted, VIN takes priority.
type LookupRequest struct {
	Vin   string
	Plate string
	State string
}

// Validate normalizes the inputs and requires a valid VIN, or a valid plate
// and state when no VIN is provided.
func (l *LookupRequest) Validate() error {
	l.Vin = strings.ToUpper(strings.TrimSpace(l.Vin))
	l.Plate = strings.ToUpper(strings.TrimSpace(l.Plate))
	l.State = strings.ToUpper(strings.TrimSpace(l.State))

	if l.Vin != "" {
		// Discard plate/state so a tampered request cannot mix lookup types.
		l.Plate = ""
		l.State = ""
		return validation.ValidateStruct(l,
			validation.Field(&l.Vin,
				validation.Length(17, 17).Error("VIN must be exactly 17 characters"),
				validation.Match(vinRegex).Error("VIN format is invalid"),
			),
		)
	}

	return validation.ValidateStruct(l,
		validation.Field(&l.Plate,
			validation.Required.Error("Plate is required"),
			validation.Match(plateRegex).Error("Plate format is invalid"),
		),
		validation.Field(&l.State,
			validation.Required.Error("State is required"),
			validation.Match(stateRegex).Error("State format is invalid"),
		),
	)
}

// VehicleInfoRequest validates a VIN or an optional plate/state pair.
type VehicleInfoRequest struct {
	Vin   string
	Plate string
	State string
}

// PlateRequest validates a license plate and US state.
type PlateRequest struct {
	Plate string
	State string
}

var (
	// VINs are 17 characters; I, O, and Q are excluded by the VIN standard.
	vinRegex = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
	// US state/territory codes used by this form are two letters.
	stateRegex = regexp.MustCompile(`^[A-Z]{2}$`)
	// Plates allow letters, digits, spaces, and hyphens after normalization.
	plateRegex = regexp.MustCompile(`^[A-Z0-9 -]{1,10}$`)
)

// Validate normalizes and validates a plate/state pair.
func (p *PlateRequest) Validate() error {
	p.Plate = strings.ToUpper(strings.TrimSpace(p.Plate))
	p.State = strings.ToUpper(strings.TrimSpace(p.State))
	return validation.ValidateStruct(p,
		validation.Field(&p.Plate, validation.Required.Error("plate is required"), validation.Match(plateRegex).Error("plate format is invalid")),
		validation.Field(&p.State, validation.Required.Error("state is required"), validation.Match(stateRegex).Error("state format is invalid")),
	)
}

// Validate normalizes and validates a vehicle lookup request.
func (v *VehicleInfoRequest) Validate() error {
	v.Vin = strings.ToUpper(strings.TrimSpace(v.Vin))
	v.Plate = strings.ToUpper(strings.TrimSpace(v.Plate))
	v.State = strings.ToUpper(strings.TrimSpace(v.State))

	if v.Vin == "" {
		return validation.ValidateStruct(v,
			validation.Field(
				&v.State,
				validation.When(
					v.State != "",
					validation.Match(stateRegex).
						Error("state format is invalid"),
				),
			),
			validation.Field(
				&v.Plate,
				validation.When(
					v.Plate != "",
					validation.Match(plateRegex).
						Error("plate format is invalid"),
				),
			),
		)
	}

	return validation.ValidateStruct(v,
		validation.Field(
			&v.Vin,
			validation.Required.
				Error("VIN is required"),
			validation.Length(17, 17).
				Error("VIN must be exactly 17 characters"),
			validation.Match(vinRegex).
				Error("VIN format is invalid"),
		),
	)
}
