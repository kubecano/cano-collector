package destslack

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDestinationSlack_Send_DelegatesToSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSenderSlack(ctrl)
	mockSender.EXPECT().Send(gomock.Any(), "test message").Return(nil).Times(1)

	d := NewDestinationSlack(mockSender)
	err := d.Send(context.Background(), "test message")
	assert.NoError(t, err)
}
