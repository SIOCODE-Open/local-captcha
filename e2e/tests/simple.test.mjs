import { expect } from 'chai';
import { CaptchaClient, getBaseUrl, launchSnapshot, teardownSnapshot, waitAvailable } from "./utils/index.mjs";

describe('simple', () => {
    const baseUrl = getBaseUrl(8080, 'TEST_PORT', 'BASE_URL');
    const client = new CaptchaClient(baseUrl);

    before(async () => {
        await launchSnapshot('simple');
        await waitAvailable(baseUrl);
    });

    after(async () => {
        await teardownSnapshot('simple');
    });

    beforeEach(async () => {
        await client.resetTestingState();
    });

    describe('Health Check', () => {
        it('GET /healthz should return 200 OK with status OK', async () => {
            const health = await client.healthz();
            expect(health).to.have.property('status', 'OK');
        });

        it('Unknown route should return 404 Not Found', async () => {
            const res = await fetch(`${baseUrl}/non-existent-route`);
            expect(res.status).to.equal(404);
        });

        it('OPTIONS request should return 204 No Content with CORS headers', async () => {
            const res = await fetch(`${baseUrl}/recaptcha/api/siteverify`, {
                method: 'OPTIONS',
                headers: {
                    'Origin': 'http://localhost:3000',
                    'Access-Control-Request-Method': 'POST',
                },
            });
            expect(res.status).to.equal(204);
            expect(res.headers.get('Access-Control-Allow-Origin')).to.exist;
        });
    });

    describe('Injectable JavaScript SDK', () => {
        it('GET /recaptcha/api.js should return JavaScript SDK defining window.grecaptcha', async () => {
            const res = await client.getApiJs({}, '/recaptcha/api.js');
            expect(res.status).to.equal(200);
            expect(res.headers.get('content-type')).to.include('application/javascript');
            expect(res.text).to.include('grecaptcha');
            expect(res.text).to.include('ready');
            expect(res.text).to.include('render');
            expect(res.text).to.include('execute');
            expect(res.text).to.include('getResponse');
            expect(res.text).to.include('reset');
            expect(res.text).to.include('/recaptcha/anchor');
            expect(res.text).to.include('/challenge/solve');
        });

        it('GET /captcha.js alias should return identical JavaScript SDK', async () => {
            const res1 = await client.getApiJs({}, '/recaptcha/api.js');
            const res2 = await client.getApiJs({}, '/captcha.js');
            expect(res2.status).to.equal(200);
            expect(res2.headers.get('content-type')).to.include('application/javascript');
            expect(res2.text).to.equal(res1.text);
        });

        it('GET /recaptcha/api.js should support render, onload, and hl query parameters', async () => {
            const resExplicit = await client.getApiJs({
                render: 'explicit',
                onload: 'myOnloadCallback',
                hl: 'en',
            });
            expect(resExplicit.status).to.equal(200);
            expect(resExplicit.text).to.include('explicit');
            expect(resExplicit.text).to.include('myOnloadCallback');

            const resV3 = await client.getApiJs({
                render: 'site-key-1',
                onload: 'onV3Ready',
                hl: 'fr',
            });
            expect(resV3.status).to.equal(200);
            expect(resV3.text).to.include('site-key-1');
            expect(resV3.text).to.include('onV3Ready');
        });
    });

    describe('Interactive Challenge UI', () => {
        it('GET /recaptcha/anchor should return HTML with required DOM elements in auto_pass mode', async () => {
            const res = await client.getAnchor({
                k: 'site-key-1',
                size: 'normal',
                theme: 'light',
                action: 'login',
                widget_id: '0',
            }, '/recaptcha/anchor');
            expect(res.status).to.equal(200);
            expect(res.headers.get('content-type')).to.include('text/html');
            expect(res.text).to.include('id="recaptcha-anchor"');
            expect(res.text).to.include('id="recaptcha-challenge"');
            expect(res.text).to.include('id="solve-challenge-btn"');
            expect(res.text).to.include('id="challenge-answer"');
            expect(res.text).to.include('id="recaptcha-token"');
            expect(res.text).to.include('id="recaptcha-error"');
            expect(res.text).to.include('local-captcha-solved');
            expect(res.text).to.include('local-captcha-error');
        });

        it('GET /challenge alias should return HTML with required DOM elements', async () => {
            const res = await client.getAnchor({
                sitekey: 'site-key-challenge',
                size: 'compact',
                theme: 'dark',
                action: 'register',
                widget_id: '2',
            }, '/challenge');
            expect(res.status).to.equal(200);
            expect(res.headers.get('content-type')).to.include('text/html');
            expect(res.text).to.include('id="recaptcha-anchor"');
            expect(res.text).to.include('id="recaptcha-challenge"');
            expect(res.text).to.include('id="solve-challenge-btn"');
            expect(res.text).to.include('id="challenge-answer"');
            expect(res.text).to.include('id="recaptcha-token"');
            expect(res.text).to.include('id="recaptcha-error"');
        });

        it('GET /recaptcha/anchor in always_fail mode should include error element', async () => {
            const res = await client.getAnchor({
                k: 'site-key-fail',
            });
            expect(res.status).to.equal(200);
            expect(res.text).to.include('id="recaptcha-error"');
        });

        it('GET /recaptcha/anchor without parameters should default to default-sitekey', async () => {
            const res = await client.getAnchor({});
            expect(res.status).to.equal(200);
            expect(res.text).to.include('default-sitekey');
        });
    });

    describe('Token Issuance (POST /challenge/solve & POST /recaptcha/reload)', () => {
        it('POST /challenge/solve via JSON should issue a valid token with site defaults', async () => {
            const res = await client.solveJson({
                sitekey: 'site-key-1',
                action: 'login',
            }, '/challenge/solve');
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('token').that.is.a('string').and.not.empty;
            expect(res.data).to.have.property('sitekey', 'site-key-1');
            expect(res.data).to.have.property('action', 'login');
            expect(res.data).to.have.property('score', 0.9);
            expect(res.data).to.have.property('hostname', 'app.example.local');
            expect(res.data).to.have.property('challenge_ts').that.is.a('string');
            expect(res.data).to.have.property('expires_in', 300);
        });

        it('POST /recaptcha/reload alias should issue a valid token', async () => {
            const res = await client.solveJson({
                sitekey: 'site-key-bot',
                action: 'scrape',
            }, '/recaptcha/reload');
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('token').that.is.a('string');
            expect(res.data).to.have.property('sitekey', 'site-key-bot');
            expect(res.data).to.have.property('action', 'scrape');
            expect(res.data).to.have.property('score', 0.1);
            expect(res.data).to.have.property('hostname', 'localhost');
        });

        it('POST /challenge/solve via application/x-www-form-urlencoded should work and accept k alias', async () => {
            const res = await client.solveForm({
                k: 'site-key-challenge',
                action: 'comment',
                challenge_answer: 'human-answer',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('sitekey', 'site-key-challenge');
            expect(res.data).to.have.property('action', 'comment');
            expect(res.data).to.have.property('score', 0.8);
        });

        it('POST /challenge/solve via query parameters should work and apply default-sitekey & submit action', async () => {
            const res = await client.solveQuery({});
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('sitekey', 'default-sitekey');
            expect(res.data).to.have.property('action', 'submit');
            expect(res.data).to.have.property('score', 0.9);
            expect(res.data).to.have.property('hostname', 'localhost');
        });

        it('POST /challenge/solve should support custom score and hostname overrides', async () => {
            const res = await client.solveJson({
                sitekey: 'site-key-1',
                action: 'custom_act',
                score: 0.42,
                hostname: 'override.example.org',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', true);
            expect(res.data).to.have.property('score', 0.42);
            expect(res.data).to.have.property('hostname', 'override.example.org');

            // Verify the issued token carries the overridden score and hostname
            const verifyRes = await client.verifyForm({
                secret: 'secret-key-1',
                response: res.data.token,
            });
            expect(verifyRes.status).to.equal(200);
            expect(verifyRes.data).to.have.property('success', true);
            expect(verifyRes.data).to.have.property('score', 0.42);
            expect(verifyRes.data).to.have.property('hostname', 'override.example.org');
            expect(verifyRes.data).to.have.property('action', 'custom_act');
        });

        it('POST /challenge/solve with challenge_answer="fail" should return success=false', async () => {
            const res = await client.solveJson({
                sitekey: 'site-key-challenge',
                challenge_answer: 'fail',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data).to.have.property('error-codes').that.includes('invalid-input-response');
        });

        it('POST /challenge/solve for a site in always_fail mode should return success=false with site error_codes', async () => {
            const res = await client.solveJson({
                sitekey: 'site-key-fail',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data).to.have.property('error-codes').that.is.an('array').and.not.empty;
        });

        it('POST /challenge/solve when global mode is always_fail should return success=false', async () => {
            await client.updateTestingConfig({ mode: 'always_fail' });
            const res = await client.solveJson({
                sitekey: 'site-key-1',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data).to.have.property('error-codes').that.includes('invalid-input-response');
        });

        it('POST /challenge/solve when force_error_codes is set should return success=false', async () => {
            await client.updateTestingConfig({ force_error_codes: ['forced-solve-error'] });
            const res = await client.solveJson({
                sitekey: 'site-key-1',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data).to.have.property('error-codes').that.includes('forced-solve-error');
        });
    });

    describe('CAPTCHA Verification API (siteverify, /siteverify, /verify)', () => {
        it('Should verify an issued token via POST form-urlencoded on /recaptcha/api/siteverify', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-1', action: 'checkout' });
            const verified = await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
                remoteip: '192.168.1.50',
            }, '/recaptcha/api/siteverify');

            expect(verified.status).to.equal(200);
            expect(verified.data).to.have.property('success', true);
            expect(verified.data).to.have.property('score', 0.9);
            expect(verified.data).to.have.property('action', 'checkout');
            expect(verified.data).to.have.property('hostname', 'app.example.local');
            expect(verified.data).to.have.property('challenge_ts');
        });

        it('Should verify an issued token via POST JSON on /siteverify alias', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-challenge', action: 'signup' });
            const verified = await client.verifyJson({
                secret: 'secret-key-challenge',
                response: solved.data.token,
            }, '/siteverify');

            expect(verified.status).to.equal(200);
            expect(verified.data).to.have.property('success', true);
            expect(verified.data).to.have.property('score', 0.8);
            expect(verified.data).to.have.property('action', 'signup');
            expect(verified.data).to.have.property('hostname', 'localhost');
        });

        it('Should verify an issued token via POST query params on /verify alias', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-bot', action: 'ping' });
            const verified = await client.verifyQuery({
                secret: 'secret-key-bot',
                response: solved.data.token,
            }, '/verify', 'POST');

            expect(verified.status).to.equal(200);
            expect(verified.data).to.have.property('success', true);
            expect(verified.data).to.have.property('score', 0.1);
            expect(verified.data).to.have.property('action', 'ping');
        });

        it('Should verify an issued token via GET query params on all three verification paths', async () => {
            for (const pathAlias of ['/recaptcha/api/siteverify', '/siteverify', '/verify']) {
                const solved = await client.solveJson({ sitekey: 'site-key-1', action: 'get_verify' });
                const verified = await client.verifyQuery({
                    secret: 'secret-key-1',
                    response: solved.data.token,
                }, pathAlias, 'GET');

                expect(verified.status).to.equal(200);
                expect(verified.data).to.have.property('success', true);
                expect(verified.data).to.have.property('action', 'get_verify');
            }
        });
    });

    describe('5-Step Verification Evaluation Order & Error Codes', () => {
        it('Step 1: Missing both secret and response should return missing-input-secret and missing-input-response', async () => {
            const res = await client.verifyForm({});
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['missing-input-secret', 'missing-input-response']);
        });

        it('Step 1: Missing secret only should return missing-input-secret', async () => {
            const res = await client.verifyForm({ response: 'always-pass-token' });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['missing-input-secret']);
        });

        it('Step 1: Missing response only should return missing-input-response', async () => {
            const res = await client.verifyForm({ secret: 'secret-key-1' });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['missing-input-response']);
        });

        it('Step 2: Global force_error_codes takes precedence over magic tokens and valid secrets', async () => {
            await client.updateTestingConfig({
                force_error_codes: ['custom-forced-1', 'custom-forced-2'],
            });
            const res = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'always-pass-token',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['custom-forced-1', 'custom-forced-2']);
        });

        it('Step 2: Global mode="always_fail" fails even with always-pass-token', async () => {
            await client.updateTestingConfig({
                mode: 'always_fail',
            });
            const res = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'always-pass-token',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['invalid-input-response']);
        });

        it('Step 3: Secret matching a site with mode="always_fail" fails with site error_codes', async () => {
            const res = await client.verifyForm({
                secret: 'secret-key-fail',
                response: 'always-pass-token',
            });
            expect(res.status).to.equal(200);
            expect(res.data).to.have.property('success', false);
            expect(res.data['error-codes']).to.deep.equal(['custom-site-fail-code']);
        });

        it('Step 4: Built-in magic tokens should return deterministic results and not be consumed by verify_once', async () => {
            // always-pass-token (verify twice to confirm not consumed by verify_once)
            const pass1 = await client.verifyForm({ secret: 'secret-key-1', response: 'always-pass-token' });
            expect(pass1.data).to.include({ success: true, score: 0.9, action: 'submit', hostname: 'app.example.local' });
            const pass2 = await client.verifyForm({ secret: 'secret-key-1', response: 'always-pass-token' });
            expect(pass2.data).to.include({ success: true, score: 0.9, action: 'submit' });

            // valid-token
            const validTok = await client.verifyForm({ secret: 'secret-key-challenge', response: 'valid-token' });
            expect(validTok.data).to.include({ success: true, score: 0.8, action: 'submit' });

            // low-score-token
            const lowScore = await client.verifyForm({ secret: 'secret-key-1', response: 'low-score-token' });
            expect(lowScore.data).to.include({ success: true, score: 0.1, action: 'submit' });

            // high-score-token
            const highScore = await client.verifyForm({ secret: 'secret-key-1', response: 'high-score-token' });
            expect(highScore.data).to.include({ success: true, score: 1.0, action: 'submit' });

            // always-fail-token
            const alwaysFail = await client.verifyForm({ secret: 'secret-key-1', response: 'always-fail-token' });
            expect(alwaysFail.data.success).to.equal(false);
            expect(alwaysFail.data['error-codes']).to.deep.equal(['invalid-input-response']);

            // invalid-token
            const invalidTok = await client.verifyForm({ secret: 'secret-key-1', response: 'invalid-token' });
            expect(invalidTok.data.success).to.equal(false);
            expect(invalidTok.data['error-codes']).to.deep.equal(['invalid-input-response']);

            // expired-token
            const expiredTok = await client.verifyForm({ secret: 'secret-key-1', response: 'expired-token' });
            expect(expiredTok.data.success).to.equal(false);
            expect(expiredTok.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);

            // duplicate-token
            const dupTok = await client.verifyForm({ secret: 'secret-key-1', response: 'duplicate-token' });
            expect(dupTok.data.success).to.equal(false);
            expect(dupTok.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);

            // bad-request-token
            const badReqTok = await client.verifyForm({ secret: 'secret-key-1', response: 'bad-request-token' });
            expect(badReqTok.data.success).to.equal(false);
            expect(badReqTok.data['error-codes']).to.deep.equal(['bad-request']);
        });

        it('Step 4: Custom magic_tokens from YAML config should be evaluated accurately', async () => {
            const customPass = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'custom-magic-pass',
            });
            expect(customPass.status).to.equal(200);
            expect(customPass.data).to.include({
                success: true,
                score: 0.75,
                action: 'custom_action',
                hostname: 'magic.example.local',
            });

            const customFail = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'custom-magic-fail',
            });
            expect(customFail.status).to.equal(200);
            expect(customFail.data.success).to.equal(false);
            expect(customFail.data['error-codes']).to.deep.equal(['custom-magic-error']);
        });

        it('Step 5: Non-existent token should return invalid-input-response', async () => {
            const res = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'tok_does_not_exist_99999',
            });
            expect(res.status).to.equal(200);
            expect(res.data.success).to.equal(false);
            expect(res.data['error-codes']).to.deep.equal(['invalid-input-response']);
        });

        it('Step 5: Site-bound token verified with wrong secret should fail with invalid-input-secret, unless Google test secret is used', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-1' });
            const wrongSecretRes = await client.verifyForm({
                secret: 'wrong-secret',
                response: solved.data.token,
            });
            expect(wrongSecretRes.status).to.equal(200);
            expect(wrongSecretRes.data.success).to.equal(false);
            expect(wrongSecretRes.data['error-codes']).to.deep.equal(['invalid-input-secret']);

            // Token should still be usable with Google's official test secret
            const googleSecretRes = await client.verifyForm({
                secret: '6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe',
                response: solved.data.token,
            });
            expect(googleSecretRes.status).to.equal(200);
            expect(googleSecretRes.data.success).to.equal(true);
        });

        it('Step 5: verify_once=true should reject second verification of same token with timeout-or-duplicate', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-1' });
            const first = await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
            });
            expect(first.data.success).to.equal(true);

            const second = await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
            });
            expect(second.data.success).to.equal(false);
            expect(second.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);
        });

        it('Step 5: verify_once=false should allow verifying the same token multiple times', async () => {
            await client.updateTestingConfig({ verify_once: false });
            const solved = await client.solveJson({ sitekey: 'site-key-1' });

            const first = await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
            });
            expect(first.data.success).to.equal(true);

            const second = await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
            });
            expect(second.data.success).to.equal(true);
        });
    });

    describe('Testing & Inspection API (/testing/...)', () => {
        it('POST /testing/tokens should pre-seed custom tokens (201 Created) and support error_codes, expired, and used flags', async () => {
            // 1. Custom valid token
            const seed1 = await client.createTestingToken({
                token: 'seeded-valid-1',
                sitekey: 'site-key-1',
                action: 'payment',
                score: 0.33,
                hostname: 'pay.example.local',
                ttl_seconds: 600,
            });
            expect(seed1.status).to.equal(201);
            expect(seed1.data).to.include({
                token: 'seeded-valid-1',
                sitekey: 'site-key-1',
                action: 'payment',
                score: 0.33,
                hostname: 'pay.example.local',
                used: false,
                expired: false,
            });

            const verify1 = await client.verifyForm({
                secret: 'secret-key-1',
                response: 'seeded-valid-1',
            });
            expect(verify1.data).to.include({
                success: true,
                action: 'payment',
                score: 0.33,
                hostname: 'pay.example.local',
            });

            // 2. Pre-seeded expired token
            const seedExpired = await client.createTestingToken({
                token: 'seeded-expired-1',
                expired: true,
            });
            expect(seedExpired.status).to.equal(201);
            expect(seedExpired.data.expired).to.equal(true);

            const verifyExpired = await client.verifyForm({
                secret: 'any-secret',
                response: 'seeded-expired-1',
            });
            expect(verifyExpired.data.success).to.equal(false);
            expect(verifyExpired.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);

            // 3. Pre-seeded used token
            const seedUsed = await client.createTestingToken({
                token: 'seeded-used-1',
                used: true,
            });
            expect(seedUsed.status).to.equal(201);
            expect(seedUsed.data.used).to.equal(true);

            const verifyUsed = await client.verifyForm({
                secret: 'any-secret',
                response: 'seeded-used-1',
            });
            expect(verifyUsed.data.success).to.equal(false);
            expect(verifyUsed.data['error-codes']).to.deep.equal(['timeout-or-duplicate']);

            // 4. Pre-seeded token with error_codes
            const seedErrors = await client.createTestingToken({
                token: 'seeded-error-1',
                error_codes: ['seeded-err-code'],
            });
            expect(seedErrors.status).to.equal(201);

            const verifyErrors = await client.verifyForm({
                secret: 'any-secret',
                response: 'seeded-error-1',
            });
            expect(verifyErrors.data.success).to.equal(false);
            expect(verifyErrors.data['error-codes']).to.deep.equal(['seeded-err-code']);
        });

        it('POST /testing/tokens with invalid JSON should return 400 Bad Request', async () => {
            const res = await fetch(`${baseUrl}/testing/tokens`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: '{invalid-json',
            });
            expect(res.status).to.equal(400);
        });

        it('GET /testing/tokens should list tokens and filter by sitekey and used', async () => {
            const t1 = await client.solveJson({ sitekey: 'site-key-1' });
            const t2 = await client.solveJson({ sitekey: 'site-key-bot' });
            await client.verifyForm({ secret: 'secret-key-1', response: t1.data.token });

            const allTokens = await client.listTestingTokens();
            expect(allTokens.status).to.equal(200);
            expect(allTokens.data.tokens).to.be.an('array').with.lengthOf(2);

            const site1Tokens = await client.listTestingTokens({ sitekey: 'site-key-1' });
            expect(site1Tokens.data.tokens).to.have.lengthOf(1);
            expect(site1Tokens.data.tokens[0].token).to.equal(t1.data.token);

            const usedTokens = await client.listTestingTokens({ used: 'true' });
            expect(usedTokens.data.tokens).to.have.lengthOf(1);
            expect(usedTokens.data.tokens[0].token).to.equal(t1.data.token);

            const unusedTokens = await client.listTestingTokens({ used: 'false' });
            expect(unusedTokens.data.tokens).to.have.lengthOf(1);
            expect(unusedTokens.data.tokens[0].token).to.equal(t2.data.token);
        });

        it('GET /testing/verifications should record all attempts and filter by response and success', async () => {
            const solved = await client.solveJson({ sitekey: 'site-key-1', action: 'transfer' });
            await client.verifyForm({
                secret: 'secret-key-1',
                response: solved.data.token,
                remoteip: '10.0.0.1',
            });
            await client.verifyForm({
                secret: 'secret-key-1',
                response: 'always-fail-token',
                remoteip: '10.0.0.2',
            });

            const all = await client.listTestingVerifications();
            expect(all.status).to.equal(200);
            expect(all.data.verifications).to.be.an('array').with.lengthOf(2);
            expect(all.data.verifications[0]).to.include({
                secret: 'secret-key-1',
                response: solved.data.token,
                remoteip: '10.0.0.1',
                success: true,
                score: 0.9,
                action: 'transfer',
                hostname: 'app.example.local',
            });
            expect(all.data.verifications[0].timestamp).to.be.a('string');

            const byResp = await client.listTestingVerifications({ response: 'always-fail-token' });
            expect(byResp.data.verifications).to.have.lengthOf(1);
            expect(byResp.data.verifications[0].success).to.equal(false);
            expect(byResp.data.verifications[0].error_codes).to.include('invalid-input-response');

            const bySuccess = await client.listTestingVerifications({ success: 'true' });
            expect(bySuccess.data.verifications).to.have.lengthOf(1);
            expect(bySuccess.data.verifications[0].response).to.equal(solved.data.token);

            const byFailure = await client.listTestingVerifications({ success: 'false' });
            expect(byFailure.data.verifications).to.have.lengthOf(1);
            expect(byFailure.data.verifications[0].response).to.equal('always-fail-token');
        });

        it('GET, PUT, and POST /testing/config should inspect and dynamically mutate runtime configuration', async () => {
            const initial = await client.getTestingConfig();
            expect(initial.status).to.equal(200);
            expect(initial.data).to.include({
                mode: 'auto_pass',
                default_score: 0.9,
                token_ttl_seconds: 300,
                verify_once: true,
                allow_any_site: true,
                allow_any_secret: true,
                hostname: 'localhost',
            });
            expect(initial.data.sites).to.be.an('array').with.lengthOf(4);

            // Update via PUT
            const updatedPut = await client.updateTestingConfig({
                mode: 'challenge',
                default_score: 0.55,
                hostname: 'dynamic.example.local',
                token_ttl_seconds: 120,
            }, 'PUT');
            expect(updatedPut.status).to.equal(200);
            expect(updatedPut.data).to.include({
                mode: 'challenge',
                default_score: 0.55,
                hostname: 'dynamic.example.local',
                token_ttl_seconds: 120,
            });

            // Update via POST
            const updatedPost = await client.updateTestingConfig({
                sites: [
                    {
                        sitekey: 'new-dynamic-site',
                        secret: 'new-dynamic-secret',
                        mode: 'auto_pass',
                        default_score: 0.77,
                    },
                ],
            }, 'POST');
            expect(updatedPost.status).to.equal(200);
            expect(updatedPost.data.sites).to.have.lengthOf(1);
            expect(updatedPost.data.sites[0].sitekey).to.equal('new-dynamic-site');

            // Invalid mode should return 400 Bad Request
            const badMode = await client.updateTestingConfig({
                mode: 'invalid_mode_name',
            });
            expect(badMode.status).to.equal(400);
        });

        it('POST /testing/reset should clear tokens, verifications, and restore initial config', async () => {
            await client.solveJson({ sitekey: 'site-key-1' });
            await client.verifyForm({ secret: 'secret-key-1', response: 'always-pass-token' });
            await client.updateTestingConfig({ mode: 'always_fail', default_score: 0.12 });

            const resetRes = await client.resetTestingState();
            expect(resetRes.status).to.equal(200);
            expect(resetRes.data).to.include({
                status: 'OK',
                message: 'State reset',
            });

            const tokensAfter = await client.listTestingTokens();
            expect(tokensAfter.data.tokens).to.be.empty;

            const verificationsAfter = await client.listTestingVerifications();
            expect(verificationsAfter.data.verifications).to.be.empty;

            const configAfter = await client.getTestingConfig();
            expect(configAfter.data.mode).to.equal('auto_pass');
            expect(configAfter.data.default_score).to.equal(0.9);
            expect(configAfter.data.sites).to.have.lengthOf(4);
        });
    });
});
