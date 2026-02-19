package gapi

import (
	"context"
	"fmt"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/Ian-Balijawa/simplebank/token"
)

func (server *Server) getOwnedAccount(ctx context.Context, accountID int64, payload *token.Payload) (db.Account, error) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		return account, err
	}

	if account.Owner != payload.Username {
		return account, fmt.Errorf("%w: account doesn't belong to the authenticated user", errPermissionDenied)
	}

	return account, nil
}
