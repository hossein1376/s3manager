package handlers

import (
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/validator"
)

func (h *Handler) CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := grape.ReadJSON[CreateBucketRequest](w, r)
	if err != nil {
		grape.ExtractFromErr(ctx, w, errs.BadRequest(errs.WithMsg(err.Error())))
		return
	}

	err = h.service.CreateBucket(ctx, req.Name)
	if err != nil {
		grape.ExtractFromErr(ctx, w, err)
		return
	}

	grape.WriteJSON(ctx, w, grape.WithStatus(http.StatusNoContent))
}

type CreateBucketRequest struct {
	Name string `json:"name"`
}

func (c CreateBucketRequest) Validate() error {
	v := validator.New()
	v.Check(
		"name",
		validator.Case{
			Cond: validator.LengthMin(c.Name, 3),
			Msg:  "Bucket name cannot be shorter than 3 characters",
		},
		validator.Case{
			Cond: validator.Not(validator.Contains(c.Name, "/")),
			Msg:  "Bucket name cannot contain invalid characters",
		},
	)
	if ok := v.Validate(); !ok {
		return errs.BadRequest(errs.WithMsg(v.Errors.Error()))
	}
	return nil
}
