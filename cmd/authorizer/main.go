package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/token"
)

type authorizer struct {
	issuer *token.Issuer
	logger *slog.Logger
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	a := &authorizer{
		issuer: token.NewIssuer(os.Getenv("JWT_SECRET"), time.Minute),
		logger: logger,
	}

	lambda.Start(a.handle)
}

func (a *authorizer) handle(_ context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	logger := a.logger.With("correlation_id", event.RequestContext.RequestID, "route", event.RouteKey)

	signed, ok := bearerToken(event.Headers)
	if !ok {
		logger.Warn("authorization header ausente ou malformado")
		return deny(), nil
	}

	claims, err := a.issuer.Verify(signed)
	if err != nil {
		logger.Warn("token rejeitado", "error", err)
		return deny(), nil
	}

	logger.Info("token aceito", "subject", claims.Subject, "role", claims.Role)

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context: map[string]interface{}{
			"subject":  claims.Subject,
			"role":     claims.Role,
			"document": claims.Document,
		},
	}, nil
}

func bearerToken(headers map[string]string) (string, bool) {
	for key, value := range headers {
		if !strings.EqualFold(key, "authorization") {
			continue
		}

		parts := strings.SplitN(value, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
			return "", false
		}

		return parts[1], true
	}

	return "", false
}

func deny() events.APIGatewayV2CustomAuthorizerSimpleResponse {
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}
}
