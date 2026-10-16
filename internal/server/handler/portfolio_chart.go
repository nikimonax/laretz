package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
)

// chartRow is the JSON row for uPlot: aligned arrays of hours (epoch
// seconds) and values (null where the hour is not held), with the
// display-time flags of each hour.
type chartRow struct {
	Hours   []int64    `json:"ts"`
	Values  []*float64 `json:"value"`
	Stale   []bool     `json:"stale"`
	Missing []bool     `json:"missing"`
}

func newChartRow(hours []domain.HourValue) chartRow {
	row := chartRow{
		Hours:   make([]int64, 0, len(hours)),
		Values:  make([]*float64, 0, len(hours)),
		Stale:   make([]bool, 0, len(hours)),
		Missing: make([]bool, 0, len(hours)),
	}

	for _, hour := range hours {
		row.Hours = append(row.Hours, hour.Hour.Unix())

		if hour.Held {
			value := hour.Amount.InexactFloat64()
			row.Values = append(row.Values, &value)
		} else {
			row.Values = append(row.Values, nil)
		}

		row.Stale = append(row.Stale, hour.Stale)
		row.Missing = append(row.Missing, hour.MissingPrices)
	}

	return row
}

type PortfolioChartHandler struct {
	logger logr.Logger
	chart  *usecase.GetPortfolioChartUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *PortfolioChartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || portfolioID <= 0 {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	hours, err := handler.chart.Execute(r.Context(), usecase.GetPortfolioChart{ID: domain.PortfolioID(portfolioID)})

	if err != nil {
		if errors.Is(err, app.ErrEntityNotFound) {
			renderNotFoundPage(handler.logger, w, r)
			return
		}

		handler.logger.Error(err, "render chart")
		http.Error(w, messageSomethingWentWrong, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(newChartRow(hours)); err != nil {
		handler.logger.Error(err, "encode chart")
	}
}

func NewPortfolioChartHandler(logger logr.Logger, chart *usecase.GetPortfolioChartUseCase) http.Handler {
	return &PortfolioChartHandler{logger: logger, chart: chart}
}
