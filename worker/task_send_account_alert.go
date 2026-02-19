package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TaskSendAccountAlert = "task:send_account_alert"

	AlertDirectionLow  = "low"
	AlertDirectionHigh = "high"
)

type PayloadAccountAlert struct {
	Username  string `json:"username"`
	AccountID int64  `json:"account_id"`
	Balance   int64  `json:"balance"`
	Threshold int64  `json:"threshold"`
	Direction string `json:"direction"`
	Currency  string `json:"currency"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendAccountAlert(
	ctx context.Context,
	payload *PayloadAccountAlert,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendAccountAlert, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendAccountAlert(ctx context.Context, task *asynq.Task) error {
	var payload PayloadAccountAlert
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	subject := "Simple Bank balance alert"
	if payload.Direction == AlertDirectionHigh {
		subject = "Simple Bank high balance alert"
	}

	content := fmt.Sprintf(`Hello %s,<br/>
Your account %d balance is now %d %s.<br/>
Alert threshold: %d %s.<br/>
`, user.FullName, payload.AccountID, payload.Balance, payload.Currency, payload.Threshold, payload.Currency)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send account alert: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
