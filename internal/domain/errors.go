package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("usuário não encontrado")
	ErrUserAlreadyExists     = errors.New("usuário já cadastrado com este e-mail")
	ErrInvalidCredentials    = errors.New("credenciais inválidas")
	ErrInvalidEmail          = errors.New("e-mail inválido")
	ErrPasswordTooShort      = errors.New("a senha deve conter pelo menos 6 caracteres")
	ErrUnauthorized          = errors.New("não autorizado")
	ErrVideoNotFound         = errors.New("vídeo não encontrado")
	ErrVideoForbidden        = errors.New("acesso negado ao vídeo")
	ErrUnsupportedVideoType  = errors.New("formato de vídeo não suportado")
	ErrVideoProcessingFailed = errors.New("falha no processamento do vídeo")
	ErrVideoNotReady         = errors.New("o processamento do vídeo ainda não foi concluído")
)
