package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/rpc"
	readproto "codepix/example-bank-api/proto/codepix/transaction/read"
	writeproto "codepix/example-bank-api/proto/codepix/transaction/write"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	WriteServiceClient writeproto.ServiceClient
	ReadServiceClient  readproto.ServiceClient
}

type SendParams struct {
	AccountID uuid.UUID `param:"account-id"`
}
type Send struct {
	ReceiverKey string `json:"receiver_key"`
	Amount      uint64 `json:"amount"`
	Description string `json:"description"`
}
type Sent struct {
	ID uuid.UUID `json:"id"`
}

func (s *Service) Send(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, SendParams{})
	body := httputils.Body(r, Send{})

	request := &writeproto.StartRequest{
		SenderId:    params.AccountID[:],
		ReceiverKey: body.ReceiverKey,
		Amount:      body.Amount,
		Description: body.Description,
	}
	pbReply, err := s.WriteServiceClient.Start(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	ID, _ := uuid.FromBytes(pbReply.Id)
	reply := Sent{
		ID: ID,
	}
	httputils.Json(w, reply, http.StatusOK)
}

type Find struct {
	AccountID uuid.UUID `param:"account-id"`
	ID        uuid.UUID `param:"pix-transaction-id"`
}
type FindResult struct {
	ID               uuid.UUID        `json:"id"`
	Sender           uuid.UUID        `json:"sender"`
	SenderBank       uuid.UUID        `json:"sender_bank"`
	Receiver         uuid.UUID        `json:"receiver"`
	ReceiverBank     uuid.UUID        `json:"receiver_bank"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	Amount           uint64           `json:"amount"`
	Description      string           `json:"description"`
	Status           readproto.Status `json:"status"`
	ReasonForFailing string           `json:"reason_for_failing"`
}

func (s Service) Find(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Find{})

	request := &readproto.FindRequest{
		Id: params.ID[:],
	}
	pb, err := s.ReadServiceClient.Find(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	reply := FindReply(pb)
	if !(params.AccountID == reply.Sender || params.AccountID == reply.Receiver) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	httputils.Json(w, reply, http.StatusOK)
}

func FindReply(pb *readproto.FindReply) FindResult {
	ID, _ := uuid.FromBytes(pb.Id)
	Sender, _ := uuid.FromBytes(pb.Sender)
	SenderBank, _ := uuid.FromBytes(pb.SenderBank)
	Receiver, _ := uuid.FromBytes(pb.Receiver)
	ReceiverBank, _ := uuid.FromBytes(pb.ReceiverBank)
	return FindResult{
		ID:               ID,
		Sender:           Sender,
		SenderBank:       SenderBank,
		Receiver:         Receiver,
		ReceiverBank:     ReceiverBank,
		CreatedAt:        pb.CreatedAt.AsTime(),
		UpdatedAt:        pb.UpdatedAt.AsTime(),
		Amount:           pb.Amount,
		Description:      pb.Description,
		Status:           pb.Status,
		ReasonForFailing: pb.ReasonForFailing,
	}
}

type List struct {
	AccountID    uuid.UUID `param:"account-id"`
	CreatedAfter time.Time `query:"after"`
	SenderID     uuid.UUID `query:"sender"`
	ReceiverID   uuid.UUID `query:"receiver"`
	Limit        uint64    `query:"limit"`
	Skip         uint64    `query:"skip"`
}
type ListItem = FindResult
type ListResult struct {
	Transactions []ListItem `json:"transactions"`
}

func (s Service) List(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, List{})

	if !(params.AccountID == params.SenderID || params.AccountID == params.ReceiverID) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	request := &readproto.ListRequest{
		CreatedAfter: timestamppb.New(params.CreatedAfter),
		SenderId:     params.SenderID[:],
		ReceiverId:   params.ReceiverID[:],
		Limit:        params.Limit,
		Skip:         params.Skip,
	}
	pbReply, err := s.ReadServiceClient.List(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	transactions := []ListItem{}
	for _, pb := range pbReply.Items {
		transactions = append(transactions, ListItemReply(pb))
	}
	reply := ListResult{
		Transactions: transactions,
	}
	httputils.Json(w, reply, http.StatusOK)
}

func ListItemReply(pb *readproto.ListItem) ListItem {
	ID, _ := uuid.FromBytes(pb.Id)
	Sender, _ := uuid.FromBytes(pb.Sender)
	SenderBank, _ := uuid.FromBytes(pb.SenderBank)
	Receiver, _ := uuid.FromBytes(pb.Receiver)
	ReceiverBank, _ := uuid.FromBytes(pb.ReceiverBank)
	return ListItem{
		ID:               ID,
		Sender:           Sender,
		SenderBank:       SenderBank,
		Receiver:         Receiver,
		ReceiverBank:     ReceiverBank,
		CreatedAt:        pb.CreatedAt.AsTime(),
		UpdatedAt:        pb.UpdatedAt.AsTime(),
		Amount:           pb.Amount,
		Description:      pb.Description,
		Status:           pb.Status,
		ReasonForFailing: pb.ReasonForFailing,
	}
}
