// Package admin implements the isolated admin dashboard.
package admin

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/app/services/admin_auth"
	"carf-chorf.com/app/app/services/admin_service"
	"carf-chorf.com/app/app/services/mail_service"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
	"github.com/gorilla/csrf"
)

const pageSize = 25

// PageData describes a paginated result and its navigation URLs.
type PageData struct {
	Number     int
	TotalPages int
	Total      int64
	TotalPaid  uint64
	HasPrev    bool
	HasNext    bool
	PrevURL    string
	NextURL    string
}

// RequireAuth protects admin pages.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !admin_auth.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoginPage renders the admin key form.
func LoginPage(w http.ResponseWriter, r *http.Request) {
	if admin_auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderAdmin(w, r, "admin/login.html", http.StatusOK, pongo2.Context{})
}

// LoginSubmit validates the admin key with a global failure limiter.
func LoginSubmit(limiter *admin_auth.FailureLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if limiter.Blocked() {
			retry := int(limiter.RetryAfter().Seconds())
			if retry < 1 {
				retry = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			renderAdmin(w, r, "admin/login.html", http.StatusTooManyRequests, pongo2.Context{
				"error": "Too many failed attempts. Please try again later.",
			})
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form submission", http.StatusBadRequest)
			return
		}

		if !adminKeyEqual(r.PostFormValue("admin_key"), kernel.Dashboard.AdminKey) {
			limiter.Fail()
			renderAdmin(w, r, "admin/login.html", http.StatusUnauthorized, pongo2.Context{
				"error": "Invalid admin key.",
			})
			return
		}

		limiter.Reset()
		if err := admin_auth.SetAuthenticated(w, r); err != nil {
			simple_logger.Errorf("admin: set session: %v", err)
			renderAdmin(w, r, "admin/login.html", http.StatusInternalServerError, pongo2.Context{
				"error": "Something went wrong. Please try again.",
			})
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Logout clears the admin session.
func Logout(w http.ResponseWriter, r *http.Request) {
	if err := admin_auth.Clear(w, r); err != nil {
		simple_logger.Errorf("admin: clear session: %v", err)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Index renders the dashboard summary.
func Index(w http.ResponseWriter, r *http.Request) {
	summary, err := admin_service.GetSummary()
	if err != nil {
		simple_logger.Errorf("admin: summary: %v", err)
		http.Error(w, "Could not load dashboard.", http.StatusInternalServerError)
		return
	}
	renderAdmin(w, r, "admin/index.html", http.StatusOK, pongo2.Context{"summary": summary})
}

// Transactions renders the filtered transaction list.
func Transactions(w http.ResponseWriter, r *http.Request) {
	filters := admin_service.TransactionFilters{
		Email:  r.URL.Query().Get("email"),
		Token:  r.URL.Query().Get("token"),
		Pid:    r.URL.Query().Get("pid"),
		Status: r.URL.Query().Get("status"),
		From:   parseDate(r.URL.Query().Get("from")),
		To:     parseDate(r.URL.Query().Get("to")),
	}
	page := pageNumber(r)

	transactions, total, paid, err := admin_service.ListTransactions(filters, page, pageSize)
	if err != nil {
		simple_logger.Errorf("admin: transactions: %v", err)
		http.Error(w, "Could not load transactions.", http.StatusInternalServerError)
		return
	}

	renderAdmin(w, r, "admin/transactions.html", http.StatusOK, pongo2.Context{
		"transactions": transactions,
		"page":         buildPage(r, total, page, pageSize),
		"total_paid":   paid,
		"filters": map[string]string{
			"email":  filters.Email,
			"token":  filters.Token,
			"pid":    filters.Pid,
			"status": filters.Status,
			"from":   formatDate(filters.From),
			"to":     formatDate(filters.To),
		},
	})
}

// Orders renders the filtered report order list.
func Orders(w http.ResponseWriter, r *http.Request) {
	filters := admin_service.OrderFilters{
		Email: r.URL.Query().Get("email"),
		VIN:   r.URL.Query().Get("vin"),
	}
	page := pageNumber(r)

	orders, total, err := admin_service.ListOrders(filters, page, pageSize)
	if err != nil {
		simple_logger.Errorf("admin: orders: %v", err)
		http.Error(w, "Could not load orders.", http.StatusInternalServerError)
		return
	}

	renderAdmin(w, r, "admin/orders.html", http.StatusOK, pongo2.Context{
		"orders": orders,
		"page":   buildPage(r, total, page, pageSize),
		"filters": map[string]string{
			"email": filters.Email,
			"vin":   filters.VIN,
		},
	})
}

// DailyTransactions renders paid transactions for the selected day.
func DailyTransactions(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	day := parseDailyDate(r.URL.Query().Get("date"), now)

	groups, total, err := admin_service.DailyReport(day)
	if err != nil {
		simple_logger.Errorf("admin: daily transactions: %v", err)
		http.Error(w, "Could not load daily transactions.", http.StatusInternalServerError)
		return
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	isToday := day.Equal(today)
	prev := day.AddDate(0, 0, -1)
	next := day.AddDate(0, 0, 1)

	ctx := pongo2.Context{
		"groups":        groups,
		"total":         total,
		"day":           day.Format("Monday, January 2, 2006"),
		"selected_date": day.Format("2006-01-02"),
		"today_date":    today.Format("2006-01-02"),
		"is_today":      isToday,
		"prev_url":      "/daily-transactions?date=" + prev.Format("2006-01-02"),
	}
	if !isToday {
		ctx["next_url"] = "/daily-transactions?date=" + next.Format("2006-01-02")
	}

	renderAdmin(w, r, "admin/daily_transactions.html", http.StatusOK, ctx)
}

func parseDailyDate(raw string, now time.Time) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if strings.TrimSpace(raw) == "" {
		return today
	}
	day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), now.Location())
	if err != nil || day.After(today) {
		return today
	}
	return day
}

// Users renders the filtered user list.
func Users(w http.ResponseWriter, r *http.Request) {
	filters, filterValues := parseUserFilters(r.URL.Query())
	page := pageNumber(r)

	users, total, err := admin_service.ListUsers(filters, page, pageSize)
	if err != nil {
		simple_logger.Errorf("admin: users: %v", err)
		http.Error(w, "Could not load users.", http.StatusInternalServerError)
		return
	}

	recipientCount, err := admin_service.CountReminderRecipients(filters)
	if err != nil {
		simple_logger.Errorf("admin: count reminder recipients: %v", err)
		recipientCount = 0
	}

	renderAdmin(w, r, "admin/users.html", http.StatusOK, pongo2.Context{
		"users":           users,
		"page":            buildPage(r, total, page, pageSize),
		"notice":          r.URL.Query().Get("notice"),
		"filters":         filterValues,
		"recipient_count": recipientCount,
	})
}

// RemindUsers sends reminders to every user matching the current filters.
func RemindUsers(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}

	filters, filterValues := parseUserFilters(r.PostForm)
	users, err := admin_service.ListReminderRecipients(filters)
	if err != nil {
		simple_logger.Errorf("admin: list reminder recipients: %v", err)
		redirectUsersNotice(w, r, filterValues, "Could not load matching users.")
		return
	}
	if len(users) == 0 {
		redirectUsersNotice(w, r, filterValues, "No active users with unused credits match these filters.")
		return
	}

	sent, failed := 0, 0
	for _, user := range users {
		if err := mail_service.SendCreditReminder(user.Email, user.Credits); err != nil {
			failed++
			simple_logger.Errorf("admin: send credit reminder to %s: %v", user.Email, err)
			continue
		}
		sent++
	}

	redirectUsersNotice(w, r, filterValues, "Reminder emails sent: "+strconv.Itoa(sent)+". Failed: "+strconv.Itoa(failed)+".")
}

func parseUserFilters(values url.Values) (admin_service.UserFilters, map[string]string) {
	rawMinCredits := strings.TrimSpace(values.Get("min_credits"))
	filters := admin_service.UserFilters{
		Email:  values.Get("email"),
		Token:  values.Get("token"),
		Active: values.Get("active"),
		From:   parseDate(values.Get("from")),
		To:     parseDate(values.Get("to")),
	}
	if rawMinCredits != "" {
		if value, err := strconv.ParseUint(rawMinCredits, 10, 32); err == nil {
			credits := uint(value)
			filters.MinCredits = &credits
		}
	}

	return filters, map[string]string{
		"email":       filters.Email,
		"token":       filters.Token,
		"min_credits": rawMinCredits,
		"active":      filters.Active,
		"from":        formatDate(filters.From),
		"to":          formatDate(filters.To),
	}
}

func renderAdmin(w http.ResponseWriter, r *http.Request, name string, status int, ctx pongo2.Context) {
	ctx["csrf_token"] = csrf.Token(r)
	ctx["current_path"] = r.URL.Path
	ctx["public_domain"] = strings.TrimRight(kernel.Server.Domain, "/")
	html, err := renderer.Render(name, ctx)
	if err != nil {
		simple_logger.Errorf("admin: render %s: %v", name, err)
		http.Error(w, "renderer error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(html))
}

func adminKeyEqual(got, want string) bool {
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotHash[:], wantHash[:]) == 1
}

func pageNumber(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func buildPage(r *http.Request, total int64, page, size int) PageData {
	pages := int((total + int64(size) - 1) / int64(size))
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}

	data := PageData{
		Number:     page,
		TotalPages: pages,
		Total:      total,
		HasPrev:    page > 1,
		HasNext:    page < pages,
	}
	if data.HasPrev {
		data.PrevURL = pageURL(r, page-1)
	}
	if data.HasNext {
		data.NextURL = pageURL(r, page+1)
	}
	return data
}

func pageURL(r *http.Request, page int) string {
	values := r.URL.Query()
	values.Set("page", strconv.Itoa(page))
	return r.URL.Path + "?" + values.Encode()
}

func parseDate(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	date, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return time.Time{}
	}
	return date
}

func formatDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}

func redirectUsersNotice(w http.ResponseWriter, r *http.Request, filters map[string]string, message string) {
	values := url.Values{}
	for key, value := range filters {
		if value != "" {
			values.Set(key, value)
		}
	}
	values.Set("notice", message)
	http.Redirect(w, r, "/users?"+values.Encode(), http.StatusSeeOther)
}
