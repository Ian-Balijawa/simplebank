package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/Ian-Balijawa/simplebank/token"
	"github.com/Ian-Balijawa/simplebank/worker"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	toAccount, valid := server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	if err := server.checkDailyTransferLimit(ctx, req.FromAccountID, req.Amount); err != nil {
		if errors.Is(err, db.ErrDailyTransferLimitExceeded) {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	server.sendBalanceAlerts(ctx, fromAccount, result.FromAccount)
	server.sendBalanceAlerts(ctx, toAccount, result.ToAccount)

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}

func (server *Server) checkDailyTransferLimit(ctx *gin.Context, accountID int64, amount int64) error {
	limit, err := server.store.GetAccountLimit(ctx, accountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if limit.DailyTransferLimit <= 0 {
		return nil
	}

	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	total, err := server.store.GetDailyTransferTotal(ctx, db.GetDailyTransferTotalParams{
		FromAccountID: accountID,
		CreatedAt:     dayStart,
		CreatedAt_2:   dayEnd,
	})
	if err != nil {
		return err
	}

	if total+amount > limit.DailyTransferLimit {
		return db.ErrDailyTransferLimitExceeded
	}

	return nil
}

func (server *Server) sendBalanceAlerts(ctx *gin.Context, before db.Account, after db.Account) {
	if server.taskDistributor == nil {
		return
	}

	alert, err := server.store.GetAccountAlert(ctx, after.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return
		}
		return
	}

	if alert.LowBalanceThreshold > 0 &&
		before.Balance > alert.LowBalanceThreshold &&
		after.Balance <= alert.LowBalanceThreshold {
		_ = server.taskDistributor.DistributeTaskSendAccountAlert(ctx, &worker.PayloadAccountAlert{
			Username:  after.Owner,
			AccountID: after.ID,
			Balance:   after.Balance,
			Threshold: alert.LowBalanceThreshold,
			Direction: worker.AlertDirectionLow,
			Currency:  after.Currency,
		})
	}

	if alert.HighBalanceThreshold > 0 &&
		before.Balance < alert.HighBalanceThreshold &&
		after.Balance >= alert.HighBalanceThreshold {
		_ = server.taskDistributor.DistributeTaskSendAccountAlert(ctx, &worker.PayloadAccountAlert{
			Username:  after.Owner,
			AccountID: after.ID,
			Balance:   after.Balance,
			Threshold: alert.HighBalanceThreshold,
			Direction: worker.AlertDirectionHigh,
			Currency:  after.Currency,
		})
	}
}
