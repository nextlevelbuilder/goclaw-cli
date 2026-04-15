package output

import "testing"

func TestMapServerCode_AuthCodes(t *testing.T) {
	for _, code := range []string{"UNAUTHORIZED", "NOT_PAIRED", "TENANT_ACCESS_REVOKED"} {
		if got := MapServerCode(code); got != ExitAuth {
			t.Errorf("MapServerCode(%q) = %d, want %d", code, got, ExitAuth)
		}
	}
}

func TestMapServerCode_NotFoundCodes(t *testing.T) {
	for _, code := range []string{"NOT_FOUND", "NOT_LINKED"} {
		if got := MapServerCode(code); got != ExitNotFound {
			t.Errorf("MapServerCode(%q) = %d, want %d", code, got, ExitNotFound)
		}
	}
}

func TestMapServerCode_ValidationCodes(t *testing.T) {
	for _, code := range []string{"INVALID_REQUEST", "FAILED_PRECONDITION", "ALREADY_EXISTS"} {
		if got := MapServerCode(code); got != ExitValidation {
			t.Errorf("MapServerCode(%q) = %d, want %d", code, got, ExitValidation)
		}
	}
}

func TestMapServerCode_ServerCodes(t *testing.T) {
	for _, code := range []string{"INTERNAL", "UNAVAILABLE", "AGENT_TIMEOUT"} {
		if got := MapServerCode(code); got != ExitServer {
			t.Errorf("MapServerCode(%q) = %d, want %d", code, got, ExitServer)
		}
	}
}

func TestMapServerCode_ResourceCodes(t *testing.T) {
	if got := MapServerCode("RESOURCE_EXHAUSTED"); got != ExitResource {
		t.Errorf("MapServerCode(RESOURCE_EXHAUSTED) = %d, want %d", got, ExitResource)
	}
}

func TestMapServerCode_Unknown(t *testing.T) {
	if got := MapServerCode("SOME_UNKNOWN_CODE"); got != ExitGeneric {
		t.Errorf("MapServerCode(unknown) = %d, want %d", got, ExitGeneric)
	}
}

func TestMapHTTPStatus(t *testing.T) {
	cases := []struct {
		status int
		want   int
	}{
		{200, ExitGeneric},
		{400, ExitValidation},
		{401, ExitAuth},
		{403, ExitAuth},
		{404, ExitNotFound},
		{409, ExitValidation},
		{422, ExitValidation},
		{429, ExitResource},
		{500, ExitServer},
		{502, ExitServer},
		{503, ExitServer},
	}
	for _, tc := range cases {
		if got := MapHTTPStatus(tc.status); got != tc.want {
			t.Errorf("MapHTTPStatus(%d) = %d, want %d", tc.status, got, tc.want)
		}
	}
}
