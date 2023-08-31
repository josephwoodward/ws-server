package ws_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ws "github.com/josephwoodward/go-websocket-server/websocket"
)

func TestMethodMustBeGet(t *testing.T) {
	t.Run("returns Pepper's score", func(t *testing.T) {
		// Arrange
		request, _ := http.NewRequest(http.MethodPost, "/ws", nil)
		resp := httptest.NewRecorder()

		// Act
		ws.Upgrade(resp, request)

		// Assert
		want := http.StatusMethodNotAllowed
		got := resp.Result().StatusCode
		if got != want {
			t.Errorf("wanted %d but got %d", got, want)
		}
	})
}

func TestNonGetMethodsShouldFail(t *testing.T) {
	tests := []struct {
		name   string
		action string
		want   int
	}{
		{name: "POST should result in status 405", action: http.MethodPost, want: http.StatusMethodNotAllowed},
		{name: "PATCH should result in status 405", action: http.MethodPatch, want: http.StatusMethodNotAllowed},
		{name: "PUT should result in status 405", action: http.MethodPut, want: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		// Arrange
		request, _ := http.NewRequest(tc.action, "/ws", nil)
		resp := httptest.NewRecorder()

		// Act
		ws.Upgrade(resp, request)

		// Assert
		got := resp.Result().StatusCode
		if tc.want != got {
			t.Errorf("wanted %d but got %d", tc.want, got)
		}
	}
}
