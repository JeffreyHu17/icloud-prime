package createjob

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type fakeCreator struct {
	created int
	fail    error
	labels  []string
}

func (f *fakeCreator) CreateAlias(ctx context.Context, accountID, label string) (*CreateResult, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	f.created++
	f.labels = append(f.labels, label)
	return &CreateResult{
		Email:     "alias" + strconv.Itoa(f.created) + "@icloud.com",
		Label:     label,
		CreatedAt: time.Now().Format(time.RFC3339),
		AccountID: accountID,
	}, nil
}

func TestCreateOneUsesExactLabel(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	creator := &fakeCreator{}
	s := newTestScheduler(t, creator, now)

	result, err := s.CreateOne(context.Background(), "acc_1", "GitHub 注册")
	if err != nil {
		t.Fatalf("create one failed: %v", err)
	}
	if result.Label != "GitHub 注册" {
		t.Fatalf("expected exact result label, got %q", result.Label)
	}
	if len(creator.labels) != 1 || creator.labels[0] != "GitHub 注册" {
		t.Fatalf("expected exact creator label, got %+v", creator.labels)
	}
}

func newTestScheduler(t *testing.T, creator AliasCreator, now time.Time) *Scheduler {
	t.Helper()
	s, err := NewScheduler(Config{
		StorePath: filepath.Join(t.TempDir(), "create_jobs.json"),
		Creator:   creator,
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("new scheduler failed: %v", err)
	}
	return s
}

func TestBatchCreateRejectsCountAboveFive(t *testing.T) {
	s := newTestScheduler(t, &fakeCreator{}, time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local))

	_, err := s.BatchCreate(context.Background(), BatchRequest{AccountID: "acc_1", Count: 6})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBatchCreateUsesRemainingQuota(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	creator := &fakeCreator{}
	s := newTestScheduler(t, creator, now)

	first, err := s.BatchCreate(context.Background(), BatchRequest{AccountID: "acc_1", Count: 3, LabelPrefix: "手动"})
	if err != nil {
		t.Fatalf("first batch failed: %v", err)
	}
	second, err := s.BatchCreate(context.Background(), BatchRequest{AccountID: "acc_1", Count: 5, LabelPrefix: "手动"})
	if err != nil {
		t.Fatalf("second batch failed: %v", err)
	}

	if first.CreatedCount != 3 || second.CreatedCount != 2 {
		t.Fatalf("expected 3 then 2 created, got %d then %d", first.CreatedCount, second.CreatedCount)
	}
	if second.SkippedCount != 3 || second.RemainingThisHour != 0 {
		t.Fatalf("expected 3 skipped and 0 remaining, got %+v", second)
	}
	if creator.created != 5 {
		t.Fatalf("expected 5 real creates, got %d", creator.created)
	}
}

func TestHourlyQuotaPersistsAcrossSchedulerRestart(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 30, 0, 0, time.Local)
	storePath := filepath.Join(t.TempDir(), "create_jobs.json")
	first, err := NewScheduler(Config{
		StorePath: storePath,
		Creator:   &fakeCreator{},
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("new scheduler failed: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := first.CreateOne(context.Background(), "acc_1", "手动"); err != nil {
			t.Fatalf("create one %d failed: %v", i+1, err)
		}
	}

	restarted, err := NewScheduler(Config{
		StorePath: storePath,
		Creator:   &fakeCreator{},
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("restart scheduler failed: %v", err)
	}
	if got := restarted.RemainingThisHour("acc_1"); got != 2 {
		t.Fatalf("expected persisted quota to leave 2 remaining, got %d", got)
	}
	resp, err := restarted.BatchCreate(context.Background(), BatchRequest{AccountID: "acc_1", Count: 5, LabelPrefix: "重启后"})
	if err != nil {
		t.Fatalf("batch after restart failed: %v", err)
	}
	if resp.CreatedCount != 2 || resp.SkippedCount != 3 || resp.RemainingThisHour != 0 {
		t.Fatalf("expected persisted quota to allow only 2 more creates, got %+v", resp)
	}
}

func TestDurationJobCompletesAfterEndTime(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	creator := &fakeCreator{}
	s := newTestScheduler(t, creator, now)

	job, err := s.UpsertJob(JobRequest{AccountID: "acc_1", Mode: ModeDuration, DurationHours: 1})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	s.now = func() time.Time { return now.Add(2 * time.Hour) }

	if err := s.RunDue(context.Background()); err != nil {
		t.Fatalf("run due failed: %v", err)
	}
	updated, ok := s.GetJob(job.ID)
	if !ok {
		t.Fatal("job missing")
	}
	if updated.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", updated.Status)
	}
}

func TestDailyWindowAllowsCrossMidnight(t *testing.T) {
	now := time.Date(2026, 8, 9, 23, 15, 0, 0, time.Local)
	s := newTestScheduler(t, &fakeCreator{}, now)

	if !s.isInDailyWindow(now, "22:00", "02:00") {
		t.Fatal("expected 23:15 to be inside cross-midnight window")
	}
	if s.isInDailyWindow(time.Date(2026, 8, 9, 15, 0, 0, 0, time.Local), "22:00", "02:00") {
		t.Fatal("expected 15:00 outside cross-midnight window")
	}
}

func TestJobRecordsErrorWithoutConsumingQuota(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	s := newTestScheduler(t, &fakeCreator{fail: errors.New("HTTP 401")}, now)

	_, err := s.UpsertJob(JobRequest{AccountID: "acc_1", Mode: ModeDuration, DurationHours: 1})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if err := s.RunDue(context.Background()); err != nil {
		t.Fatalf("run due failed: %v", err)
	}
	jobs := s.ListJobs("")
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Status != StatusError || jobs[0].LastError == "" {
		t.Fatalf("expected error status with last error, got %+v", jobs[0])
	}
	if got := s.RemainingThisHour("acc_1"); got != 5 {
		t.Fatalf("failed create should not consume quota, got %d", got)
	}
}

func TestJobBacksOffAfterHTTP421WithoutStopping(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	s := newTestScheduler(t, &fakeCreator{fail: errors.New("HTTP 421: request throttled")}, now)

	job, err := s.UpsertJob(JobRequest{AccountID: "acc_1", Mode: ModeDuration, DurationHours: 1})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if err := s.RunDue(context.Background()); err != nil {
		t.Fatalf("run due failed: %v", err)
	}
	updated, ok := s.GetJob(job.ID)
	if !ok {
		t.Fatal("job missing")
	}
	if updated.Status != StatusRunning {
		t.Fatalf("expected transient 421 to keep job running, got %s", updated.Status)
	}
	if updated.LastError == "" {
		t.Fatal("expected last error to be recorded")
	}
	if updated.NextRunAt == nil || !updated.NextRunAt.After(now) {
		t.Fatalf("expected retry to be scheduled after now, got %v", updated.NextRunAt)
	}
	if got := s.RemainingThisHour("acc_1"); got != 5 {
		t.Fatalf("failed create should not consume quota, got %d", got)
	}
}

func TestAutomaticJobPacesNextRunAcrossRemainingHourlyQuota(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	s := newTestScheduler(t, &fakeCreator{}, now)

	job, err := s.UpsertJob(JobRequest{AccountID: "acc_1", Mode: ModeDuration, DurationHours: 1})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if err := s.RunDue(context.Background()); err != nil {
		t.Fatalf("run due failed: %v", err)
	}
	updated, ok := s.GetJob(job.ID)
	if !ok {
		t.Fatal("job missing")
	}
	expected := now.Add(12 * time.Minute)
	if updated.NextRunAt == nil || !updated.NextRunAt.Equal(expected) {
		t.Fatalf("expected next run at %v, got %v", expected, updated.NextRunAt)
	}
}

func TestAutomaticJobsUseSeparateHourlyQuotaPerAccount(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	s := newTestScheduler(t, &fakeCreator{}, now)

	if _, err := s.BatchCreate(context.Background(), BatchRequest{AccountID: "acc_1", Count: 5, LabelPrefix: "满额"}); err != nil {
		t.Fatalf("exhaust first account quota failed: %v", err)
	}
	job, err := s.UpsertJob(JobRequest{AccountID: "acc_2", Mode: ModeDuration, DurationHours: 1})
	if err != nil {
		t.Fatalf("upsert second account job failed: %v", err)
	}
	if err := s.RunDue(context.Background()); err != nil {
		t.Fatalf("run due failed: %v", err)
	}
	updated, ok := s.GetJob(job.ID)
	if !ok {
		t.Fatal("job missing")
	}
	if updated.CreatedCount != 1 || updated.Status != StatusRunning {
		t.Fatalf("expected second account job to keep running after creating one alias, got %+v", updated)
	}
	if got := s.RemainingThisHour("acc_1"); got != 0 {
		t.Fatalf("expected first account quota exhausted, got %d remaining", got)
	}
	if got := s.RemainingThisHour("acc_2"); got != 4 {
		t.Fatalf("expected second account to keep independent quota, got %d remaining", got)
	}
}

func TestSchedulerStartRunsDueJobs(t *testing.T) {
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.Local)
	creator := &fakeCreator{}
	s := newTestScheduler(t, creator, now)

	if _, err := s.UpsertJob(JobRequest{AccountID: "acc_1", Mode: ModeDuration, DurationHours: 1}); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx, 10*time.Millisecond)
	deadline := time.After(500 * time.Millisecond)
	for creator.created == 0 {
		select {
		case <-deadline:
			t.Fatal("expected background scheduler to create an alias")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
