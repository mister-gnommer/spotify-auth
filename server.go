// 🤖 AI-generated
package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
)

const (
	addr        = "127.0.0.1:8888"
	callbackURI = "http://127.0.0.1:8888/callback"
)

// startServer starts the local OAuth callback server and returns a channel that
// receives exactly one callbackResult when the /callback route is hit.
func startServer(expectedState string) (<-chan callbackResult, *http.Server) {
	resultCh := make(chan callbackResult, 1)
	var once sync.Once

	send := func(r callbackResult) {
		once.Do(func() { resultCh <- r })
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			writeHTML(w, http.StatusBadRequest, "Authorization denied",
				"You denied access. You can close this tab.")
			send(callbackResult{err: fmt.Errorf("authorization denied by user: %s", errParam)})
			return
		}

		if q.Get("state") != expectedState {
			writeHTML(w, http.StatusBadRequest, "State mismatch",
				"CSRF state mismatch. You can close this tab.")
			send(callbackResult{err: fmt.Errorf("CSRF state mismatch — possible redirect interception")})
			return
		}

		code := q.Get("code")
		if code == "" {
			writeHTML(w, http.StatusBadRequest, "Missing code",
				"No authorization code in callback. You can close this tab.")
			send(callbackResult{err: fmt.Errorf("no authorization code in callback")})
			return
		}

		writeHTML(w, http.StatusOK, "Authorization successful",
			"Authorization successful. You can close this tab.")
		send(callbackResult{code: code})
	})

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			send(callbackResult{err: fmt.Errorf("server error: %w", err)})
		}
	}()

	return resultCh, srv
}

func writeHTML(w http.ResponseWriter, status int, title, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>%s</title></head><body><h2>%s</h2></body></html>`,
		title, body)
}

// openBrowser opens url in the default system browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}
