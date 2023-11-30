package ws_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ws "github.com/josephwoodward/go-websocket-server/websocket"
)

func TestWriteCreatesCorrectFrames(t *testing.T) {
	r := &ws.WsUpgradeResult{}
	f := ws.Frame{
		IsFragment: false,
		Opcode:     0,
		Reserved:   0,
		IsMasked:   false,
		Length:     0,
		Payload:    []byte{},
		MaskingKey: []byte{},
	}
	r.Write(f)
}

func TestMethodMustBeGet(t *testing.T) {
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

func TestReturnsCorrectSecWebSocketKey(t *testing.T) {
	// Arrange
	request, _ := http.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Add("Upgrade", "websocket")
	request.Header.Add("Connection", "Upgrade")
	request.Header.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	resp := httptest.NewRecorder()

	// Act
	ws.Upgrade(resp, request)

	// Assert
	want := http.StatusSwitchingProtocols
	got := resp.Result().StatusCode
	if got != want {
		t.Errorf("wanted %d but got %d", want, got)
	}

	got2 := resp.Result().Header.Get("Sec-WebSocket-Accept")
	want2 := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got2 != want2 {
		t.Errorf("wanted %s but got %s", want2, got2)
	}
}

func assert(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("\nexpected: %v \ngot: %v", expected, actual)
	}
}

func TestGetAcceptHash(t *testing.T) {
	assert(t, ws.GenerateAcceptHash("dGhlIHNhbXBsZSBub25jZQ=="), "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=")
}
