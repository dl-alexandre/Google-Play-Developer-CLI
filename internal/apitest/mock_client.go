package apitest

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/games/v1"
	gamesmanagement "google.golang.org/api/gamesmanagement/v1management"
	"google.golang.org/api/playcustomapp/v1"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
	"google.golang.org/api/playintegrity/v1"
)

type MockClient struct {
	mu sync.RWMutex

	publisherSvc       *androidpublisher.Service
	reportingSvc       *playdeveloperreporting.Service
	gamesManagementSvc *gamesmanagement.Service
	gamesSvc           *games.Service
	playIntegritySvc   *playintegrity.Service
	customAppSvc       *playcustomapp.Service

	PublisherResponses *PublisherResponses
	ReportingResponses *ReportingResponses
	GamesResponses     *GamesResponses
	IntegrityResponses *IntegrityResponses
	CustomAppResponses *CustomAppResponses

	Calls []Call
}

type Call struct {
	Service   string
	Method    string
	Arguments map[string]interface{}
}

type PublisherResponses struct {
	Edits    *MockEditsService
	Tracks   map[string]*androidpublisher.Track
	Testers  map[string]*androidpublisher.Testers
	Bundles  map[int64]*androidpublisher.Bundle
	APKs     map[int64]*androidpublisher.Apk
	Listings map[string]*androidpublisher.Listing
	Images   map[string][]*androidpublisher.Image
}

type MockEditsService struct {
	InsertFunc   func(packageName string) (*androidpublisher.AppEdit, error)
	GetFunc      func(packageName, editId string) (*androidpublisher.AppEdit, error)
	CommitFunc   func(packageName, editId string) (*androidpublisher.AppEdit, error)
	ValidateFunc func(packageName, editId string) error
	DeleteFunc   func(packageName, editId string) error
	Bundles      *MockBundlesService
	Apks         *MockApksService
	Tracks       *MockTracksService
	Testers      *MockTestersService
	Listings     *MockListingsService
	Images       *MockImagesService
}

type MockBundlesService struct {
	UploadFunc func(packageName, editId string) (*androidpublisher.Bundle, error)
	ListFunc   func(packageName, editId string) (*androidpublisher.BundlesListResponse, error)
	GetFunc    func(packageName, editId string, versionCode int64) (*androidpublisher.Bundle, error)
}

type MockApksService struct {
	UploadFunc func(packageName, editId string) (*androidpublisher.Apk, error)
	ListFunc   func(packageName, editId string) (*androidpublisher.ApksListResponse, error)
	GetFunc    func(packageName, editId string, versionCode int64) (*androidpublisher.Apk, error)
}

type MockTracksService struct {
	GetFunc    func(packageName, editId, track string) (*androidpublisher.Track, error)
	ListFunc   func(packageName, editId string) (*androidpublisher.TracksListResponse, error)
	UpdateFunc func(packageName, editId, track string, trackConfig *androidpublisher.Track) (*androidpublisher.Track, error)
}

type MockTestersService struct {
	GetFunc    func(packageName, editId, track string) (*androidpublisher.Testers, error)
	UpdateFunc func(packageName, editId, track string, testers *androidpublisher.Testers) (*androidpublisher.Testers, error)
}

type MockListingsService struct {
	GetFunc    func(packageName, editId, language string) (*androidpublisher.Listing, error)
	UpdateFunc func(packageName, editId, language string, listing *androidpublisher.Listing) (*androidpublisher.Listing, error)
}

type MockImagesService struct {
	ListFunc      func(packageName, editId, language, imageType string) (*androidpublisher.ImagesListResponse, error)
	UploadFunc    func(packageName, editId, language, imageType string) (*androidpublisher.Image, error)
	DeleteFunc    func(packageName, editId, language, imageType, imageId string) error
	DeleteAllFunc func(packageName, editId, language, imageType string) error
}

type ReportingResponses struct {
	Vitals *MockVitalsService
}

type MockVitalsService struct {
	CrashRateQueryFunc    func(packageName string) (*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetResponse, error)
	ANRRateQueryFunc      func(packageName string) (*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetResponse, error)
	ErrorIssuesSearchFunc func(packageName string) (*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchErrorIssuesResponse, error)
}

type GamesResponses struct {
	Achievements *MockAchievementsService
	Scores       *MockScoresService
	Events       *MockEventsService
	Players      *MockPlayersService
}

type MockAchievementsService struct {
	ListFunc  func(playerId string) (*games.PlayerAchievementListResponse, error)
	ResetFunc func(achievementId string) error
}

type MockScoresService struct {
	ListFunc  func(playerId, leaderboardId string) (*games.PlayerScoreResponse, error)
	ResetFunc func(leaderboardId string) error
}

type MockEventsService struct {
	ListFunc  func(playerId string) (*games.PlayerEventListResponse, error)
	ResetFunc func(eventId string) error
}

type MockPlayersService struct {
	GetFunc    func(playerId string) (*games.Player, error)
	HideFunc   func(playerId string) error
	UnhideFunc func(playerId string) error
}

type IntegrityResponses struct {
	DecodeTokenFunc func(token string) (*playintegrity.DecodeIntegrityTokenResponse, error)
}

type CustomAppResponses struct {
	CreateFunc func(app *playcustomapp.CustomApp) (*playcustomapp.CustomApp, error)
}

func NewMockClient() *MockClient {
	return &MockClient{
		PublisherResponses: &PublisherResponses{
			Edits: &MockEditsService{
				Bundles:  &MockBundlesService{},
				Apks:     &MockApksService{},
				Tracks:   &MockTracksService{},
				Testers:  &MockTestersService{},
				Listings: &MockListingsService{},
				Images:   &MockImagesService{},
			},
			Tracks:   make(map[string]*androidpublisher.Track),
			Testers:  make(map[string]*androidpublisher.Testers),
			Bundles:  make(map[int64]*androidpublisher.Bundle),
			APKs:     make(map[int64]*androidpublisher.Apk),
			Listings: make(map[string]*androidpublisher.Listing),
			Images:   make(map[string][]*androidpublisher.Image),
		},
		ReportingResponses: &ReportingResponses{
			Vitals: &MockVitalsService{},
		},
		GamesResponses: &GamesResponses{
			Achievements: &MockAchievementsService{},
			Scores:       &MockScoresService{},
			Events:       &MockEventsService{},
			Players:      &MockPlayersService{},
		},
		IntegrityResponses: &IntegrityResponses{},
		CustomAppResponses: &CustomAppResponses{},
		Calls:              make([]Call, 0),
	}
}

func (m *MockClient) TrackCall(service, method string, args map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, Call{
		Service:   service,
		Method:    method,
		Arguments: args,
	})
}

func (m *MockClient) AndroidPublisher() (*androidpublisher.Service, error) {
	return m.publisherSvc, nil
}

func (m *MockClient) PlayReporting() (*playdeveloperreporting.Service, error) {
	return m.reportingSvc, nil
}

func (m *MockClient) GamesManagement() (*gamesmanagement.Service, error) {
	return m.gamesManagementSvc, nil
}

func (m *MockClient) Games() (*games.Service, error) {
	return m.gamesSvc, nil
}

func (m *MockClient) PlayIntegrity() (*playintegrity.Service, error) {
	return m.playIntegritySvc, nil
}

func (m *MockClient) PlayCustomApp() (*playcustomapp.Service, error) {
	return m.customAppSvc, nil
}

func (m *MockClient) Acquire(ctx context.Context) error {
	return nil
}

func (m *MockClient) Release() {}

func (m *MockClient) AcquireForUpload(ctx context.Context) error {
	return nil
}

func (m *MockClient) ReleaseForUpload() {}

func (m *MockClient) DoWithRetry(ctx context.Context, fn func() error) error {
	return fn()
}

func (m *MockClient) RetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
	}
}

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

type MockTokenSource struct {
	TokenFunc func() (*oauth2.Token, error)
}

func (m *MockTokenSource) Token() (*oauth2.Token, error) {
	if m.TokenFunc != nil {
		return m.TokenFunc()
	}
	return &oauth2.Token{
		AccessToken: "mock-token",
		TokenType:   "Bearer",
	}, nil
}

var ErrMockNotConfigured = errors.New("mock response not configured")

func (m *MockClient) SetPublisherResponse(editID string, track *androidpublisher.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.PublisherResponses.Tracks == nil {
		m.PublisherResponses.Tracks = make(map[string]*androidpublisher.Track)
	}
	m.PublisherResponses.Tracks[editID] = track
}

func (m *MockClient) GetCallCount(service, method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, call := range m.Calls {
		if call.Service == service && call.Method == method {
			count++
		}
	}
	return count
}

func (m *MockClient) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]Call, 0)
}
