import { expect } from 'chai';
import { CaptchaClient, getBaseUrl, launchSnapshot, teardownSnapshot, waitAvailable } from "./utils/index.mjs";

describe('strict-mode', () => {
    const baseUrl = getBaseUrl(8081, 'STRICT_TEST_PORT', 'STRICT_BASE_URL');
    const client = new CaptchaClient(baseUrl);

    before(async () => {
        await launchSnapshot('strict-mode');
        await waitAvailable(baseUrl);
    });

    after(async () => {
        await teardownSnapshot('strict-mode');
    });

    beforeEach(async () => {
        await client.resetTestingState();
    });

    describe('Strict Sitekey Validation (allow_any_site: false)', () => {
        it('GET /recaptcha/anchor with unknown sitekey should return 400 Bad Request', async () => {
            const res = await client.getAnchor({ k: 'unregistered-sitekey' });
            expect(res.status).to.equal(400);
        });

        it('GET /challenge with unknown sitekey should return 400 Bad Request', async () => {
            const res = await client.getAnchor({ sitekey: 'unregistered-sitekey' }, '/challenge');
            expect(res.status).to.equal(400);
        });

        it('GET /recaptcha/anchor with registered sitekey should return 200 OK', async () => {
            const res = await client.getAnchor({ k: 'strict-site-1' });
            expect(res.status).to.equal(200);
            expect(res.text).to.include('id="recaptcha-anchor"');
        });

        it('GET /recaptcha/anchor with Google official test sitekey should return 200 OK even when allow_any_site is false', async () => {
            const res = await client.getAnchor({
                k: '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI',
            });
            expect(res.status).to.equal(200);
            expect(res.text).to.include('id="recaptcha-anchor"');
        });

        it('POST /challenge/solve with unknown sitekey should return 400 Bad Request with invalid-sitekey', async () => {
            const res = await client.solveJson({ sitekey: 'unregistered-sitekey' });
            expect(res.status).to.equal(400);
            expect(res.data).to.have.property('success', false);
            expect(res.data).to.have.property('error', 'invalid-sitekey');
            expect(res.data['error-codes']).to.include('invalid-sitekey');
        });

        it('POST /challenge/solve with missing sitekey (defaults to default-sitekey) should return 400 Bad Request when not registered', async () => {
            const res = await client.solveJson({});
            expect(res.status).to.equal(400);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.include('invalid-sitekey');
        });

        it('POST /challenge/solve with Google official test sitekey should succeed even when allow_any_site is false', async () => {
            const res = await client.solveJson({
                sitekey: '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI',
                action: 'test_google_key',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('token').that.is.a('string');
        });
    });

    describe('Strict Secret Validation (allow_any_secret: false)', () => {
        it('POST /recaptcha/api/siteverify with unknown secret should return invalid-input-secret', async () => {
            const res = await client.verifyForm({
                secret: 'unknown-secret-key',
                response: 'always-pass-token',
            });
            expect(res.status).to.equal(200);
            expect(res.data.success).to.equal(false);
            expect(res.data['error-codes']).to.deep.equal(['invalid-input-secret']);
        });

        it('POST /recaptcha/api/siteverify with configured secret should succeed', async () => {
            const res = await client.verifyForm({
                secret: 'strict-secret-1',
                response: 'always-pass-token',
            });
            expect(res.status).to.equal(200);
            expect(res.data.success).to.equal(true);
            expect(res.data.score).to.equal(0.95);
            expect(res.data.hostname).to.equal('strict1.example.local');
        });

        it('POST /recaptcha/api/siteverify with Google official test secret should succeed even when allow_any_secret is false', async () => {
            const solved = await client.solveJson({
                sitekey: '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI',
            });
            const verified = await client.verifyForm({
                secret: '6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe',
                response: solved.data.token,
            });
            expect(verified.status).to.equal(200);
            expect(verified.data.success).to.equal(true);
        });
    });

    describe('Token TTL Expiration (token_ttl_seconds: 1)', () => {
        it('Issued token should verify immediately and fail with timeout-or-duplicate after TTL expires', async () => {
            await client.updateTestingConfig({ verify_once: false });
            const solved = await client.solveJson({
                sitekey: 'strict-site-1',
                action: 'quick_action',
            });
            expect(solved.status).to.equal(200);
            expect(solved.data.expires_in).to.equal(1);

            // Immediate verification succeeds
            const immediate = await client.verifyForm({
                secret: 'strict-secret-1',
                response: solved.data.token,
            });
            expect(immediate.data.success).to.equal(true);

            // Wait for 1s TTL + buffer
            await new Promise(resolve => setTimeout(resolve, 1500));

            // Subsequent verification fails due to expiration
            const expired = await client.verifyForm({
                secret: 'strict-secret-1',
                response: solved.data.token,
            });
            expect(expired.data.success).to.equal(false);
            expect(expired.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);
        });
    });

    describe('Configured CORS Allowed Origins', () => {
        it('Should return matching Access-Control-Allow-Origin for configured origin', async () => {
            const res = await fetch(`${baseUrl}/healthz`, {
                headers: {
                    'Origin': 'http://localhost:3000',
                },
            });
            expect(res.status).to.equal(200);
            expect(res.headers.get('Access-Control-Allow-Origin')).to.include('http://localhost:3000');
        });
    });
});
