package gapi

import (
	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/Ian-Balijawa/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}

func convertAccount(account db.Account) *pb.Account {
	return &pb.Account{
		Id:        account.ID,
		Owner:     account.Owner,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
}

func convertEntry(entry db.Entry) *pb.Entry {
	return &pb.Entry{
		Id:        entry.ID,
		AccountId: entry.AccountID,
		Amount:    entry.Amount,
		CreatedAt: timestamppb.New(entry.CreatedAt),
	}
}

func convertTransfer(transfer db.Transfer) *pb.Transfer {
	return &pb.Transfer{
		Id:            transfer.ID,
		FromAccountId: transfer.FromAccountID,
		ToAccountId:   transfer.ToAccountID,
		Amount:        transfer.Amount,
		CreatedAt:     timestamppb.New(transfer.CreatedAt),
	}
}

func convertAccountLimit(limit db.AccountLimit) *pb.AccountLimit {
	return &pb.AccountLimit{
		AccountId:          limit.AccountID,
		DailyTransferLimit: limit.DailyTransferLimit,
		CreatedAt:          timestamppb.New(limit.CreatedAt),
		UpdatedAt:          timestamppb.New(limit.UpdatedAt),
	}
}

func convertAccountAlert(alert db.AccountAlert) *pb.AccountAlert {
	return &pb.AccountAlert{
		AccountId:            alert.AccountID,
		LowBalanceThreshold:  alert.LowBalanceThreshold,
		HighBalanceThreshold: alert.HighBalanceThreshold,
		CreatedAt:            timestamppb.New(alert.CreatedAt),
		UpdatedAt:            timestamppb.New(alert.UpdatedAt),
	}
}
