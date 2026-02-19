package gapi

import (
	"context"
	"errors"
	"time"

	db "github.com/Ian-Balijawa/simplebank/db/sqlc"
	"github.com/Ian-Balijawa/simplebank/pb"
	"github.com/Ian-Balijawa/simplebank/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListTransfers(ctx context.Context, req *pb.ListTransfersRequest) (*pb.ListTransfersResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateListTransfersRequest(req)
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

	minAmount := optionalInt64(req.MinAmount, 0)
	maxAmount := optionalInt64(req.MaxAmount, maxAmountDefault)

	fromTime, err := optionalTime(req.FromTime, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from_time: %s", err)
	}
	toTime, err := optionalTime(req.ToTime, time.Now().UTC())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to_time: %s", err)
	}
	if fromTime.After(toTime) {
		return nil, status.Errorf(codes.InvalidArgument, "from_time must be before or equal to to_time")
	}

	direction := "any"
	switch req.Direction {
	case pb.TransferDirection_TRANSFER_DIRECTION_IN:
		direction = "in"
	case pb.TransferDirection_TRANSFER_DIRECTION_OUT:
		direction = "out"
	}

	arg := db.ListTransfersFilteredDescParams{
		AccountID:   req.GetAccountId(),
		Direction:   direction,
		Amount:      minAmount,
		Amount_2:    maxAmount,
		CreatedAt:   fromTime,
		CreatedAt_2: toTime,
		Limit:       req.GetPageSize(),
		Offset:      (req.GetPageId() - 1) * req.GetPageSize(),
	}

	var transfers []db.Transfer
	if req.SortOrder == pb.SortOrder_SORT_ORDER_ASC {
		argAsc := db.ListTransfersFilteredAscParams(arg)
		transfers, err = server.store.ListTransfersFilteredAsc(ctx, argAsc)
	} else {
		transfers, err = server.store.ListTransfersFilteredDesc(ctx, arg)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transfers: %s", err)
	}

	rsp := &pb.ListTransfersResponse{
		Transfers: make([]*pb.Transfer, 0, len(transfers)),
	}
	for _, transfer := range transfers {
		rsp.Transfers = append(rsp.Transfers, convertTransfer(transfer))
	}

	return rsp, nil
}

func validateListTransfersRequest(req *pb.ListTransfersRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.GetAccountId() <= 0 {
		violations = append(violations, fieldViolation("account_id", errors.New("must be a positive integer")))
	}
	if req.GetPageId() <= 0 {
		violations = append(violations, fieldViolation("page_id", errors.New("must be a positive integer")))
	}
	if req.GetPageSize() < 5 || req.GetPageSize() > 50 {
		violations = append(violations, fieldViolation("page_size", errors.New("must be between 5 and 50")))
	}

	minAmount := optionalInt64(req.MinAmount, 0)
	maxAmount := optionalInt64(req.MaxAmount, maxAmountDefault)
	if minAmount > maxAmount {
		violations = append(violations, fieldViolation("min_amount", errors.New("must be less than or equal to max_amount")))
	}

	if req.FromTime != nil && req.FromTime.CheckValid() != nil {
		violations = append(violations, fieldViolation("from_time", errors.New("must be a valid timestamp")))
	}
	if req.ToTime != nil && req.ToTime.CheckValid() != nil {
		violations = append(violations, fieldViolation("to_time", errors.New("must be a valid timestamp")))
	}

	return violations
}
