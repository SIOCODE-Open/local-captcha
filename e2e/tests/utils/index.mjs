import path from 'path';
import fs from 'fs';
import cp from 'child_process';
import { z } from 'zod';

export function getBaseUrl(defaultPort = 8080, envPortKey = 'TEST_PORT', envUrlKey = 'BASE_URL') {
    if (process.env[envUrlKey]) {
        return process.env[envUrlKey];
    }
    const port = process.env[envPortKey] || defaultPort;
    return `http://localhost:${port}`;
}

export async function waitAvailable(baseUrl, timeoutMs = 120000, intervalMs = 500) {
    const startTime = Date.now();
    while (true) {
        try {
            const response = await fetch(`${baseUrl}/healthz`);
            if (response.ok) {
                return;
            }
        } catch (err) {
            // Ignore errors and retry
        }
        if (Date.now() - startTime > timeoutMs) {
            throw new Error(`Service at ${baseUrl} did not become available within ${timeoutMs} ms`);
        }
        await new Promise(resolve => setTimeout(resolve, intervalMs));
    }
}

export async function launchSnapshot(snapshotName) {
    const snapshotsDir = path.resolve(process.cwd(), 'snapshots', snapshotName);
    if (!fs.existsSync(snapshotsDir)) {
        throw new Error(`Snapshot directory does not exist: ${snapshotsDir}`);
    }

    const snapshotDockerCompose = path.join(snapshotsDir, 'docker-compose.yaml');
    if (!fs.existsSync(snapshotDockerCompose)) {
        throw new Error(`docker-compose.yaml not found in snapshot directory: ${snapshotDockerCompose}`);
    }

    const projectName = `local-captcha-${snapshotName}-${process.env.TEST_PORT || 'default'}`;

    return new Promise((resolve, reject) => {
        try {
            const dockerComposeProcess = cp.spawn('docker', ['compose', '-p', projectName, '-f', snapshotDockerCompose, 'up', '-d', '--build'], {
                cwd: snapshotsDir,
                env: process.env,
            });
            dockerComposeProcess.on('close', (code) => {
                if (code === 0) {
                    resolve();
                } else {
                    reject(new Error(`docker-compose process exited with code ${code}`));
                }
            });
        } catch (err) {
            reject(err);
        }
    });
}

export async function teardownSnapshot(snapshotName) {
    const snapshotsDir = path.resolve(process.cwd(), 'snapshots', snapshotName);
    const snapshotDockerCompose = path.join(snapshotsDir, 'docker-compose.yaml');
    const projectName = `local-captcha-${snapshotName}-${process.env.TEST_PORT || 'default'}`;

    return new Promise((resolve, reject) => {
        try {
            const dockerComposeProcess = cp.spawn('docker', ['compose', '-p', projectName, '-f', snapshotDockerCompose, 'down', '-v'], {
                cwd: snapshotsDir,
                env: process.env,
            });
            dockerComposeProcess.on('close', (code) => {
                if (code === 0) {
                    resolve();
                } else {
                    reject(new Error(`docker-compose process exited with code ${code}`));
                }
            });
        } catch (err) {
            reject(err);
        }
    });
}

const Z_ChallengeSolveRequest = z.object({
    sitekey: z.string().optional(),
    k: z.string().optional(),
    action: z.string().optional(),
    challenge_answer: z.string().optional(),
    score: z.number().optional(),
    hostname: z.string().optional(),
});

const Z_SiteVerifyRequest = z.object({
    secret: z.string().optional(),
    response: z.string().optional(),
    remoteip: z.string().optional(),
});

const Z_CreateTokenRequest = z.object({
    token: z.string().optional(),
    sitekey: z.string().optional(),
    action: z.string().optional(),
    score: z.number().optional(),
    hostname: z.string().optional(),
    ttl_seconds: z.number().optional(),
    expired: z.boolean().optional(),
    used: z.boolean().optional(),
    error_codes: z.array(z.string()).optional(),
});

export class CaptchaClient {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
    }

    async healthz() {
        const response = await fetch(`${this.baseUrl}/healthz`);
        if (!response.ok) {
            throw new Error(`Health check failed with status ${response.status}`);
        }
        return await response.json();
    }

    async getApiJs(params = {}, pathAlias = '/recaptcha/api.js') {
        const qs = new URLSearchParams(params).toString();
        const url = qs ? `${this.baseUrl}${pathAlias}?${qs}` : `${this.baseUrl}${pathAlias}`;
        const response = await fetch(url);
        return {
            status: response.status,
            headers: response.headers,
            text: await response.text(),
        };
    }

    async getAnchor(params = {}, pathAlias = '/recaptcha/anchor') {
        const qs = new URLSearchParams(params).toString();
        const url = qs ? `${this.baseUrl}${pathAlias}?${qs}` : `${this.baseUrl}${pathAlias}`;
        const response = await fetch(url);
        const text = await response.text();
        let json = null;
        try {
            json = JSON.parse(text);
        } catch (_) {
            // HTML response
        }
        return {
            status: response.status,
            headers: response.headers,
            text,
            json,
        };
    }

    async solveJson(params = {}, pathAlias = '/challenge/solve') {
        const parsed = Z_ChallengeSolveRequest.parse(params);
        const response = await fetch(`${this.baseUrl}${pathAlias}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(parsed),
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async solveForm(params = {}, pathAlias = '/challenge/solve') {
        const body = new URLSearchParams();
        for (const [k, v] of Object.entries(params)) {
            if (v !== undefined) body.set(k, String(v));
        }
        const response = await fetch(`${this.baseUrl}${pathAlias}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body,
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async solveQuery(params = {}, pathAlias = '/challenge/solve') {
        const qs = new URLSearchParams();
        for (const [k, v] of Object.entries(params)) {
            if (v !== undefined) qs.set(k, String(v));
        }
        const response = await fetch(`${this.baseUrl}${pathAlias}?${qs.toString()}`, {
            method: 'POST',
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async verifyForm(params = {}, pathAlias = '/recaptcha/api/siteverify') {
        const parsed = Z_SiteVerifyRequest.parse(params);
        const body = new URLSearchParams();
        for (const [k, v] of Object.entries(parsed)) {
            if (v !== undefined) body.set(k, String(v));
        }
        const response = await fetch(`${this.baseUrl}${pathAlias}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body,
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async verifyJson(params = {}, pathAlias = '/recaptcha/api/siteverify') {
        const parsed = Z_SiteVerifyRequest.parse(params);
        const response = await fetch(`${this.baseUrl}${pathAlias}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(parsed),
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async verifyQuery(params = {}, pathAlias = '/recaptcha/api/siteverify', method = 'POST') {
        const parsed = Z_SiteVerifyRequest.parse(params);
        const qs = new URLSearchParams();
        for (const [k, v] of Object.entries(parsed)) {
            if (v !== undefined) qs.set(k, String(v));
        }
        const response = await fetch(`${this.baseUrl}${pathAlias}?${qs.toString()}`, {
            method,
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async createTestingToken(params = {}) {
        const parsed = Z_CreateTokenRequest.parse(params);
        const response = await fetch(`${this.baseUrl}/testing/tokens`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(parsed),
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async listTestingTokens(params = {}) {
        const qs = new URLSearchParams(params).toString();
        const url = qs ? `${this.baseUrl}/testing/tokens?${qs}` : `${this.baseUrl}/testing/tokens`;
        const response = await fetch(url);
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async listTestingVerifications(params = {}) {
        const qs = new URLSearchParams(params).toString();
        const url = qs ? `${this.baseUrl}/testing/verifications?${qs}` : `${this.baseUrl}/testing/verifications`;
        const response = await fetch(url);
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async getTestingConfig() {
        const response = await fetch(`${this.baseUrl}/testing/config`);
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async updateTestingConfig(params = {}, method = 'PUT') {
        const response = await fetch(`${this.baseUrl}/testing/config`, {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(params),
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }

    async resetTestingState() {
        const response = await fetch(`${this.baseUrl}/testing/reset`, {
            method: 'POST',
        });
        return {
            status: response.status,
            data: await response.json(),
        };
    }
}
