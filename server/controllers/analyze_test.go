package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
)

func analyzeServer(t *testing.T, endpoint string) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	config.AppConfig = &config.App{Site: config.Site{AnalyzeEndpoint: endpoint}}
	r := gin.New()
	r.POST("/v1/analyze", Analyze)
	// A real server: the reverse proxy needs CloseNotify, which httptest.ResponseRecorder lacks.
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func postRecording(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("audio", "recording.wav")
	fw.Write([]byte{1, 2, 3})
	mw.WriteField("meta", `{"save":false}`)
	mw.Close()
	req, _ := http.NewRequest(http.MethodPost, url, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Cookie", "sf_access=secret")
	req.Header.Set("Authorization", "Bearer secret")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func TestAnalyzeDisabled(t *testing.T) {
	srv := analyzeServer(t, "")
	res, body := postRecording(t, srv.URL+"/v1/analyze")
	if res.StatusCode != http.StatusServiceUnavailable || !strings.Contains(body, "8001") {
		t.Fatalf("disabled analyze = %d %s", res.StatusCode, body)
	}
}

func TestAnalyzeForwardsRecordingWithoutSession(t *testing.T) {
	var gotType, gotBody, gotCookie, gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotType = r.Header.Get("Content-Type")
		gotCookie = r.Header.Get("Cookie")
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"msg":"success","results":[["Parus major_Great Tit",0.9]]}`))
	}))
	defer upstream.Close()

	srv := analyzeServer(t, upstream.URL+"/analyze")
	res, body := postRecording(t, srv.URL+"/v1/analyze")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("analyze = %d %s", res.StatusCode, body)
	}
	var out struct {
		Results [][]interface{} `json:"results"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil || len(out.Results) != 1 {
		t.Fatalf("response not passed through: %s", body)
	}
	if !strings.HasPrefix(gotType, "multipart/form-data; boundary=") || !strings.Contains(gotBody, `filename="recording.wav"`) {
		t.Errorf("upload not forwarded unchanged: %s", gotType)
	}
	if gotCookie != "" || gotAuth != "" {
		t.Errorf("session sent to the service: cookie %q, authorization %q", gotCookie, gotAuth)
	}
}

func TestAnalyzeServiceDown(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	endpoint := upstream.URL + "/analyze"
	upstream.Close()

	srv := analyzeServer(t, endpoint)
	res, body := postRecording(t, srv.URL+"/v1/analyze")
	if res.StatusCode != http.StatusBadGateway || !strings.Contains(body, "8001") {
		t.Fatalf("service down = %d %s", res.StatusCode, body)
	}
}
