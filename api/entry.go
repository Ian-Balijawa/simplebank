package api

import (
	"net/http"
	"time"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type listEntriesUri struct {
	AccountID int64 `uri:"id" binding:"required,min=1"`
}

type listEntriesQuery struct {
	PageID   int32  `form:"page_id" binding:"required,min=1"`
	PageSize int32  `form:"page_size" binding:"required,min=5,max=50"`
	Min      string `form:"min_amount"`
	Max      string `form:"max_amount"`
	From     string `form:"from_date"`
	To       string `form:"to_date"`
	Sort     string `form:"sort"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var uri listEntriesUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var q listEntriesQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if _, ok := server.getOwnedAccount(ctx, uri.AccountID); !ok {
		return
	}

	minAmount, err := parseOptionalInt64(q.Min, minAmountDefault)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	maxAmount, err := parseOptionalInt64(q.Max, maxAmountDefault)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if minAmount > maxAmount {
		ctx.JSON(http.StatusBadRequest, errorResponse(errInvalidAmountRange))
		return
	}

	fromTime, err := parseOptionalTime(q.From, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	toTime, err := parseOptionalTime(q.To, time.Now().UTC())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if fromTime.After(toTime) {
		ctx.JSON(http.StatusBadRequest, errorResponse(errInvalidDateRange))
		return
	}

	sortOrder, err := parseSortOrder(q.Sort)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	limit := q.PageSize
	offset := (q.PageID - 1) * q.PageSize

	if sortOrder == "asc" {
		entries, err := server.store.ListEntriesFilteredAsc(ctx, db.ListEntriesFilteredAscParams{
			AccountID:   uri.AccountID,
			Amount:      minAmount,
			Amount_2:    maxAmount,
			CreatedAt:   fromTime,
			CreatedAt_2: toTime,
			Limit:       limit,
			Offset:      offset,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, entries)
		return
	}

	entries, err := server.store.ListEntriesFilteredDesc(ctx, db.ListEntriesFilteredDescParams{
		AccountID:   uri.AccountID,
		Amount:      minAmount,
		Amount_2:    maxAmount,
		CreatedAt:   fromTime,
		CreatedAt_2: toTime,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
