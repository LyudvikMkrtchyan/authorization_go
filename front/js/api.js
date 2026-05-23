const BASE_URL = 'http://localhost:8080';

async function register(login, password) {
    try {
        const res = await fetch(`${BASE_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ login, password }),
        });
        if (res.ok) return { ok: true };
        if (res.status === 409) return { ok: false, error: 'user already exists' };
        return { ok: false, error: 'server error' };
    } catch {
        return { ok: false, error: 'network error' };
    }
}

async function login(login, password) {
    try {
        const res = await fetch(`${BASE_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ login, password }),
        });
        if (res.ok) {
            const data = await res.json();
            return { ok: true, data };
        }
        if (res.status === 401) return { ok: false, error: 'invalid credentials' };
        return { ok: false, error: 'server error' };
    } catch {
        return { ok: false, error: 'network error' };
    }
}

async function refreshToken(refreshToken) {
    try {
        const res = await fetch(`${BASE_URL}/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (res.ok) {
            const data = await res.json();
            return { ok: true, data };
        }
        const body = await res.json().catch(() => ({}));
        return { ok: false, error: body.error || 'server error' };
    } catch {
        return { ok: false, error: 'network error' };
    }
}

window.api = { register, login, refreshToken };
