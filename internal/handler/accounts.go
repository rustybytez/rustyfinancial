package handler

import (
	"context"
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
	Balance int64
}

type dashboardData struct {
	ByType map[string][]accountWithBalance
	Totals map[string]int64
}

func (h *AccountHandler) Dashboard(c echo.Context) error {
	accounts, err := h.store.ListAccounts(context.Background())
	if err != nil {
		return err
	}

	byType := map[string][]accountWithBalance{}
	totals := map[string]int64{}

	for _, a := range accounts {
		raw, err := h.store.GetAccountBalance(context.Background(), a.ID)
		if err != nil {
			return err
		}
		bal := toInt64(raw)
		awb := accountWithBalance{Account: a, Balance: bal}
		byType[a.Type] = append(byType[a.Type], awb)
		totals[a.Type] += bal
	}

	return render(c, "accounts/dashboard.html", dashboardData{ByType: byType, Totals: totals})
}

func (h *AccountHandler) NewForm(c echo.Context) error {
	return render(c, "accounts/form.html", nil)
}

func (h *AccountHandler) Create(c echo.Context) error {
	params := db.CreateAccountParams{
		Name:     c.FormValue("name"),
		Type:     c.FormValue("type"),
		Currency: c.FormValue("currency"),
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
		Name:     c.FormValue("name"),
		Type:     c.FormValue("type"),
		Currency: c.FormValue("currency"),
		ID:       id,
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
