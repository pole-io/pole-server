const baseURL = process.env.POLE_BASE_URL || 'http://127.0.0.1:8080';
const user = process.env.POLE_USER || 'admin';
const password = process.env.POLE_PASSWORD || 'admin123';
const keepFixture = process.env.POLE_KEEP_FIXTURE === '1';

const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(4, 14);
const namespace = 'default';
const group = `codex-agent-${stamp}`;
const fileName = 'app.yaml';
const baselineContent = 'feature:\n  enabled: false\n';
const desiredContent = 'feature:\n  enabled: true\n';
let authHeaders = {};
let cleanupFile = false;
let cleanupGroup = false;

async function request(method, path, body) {
  const response = await fetch(`${baseURL}${path}`, {
    method,
    headers: {
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      ...authHeaders,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok || payload.code !== 200000) {
    throw new Error(`${method} ${path} failed: http=${response.status} body=${text}`);
  }
  return payload;
}

function assertBatchSuccess(label, payload) {
  const failures = (payload.responses || []).filter((item) => item.code !== 200000);
  if (failures.length > 0) {
    throw new Error(`${label} failed: ${JSON.stringify(failures)}`);
  }
}

async function cleanup() {
  if (keepFixture) return;
  if (cleanupFile) {
    await request('POST', '/config/v1/files/delete', [{ namespace, group, name: fileName }]).catch(() => {});
  }
  if (cleanupGroup) {
    await request('POST', '/config/v1/groups/delette', [{ namespace, name: group }]).catch(() => {});
  }
}

async function main() {
  const login = await request('POST', '/auth/v1/user/login', { name: user, password });
  authHeaders = {
    Authorization: login.data.token,
    'X-Pole-User': login.data.user_id,
  };

  const createGroup = await request('POST', '/config/v1/groups', [{ namespace, name: group, comment: 'Agent smoke fixture' }]);
  assertBatchSuccess('create group', createGroup);
  cleanupGroup = true;

  const createFile = await request('POST', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: baselineContent,
    format: 'yaml',
    comment: 'baseline',
    labels: { source: 'agent-smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);
  assertBatchSuccess('create file', createFile);
  cleanupFile = true;

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: `baseline-${stamp}`,
    release_description: 'Agent smoke baseline',
    release_type: 'normal',
  });

  const activeBefore = await request('GET', `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${fileName}`);
  if (activeBefore.data.content !== baselineContent) throw new Error('baseline release content mismatch');

  const proposal = await request('POST', '/ai/agent/v1/proposals/config-file', {
    namespace,
    group,
    name: fileName,
    desiredContent,
    comment: 'prepared by Agent smoke',
  });
  if (proposal.data.status !== 'preview_ready') throw new Error(`unexpected proposal status: ${proposal.data.status}`);

  const draftBeforeConfirm = await request('GET', `/config/v1/files/detail?namespace=${namespace}&group=${group}&name=${fileName}`);
  const activeBeforeConfirm = await request('GET', `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${fileName}`);
  if (draftBeforeConfirm.data.content !== baselineContent || activeBeforeConfirm.data.content !== baselineContent) {
    throw new Error('preview must not mutate draft or active release');
  }

  const idempotencyKey = `agent-smoke-${stamp}`;
  const confirmation = {
    proposalVersion: proposal.data.version,
    previewHash: proposal.data.previewHash,
    idempotencyKey,
  };
  const receipt = await request('POST', `/ai/agent/v1/proposals/${proposal.data.id}/confirm`, confirmation);
  const repeatedReceipt = await request('POST', `/ai/agent/v1/proposals/${proposal.data.id}/confirm`, confirmation);
  if (receipt.data.status !== 'waiting_for_publish' || repeatedReceipt.data.proposalId !== receipt.data.proposalId) {
    throw new Error('confirmation must be idempotent and wait for publish');
  }

  const draftAfterConfirm = await request('GET', `/config/v1/files/detail?namespace=${namespace}&group=${group}&name=${fileName}`);
  const activeAfterConfirm = await request('GET', `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${fileName}`);
  if (draftAfterConfirm.data.content !== desiredContent) throw new Error('confirmation did not save desired draft');
  if (activeAfterConfirm.data.content !== baselineContent) throw new Error('Agent confirmation must not publish');

  console.log(JSON.stringify({
    ok: true,
    namespace,
    group,
    fileName,
    proposalStatus: proposal.data.status,
    receiptStatus: receipt.data.status,
    draftChanged: true,
    activeReleaseUnchanged: true,
    keepFixture,
  }));
}

main().finally(cleanup).catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
