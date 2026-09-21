package httpapi_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/httpapi"
	"video-processor/internal/httpapi/dto"
)

func setupTestApp() (*httpapi.ServerConfig, http.Handler) {
	userRepo := memory.NewMemoryUserRepository()
	videoRepo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	authUC := application.NewAuthUseCase(userRepo, "test-jwt-secret-key-12345", 2*time.Hour)
	videoUC := application.NewVideoUseCase(videoRepo, storage, queue)

	cfg := httpapi.ServerConfig{
		AuthUseCase:  authUC,
		VideoUseCase: videoUC,
	}

	router := httpapi.SetupRouter(cfg)
	return &cfg, router
}

func TestAPI_HealthEndpoints(t *testing.T) {
	_, router := setupTestApp()

	// Live
	wLive := httptest.NewRecorder()
	reqLive, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
	router.ServeHTTP(wLive, reqLive)
	assert.Equal(t, http.StatusOK, wLive.Code)
	assert.Contains(t, wLive.Body.String(), "UP")

	// Ready
	wReady := httptest.NewRecorder()
	reqReady, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	router.ServeHTTP(wReady, reqReady)
	assert.Equal(t, http.StatusOK, wReady.Code)
	assert.Contains(t, wReady.Body.String(), "READY")

	// Web UI
	wUI := httptest.NewRecorder()
	reqUI, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(wUI, reqUI)
	assert.Equal(t, http.StatusOK, wUI.Code)
	assert.Contains(t, wUI.Body.String(), "FIAP X")
}

func TestAPI_Auth_RegisterAndLogin(t *testing.T) {
	_, router := setupTestApp()

	// 1. Registro
	regPayload := `{"name":"Teste Investidor","email":"investidor@fiap.com","password":"senhaSegura123"}`
	wReg := httptest.NewRecorder()
	reqReg, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(regPayload))
	reqReg.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReg, reqReg)

	require.Equal(t, http.StatusCreated, wReg.Code)
	var authResp dto.AuthResponse
	err := json.Unmarshal(wReg.Body.Bytes(), &authResp)
	require.NoError(t, err)
	assert.NotEmpty(t, authResp.Token)
	assert.Equal(t, "investidor@fiap.com", authResp.User.Email)

	// 2. Login
	loginPayload := `{"email":"investidor@fiap.com","password":"senhaSegura123"}`
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(loginPayload))
	reqLogin.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLogin, reqLogin)

	require.Equal(t, http.StatusOK, wLogin.Code)
	var loginResp dto.AuthResponse
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token)

	// 3. Login com senha errada
	badLoginPayload := `{"email":"investidor@fiap.com","password":"senhaIncorreta"}`
	wBad := httptest.NewRecorder()
	reqBad, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(badLoginPayload))
	reqBad.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusUnauthorized, wBad.Code)
}

func TestAPI_Video_UnauthorizedAccess(t *testing.T) {
	_, router := setupTestApp()

	// Tentativa de acessar rota protegida sem token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPI_Video_FullWorkflow(t *testing.T) {
	cfg, router := setupTestApp()

	// 1. Cadastra usuário e obtém token
	out, err := cfg.AuthUseCase.Register(t.Context(), application.RegisterInput{
		Name:     "Investidor A",
		Email:    "a@fiap.com",
		Password: "senhaSegura123",
	})
	require.NoError(t, err)
	token := out.Token

	// 2. Upload de vídeo (multipart/form-data)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video", "apresentacao.mp4")
	require.NoError(t, err)
	_, _ = part.Write([]byte("fake video stream bytes"))
	_ = writer.Close()

	wUpload := httptest.NewRecorder()
	reqUpload, _ := http.NewRequest(http.MethodPost, "/api/v1/videos/upload", body)
	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	reqUpload.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wUpload, reqUpload)

	require.Equal(t, http.StatusAccepted, wUpload.Code)
	var uploadResp dto.UploadVideoResponse
	err = json.Unmarshal(wUpload.Body.Bytes(), &uploadResp)
	require.NoError(t, err)
	assert.NotEmpty(t, uploadResp.JobID)
	assert.Equal(t, "PENDING", uploadResp.Status)

	// 3. Listar vídeos do usuário
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	reqList.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wList, reqList)

	require.Equal(t, http.StatusOK, wList.Code)
	var listResp dto.VideoListResponse
	err = json.Unmarshal(wList.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Equal(t, 1, listResp.Total)
	assert.Equal(t, uploadResp.JobID, listResp.Videos[0].ID)
	assert.Equal(t, "apresentacao.mp4", listResp.Videos[0].OriginalName)

	// 4. Detalhes do vídeo
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/videos/"+uploadResp.JobID, nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wGet, reqGet)

	require.Equal(t, http.StatusOK, wGet.Code)
	var detailResp dto.VideoJobResponse
	err = json.Unmarshal(wGet.Body.Bytes(), &detailResp)
	require.NoError(t, err)
	assert.Equal(t, uploadResp.JobID, detailResp.ID)
}
