package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
)

const layoutDatetimeLocal = "2006-01-02T15:04"

func renderPortfolioPage(
	w http.ResponseWriter,
	r *http.Request,
	logger logr.Logger,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	portfolioID domain.PortfolioID,
	message string,
	dialog string,
) {
	responseRenderer := renderer.Portfolio(portfolioGet, assetsList, portfolioID, parsePage(r, "ops_page"))

	if err := responseRenderer.Render(r.Context(), w, renderer.PortfolioWithError(message, dialog)); err != nil {
		logger.Error(err, "render")
	}
}

func parseOperationForm(r *http.Request, portfolioID domain.PortfolioID) (usecase.CreateOperation, string) {
	cmd := usecase.CreateOperation{PortfolioID: portfolioID}

	if err := r.ParseForm(); err != nil {
		return cmd, "Invalid form"
	}

	operationType := domain.OperationType(r.FormValue("type"))

	if operationType != domain.OperationBuy && operationType != domain.OperationSell {
		return cmd, "Type must be buy or sell."
	}

	cmd.Type = operationType

	assetID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("asset_id")), 10, 64)

	if err != nil || assetID <= 0 {
		return cmd, "Asset must be selected."
	}

	cmd.AssetID = domain.AssetID(assetID)

	executedAt, err := time.ParseInLocation(layoutDatetimeLocal, r.FormValue("executed_at"), time.Local)

	if err != nil {
		return cmd, "Invalid time."
	}

	if executedAt.After(time.Now()) {
		return cmd, "Time must not be in the future."
	}

	cmd.ExecutedAt = executedAt

	quantity, err := decimal.NewFromString(strings.TrimSpace(r.FormValue("quantity")))

	if err != nil {
		return cmd, "Invalid quantity."
	}

	if quantity.LessThanOrEqual(decimal.Zero) {
		return cmd, "Quantity must be greater than zero."
	}

	cmd.Quantity = quantity

	totalUSD, err := decimal.NewFromString(strings.TrimSpace(r.FormValue("total_usd")))

	if err != nil {
		return cmd, "Invalid total."
	}

	if totalUSD.LessThanOrEqual(decimal.Zero) {
		return cmd, "Total must be greater than zero."
	}

	cmd.TotalUSD = totalUSD

	if message := parseOperationFee(r, &cmd); message != "" {
		return cmd, message
	}

	cmd.Note = strings.TrimSpace(r.FormValue("note"))

	return cmd, ""
}

func parseOperationFee(r *http.Request, cmd *usecase.CreateOperation) string {
	feeAmountValue := strings.TrimSpace(r.FormValue("fee_amount"))
	feeAssetIDValue := strings.TrimSpace(r.FormValue("fee_asset_id"))

	if feeAmountValue == "" && feeAssetIDValue == "" {
		return ""
	}

	if feeAmountValue == "" || feeAssetIDValue == "" {
		return "Fee requires both amount and asset."
	}

	feeAmount, err := decimal.NewFromString(feeAmountValue)

	if err != nil {
		return "Invalid fee amount."
	}

	feeAssetID, err := strconv.ParseInt(feeAssetIDValue, 10, 64)

	if err != nil || feeAssetID <= 0 {
		return "Invalid fee asset."
	}

	cmd.FeeAmount = &feeAmount

	feeAsset := domain.AssetID(feeAssetID)
	cmd.FeeAssetID = &feeAsset

	return ""
}
