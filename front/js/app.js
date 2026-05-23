const STORAGE_KEYS = {
    userId:       'auth_user_id',
    jwtToken:     'auth_jwt_token',
    refreshToken: 'auth_refresh_token',
};

const VIEWS = ['login-view', 'register-view', 'dashboard-view'];

function showView(viewId) {
    VIEWS.forEach(id => {
        document.getElementById(id).style.display = id === viewId ? 'block' : 'none';
    });
}

function showError(elementId, message) {
    const el = document.getElementById(elementId);
    el.textContent = message;
    el.style.display = 'block';
}

document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem(STORAGE_KEYS.jwtToken);
    if (token) {
        const userId = localStorage.getItem(STORAGE_KEYS.userId);
        document.getElementById('welcome-msg').textContent = 'Welcome! Your user ID is: ' + userId;
        showView('dashboard-view');
    } else {
        showView('login-view');
    }

    document.getElementById('login-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const login    = document.getElementById('login-input').value;
        const password = document.getElementById('login-password-input').value;
        const btn      = e.target.querySelector('button[type="submit"]');

        document.getElementById('login-error').style.display = 'none';
        btn.disabled = true;

        const result = await window.api.login(login, password);

        if (result.ok) {
            localStorage.setItem(STORAGE_KEYS.userId,       result.data.user_id);
            localStorage.setItem(STORAGE_KEYS.jwtToken,     result.data.jwt_token);
            localStorage.setItem(STORAGE_KEYS.refreshToken, result.data.refresh_token);
            document.getElementById('welcome-msg').textContent = 'Welcome! Your user ID is: ' + result.data.user_id;
            showView('dashboard-view');
        } else {
            showError('login-error', result.error);
            btn.disabled = false;
        }
    });

    document.getElementById('register-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const login    = document.getElementById('register-input').value;
        const password = document.getElementById('register-password-input').value;

        document.getElementById('register-error').style.display   = 'none';
        document.getElementById('register-success').style.display = 'none';

        if (!login || !password) {
            showError('register-error', 'All fields required');
            return;
        }

        const result = await window.api.register(login, password);

        if (result.ok) {
            const successEl = document.getElementById('register-success');
            successEl.textContent = 'Registered! You can now log in.';
            successEl.style.display = 'block';
            setTimeout(() => showView('login-view'), 1500);
        } else {
            showError('register-error', result.error);
        }
    });

    document.getElementById('logout-btn').addEventListener('click', () => {
        localStorage.removeItem(STORAGE_KEYS.userId);
        localStorage.removeItem(STORAGE_KEYS.jwtToken);
        localStorage.removeItem(STORAGE_KEYS.refreshToken);
        showView('login-view');
    });

    document.getElementById('go-to-register').addEventListener('click', (e) => {
        e.preventDefault();
        showView('register-view');
    });

    document.getElementById('go-to-login').addEventListener('click', (e) => {
        e.preventDefault();
        showView('login-view');
    });
});
