package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/cpf"
	"github.com/Kc1t/postech-tc3-lambda-auth/internal/requester"
	"github.com/Kc1t/postech-tc3-lambda-auth/internal/token"
)

type request struct {
	CPF string `json:"cpf"`
}

type response struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   string `json:"expires_at"`
}

type handler struct {
	requesters *requester.Repository
	issuer     *token.Issuer
	logger     *slog.Logger
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("falha ao abrir conexao", "error", err)
		os.Exit(1)
	}

	ttl, err := time.ParseDuration(envOrDefault("JWT_TTL", "15m"))
	if err != nil {
		logger.Error("JWT_TTL invalido", "error", err)
		os.Exit(1)
	}

	h := &handler{
		requesters: requester.NewRepository(db),
		issuer:     token.NewIssuer(os.Getenv("JWT_SECRET"), ttl),
		logger:     logger,
	}

	lambda.Start(h.handle)
}

func (h *handler) handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	logger := h.logger.With("correlation_id", correlationID(event))

	var body request
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		logger.Warn("payload invalido")
		return reply(400, map[string]string{"error": "payload invalido"})
	}

	document, err := cpf.Validate(body.CPF)
	if err != nil {
		logger.Warn("cpf invalido")
		return reply(400, map[string]string{"error": err.Error()})
	}

	found, err := h.requesters.FindByDocument(ctx, document)
	switch {
	case errors.Is(err, requester.ErrNotFound):
		logger.Warn("cliente nao encontrado")
		return reply(404, map[string]string{"error": err.Error()})
	case errors.Is(err, requester.ErrInactive):
		logger.Warn("cliente inativo", "requester_id", found.ID, "status", found.Status)
		return reply(403, map[string]string{"error": err.Error()})
	case err != nil:
		logger.Error("falha ao consultar cliente", "error", err)
		return reply(500, map[string]string{"error": "erro interno"})
	}

	subject := token.Subject{
		ID:       found.ID,
		Name:     found.Name,
		Email:    found.Email,
		Document: document,
	}

	signed, expiresAt, err := h.issuer.Issue(subject, time.Now())
	if err != nil {
		logger.Error("falha ao assinar token", "error", err)
		return reply(500, map[string]string{"error": "erro interno"})
	}

	logger.Info("token emitido", "requester_id", found.ID)

	return reply(200, response{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	})
}

func correlationID(event events.APIGatewayV2HTTPRequest) string {
	if id := event.Headers["x-correlation-id"]; id != "" {
		return id
	}

	return event.RequestContext.RequestID
}

func reply(status int, payload any) (events.APIGatewayV2HTTPResponse, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500}, err
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(encoded),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
