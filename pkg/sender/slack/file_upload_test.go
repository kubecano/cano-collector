package slack

import (
	"errors"
	"strings"
	"testing"

	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/logger"
)

// MockSlackClient is a mock implementation of SlackClientInterface for testing
type MockSlackClient struct {
	mock.Mock
}

func (m *MockSlackClient) PostMessage(channelID string, options ...slackapi.MsgOption) (string, string, error) {
	args := m.Called(channelID, options)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockSlackClient) AuthTest() (*slackapi.AuthTestResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slackapi.AuthTestResponse), args.Error(1)
}

func (m *MockSlackClient) UploadFileV2(params slackapi.UploadFileV2Parameters) (*slackapi.FileSummary, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slackapi.FileSummary), args.Error(1)
}

func (m *MockSlackClient) GetFileInfo(fileID string, count, page int) (*slackapi.File, []slackapi.Comment, *slackapi.Paging, error) {
	args := m.Called(fileID, count, page)
	if args.Get(0) == nil {
		return nil, args.Get(1).([]slackapi.Comment), args.Get(2).(*slackapi.Paging), args.Error(3)
	}
	return args.Get(0).(*slackapi.File), args.Get(1).([]slackapi.Comment), args.Get(2).(*slackapi.Paging), args.Error(3)
}

func (m *MockSlackClient) GetConversations(params *slackapi.GetConversationsParameters) ([]slackapi.Channel, string, error) {
	args := m.Called(params)
	return args.Get(0).([]slackapi.Channel), args.String(1), args.Error(2)
}

func (m *MockSlackClient) GetConversationHistory(params *slackapi.GetConversationHistoryParameters) (*slackapi.GetConversationHistoryResponse, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slackapi.GetConversationHistoryResponse), args.Error(1)
}

func (m *MockSlackClient) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	args := m.Called(channelID, timestamp, options)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func TestUploadFileToSlack_EmptyFile(t *testing.T) {
	mockClient := new(MockSlackClient)
	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("empty.log", []byte{})

	require.Error(t, err)
	assert.Empty(t, fileID)
	assert.Empty(t, permalink)
	assert.Contains(t, err.Error(), "file is empty")
	mockClient.AssertNotCalled(t, "UploadFileV2")
}

func TestUploadFileToSlack_DirectSuccess(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("test log content")

	fileSummary := &slackapi.FileSummary{
		ID:    "F12345",
		Title: "test.log",
	}

	fileInfo := &slackapi.File{
		ID:        "F12345",
		Permalink: "https://slack.com/files/test/F12345",
	}

	mockClient.On("UploadFileV2", mock.MatchedBy(func(params slackapi.UploadFileV2Parameters) bool {
		return params.Filename == "test.log" &&
			params.FileSize == len(content) &&
			params.Channel == "" // Upload without channel, attachment via unfurl
	})).Return(fileSummary, nil)

	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("test.log", content)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestUploadFileToSlack_DirectFailTempFileSuccess(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("test log content")

	fileSummary := &slackapi.FileSummary{
		ID:    "F12345",
		Title: "test.log",
	}

	fileInfo := &slackapi.File{
		ID:        "F12345",
		Permalink: "https://slack.com/files/test/F12345",
	}

	// First call (direct) fails
	mockClient.On("UploadFileV2", mock.Anything).Return(nil, errors.New("direct upload failed")).Once()

	// Second call (temp file) succeeds
	mockClient.On("UploadFileV2", mock.Anything).Return(fileSummary, nil).Once()

	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("test.log", content)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestUploadFileToSlack_BothStrategiesFail(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("test log content")

	// Both calls fail
	mockClient.On("UploadFileV2", mock.Anything).Return(nil, errors.New("upload failed"))

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("test.log", content)

	require.Error(t, err)
	assert.Empty(t, fileID)
	assert.Empty(t, permalink)
	assert.Contains(t, err.Error(), "file upload failed")
	mockClient.AssertExpectations(t)
}

func TestUploadFileToSlack_GetFileInfoFails(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("test log content")

	fileSummary := &slackapi.FileSummary{
		ID:    "F12345",
		Title: "test.log",
	}

	mockClient.On("UploadFileV2", mock.Anything).Return(fileSummary, nil)
	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(nil, []slackapi.Comment{}, &slackapi.Paging{}, errors.New("get file info failed"))

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("test.log", content)

	require.Error(t, err)
	assert.Empty(t, fileID)
	assert.Empty(t, permalink)
	mockClient.AssertExpectations(t)
}

func TestUploadFileToSlack_ChannelParameterNotSet(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("test content")

	fileSummary := &slackapi.FileSummary{ID: "F12345", Title: "test.log"}
	fileInfo := &slackapi.File{ID: "F12345", Permalink: "https://slack.com/files/test/F12345"}

	// Verify that Channel parameter is NOT set
	// File will be attached via permalink unfurling in message text
	mockClient.On("UploadFileV2", mock.MatchedBy(func(params slackapi.UploadFileV2Parameters) bool {
		return params.Channel == ""
	})).Return(fileSummary, nil)

	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "my-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.uploadFileToSlack("test.log", content)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestTryUploadDirect(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("direct upload test")

	fileSummary := &slackapi.FileSummary{ID: "F12345", Title: "test.log"}
	fileInfo := &slackapi.File{ID: "F12345", Permalink: "https://slack.com/files/test/F12345"}

	mockClient.On("UploadFileV2", mock.Anything).Return(fileSummary, nil)
	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.tryUploadDirect("test.log", content)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestTryUploadViaTempFile(t *testing.T) {
	mockClient := new(MockSlackClient)
	content := []byte("temp file upload test")

	fileSummary := &slackapi.FileSummary{ID: "F12345", Title: "test.log"}
	fileInfo := &slackapi.File{ID: "F12345", Permalink: "https://slack.com/files/test/F12345"}

	mockClient.On("UploadFileV2", mock.Anything).Return(fileSummary, nil)
	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.tryUploadViaTempFile("test.log", content)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestExecuteUpload(t *testing.T) {
	mockClient := new(MockSlackClient)

	fileSummary := &slackapi.FileSummary{ID: "F12345", Title: "test.log"}
	fileInfo := &slackapi.File{ID: "F12345", Permalink: "https://slack.com/files/test/F12345"}

	params := slackapi.UploadFileV2Parameters{
		Filename: "test.log",
		FileSize: 100,
		Channel:  "test-channel",
	}

	mockClient.On("UploadFileV2", params).Return(fileSummary, nil)
	mockClient.On("GetFileInfo", "F12345", 0, 0).Return(fileInfo, []slackapi.Comment{}, &slackapi.Paging{}, nil)

	sender := &SenderSlack{
		channel:     "test-channel",
		slackClient: mockClient,
		logger:      logger.NewLogger("test", "debug"),
	}

	fileID, permalink, err := sender.executeUpload(params)

	require.NoError(t, err)
	assert.Equal(t, "F12345", fileID)
	assert.Equal(t, "https://slack.com/files/test/F12345", permalink)
	mockClient.AssertExpectations(t)
}

func TestCreateLogSnippet_EmptyContent(t *testing.T) {
	sender := &SenderSlack{
		logger: logger.NewLogger("test", "debug"),
	}

	snippet := sender.createLogSnippet([]byte{}, 50, 2500)
	assert.Empty(t, snippet)
}

func TestCreateLogSnippet_ShortContent(t *testing.T) {
	content := []byte("line1\nline2\nline3")
	sender := &SenderSlack{
		logger: logger.NewLogger("test", "debug"),
	}

	snippet := sender.createLogSnippet(content, 50, 2500)
	assert.Equal(t, "line1\nline2\nline3", snippet)
}

func TestCreateLogSnippet_TruncateLines(t *testing.T) {
	// Generate 100 lines
	lines := make([]string, 100)
	for i := 0; i < 100; i++ {
		lines[i] = "line " + string(rune('0'+i))
	}
	content := []byte(strings.Join(lines, "\n"))

	sender := &SenderSlack{
		logger: logger.NewLogger("test", "debug"),
	}

	// Request only last 10 lines
	snippet := sender.createLogSnippet(content, 10, 5000)

	// Should contain truncation indicator
	assert.Contains(t, snippet, "[... showing last 10 lines of 100 total ...]")

	// Should contain the last line
	assert.Contains(t, snippet, lines[99])

	// Should NOT contain the first line
	assert.NotContains(t, snippet, lines[0])
}

func TestCreateLogSnippet_TruncateChars(t *testing.T) {
	// Create content with multiple long lines
	longLine := strings.Repeat("x", 500)
	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = longLine
	}
	content := []byte(strings.Join(lines, "\n"))

	sender := &SenderSlack{
		logger: logger.NewLogger("test", "debug"),
	}

	// Request max 1000 chars - should fit only 1-2 lines
	snippet := sender.createLogSnippet(content, 50, 1000)

	// Snippet should be truncated to fit maxChars
	assert.LessOrEqual(t, len(snippet), 1000)

	// Should contain truncation indicator since we have 10 lines but showing only 1-2
	assert.Contains(t, snippet, "[... showing last")

	// Should contain at least one complete line (last line)
	assert.Contains(t, snippet, "xxx")
}

func TestCreateLogSnippet_LastLinesPreserved(t *testing.T) {
	content := []byte("line1\nline2\nline3\nline4\nline5")

	sender := &SenderSlack{
		logger: logger.NewLogger("test", "debug"),
	}

	// Request last 3 lines
	snippet := sender.createLogSnippet(content, 3, 5000)

	// Should show last 3 lines
	assert.Contains(t, snippet, "line3")
	assert.Contains(t, snippet, "line4")
	assert.Contains(t, snippet, "line5")

	// Should NOT show first 2 lines
	assert.NotContains(t, snippet, "line1")
	assert.NotContains(t, snippet, "line2")
}
