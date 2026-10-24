package renderer

import (
	"fmt"
	"net/http"

	"carf-chorf.com/app/app/kernel"
	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
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
}

func baseContext(w http.ResponseWriter, r *http.Request, extra pongo2.Context) pongo2.Context {
	state, authenticated := auth_service.Current(r)

	ctx := pongo2.Context{
		"flashes":            flashes_service.GetFlashes(w, r, "success"),
		"errors":             flashes_service.GetFlashes(w, r, "error"),
		"csrf_token":         csrf.Token(r),
		"authenticated":      authenticated,
		"current_user_email": state.Email,
	}

	if extra != nil {
		ctx.Update(extra)
	}

	return ctx
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

func pongoError(w http.ResponseWriter, status int, err error) {
	if kernel.IsDev() {
		response.Error(w, http.StatusInternalServerError, err.Error())
	}
	response.Error(w, http.StatusInternalServerError, "renderer error")
}
