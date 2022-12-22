package pkgservice

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ErrorService struct {
	Code      Code
	Message   string
	Detail    string
	Raw       error
	Attribute map[string]string
}

func (e *ErrorService) GetCode() Code {
	return e.Code
}

func (e *ErrorService) String() string {
	return e.Message
}

func (e *ErrorService) Error() string {
	return e.Message
}

func (e *ErrorService) GetDetail() string {
	return e.Detail
}

func (e *ErrorService) GetRaw() error {
	return e.Raw
}

func (e *ErrorService) GetAttr() map[string]string {
	return e.Attribute
}

func NewErrorService(code Code) *ErrorService {
	return &ErrorService{
		Code:    code,
		Message: code.String(),
	}
}

func NewErrorServiceD(code Code, detail string) *ErrorService {
	return &ErrorService{
		Code:    code,
		Message: code.String(),
		Detail:  detail,
	}
}

func NewErrorServiceC(ctx context.Context, code Code, detail string) *ErrorService {
	errT := grpc.SetTrailer(ctx, getDetailErrCode(code))
	if errT != nil {
		log.Println(ctx, "[service_CashOutWithdrawal] Failed to set data trailing", errT)
	}
	return &ErrorService{
		Code:    code,
		Message: code.String(),
		Detail:  detail,
	}
}

func getDetailErrCode(code Code) metadata.MD {
	return metadata.New(map[string]string{
		"error_code": code.String(),
	})
}
func NewErrorServiceDA(code Code, detail string, attr map[string]string) *ErrorService {
	return &ErrorService{
		Code:      code,
		Message:   code.String(),
		Detail:    detail,
		Attribute: attr,
	}
}
