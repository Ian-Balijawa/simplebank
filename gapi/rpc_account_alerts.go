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

func (server *Server) GetAccountAlert(ctx context.Context, req *pb.GetAccountAlertRequest) (*pb.GetAccountAlertResponse, error) {
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

	alert, err := server.store.GetAccountAlert(ctx, req.GetAccountId())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account alert not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get account alert: %s", err)
	}

	return &pb.GetAccountAlertResponse{Alert: convertAccountAlert(alert)}, nil
}

func (server *Server) UpsertAccountAlert(ctx context.Context, req *pb.UpsertAccountAlertRequest) (*pb.UpsertAccountAlertResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpsertAccountAlertRequest(req)
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

	alert, err := server.store.UpsertAccountAlert(ctx, db.UpsertAccountAlertParams{
		AccountID:            req.GetAccountId(),
		LowBalanceThreshold:  req.GetLowBalanceThreshold(),
		HighBalanceThreshold: req.GetHighBalanceThreshold(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert account alert: %s", err)
	}

	return &pb.UpsertAccountAlertResponse{Alert: convertAccountAlert(alert)}, nil
}

func validateUpsertAccountAlertRequest(req *pb.UpsertAccountAlertRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	violations = validateAccountId(req.GetAccountId())
	if req.GetLowBalanceThreshold() < 0 {
		violations = append(violations, fieldViolation("low_balance_threshold", errors.New("must be greater than or equal to 0")))
	}
	if req.GetHighBalanceThreshold() < 0 {
		violations = append(violations, fieldViolation("high_balance_threshold", errors.New("must be greater than or equal to 0")))
	}
	if req.GetLowBalanceThreshold() > 0 && req.GetHighBalanceThreshold() > 0 &&
		req.GetLowBalanceThreshold() > req.GetHighBalanceThreshold() {
		violations = append(violations, fieldViolation("low_balance_threshold", errors.New("must be less than or equal to high_balance_threshold")))
	}
	return violations
}
