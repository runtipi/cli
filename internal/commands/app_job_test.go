package commands

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPollAppJobWaitsForTerminalResult(t *testing.T) {
	states := []string{"queued", "running", "cancelling", "cancelled"}
	calls := 0
	job, err := pollAppJob(context.Background(), time.Millisecond, func() (*appJob, error) {
		state := states[calls]
		calls++
		return &appJob{Status: state, Message: "Job cancelled"}, nil
	}, func(*appJob) {})

	if err == nil || job.Status != "cancelled" || calls != 4 {
		t.Fatalf("expected cancellation after four polls, got job=%+v err=%v calls=%d", job, err, calls)
	}
}

func TestPollAppJobReportsFailedBackup(t *testing.T) {
	job, err := pollAppJob(context.Background(), time.Millisecond, func() (*appJob, error) {
		return &appJob{Status: "failed", Message: "tar terminated by SIGTERM"}, nil
	}, func(*appJob) {})

	if job == nil || err == nil || err.Error() != "tar terminated by SIGTERM" {
		t.Fatalf("expected failed backup, got job=%+v err=%v", job, err)
	}
}

func TestPollAppJobDoesNotTreatLossOfProgressAsJobFailure(t *testing.T) {
	connectionError := errors.New("connection lost")
	job, err := pollAppJob(context.Background(), time.Millisecond, func() (*appJob, error) {
		return nil, connectionError
	}, func(*appJob) {})

	if job != nil || !errors.Is(err, connectionError) {
		t.Fatalf("expected unknown job outcome on connection loss, got job=%+v err=%v", job, err)
	}
}

func TestPollAppJobStopsFollowingOnInterrupt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	job, err := pollAppJob(ctx, time.Hour, func() (*appJob, error) {
		cancel()
		return &appJob{Status: "running"}, nil
	}, func(*appJob) {})

	if job != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected interrupted progress polling, got job=%+v err=%v", job, err)
	}
}
