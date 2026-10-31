package renderer

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"carf-chorf.com/app/app/kernel"
	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/pkg/money"
	"carf-chorf.com/app/pkg/response"
	"github.com/flosch/pongo2/v7"
	"github.com/gorilla/csrf"
)

var templateSet *pongo2.TemplateSet

func InitTemplates(dir string) {
	loader := pongo2.MustNewLocalFileSystemLoader(dir)
	templateSet = pongo2.NewSet(dir, loader)
	registerFilters()
}

func registerFilters() {
	templateSet.RegisterFilter("is_defined", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		return pongo2.AsValue(!in.IsNil()), nil
	})
	templateSet.RegisterFilter("not_empty", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		if in.IsNil() {
			return pongo2.AsValue(false), nil
		}
		switch v := in.Interface().(type) {
		case string:
			return pongo2.AsValue(v != ""), nil
		default:
			return pongo2.AsValue(true), nil
		}
	})
	templateSet.RegisterFilter("format", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		if param == nil || param.IsNil() {
			return in, nil
		}
		return pongo2.AsValue(fmt.Sprintf(in.String(), param.Interface())), nil
	})
	templateSet.RegisterFilter("string", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		if in.IsNil() {
			return pongo2.AsValue(""), nil
		}
		return pongo2.AsValue(in.String()), nil
	})
	templateSet.RegisterFilter("money", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		switch v := in.Interface().(type) {
		case uint:
			return pongo2.AsValue(money.FormatCents(v)), nil
		case uint64:
			return pongo2.AsValue(money.FormatCents(uint(v))), nil
		case int:
			return pongo2.AsValue(money.FormatCents(uint(v))), nil
		default:
			return in, nil
		}
	})
	templateSet.RegisterFilter("date", func(in, param *pongo2.Value) (*pongo2.Value, error) {
		t, ok := in.Interface().(time.Time)
		if !ok {
			return in, nil
		}
		layout := "Jan 2, 2006"
		if param != nil && !param.IsNil() && param.String() != "" {
			layout = param.String()
		}
		return pongo2.AsValue(t.Format(layout)), nil
	})
}

func baseContext(w http.ResponseWriter, r *http.Request, extra pongo2.Context) pongo2.Context {
	state, authenticated := auth_service.Current(r)

	ctx := pongo2.Context{
		"flashes":            flashes_service.GetFlashes(w, r, "success"),
		"errors":             flashes_service.GetFlashes(w, r, "error"),
		"csrf_token":         csrf.Token(r),
		"authenticated":      authenticated,
		"current_user_email": state.Email,
		"user_initial":       userInitial(state.Email),
		"current_path":       r.URL.Path,
		"year":               time.Now().Year(),
		"turnstile_site_key": kernel.Turnstile.SiteKey,
	}

	if extra != nil {
		ctx.Update(extra)
	}

	return ctx
}

// Render executes a template with the given context without an HTTP request.
// It is used for templates that are not page responses, such as emails.
func Render(name string, ctx pongo2.Context) (string, error) {
	if templateSet == nil {
		return "", fmt.Errorf("renderer: templates are not initialized")
	}
	tpl, err := templateSet.FromCache(name)
	if err != nil {
		return "", err
	}
	return tpl.Execute(ctx)
}

func Pongo(w http.ResponseWriter, r *http.Request, name string, status int, ctx pongo2.Context) {
	if templateSet == nil {
		response.Error(w, http.StatusInternalServerError, "please initialize templates with renderer.InitTemplates(dir)")
		return
	}

	tpl, err := templateSet.FromCache(name)
	if err != nil {
		pongoError(w, http.StatusInternalServerError, err)
		return
	}

	ctx = baseContext(w, r, ctx)
	html, err := tpl.ExecuteBytes(ctx)

	if err != nil {
		pongoError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(html)
}

func userInitial(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "?"
	}
	return strings.ToUpper(email[:1])
}

func pongoError(w http.ResponseWriter, status int, err error) {
	if kernel.IsDev() {
		response.Error(w, http.StatusInternalServerError, err.Error())
	}
	response.Error(w, http.StatusInternalServerError, "renderer error")
}
