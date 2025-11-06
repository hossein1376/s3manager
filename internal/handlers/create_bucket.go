package handlers

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
)

func (h *Handler) CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createBucketRequest
	err := grape.ReadJson(w, r, &req)
	if err != nil {
		resp := grape.Response{
			Message: fmt.Sprintf("decoding body JSON: %s", err),
		}
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	v := validator.New()
	v.Check(
		"name",
		validator.Case{
			Cond: validator.LengthMin(req.Name, 3),
			Msg:  "Bucket name cannot be shorter than 3 characters",
		},
		validator.Case{
			Cond: validator.Contains(req.Name, "/"),
			Msg:  "Bucket name cannot contain invalid characters",
		},
	)
	if ok := v.Validate(); !ok {
		resp := grape.Response{Message: "Bad input", Data: v.Errors}
		grape.WriteJson(
			ctx, w, grape.WithStatus(http.StatusBadRequest), grape.WithData(resp),
		)
		return
	}

	err = h.service.CreateBucket(ctx, req.Name)
	if err != nil {
		grape.RespondFromErr(ctx, w, err)
		return
	}

	grape.WriteJson(ctx, w, grape.WithStatus(http.StatusNoContent))
}

type createBucketRequest struct {
	Name string `json:"name"`
}
