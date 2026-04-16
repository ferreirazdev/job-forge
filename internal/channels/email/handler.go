package email

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"job-forge/internal/domain"
)

type Handler struct{}

type EmailPayload struct {
	To         string `json:"to"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	ShouldFail bool   `json:"should_fail,omitempty"` // só pra testar o caminho de erro
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context, job domain.Job) error {
	var p EmailPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil {
		return err
	}

	if p.ShouldFail {
		return errors.New("simulated email provider failure")
	}

	log.Printf("[email] sending: to=%s subject=%q job_id=%s attempt=%d",
		p.To, p.Subject, job.ID, job.Attempt)

	// F1: simula trabalho; no futuro isso vira integração real.
	return nil
}
