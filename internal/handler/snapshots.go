package handler

import (
	"context"

	"rustyfinancial/internal/store"

	"github.com/labstack/echo/v4"
)

type SnapshotHandler struct {
	store *store.Store
}

func NewSnapshotHandler(s *store.Store) *SnapshotHandler {
	return &SnapshotHandler{store: s}
}

type snapshotRow struct {
	Date     string
	Totals   map[string]int64
	NetWorth int64
}

var snapshotTypeOrder = []string{"cash", "investment", "real_estate", "liability", "other"}

func (h *SnapshotHandler) History(c echo.Context) error {
	ctx := context.Background()

	accounts, err := h.store.ListAccounts(ctx)
	if err != nil {
		return err
	}
	accountTypes := make(map[int64]string, len(accounts))
	for _, a := range accounts {
		accountTypes[a.ID] = a.Type
	}

	snapshots, err := h.store.ListAllSnapshots(ctx)
	if err != nil {
		return err
	}

	// Group by date, sum balance per account type.
	rowIndex := map[string]int{}
	var rows []snapshotRow
	for _, s := range snapshots {
		date := s.Date.Format("2006-01-02")
		idx, ok := rowIndex[date]
		if !ok {
			idx = len(rows)
			rowIndex[date] = idx
			rows = append(rows, snapshotRow{
				Date:   date,
				Totals: map[string]int64{},
			})
		}
		t := accountTypes[s.AccountID]
		rows[idx].Totals[t] += s.Balance
	}

	// Compute net worth per row.
	for i, row := range rows {
		var nw int64
		for t, total := range row.Totals {
			if t == "liability" {
				nw -= total
			} else {
				nw += total
			}
		}
		rows[i].NetWorth = nw
	}

	return render(c, "snapshots/history.html", map[string]any{
		"Rows":      rows,
		"TypeOrder": snapshotTypeOrder,
	})
}

func (h *SnapshotHandler) Take(c echo.Context) error {
	if err := h.store.TakeSnapshot(context.Background()); err != nil {
		return err
	}
	return c.Redirect(302, "/snapshots")
}
