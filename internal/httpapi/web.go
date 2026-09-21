package httpapi

func GetAppHTML() string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FIAP X — Sistema de Processamento de Vídeos</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #6366f1;
            --primary-hover: #4f46e5;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --bg: #0f172a;
            --card-bg: #1e293b;
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --border: #334155;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: 'Inter', sans-serif;
        }

        body {
            background-color: var(--bg);
            color: var(--text-main);
            min-height: 100vh;
            padding: 40px 20px;
            display: flex;
            justify-content: center;
        }

        .container {
            width: 100%;
            max-width: 900px;
        }

        header {
            text-align: center;
            margin-bottom: 30px;
        }

        .badge-brand {
            display: inline-block;
            background: rgba(99, 102, 241, 0.15);
            color: #a5b4fc;
            padding: 4px 12px;
            border-radius: 9999px;
            font-size: 0.85rem;
            font-weight: 600;
            margin-bottom: 12px;
            border: 1px solid rgba(99, 102, 241, 0.3);
        }

        h1 {
            font-size: 2.2rem;
            font-weight: 700;
            letter-spacing: -0.5px;
            margin-bottom: 8px;
            background: linear-gradient(135deg, #fff 40%, #94a3b8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        header p {
            color: var(--text-muted);
            font-size: 1rem;
        }

        .card {
            background-color: var(--card-bg);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 28px;
            margin-bottom: 24px;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
        }

        .card h2 {
            font-size: 1.25rem;
            margin-bottom: 18px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .form-group {
            margin-bottom: 16px;
        }

        label {
            display: block;
            margin-bottom: 6px;
            font-size: 0.875rem;
            font-weight: 500;
            color: var(--text-muted);
        }

        input[type="text"], input[type="email"], input[type="password"] {
            width: 100%;
            padding: 12px 16px;
            background: #0f172a;
            border: 1px solid var(--border);
            border-radius: 8px;
            color: var(--text-main);
            font-size: 0.95rem;
            transition: border-color 0.2s;
        }

        input:focus {
            outline: none;
            border-color: var(--primary);
        }

        .btn {
            background-color: var(--primary);
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 0.95rem;
            font-weight: 600;
            transition: background 0.2s, transform 0.1s;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .btn:hover {
            background-color: var(--primary-hover);
        }

        .btn:active {
            transform: scale(0.98);
        }

        .btn-outline {
            background: transparent;
            border: 1px solid var(--border);
            color: var(--text-main);
        }

        .btn-outline:hover {
            background: rgba(255, 255, 255, 0.05);
        }

        .btn-download {
            background-color: var(--success);
            color: #064e3b;
            font-weight: 700;
            padding: 6px 14px;
            font-size: 0.825rem;
            border-radius: 6px;
            text-decoration: none;
        }

        .btn-download:hover {
            opacity: 0.9;
        }

        .tabs {
            display: flex;
            gap: 12px;
            margin-bottom: 16px;
        }

        .tab-btn {
            background: transparent;
            border: none;
            color: var(--text-muted);
            font-size: 0.95rem;
            font-weight: 600;
            padding: 8px 16px;
            cursor: pointer;
            border-bottom: 2px solid transparent;
        }

        .tab-btn.active {
            color: var(--primary);
            border-bottom-color: var(--primary);
        }

        .user-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #0f172a;
            padding: 12px 18px;
            border-radius: 8px;
            border: 1px solid var(--border);
            margin-bottom: 20px;
        }

        .drop-zone {
            border: 2px dashed var(--border);
            border-radius: 12px;
            padding: 36px 20px;
            text-align: center;
            cursor: pointer;
            transition: all 0.2s;
            background: rgba(15, 23, 42, 0.6);
        }

        .drop-zone:hover {
            border-color: var(--primary);
            background: rgba(99, 102, 241, 0.05);
        }

        .drop-zone p {
            color: var(--text-muted);
            margin-top: 8px;
            font-size: 0.875rem;
        }

        .alert {
            padding: 12px 16px;
            border-radius: 8px;
            font-size: 0.9rem;
            margin-bottom: 16px;
            display: none;
        }

        .alert-error {
            background: rgba(239, 68, 68, 0.15);
            border: 1px solid rgba(239, 68, 68, 0.3);
            color: #fca5a5;
        }

        .alert-success {
            background: rgba(16, 185, 129, 0.15);
            border: 1px solid rgba(16, 185, 129, 0.3);
            color: #86efac;
        }

        .job-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 12px;
        }

        .job-table th, .job-table td {
            padding: 12px 14px;
            text-align: left;
            border-bottom: 1px solid var(--border);
            font-size: 0.875rem;
        }

        .job-table th {
            color: var(--text-muted);
            font-weight: 600;
        }

        .status-badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
        }

        .status-PENDING { background: rgba(245, 158, 11, 0.15); color: #fbbf24; border: 1px solid rgba(245, 158, 11, 0.3); }
        .status-PROCESSING { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; border: 1px solid rgba(99, 102, 241, 0.3); }
        .status-COMPLETED { background: rgba(16, 185, 129, 0.15); color: #86efac; border: 1px solid rgba(16, 185, 129, 0.3); }
        .status-FAILED { background: rgba(239, 68, 68, 0.15); color: #fca5a5; border: 1px solid rgba(239, 68, 68, 0.3); }

        .hidden { display: none !important; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <span class="badge-brand">FIAP X — HACKATHON FASE 5</span>
            <h1>Processamento Distribuído de Vídeos</h1>
            <p>Upload assíncrono com extração de frames e compactação em .ZIP resiliente a picos</p>
        </header>

        <div id="alertBox" class="alert"></div>

        <!-- SEÇÃO DE AUTENTICAÇÃO -->
        <div id="authSection" class="card">
            <div class="tabs">
                <button class="tab-btn active" id="tabLogin" onclick="switchAuthTab('login')">Entrar</button>
                <button class="tab-btn" id="tabRegister" onclick="switchAuthTab('register')">Criar Conta</button>
            </div>

            <form id="loginForm" onsubmit="handleLogin(event)">
                <div class="form-group">
                    <label>E-mail</label>
                    <input type="email" id="loginEmail" placeholder="exemplo@fiapx.com" required>
                </div>
                <div class="form-group">
                    <label>Senha</label>
                    <input type="password" id="loginPassword" placeholder="••••••••" required>
                </div>
                <button type="submit" class="btn">Acessar Plataforma ➔</button>
            </form>

            <form id="registerForm" class="hidden" onsubmit="handleRegister(event)">
                <div class="form-group">
                    <label>Nome Completo</label>
                    <input type="text" id="regName" placeholder="Investidor FIAP X" required>
                </div>
                <div class="form-group">
                    <label>E-mail</label>
                    <input type="email" id="regEmail" placeholder="investidor@fiapx.com" required>
                </div>
                <div class="form-group">
                    <label>Senha (mínimo 6 caracteres)</label>
                    <input type="password" id="regPassword" placeholder="••••••••" required>
                </div>
                <button type="submit" class="btn">Cadastrar e Entrar ➔</button>
            </form>
        </div>

        <!-- SEÇÃO DO USUÁRIO LOGADO -->
        <div id="appSection" class="hidden">
            <div class="user-bar">
                <div>
                    <span style="color: var(--text-muted); font-size: 0.85rem;">Conectado como:</span>
                    <strong id="loggedUserName" style="margin-left: 6px;"></strong>
                    <span id="loggedUserEmail" style="color: var(--text-muted); font-size: 0.85rem; margin-left: 6px;"></span>
                </div>
                <button onclick="handleLogout()" class="btn btn-outline" style="padding: 6px 14px; font-size: 0.825rem;">Sair</button>
            </div>

            <!-- UPLOAD DE VÍDEO -->
            <div class="card">
                <h2>📤 Enviar Vídeo para Processamento</h2>
                <p style="color: var(--text-muted); font-size: 0.875rem; margin-bottom: 16px;">
                    O vídeo será enviado para o broker de mensageria assíncrona para extração de frames (1 frame/segundo) em background.
                </p>
                <form id="uploadForm" onsubmit="handleUpload(event)">
                    <div class="drop-zone" onclick="document.getElementById('videoFileInput').click()">
                        <input type="file" id="videoFileInput" style="display: none" accept=".mp4,.avi,.mov,.mkv,.wmv,.flv,.webm" onchange="updateSelectedFile(this)">
                        <div style="font-size: 2.2rem; margin-bottom: 6px;">🎬</div>
                        <strong id="selectedFileName">Clique para selecionar ou arraste o vídeo aqui</strong>
                        <p>Formatos suportados: MP4, AVI, MOV, MKV, WMV, WEBM</p>
                    </div>
                    <div style="margin-top: 16px; text-align: right;">
                        <button type="submit" id="btnUpload" class="btn" disabled>🚀 Iniciar Processamento Assíncrono</button>
                    </div>
                </form>
            </div>

            <!-- LISTAGEM DE STATUS DOS VÍDEOS -->
            <div class="card">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
                    <h2>📋 Meus Vídeos Processados</h2>
                    <button onclick="loadUserVideos()" class="btn btn-outline" style="padding: 6px 12px; font-size: 0.8rem;">🔄 Atualizar</button>
                </div>
                <div style="overflow-x: auto;">
                    <table class="job-table">
                        <thead>
                            <tr>
                                <th>Arquivo</th>
                                <th>Status</th>
                                <th>Frames</th>
                                <th>Data de Envio</th>
                                <th>Ações</th>
                            </tr>
                        </thead>
                        <tbody id="jobsTableBody">
                            <tr>
                                <td colspan="5" style="text-align: center; color: var(--text-muted);">Carregando vídeos...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentUser = null;
        let token = localStorage.getItem('fiapx_token');
        let refreshTimer = null;

        function showAlert(msg, type = 'error') {
            const el = document.getElementById('alertBox');
            el.className = 'alert alert-' + type;
            el.innerText = msg;
            el.style.display = 'block';
            setTimeout(() => { el.style.display = 'none'; }, 6000);
        }

        function switchAuthTab(tab) {
            document.getElementById('tabLogin').classList.toggle('active', tab === 'login');
            document.getElementById('tabRegister').classList.toggle('active', tab === 'register');
            document.getElementById('loginForm').classList.toggle('hidden', tab !== 'login');
            document.getElementById('registerForm').classList.toggle('hidden', tab !== 'register');
        }

        async function handleLogin(e) {
            e.preventDefault();
            const email = document.getElementById('loginEmail').value;
            const password = document.getElementById('loginPassword').value;

            try {
                const res = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, password })
                });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Falha no login');

                setSession(data.token, data.user);
            } catch (err) {
                showAlert(err.message, 'error');
            }
        }

        async function handleRegister(e) {
            e.preventDefault();
            const name = document.getElementById('regName').value;
            const email = document.getElementById('regEmail').value;
            const password = document.getElementById('regPassword').value;

            try {
                const res = await fetch('/api/v1/auth/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, email, password })
                });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Falha no cadastro');

                showAlert('Conta criada com sucesso!', 'success');
                setSession(data.token, data.user);
            } catch (err) {
                showAlert(err.message, 'error');
            }
        }

        function setSession(newToken, user) {
            token = newToken;
            currentUser = user;
            localStorage.setItem('fiapx_token', token);
            localStorage.setItem('fiapx_user', JSON.stringify(user));
            updateUIState();
        }

        function handleLogout() {
            localStorage.removeItem('fiapx_token');
            localStorage.removeItem('fiapx_user');
            token = null;
            currentUser = null;
            if (refreshTimer) clearInterval(refreshTimer);
            updateUIState();
        }

        function updateUIState() {
            const hasAuth = !!token;
            document.getElementById('authSection').classList.toggle('hidden', hasAuth);
            document.getElementById('appSection').classList.toggle('hidden', !hasAuth);

            if (hasAuth) {
                const storedUser = JSON.parse(localStorage.getItem('fiapx_user') || '{}');
                document.getElementById('loggedUserName').innerText = storedUser.name || 'Usuário';
                document.getElementById('loggedUserEmail').innerText = '(' + (storedUser.email || '') + ')';
                loadUserVideos();
                if (!refreshTimer) {
                    refreshTimer = setInterval(loadUserVideos, 3000); // auto-refresh a cada 3s
                }
            }
        }

        function updateSelectedFile(input) {
            const file = input.files[0];
            const btn = document.getElementById('btnUpload');
            if (file) {
                document.getElementById('selectedFileName').innerText = 'Arquivo: ' + file.name + ' (' + (file.size / 1024 / 1024).toFixed(2) + ' MB)';
                btn.disabled = false;
            } else {
                document.getElementById('selectedFileName').innerText = 'Clique para selecionar ou arraste o vídeo aqui';
                btn.disabled = true;
            }
        }

        async function handleUpload(e) {
            e.preventDefault();
            const fileInput = document.getElementById('videoFileInput');
            if (!fileInput.files[0]) return;

            const formData = new FormData();
            formData.append('video', fileInput.files[0]);

            const btn = document.getElementById('btnUpload');
            btn.disabled = true;
            btn.innerText = 'Enviando...';

            try {
                const res = await fetch('/api/v1/videos/upload', {
                    method: 'POST',
                    headers: { 'Authorization': 'Bearer ' + token },
                    body: formData
                });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Falha no upload');

                showAlert('Vídeo enfileirado com sucesso para processamento assíncrono!', 'success');
                fileInput.value = '';
                document.getElementById('selectedFileName').innerText = 'Clique para selecionar ou arraste o vídeo aqui';
                loadUserVideos();
            } catch (err) {
                showAlert(err.message, 'error');
            } finally {
                btn.disabled = false;
                btn.innerText = '🚀 Iniciar Processamento Assíncrono';
            }
        }

        async function loadUserVideos() {
            if (!token) return;
            try {
                const res = await fetch('/api/v1/videos', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                if (res.status === 401) {
                    handleLogout();
                    return;
                }
                const data = await res.json();
                renderVideos(data.videos || []);
            } catch (err) {
                console.error('Erro ao listar vídeos:', err);
            }
        }

        function renderVideos(videos) {
            const tbody = document.getElementById('jobsTableBody');
            if (!videos || videos.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--text-muted);">Nenhum vídeo enviado ainda. Faça seu primeiro upload acima!</td></tr>';
                return;
            }

            tbody.innerHTML = videos.map(v => {
                const date = new Date(v.created_at).toLocaleString('pt-BR');
                let actionBtn = '-';
                if (v.status === 'COMPLETED' && v.download_url) {
                    actionBtn = '<a href="' + v.download_url + '?token=' + token + '" class="btn-download">⬇️ Baixar .ZIP</a>';
                } else if (v.status === 'FAILED') {
                    actionBtn = '<span style="color: var(--danger); font-size: 0.8rem;" title="' + (v.error_message || '') + '">❌ ' + (v.error_message || 'Erro') + '</span>';
                } else if (v.status === 'PROCESSING') {
                    actionBtn = '<span style="color: #a5b4fc; font-size: 0.8rem;">⚙️ Extraindo frames...</span>';
                } else {
                    actionBtn = '<span style="color: #fbbf24; font-size: 0.8rem;">⏳ Na fila...</span>';
                }

                return '<tr>' +
                    '<td><strong>' + v.original_name + '</strong></td>' +
                    '<td><span class="status-badge status-' + v.status + '">' + v.status + '</span></td>' +
                    '<td>' + (v.frame_count > 0 ? v.frame_count + ' frames' : '-') + '</td>' +
                    '<td>' + date + '</td>' +
                    '<td>' + actionBtn + '</td>' +
                '</tr>';
            }).join('');
        }

        // Inicialização
        if (token) {
            updateUIState();
        }
    </script>
</body>
</html>`
}
