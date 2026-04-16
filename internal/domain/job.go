package domain

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ChannelEmail = "email"
)

type Job struct {
	ID        string          `json:"id"`
	Channel   string          `json:"channel"`
	Payload   json.RawMessage `json:"payload"`
	Metadata  map[string]any  `json:"metadata"`
	Attempt   int             `json:"attempt"`
	CreatedAt time.Time       `json:"created_at"`
}

func DecodeJob(data []byte) (Job, error) {
	if len(data) == 0 {
		return Job{}, errors.New("data is empty")
	}

	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return Job{}, err
	}

	if job.Attempt < 0 {
		job.Attempt = 0
	}

	return job, nil
}

func EncodeJob(job Job) ([]byte, error) {
	return json.Marshal(job)
}
