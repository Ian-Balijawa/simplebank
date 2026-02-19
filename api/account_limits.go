package api

import (
	"errors"
	"net/http"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type accountIDUri struct {
	AccountID int64 `uri:"id" binding:"required,min=1"`
}

type upsertAccountLimitRequest struct {
	DailyTransferLimit int64 `json:"daily_transfer_limit" binding:"required,min=0"`
}

func (server *Server) getAccountLimit(ctx *gin.Context) {
	var uri accountIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if _, ok := server.getOwnedAccount(ctx, uri.AccountID); !ok {
		return
	}

	limit, err := server.store.GetAccountLimit(ctx, uri.AccountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, limit)
}

func (server *Server) upsertAccountLimit(ctx *gin.Context) {
	var uri accountIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req upsertAccountLimitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if _, ok := server.getOwnedAccount(ctx, uri.AccountID); !ok {
		return
	}

	limit, err := server.store.UpsertAccountLimit(ctx, db.UpsertAccountLimitParams{
		AccountID:          uri.AccountID,
		DailyTransferLimit: req.DailyTransferLimit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, limit)
}
