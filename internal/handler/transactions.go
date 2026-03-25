package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"rustyfinancial/internal/db"
	"rustyfinancial/internal/store"

	"github.com/labstack/echo/v4"
)

type TransactionHandler struct {
	store *store.Store
}

func NewTransactionHandler(s *store.Store) *TransactionHandler {
	return &TransactionHandler{store: s}
}

type transactionPageData struct {
	Account      db.Account
	Transactions []db.Transaction
	Balance      int64
}

func (h *TransactionHandler) List(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	account, err := h.store.GetAccount(context.Background(), id)
	if err != nil {
		return echo.ErrNotFound
	}
	txs, err := h.store.ListTransactionsByAccount(context.Background(), id)
	if err != nil {
		return err
	}
	rawBal, err := h.store.GetAccountBalance(context.Background(), id)
	if err != nil {
		return err
	}
	return render(c, "transactions/list.html", transactionPageData{
		Account:      account,
		Transactions: txs,
		Balance:      toInt64(rawBal),
	})
}

func (h *TransactionHandler) Create(c echo.Context) error {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	amountStr := c.FormValue("amount")
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	// convert dollars to cents
	amountCents := int64(amountFloat * 100)

	date, err := time.Parse("2006-01-02", c.FormValue("date"))
	if err != nil {
		return echo.ErrBadRequest
	}
	params := db.CreateTransactionParams{
		AccountID:   accountID,
		Amount:      amountCents,
		Description: c.FormValue("description"),
		Date:        date,
	}
	if _, err := h.store.CreateTransaction(context.Background(), params); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/accounts/"+c.Param("id"))
}

func (h *TransactionHandler) Delete(c echo.Context) error {
	accountID := c.Param("id")
	txID, err := strconv.ParseInt(c.Param("txid"), 10, 64)
	if err != nil {
		return echo.ErrBadRequest
	}
	if err := h.store.DeleteTransaction(context.Background(), txID); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/accounts/"+accountID)
}
