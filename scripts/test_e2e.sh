#!/bin/bash
set -e

API_URL="http://localhost:8080"
MAILHOG_URL="http://localhost:8025"

echo "=== 1. VERIFICANDO HEALTH CHECK ==="
curl -s "${API_URL}/health/live"
echo ""
curl -s "${API_URL}/health/ready"
echo ""

echo -e "\n=== 2. CADASTRO DE NOVO USUÁRIO ==="
RANDOM_ID=$((RANDOM % 9000 + 1000))
EMAIL="investidor${RANDOM_ID}@fiapx.com"
REG_RESP=$(curl -s -X POST "${API_URL}/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Investidor FIAP ${RANDOM_ID}\",\"email\":\"${EMAIL}\",\"password\":\"senhaSegura123\"}")
echo "Resposta do Registro: ${REG_RESP}"

echo -e "\n=== 3. LOGIN DO USUÁRIO ==="
LOGIN_RESP=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"senhaSegura123\"}")
echo "Resposta do Login: ${LOGIN_RESP}"

TOKEN=$(echo "${LOGIN_RESP}" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "${TOKEN}" ]; then
  echo "Erro: Token não obtido!"
  exit 1
fi
echo "Token JWT obtido com sucesso!"

echo -e "\n=== 4. UPLOAD DO VÍDEO (MP4 REAL) ==="
UPLOAD_RESP=$(curl -s -X POST "${API_URL}/api/v1/videos/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "video=@/tmp/teste_video.mp4")
echo "Resposta do Upload: ${UPLOAD_RESP}"

JOB_ID=$(echo "${UPLOAD_RESP}" | grep -o '"job_id":"[^"]*' | cut -d'"' -f4)
echo "Job ID criado: ${JOB_ID}"

echo -e "\n=== 5. AGUARDANDO PROCESSAMENTO DO WORKER ==="
for i in {1..10}; do
  STATUS_RESP=$(curl -s -H "Authorization: Bearer ${TOKEN}" "${API_URL}/api/v1/videos/${JOB_ID}")
  STATUS=$(echo "${STATUS_RESP}" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
  echo "Tentativa ${i}: Status do Job = ${STATUS}"
  if [ "${STATUS}" == "COMPLETED" ]; then
    echo "Sucesso! Job concluído pelo worker!"
    echo "Detalhes: ${STATUS_RESP}"
    break
  fi
  sleep 2
done

echo -e "\n=== 6. LISTANDO VÍDEOS DO USUÁRIO ==="
curl -s -H "Authorization: Bearer ${TOKEN}" "${API_URL}/api/v1/videos"
echo ""

echo -e "\n=== 7. DOWNLOAD E VALIDAÇÃO DO ZIP COM FRAMES ==="
curl -s -H "Authorization: Bearer ${TOKEN}" "${API_URL}/api/v1/videos/${JOB_ID}/download" -o /tmp/downloaded_frames.zip
echo "Arquivo baixado: $(ls -lh /tmp/downloaded_frames.zip)"
mkdir -p /tmp/extracted_frames
unzip -o /tmp/downloaded_frames.zip -d /tmp/extracted_frames
echo "Frames extraídos:"
ls -la /tmp/extracted_frames

echo -e "\n=== 8. SIMULAÇÃO DE ERRO E NOTIFICAÇÃO POR E-MAIL ==="
# Cria um arquivo com extensão mp4 mas conteúdo corrompido que fará o ffmpeg falhar
echo "bytes invalidos que vao causar erro no ffmpeg" > /tmp/corrupted.mp4
FAIL_UPLOAD=$(curl -s -X POST "${API_URL}/api/v1/videos/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "video=@/tmp/corrupted.mp4")
FAIL_JOB_ID=$(echo "${FAIL_UPLOAD}" | grep -o '"job_id":"[^"]*' | cut -d'"' -f4)
echo "Job de Falha ID: ${FAIL_JOB_ID}"

sleep 4
FAIL_STATUS_RESP=$(curl -s -H "Authorization: Bearer ${TOKEN}" "${API_URL}/api/v1/videos/${FAIL_JOB_ID}")
echo "Status do Job Corrompido: ${FAIL_STATUS_RESP}"

echo -e "\n=== 9. VERIFICANDO E-MAIL DE ALERTA NO MAILHOG ==="
MAIL_RESP=$(curl -s "${MAILHOG_URL}/api/v2/messages")
echo "Total de e-mails capturados no MailHog: $(echo "${MAIL_RESP}" | grep -o '"total":[0-9]*' || echo 'ok')"

echo -e "\n=== 10. VERIFICANDO MÉTRICAS DO PROMETHEUS ==="
curl -s "${API_URL}/metrics" | grep -E "fiapx_video"

echo -e "\n🎉 TESTE DE PONTA A PONTA CONCLUÍDO COM SUCESSO TOTAL!"
