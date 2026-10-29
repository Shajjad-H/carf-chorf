package contact

import (
	"net/http"

	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

const templateName = "contact/index.html"

func Show(w http.ResponseWriter, r *http.Request) {
	req := ContactRequest{}
	if state, ok := auth_service.Current(r); ok {
		req.Email = state.Email
	}
	renderForm(w, r, http.StatusOK, req, map[string]string{}, nil)
}

func Submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}

	req := ContactRequest{
		Email:   r.PostFormValue("email"),
		Phone:   r.PostFormValue("phone"),
		Message: r.PostFormValue("message"),
	}

	if err := req.Validate(); err != nil {
		renderForm(w, r, http.StatusUnprocessableEntity, req, fieldErrors(err),
			[]string{"Please correct the highlighted fields."})
		return
	}

	if err := send(req); err != nil {
		simple_logger.Errorf("contact: failed to send notification: %v", err)
		renderForm(w, r, http.StatusBadGateway, req, map[string]string{},
			[]string{"We could not send your message right now. Please try again."})
		return
	}

	flashes_service.AddFlash(w, r, "success", "Thanks! Your message has been sent.")
	http.Redirect(w, r, "/contact", http.StatusSeeOther)
}

func renderForm(w http.ResponseWriter, r *http.Request, status int, req ContactRequest, fieldErrs map[string]string, errs []string) {
	ctx := pongo2.Context{
		"form": map[string]string{
			"email":   req.Email,
			"phone":   req.Phone,
			"message": req.Message,
		},
		"field_errors": fieldErrs,
	}
	if len(errs) > 0 {
		ctx["errors"] = errs
	}

	renderer.Pongo(w, r, templateName, status, ctx)
}
