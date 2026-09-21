package handler_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/httpapi/handler"
	"video-processor/internal/httpapi/middleware"
)

func TestAuthHandler_Direct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := memory.NewMemoryUserRepository()
	authUC := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)
	h := handler.NewAuthHandler(authUC)

	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Register - dados inválidos (json quebrado)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader("{invalido}"))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Register - sucesso
	wSuccess := httptest.NewRecorder()
	reqSuccess, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"name":"Usuario Teste","email":"u@fiap.com","password":"senhaSegura123"}`))
	reqSuccess.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wSuccess, reqSuccess)
	assert.Equal(t, http.StatusCreated, wSuccess.Code)

	// Register - duplicado
	wDup := httptest.NewRecorder()
	reqDup, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"name":"Usuario Teste","email":"u@fiap.com","password":"senhaSegura123"}`))
	reqDup.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wDup, reqDup)
	assert.Equal(t, http.StatusConflict, wDup.Code)

	// Login - dados inválidos
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader("{invalido}"))
	r.ServeHTTP(wLogin, reqLogin)
	assert.Equal(t, http.StatusBadRequest, wLogin.Code)

	// Login - credenciais inválidas
	wBadCreds := httptest.NewRecorder()
	reqBadCreds, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"u@fiap.com","password":"senhaErrada"}`))
	reqBadCreds.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBadCreds, reqBadCreds)
	assert.Equal(t, http.StatusUnauthorized, wBadCreds.Code)

	// Login - sucesso
	wLoginOk := httptest.NewRecorder()
	reqLoginOk, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"u@fiap.com","password":"senhaSegura123"}`))
	reqLoginOk.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wLoginOk, reqLoginOk)
	assert.Equal(t, http.StatusOK, wLoginOk.Code)
}

func TestVideoHandler_DirectValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	videoUC := application.NewVideoUseCase(repo, storage, queue)
	h := handler.NewVideoHandler(videoUC)

	r := gin.New()
	// Mock middleware de auth
	r.Use(func(c *gin.Context) {
		c.Set(middleware.CtxUserIDKey, "u1")
		c.Set(middleware.CtxUserEmailKey, "u1@fiap.com")
		c.Next()
	})
	r.POST("/upload", h.Upload)
	r.GET("/videos", h.List)
	r.GET("/videos/:id", h.GetByID)
	r.GET("/videos/:id/download", h.Download)

	// 1. Upload sem arquivo no form-data
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/upload", strings.NewReader(""))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 2. Upload com formato não suportado (ex: .exe)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "malware.exe")
	_, _ = part.Write([]byte("fake"))
	_ = writer.Close()

	wInvalid := httptest.NewRecorder()
	reqInvalid, _ := http.NewRequest(http.MethodPost, "/upload", body)
	reqInvalid.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(wInvalid, reqInvalid)
	assert.Equal(t, http.StatusBadRequest, wInvalid.Code)

	// 3. Upload com formato válido (sucesso)
	bodyValid := &bytes.Buffer{}
	writerValid := multipart.NewWriter(bodyValid)
	partValid, _ := writerValid.CreateFormFile("video", "apresentacao.mp4")
	_, _ = partValid.Write([]byte("video-bytes-conteudo"))
	_ = writerValid.Close()

	wOk := httptest.NewRecorder()
	reqOk, _ := http.NewRequest(http.MethodPost, "/upload", bodyValid)
	reqOk.Header.Set("Content-Type", writerValid.FormDataContentType())
	r.ServeHTTP(wOk, reqOk)
	assert.Equal(t, http.StatusAccepted, wOk.Code)

	// 4. Listar vídeos
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest(http.MethodGet, "/videos", nil)
	r.ServeHTTP(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)

	// 5. GetByID - inexistente
	wNotFound := httptest.NewRecorder()
	reqNotFound, _ := http.NewRequest(http.MethodGet, "/videos/nao-existe", nil)
	r.ServeHTTP(wNotFound, reqNotFound)
	assert.Equal(t, http.StatusNotFound, wNotFound.Code)

	// 6. Download - inexistente
	wDl := httptest.NewRecorder()
	reqDl, _ := http.NewRequest(http.MethodGet, "/videos/nao-existe/download", nil)
	r.ServeHTTP(wDl, reqDl)
	assert.Equal(t, http.StatusNotFound, wDl.Code)

	// 7. GetByID - sucesso
	job, err := videoUC.UploadVideo(t.Context(), application.UploadVideoInput{
		UserID: "u1", UserEmail: "u1@fiap.com", Filename: "vid.mp4", Content: strings.NewReader("data"),
	})
	require.NoError(t, err)

	wGetOk := httptest.NewRecorder()
	reqGetOk, _ := http.NewRequest(http.MethodGet, "/videos/"+job.ID, nil)
	r.ServeHTTP(wGetOk, reqGetOk)
	assert.Equal(t, http.StatusOK, wGetOk.Code)

	// 8. Download - sucesso com job completado
	zipKey := "outputs/test.zip"
	_ = storage.Upload(t.Context(), zipKey, strings.NewReader("zip-content"), "application/zip")
	job.MarkCompleted(zipKey, 10)
	_ = repo.Update(t.Context(), job)

	wDlOk := httptest.NewRecorder()
	reqDlOk, _ := http.NewRequest(http.MethodGet, "/videos/"+job.ID+"/download", nil)
	r.ServeHTTP(wDlOk, reqDlOk)
	assert.Equal(t, http.StatusOK, wDlOk.Code)
	assert.Equal(t, "application/zip", wDlOk.Header().Get("Content-Type"))
}

func TestHealthHandler_Direct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHealthHandler()
	r := gin.New()
	r.GET("/live", h.Live)
	r.GET("/ready", h.Ready)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/live", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	wReady := httptest.NewRecorder()
	reqReady, _ := http.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(wReady, reqReady)
	require.Equal(t, http.StatusOK, wReady.Code)
}
