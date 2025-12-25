package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Caesarsage/bankflow/account-service/internal/service"
	pb "github.com/Caesarsage/bankflow/account-service/proto/account"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountGRPCServer struct {
	pb.UnimplementedAccountServiceServer
	accountService *service.AccountService
}

func NewAccountGRPCServer(svc *service.AccountService) *AccountGRPCServer {
	return &AccountGRPCServer{
		accountService: svc,
	}
}

// GetAccount implements the GetAccount RPC
func (s *AccountGRPCServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &pb.GetAccountResponse{
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Invalid account ID format",
			},
		}, nil
	}

	account, err := s.accountService.GetAccountByID(ctx, accountID)
	if err != nil {
		return &pb.GetAccountResponse{
			Error: &pb.Error{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:               account.ID.String(),
			AccountNumber:    account.AccountNumber,
			CustomerId:       account.CustomerID.String(),
			AccountType:      string(account.AccountType),
			Currency:         account.Currency,
			Balance:          account.Balance.InexactFloat64(),
			AvailableBalance: account.AvailableBalance.InexactFloat64(),
			Status:           string(account.Status),
			InterestRate:     account.InterestRate.InexactFloat64(),
			OpenedAt:         account.OpenedAt.Unix(),
			CreatedAt:        account.CreatedAt.Unix(),
			UpdatedAt:        account.UpdatedAt.Unix(),
		},
	}, nil
}

// GetBalance implements the GetBalance RPC
func (s *AccountGRPCServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &pb.GetBalanceResponse{
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Invalid account ID format",
			},
		}, nil
	}

	balance, availableBalance, err := s.accountService.GetBalance(ctx, accountID)
	if err != nil {
		return &pb.GetBalanceResponse{
			Error: &pb.Error{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.GetBalanceResponse{
		Balance:          balance.InexactFloat64(),
		AvailableBalance: availableBalance.InexactFloat64(),
	}, nil
}

// UpdateBalance implements the UpdateBalance RPC
func (s *AccountGRPCServer) UpdateBalance(ctx context.Context, req *pb.UpdateBalanceRequest) (*pb.UpdateBalanceResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &pb.UpdateBalanceResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Invalid account ID format",
			},
		}, nil
	}

	amount := decimal.NewFromFloat(req.Amount)
	err = s.accountService.UpdateBalance(ctx, accountID, amount)
	if err != nil {
		return &pb.UpdateBalanceResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INTERNAL",
				Message: err.Error(),
			},
		}, nil
	}

	// Get updated balance
	balance, _, _ := s.accountService.GetBalance(ctx, accountID)

	return &pb.UpdateBalanceResponse{
		Success:    true,
		NewBalance: balance.InexactFloat64(),
	}, nil
}

// CreateHold implements the CreateHold RPC
func (s *AccountGRPCServer) CreateHold(ctx context.Context, req *pb.CreateHoldRequest) (*pb.CreateHoldResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &pb.CreateHoldResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Invalid account ID format",
			},
		}, nil
	}

	amount := decimal.NewFromFloat(req.Amount)
	var transactionRef *string
	if req.TransactionRef != "" {
		transactionRef = &req.TransactionRef
	}

	hold, err := s.accountService.CreateHold(ctx, accountID, amount, req.Reason, transactionRef)
	if err != nil {
		return &pb.CreateHoldResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INTERNAL",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.CreateHoldResponse{
		HoldId:  hold.ID.String(),
		Success: true,
	}, nil
}

// ReleaseHold implements the ReleaseHold RPC
func (s *AccountGRPCServer) ReleaseHold(ctx context.Context, req *pb.ReleaseHoldRequest) (*pb.ReleaseHoldResponse, error) {
	holdID, err := uuid.Parse(req.HoldId)
	if err != nil {
		return &pb.ReleaseHoldResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Invalid hold ID format",
			},
		}, nil
	}

	err = s.accountService.ReleaseHold(ctx, holdID)
	if err != nil {
		return &pb.ReleaseHoldResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INTERNAL",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.ReleaseHoldResponse{
		Success: true,
	}, nil
}

// StreamBalanceUpdates implements server-side streaming
func (s *AccountGRPCServer) StreamBalanceUpdates(req *pb.StreamBalanceRequest, stream pb.AccountService_StreamBalanceUpdatesServer) error {
	// This would be connected to a real-time update mechanism
	// For now, we'll show the pattern
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid account ID")
	}

	// Subscribe to balance updates (this would be a real subscription)
	// For demonstration, sending periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			balance, available, err := s.accountService.GetBalance(stream.Context(), accountID)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			update := &pb.BalanceUpdate{
				AccountId:        req.AccountId,
				Balance:          balance.InexactFloat64(),
				AvailableBalance: available.InexactFloat64(),
				Timestamp:        time.Now().Unix(),
			}

			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(accountService *service.AccountService, port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor),
		grpc.StreamInterceptor(streamLoggingInterceptor),
	)

	pb.RegisterAccountServiceServer(grpcServer, NewAccountGRPCServer(accountService))

	log.Printf("gRPC server listening on port %s", port)
	return grpcServer.Serve(lis)
}

// Interceptors for logging
func unaryLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("gRPC call: %s, duration: %v, error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func streamLoggingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	err := handler(srv, ss)
	log.Printf("gRPC stream: %s, duration: %v, error: %v", info.FullMethod, time.Since(start), err)
	return err
}
