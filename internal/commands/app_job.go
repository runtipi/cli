package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/utils"
)

type appJob struct {
	RequestID    string `json:"requestId"`
	Status       string `json:"status"`
	Phase        string `json:"phase"`
	FilesCopied  int64  `json:"filesCopied"`
	BytesCopied  int64  `json:"bytesCopied"`
	ArchiveBytes int64  `json:"archiveBytes"`
	Message      string `json:"message"`
}

func pollAppJob(ctx context.Context, interval time.Duration, fetch func() (*appJob, error), progress func(*appJob)) (*appJob, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		job, err := fetch()
		if err != nil {
			return nil, err
		}
		if job == nil {
			return nil, fmt.Errorf("job record is unavailable")
		}

		progress(job)
		switch job.Status {
		case "succeeded":
			return job, nil
		case "failed", "cancelled":
			return job, fmt.Errorf("%s", job.Message)
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func handleAppJobResponse(appID string, spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string) {
	defer spin.Finish()
	if err != nil {
		spin.Fail(errorMessage)
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		spin.Fail(errorMessage)
		fmt.Printf("Response: %s\n", body)
		return
	}

	var request AppResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&request); err != nil {
		spin.Fail("Failed to decode the job response")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	jobURL := fmt.Sprintf("%s/%s/job?requestId=%s", utils.GetAPIBaseURL("app-lifecycle"), url.PathEscape(appID), url.QueryEscape(request.RequestID))

	job, err := pollAppJob(ctx, time.Second, func() (*appJob, error) {
		response, err := utils.APIRequestWithContext(ctx, jobURL, "GET", "")
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to read job progress (HTTP %d)", response.StatusCode)
		}

		var body struct {
			Job *appJob `json:"job"`
		}
		err = json.NewDecoder(response.Body).Decode(&body)
		return body.Job, err
	}, func(job *appJob) {
		spin.SetMessage(fmt.Sprintf("%s: %s — %d files, %.1f MB copied, %.1f MB compressed", appID, job.Phase, job.FilesCopied, float64(job.BytesCopied)/1024/1024, float64(job.ArchiveBytes)/1024/1024))
	})

	if job == nil {
		spin.Warn("Stopped following progress. The job may still be running; check the app page to follow or cancel it.")
		fmt.Printf("Details: %v\n", err)
	} else if err != nil {
		spin.Fail(job.Message)
	} else {
		spin.Succeed(successMessage)
	}
}
