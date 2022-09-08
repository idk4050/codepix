package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/rpc"
	proto "codepix/example-bank-api/proto/codepix/pixkey"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	PixKeyClient proto.ServiceClient
}

type RegisterParams struct {
	AccountID uuid.UUID `param:"account-id"`
}
type RegisterReq struct {
	Type proto.Type `json:"type"`
	Key  string     `json:"key"`
}
type Registered struct {
	ID uuid.UUID `json:"id"`
}

func (s Service) Register(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, RegisterParams{})
	body := httputils.Body(r, RegisterReq{})

	request := &proto.RegisterRequest{
		Type:      body.Type,
		Key:       body.Key,
		AccountId: params.AccountID[:],
	}
	pbReply, err := s.PixKeyClient.Register(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	ID, _ := uuid.FromBytes(pbReply.Id)
	reply := Registered{
		ID,
	}
	httputils.Json(w, reply, http.StatusCreated)
}

type Find struct {
	AccountID uuid.UUID `param:"account-id"`
	ID        uuid.UUID `param:"pix-key-id"`
}
type FindResult struct {
	ID   uuid.UUID  `json:"id"`
	Type proto.Type `json:"type"`
	Key  string     `json:"key"`
}

func (s Service) Find(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Find{})

	request := &proto.FindRequest{
		Id: params.ID[:],
	}
	pbReply, err := s.PixKeyClient.Find(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	accountID, _ := uuid.FromBytes(pbReply.AccountId)
	if accountID != params.AccountID {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	ID, _ := uuid.FromBytes(pbReply.Id)
	reply := FindResult{
		ID,
		pbReply.Type,
		pbReply.Key,
	}
	httputils.Json(w, reply, http.StatusOK)
}

type List struct {
	AccountID uuid.UUID `param:"account-id"`
}
type ListItem struct {
	ID   uuid.UUID  `json:"id"`
	Type proto.Type `json:"type"`
	Key  string     `json:"key"`
}
type ListResult struct {
	PixKeys []ListItem `json:"pix_keys"`
}

func (s Service) List(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, List{})

	request := &proto.ListRequest{
		AccountId: params.AccountID[:],
	}
	pbReply, err := s.PixKeyClient.List(r.Context(), request)
	if err != nil {
		rpc.ErrorToHTTP(w, r, err)
		return
	}
	pixKeys := []ListItem{}
	for _, pbPixKey := range pbReply.Items {
		ID, _ := uuid.FromBytes(pbPixKey.Id)
		pixKey := ListItem{
			ID,
			pbPixKey.Type,
			pbPixKey.Key,
		}
		pixKeys = append(pixKeys, pixKey)
	}
	reply := ListResult{
		PixKeys: pixKeys,
	}
	httputils.Json(w, reply, http.StatusOK)
}
