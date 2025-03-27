package service

import (
	"context"
	"coupon-service/internal/application"
	"coupon-service/internal/domain"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sujin1135/coupon-service-interface/protobuf/entity"
	svcpb "github.com/Sujin1135/coupon-service-interface/protobuf/service"
	"github.com/Sujin1135/coupon-service-interface/protobuf/service/serviceconnect"
)

type GreetServiceHandler struct {
	serviceconnect.UnimplementedGreetServiceHandler
	couponService *application.CouponService
}

func NewGreetServiceHandler(couponService *application.CouponService) *GreetServiceHandler {
	return &GreetServiceHandler{
		couponService: couponService,
	}
}

func (s *GreetServiceHandler) CreateCampaign(
	ctx context.Context,
	req *connect.Request[svcpb.CreateCampaignRequest],
) (*connect.Response[svcpb.CreateCampaignResponse], error) {
	// 요청 데이터 추출
	name := req.Msg.Name
	amount := req.Msg.Amount
	issuedAt := req.Msg.IssuedAt.AsTime()
	expiresAt := req.Msg.ExpiresAt.AsTime()

	campaign, err := s.couponService.CreateCoupon(ctx, name, amount, issuedAt, expiresAt)
	if err != nil {
		switch err {
		case application.FailedSaveCouponError, application.CouponDataRecoveryError,
			application.CouponCacheDataRecoveryError, application.CouponCacheError:
			message := err.Error()
			return connect.NewResponse(&svcpb.CreateCampaignResponse{
				Value: &svcpb.CreateCampaignResponse_Error_{
					Error: &svcpb.CreateCampaignResponse_Error{
						Error: &svcpb.CreateCampaignResponse_Error_InternalProblem{
							InternalProblem: &entity.InternalError{
								Message: &message,
							},
						},
					},
				},
			}), nil
		default:
			message := "internal error occurred"
			return connect.NewResponse(&svcpb.CreateCampaignResponse{
				Value: &svcpb.CreateCampaignResponse_Error_{
					Error: &svcpb.CreateCampaignResponse_Error{
						Error: &svcpb.CreateCampaignResponse_Error_InternalProblem{
							InternalProblem: &entity.InternalError{
								Message: &message,
							},
						},
					},
				},
			}), nil
		}
	}

	protoCampaign := domainCampaignToProtoCampaign(campaign)

	resp := connect.NewResponse(&svcpb.CreateCampaignResponse{
		Value: &svcpb.CreateCampaignResponse_Data_{
			Data: &svcpb.CreateCampaignResponse_Data{
				Campaign: protoCampaign,
			},
		},
	})

	return resp, nil
}

func (s *GreetServiceHandler) IssueCoupon(
	ctx context.Context,
	req *connect.Request[svcpb.IssueCouponRequest],
) (*connect.Response[svcpb.IssueCouponResponse], error) {
	campaignID := req.Msg.CampaignId
	userID := req.Msg.UserId

	err := s.couponService.IssueCoupon(ctx, campaignID, userID)
	if err != nil {
		switch err {
		case application.DataKeyNotFoundError:
			message := err.Error()
			return connect.NewResponse(&svcpb.IssueCouponResponse{
				Value: &svcpb.IssueCouponResponse_Error_{
					Error: &svcpb.IssueCouponResponse_Error{
						Error: &svcpb.IssueCouponResponse_Error_NotFound{
							NotFound: &entity.NotFoundError{
								Message: &message,
							},
						},
					},
				},
			}), nil
		case application.CouponNotStartedError, application.CouponExpiredError,
			application.DuplicatedCouponUserError, application.AllCouponIssuedError:
			message := err.Error()
			return connect.NewResponse(&svcpb.IssueCouponResponse{
				Value: &svcpb.IssueCouponResponse_Error_{
					Error: &svcpb.IssueCouponResponse_Error{
						Error: &svcpb.IssueCouponResponse_Error_BadRequest{
							BadRequest: &entity.BadRequestError{
								Message: &message,
							},
						},
					},
				},
			}), nil
		default:
			message := err.Error()
			return connect.NewResponse(&svcpb.IssueCouponResponse{
				Value: &svcpb.IssueCouponResponse_Error_{
					Error: &svcpb.IssueCouponResponse_Error{
						Error: &svcpb.IssueCouponResponse_Error_InternalProblem{
							InternalProblem: &entity.InternalError{
								Message: &message,
							},
						},
					},
				},
			}), nil
		}
	}

	resp := connect.NewResponse(&svcpb.IssueCouponResponse{
		Value: &svcpb.IssueCouponResponse_Data_{
			Data: &svcpb.IssueCouponResponse_Data{
				Result: true,
			},
		},
	})

	return resp, nil
}

func (s *GreetServiceHandler) GetCampaign(
	_ context.Context,
	req *connect.Request[svcpb.GetCampaignRequest],
) (*connect.Response[svcpb.GetCampaignResponse], error) {
	name := req.Msg.Name

	campaign, err := s.couponService.GetCoupon(name)
	if err != nil {
		message := "campaign not found"
		return connect.NewResponse(&svcpb.GetCampaignResponse{
			Value: &svcpb.GetCampaignResponse_Error_{
				Error: &svcpb.GetCampaignResponse_Error{
					Error: &svcpb.GetCampaignResponse_Error_NotFound{
						NotFound: &entity.NotFoundError{
							Message: &message,
						},
					},
				},
			},
		}), nil
	}

	protoCampaign := domainCampaignToProtoCampaign(campaign)

	resp := connect.NewResponse(&svcpb.GetCampaignResponse{
		Value: &svcpb.GetCampaignResponse_Data_{
			Data: &svcpb.GetCampaignResponse_Data{
				Campaign: protoCampaign,
			},
		},
	})

	return resp, nil
}

func domainCampaignToProtoCampaign(campaign *domain.Coupon) *entity.Campaign {
	issuedCoupons := make([]*entity.IssuedCoupon, len(campaign.IssuedCoupons))
	for i, issuedCoupon := range campaign.IssuedCoupons {
		issuedCoupons[i] = &entity.IssuedCoupon{
			Id:         issuedCoupon.ID,
			CouponId:   issuedCoupon.CouponID,
			Code:       issuedCoupon.Code,
			CreatedAt:  timestamppb.New(issuedCoupon.CreatedAt),
			ModifiedAt: timestamppb.New(issuedCoupon.ModifiedAt),
		}
	}

	return &entity.Campaign{
		Id:            campaign.ID,
		Name:          campaign.Name,
		IssueAmount:   campaign.IssueAmount,
		IssuedAt:      timestamppb.New(campaign.IssuedAt),
		ExpiresAt:     timestamppb.New(campaign.ExpiresAt),
		IssuedCoupons: issuedCoupons,
		CreatedAt:     timestamppb.New(campaign.CreatedAt),
		ModifiedAt:    timestamppb.New(campaign.ModifiedAt),
	}
}
