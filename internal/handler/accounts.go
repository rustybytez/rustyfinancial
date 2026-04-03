package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"rustyfinancial/internal/db"
	"rustyfinancial/internal/store"

	"github.com/labstack/echo/v4"
)

type AccountHandler struct {
	store *store.Store
}

func NewAccountHandler(s *store.Store) *AccountHandler {
	return &AccountHandler{store: s}
}

type accountWithBalance struct {
	db.Account
	Balance     int64
	ManuallySet bool
}

var typeOrder = []string{"cash", "investment", "real_estate", "liability", "other"}

type dashboardData struct {
	ByType    map[string][]accountWithBalance
	Totals    map[string]int64
	TypeOrder []string
	EditMode  bool
	NetWorth  int64
}

func (h *AccountHandler) Dashboard(c echo.Context) error {
	accounts, err := h.store.ListAccounts(context.Background())
	if err != nil {
		return err
	}

	byType := map[string][]accountWithBalance{}
	totals := map[string]int64{}

	for _, a := range accounts {
		var bal int64
		var manuallySet bool
		if a.ManualBalance.Valid {
			bal = a.ManualBalance.Int64
			manuallySet = true
		} else {
			raw, err := h.store.GetAccountBalance(context.Background(), a.ID)
			if err != nil {
				return err
			}
			bal = toInt64(raw)
		}
		awb := accountWithBalance{Account: a, Balance: bal, ManuallySet: manuallySet}
		byType[a.Type] = append(byType[a.Type], awb)
		totals[a.Type] += bal
	}

	var netWorth int64
	for t, total := range totals {
		if t == "liability" {
			netWorth -= total
		} else {
			netWorth += total
		}
	}

	return render(c, "accounts/dashboard.html", dashboardData{
		ByType:    byType,
		Totals:    totals,
		TypeOrder: typeOrder,
		EditMode:  c.QueryParam("edit") == "true",
		NetWorth:  netWorth,
	})
}

func (h *AccountHandler) NewForm(c echo.Context) error {
	return render(c, "accounts/form.html", nil)
}

func (h *AccountHandler) Create(c echo.Context) error {
	params := db.CreateAccountParams{
		Name:        c.FormValue("name"),
		Type:        c.FormValue("type"),
		Currency:    c.FormValue("currency"),
		Institution: c.FormValue("institution"),
	}
	if params.Currency == "" {
		params.Currency = "USD"
	}
	if _, err := h.store.CreateAccount(context.Background(), params); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (h *AccountHandler) EditForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	account, err := h.store.GetAccount(context.Background(), id)
	if err != nil {
		return echo.ErrNotFound
	}
	return render(c, "accounts/form.html", account)
}

func (h *AccountHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	params := db.UpdateAccountParams{
		Name:        c.FormValue("name"),
		Type:        c.FormValue("type"),
		Currency:    c.FormValue("currency"),
		Institution: c.FormValue("institution"),
		ID:          id,
	}
	if _, err := h.store.UpdateAccount(context.Background(), params); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (h *AccountHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	if err := h.store.DeleteAccount(context.Background(), id); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

type balanceFormData struct {
	db.Account
	CurrentBalance string // formatted as "1234.56" for the input
}

func (h *AccountHandler) BalanceForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	account, err := h.store.GetAccount(context.Background(), id)
	if err != nil {
		return echo.ErrNotFound
	}
	var current string
	if account.ManualBalance.Valid {
		current = fmt.Sprintf("%.2f", float64(account.ManualBalance.Int64)/100)
	}
	return render(c, "accounts/balance.html", balanceFormData{Account: account, CurrentBalance: current})
}

func (h *AccountHandler) SetBalance(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}

	raw := c.FormValue("balance")
	if raw == "" {
		// Clear manual balance — fall back to transactions
		if err := h.store.ClearManualBalance(context.Background(), id); err != nil {
			return err
		}
		if c.Request().Header.Get("HX-Request") == "true" {
			return c.NoContent(http.StatusNoContent)
		}
		return c.Redirect(http.StatusFound, "/")
	}

	dollars, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	cents := int64(math.Round(dollars * 100))

	if err := h.store.SetManualBalance(context.Background(), db.SetManualBalanceParams{
		ManualBalance: sql.NullInt64{Int64: cents, Valid: true},
		ID:            id,
	}); err != nil {
		return err
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.NoContent(http.StatusNoContent)
	}
	return c.Redirect(http.StatusFound, "/")
}
