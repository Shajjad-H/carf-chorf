package contact

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ContactRequest struct {
	Email   string
	Phone   string
	Message string
}

var usPhoneRegex = regexp.MustCompile(`^(\+?1[\s.-]?)?\(?([2-9]\d{2})\)?[\s.-]?(\d{3})[\s.-]?(\d{4})$`)

func (c *ContactRequest) Validate() error {
	c.Email = strings.TrimSpace(c.Email)
	c.Phone = strings.TrimSpace(c.Phone)
	c.Message = strings.TrimSpace(c.Message)

	if err := validation.ValidateStruct(c,
		validation.Field(&c.Email,
			validation.Required.Error("Email is required"),
			is.EmailFormat.Error("Enter a valid email address"),
		),
		validation.Field(&c.Phone,
			validation.Required.Error("Phone is required"),
			validation.Match(usPhoneRegex).Error("Enter a valid US phone number"),
		),
		validation.Field(&c.Message,
			validation.Required.Error("Message is required"),
			validation.Length(10, 2000).Error("Message must be between 10 and 2000 characters"),
		),
	); err != nil {
		return err
	}

	m := usPhoneRegex.FindStringSubmatch(c.Phone)
	c.Phone = fmt.Sprintf("+1 (%s) %s-%s", m[2], m[3], m[4])
	return nil
}

func fieldErrors(err error) map[string]string {
	out := map[string]string{}
	errs, ok := err.(validation.Errors)
	if !ok {
		return out
	}
	for field, e := range errs {
		out[strings.ToLower(field)] = e.Error()
	}
	return out
}
