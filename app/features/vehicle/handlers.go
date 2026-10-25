// Package vehicle implements the public vehicle lookup.
package vehicle

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"carf-chorf.com/app/app/services/carbpi_service"
	"carf-chorf.com/app/app/services/turnstile_service"
	dto "carf-chorf.com/app/pkg/dto_validations"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
	"github.com/go-chi/chi/v5/middleware"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const homeTemplate = "home/index.html"

// FormContext returns the initial context for the home-page lookup form.
func FormContext() pongo2.Context {
	return formContext(dto.LookupRequest{}, nil)
}

// Lookup validates the lookup form, verifies Turnstile, resolves or refreshes
// the cached vehicle record, and redirects to its tokenized checkout page.
func Lookup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}

	req := dto.LookupRequest{
		Vin:   r.PostFormValue("vin"),
		Plate: r.PostFormValue("plate"),
		State: r.PostFormValue("state"),
	}

	// Validate before Turnstile so bad input does not consume a challenge
	// token or reach the paid provider.
	if err := req.Validate(); err != nil {
		renderLookup(
			w, r,
			http.StatusUnprocessableEntity,
			req, fieldErrors(err),
			[]string{"Please correct the highlighted fields."},
		)
		return
	}

	if err := turnstile_service.Verify(r.Context(), r.PostFormValue("cf-turnstile-response"), clientIP(r)); err != nil {
		simple_logger.Warnf("vehicle: turnstile verification failed: %v", err)
		renderLookup(
			w, r,
			http.StatusUnprocessableEntity,
			req, nil,
			[]string{"Please complete the verification challenge and try again."},
		)
		return
	}

	record, err := lookup(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, carbpi_service.ErrNoVIN):
			// The provider returned a record but no usable VIN. The message
			// depends on which lookup path the user chose.
			message := "We couldn't verify that VIN."
			if req.Vin == "" {
				message = "We couldn't find a VIN for that plate and state."
			}
			renderLookup(w, r, http.StatusUnprocessableEntity, req, nil, []string{message})
		case carbpi_service.IsNotFound(err):
			// A plate lookup and a VIN lookup fail for different reasons.
			message := "No CARFAX records found for this vehicle."
			if req.Vin == "" {
				message = "We couldn't find a vehicle for that plate and state."
			}
			renderLookup(w, r, http.StatusNotFound, req, nil, []string{message})
		default:
			simple_logger.Errorf("vehicle: lookup failed: %v", err)
			renderLookup(w, r, http.StatusBadGateway, req, nil,
				[]string{"Vehicle lookup is temporarily unavailable. Please try again."})
		}
		return
	}

	http.Redirect(w, r, "/checkout?token="+url.QueryEscape(record.Token), http.StatusSeeOther)
}

// renderLookup re-renders the home page with submitted values and error state.
func renderLookup(w http.ResponseWriter, r *http.Request, status int, req dto.LookupRequest, fieldErrs map[string]string, errs []string) {
	ctx := formContext(req, fieldErrs)
	if len(errs) > 0 {
		ctx["errors"] = errs
	}
	renderer.Pongo(w, r, homeTemplate, status, ctx)
}

// formContext builds the template context. The active tab is inferred from the
// submitted fields so errors appear on the tab the user was using.
func formContext(req dto.LookupRequest, fieldErrs map[string]string) pongo2.Context {
	tab := "vin"
	if req.Vin == "" && (req.Plate != "" || req.State != "") {
		tab = "plate"
	}

	if fieldErrs == nil {
		fieldErrs = map[string]string{}
	}

	return pongo2.Context{
		"tab": tab,
		"form": map[string]string{
			"vin":   req.Vin,
			"plate": req.Plate,
			"state": req.State,
		},
		"states":       States(),
		"field_errors": fieldErrs,
	}
}

// fieldErrors converts ozzo-validation errors into lowercase template keys.
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

// clientIP prefers the IP resolved by chi's ClientIP middleware and falls back
// to RemoteAddr. Turnstile uses it to bind the challenge to the request source.
func clientIP(r *http.Request) string {
	if ip := middleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
