package dto_validations

import (
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type VehicleInfoRequest struct {
	Vin   string
	Plate string
	State string
}

type PlateRequest struct {
	Plate string
	State string
}

var (
	vinRegex   = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
	stateRegex = regexp.MustCompile(`^[A-Z]{2}$`)
	plateRegex = regexp.MustCompile(`^[A-Z0-9 -]{1,10}$`)
)

func (p *PlateRequest) Validate() error {
	p.Plate = strings.ToUpper(strings.TrimSpace(p.Plate))
	p.State = strings.ToUpper(strings.TrimSpace(p.State))
	return validation.ValidateStruct(p,
		validation.Field(&p.Plate, validation.Required.Error("plate is required"), validation.Match(plateRegex).Error("plate format is invalid")),
		validation.Field(&p.State, validation.Required.Error("state is required"), validation.Match(stateRegex).Error("state format is invalid")),
	)
}

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
