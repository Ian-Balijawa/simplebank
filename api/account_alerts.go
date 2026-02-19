package api

import (
	"errors"
	"net/http"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type upsertAccountAlertRequest struct {
	LowBalanceThreshold  int64 `json:"low_balance_threshold" binding:"required,min=0"`
	HighBalanceThreshold int64 `json:"high_balance_threshold" binding:"required,min=0"`
}

func (server *Server) getAccountAlert(ctx *gin.Context) {
	var uri accountIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if _, ok := server.getOwnedAccount(ctx, uri.AccountID); !ok {
		return
	}

	alert, err := server.store.GetAccountAlert(ctx, uri.AccountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, alert)
}

func (server *Server) upsertAccountAlert(ctx *gin.Context) {
	var uri accountIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req upsertAccountAlertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.LowBalanceThreshold > 0 && req.HighBalanceThreshold > 0 && req.LowBalanceThreshold > req.HighBalanceThreshold {
		ctx.JSON(http.StatusBadRequest, errorResponse(errInvalidAmountRange))
		return
	}

	if _, ok := server.getOwnedAccount(ctx, uri.AccountID); !ok {
		return
	}

	alert, err := server.store.UpsertAccountAlert(ctx, db.UpsertAccountAlertParams{
		AccountID:            uri.AccountID,
		LowBalanceThreshold:  req.LowBalanceThreshold,
		HighBalanceThreshold: req.HighBalanceThreshold,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, alert)
}
