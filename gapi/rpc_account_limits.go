package gapi

import (
	"context"
	"errors"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/Ian-Balijawa/simplebank/pb"
	"github.com/Ian-Balijawa/simplebank/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetAccountLimit(ctx context.Context, req *pb.GetAccountLimitRequest) (*pb.GetAccountLimitResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateAccountId(req.GetAccountId()); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if _, err := server.getOwnedAccount(ctx, req.GetAccountId(), authPayload); err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		if errors.Is(err, errPermissionDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "cannot access this account")
		}
		return nil, status.Errorf(codes.Internal, "failed to get account: %s", err)
	}

	limit, err := server.store.GetAccountLimit(ctx, req.GetAccountId())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account limit not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get account limit: %s", err)
	}

	return &pb.GetAccountLimitResponse{Limit: convertAccountLimit(limit)}, nil
}

func (server *Server) UpsertAccountLimit(ctx context.Context, req *pb.UpsertAccountLimitRequest) (*pb.UpsertAccountLimitResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpsertAccountLimitRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if _, err := server.getOwnedAccount(ctx, req.GetAccountId(), authPayload); err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		if errors.Is(err, errPermissionDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "cannot access this account")
		}
		return nil, status.Errorf(codes.Internal, "failed to get account: %s", err)
	}

	limit, err := server.store.UpsertAccountLimit(ctx, db.UpsertAccountLimitParams{
		AccountID:          req.GetAccountId(),
		DailyTransferLimit: req.GetDailyTransferLimit(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert account limit: %s", err)
	}

	return &pb.UpsertAccountLimitResponse{Limit: convertAccountLimit(limit)}, nil
}

func validateAccountId(accountID int64) (violations []*errdetails.BadRequest_FieldViolation) {
	if accountID <= 0 {
		violations = append(violations, fieldViolation("account_id", errors.New("must be a positive integer")))
	}
	return violations
}

func validateUpsertAccountLimitRequest(req *pb.UpsertAccountLimitRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	violations = validateAccountId(req.GetAccountId())
	if req.GetDailyTransferLimit() < 0 {
		violations = append(violations, fieldViolation("daily_transfer_limit", errors.New("must be greater than or equal to 0")))
	}
	return violations
}
