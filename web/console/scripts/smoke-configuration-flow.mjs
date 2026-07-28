const baseURL = process.env.POLE_BASE_URL || 'http://127.0.0.1:8080';
const user = process.env.POLE_USER || 'admin';
const password = process.env.POLE_PASSWORD || 'admin123';

const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(4, 14);
const namespace = 'default';
const group = `codex-flow-${stamp}`;
const fileName = 'app.yaml';
const releaseV1 = `v1-${stamp}`;
const releaseV2 = `v2-${stamp}`;
const releaseV3 = `v3-${stamp}`;
const grayA = `gray-a-${stamp}`;
const grayB = `gray-b-${stamp}`;
const promotedRelease = `promoted-${stamp}`;

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
  let payload;
  try {
    payload = text ? JSON.parse(text) : {};
  } catch (error) {
    throw new Error(`${method} ${path} returned non-json ${response.status}: ${text}`);
  }
  if (!response.ok || payload.code !== 200000) {
    throw new Error(`${method} ${path} failed: http=${response.status} body=${text}`);
  }
  return payload;
}

function assertBatchSuccess(label, payload) {
  const failures = (payload.responses || []).filter((item) => item.code !== 200000);
  if (failures.length > 0) {
    throw new Error(`${label} failed responses: ${JSON.stringify(failures)}`);
  }
}

function assertAmount(label, payload, expected) {
  if (payload.amount !== expected) {
    throw new Error(`${label} expected amount=${expected}, got ${payload.amount}: ${JSON.stringify(payload)}`);
  }
}

function exactClientLabel(key, value) {
  return [{
    key,
    value: {
      type: 0,
      value,
      value_type: 0,
    },
  }];
}

function releaseItems(payload) {
  return payload.data || payload.configFileReleases || [];
}

async function login() {
  const payload = await request('POST', '/auth/v1/user/login', { name: user, password });
  authHeaders = {
    Authorization: payload.data.token,
    'X-Pole-User': payload.data.user_id,
  };
}

async function cleanup() {
  if (cleanupFile) {
    await request('POST', '/config/v1/files/delete', [{ namespace, group, name: fileName }]).catch((error) => {
      console.warn(`[cleanup] delete file ignored: ${error.message}`);
    });
  }
  if (cleanupGroup) {
    await request('POST', '/config/v1/groups/delette', [{ namespace, name: group }]).catch((error) => {
      console.warn(`[cleanup] delete group ignored: ${error.message}`);
    });
  }
}

async function main() {
  await login();

  const groupResp = await request('POST', '/config/v1/groups', [{
    namespace,
    name: group,
    comment: 'codex configuration flow smoke',
    metadata: {},
  }]);
  assertBatchSuccess('create group', groupResp);
  cleanupGroup = true;

  const createFileResp = await request('POST', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: 1\n',
    format: 'yaml',
    comment: 'initial config file',
    labels: {},
    encrypted: false,
    encryptAlgo: '',
  }]);
  assertBatchSuccess('create file', createFileResp);
  cleanupFile = true;

  const detailV1 = await request(
    'GET',
    `/config/v1/files/detail?namespace=${namespace}&group=${group}&name=${encodeURIComponent(fileName)}`,
  );
  if (detailV1.data.name !== fileName || detailV1.data.content !== 'a: 1\n') {
    throw new Error(`unexpected file detail after create: ${JSON.stringify(detailV1)}`);
  }

  const updateFileResp = await request('PUT', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: 2\n',
    format: 'yaml',
    comment: 'updated config file',
    labels: { source: 'smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);
  assertBatchSuccess('update file', updateFileResp);

  const detailV2 = await request(
    'GET',
    `/config/v1/files/detail?namespace=${namespace}&group=${group}&name=${encodeURIComponent(fileName)}`,
  );
  if (detailV2.data.content !== 'a: 2\n' || detailV2.data.labels.source !== 'smoke') {
    throw new Error(`unexpected file detail after update: ${JSON.stringify(detailV2)}`);
  }

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: releaseV1,
    release_description: 'release v1',
    release_type: 'normal',
  });

  await request('PUT', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: 3\n',
    format: 'yaml',
    comment: 'second update',
    labels: { source: 'smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: releaseV2,
    release_description: 'release v2',
    release_type: 'normal',
  });

  await request('PUT', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: gray-a\n',
    format: 'yaml',
    comment: 'gray update a',
    labels: { source: 'smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: grayA,
    release_description: 'gray release a',
    release_type: 'gray',
    beta_labels: exactClientLabel('env', 'gray-a'),
  });

  await request('PUT', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: gray-b\n',
    format: 'yaml',
    comment: 'gray update b',
    labels: { source: 'smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: grayB,
    release_description: 'gray release b',
    release_type: 'gray',
    beta_labels: exactClientLabel('env', 'gray-b'),
  });

  await request('PUT', '/config/v1/files', [{
    namespace,
    group,
    name: fileName,
    content: 'a: 4\n',
    format: 'yaml',
    comment: 'normal update while gray active',
    labels: { source: 'smoke' },
    encrypted: false,
    encryptAlgo: '',
  }]);

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: releaseV3,
    release_description: 'release v3 while gray active',
    release_type: 'normal',
  });

  const activeRelease = await request(
    'GET',
    `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );
  if (activeRelease.data.name !== releaseV3 || activeRelease.data.content !== 'a: 4\n') {
    throw new Error(`unexpected active release: ${JSON.stringify(activeRelease)}`);
  }

  const releases = await request(
    'GET',
    `/config/v1/files/releases?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}&offset=0&limit=10`,
  );
  assertAmount('release list', releases, 5);
  const activeGray = releaseItems(releases).filter((item) => item.release_type === 'gray' && item.active);
  if (activeGray.length !== 2) {
    throw new Error(`expected two active gray releases, got ${activeGray.length}: ${JSON.stringify(releases)}`);
  }

  const versions = await request(
    'GET',
    `/config/v1/files/release/versions?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );
  assertAmount('release versions', versions, 5);

  await request(
    'GET',
    `/config/v1/files/subscribers?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );

  await request('POST', '/config/v1/files/releases/promote-gray', {
    namespace,
    group,
    file_name: fileName,
    name: grayA,
    release_type: 'gray',
  });

  const draftAfterPromote = await request(
    'GET',
    `/config/v1/files/detail?namespace=${namespace}&group=${group}&name=${encodeURIComponent(fileName)}`,
  );
  if (draftAfterPromote.data.content !== 'a: gray-a\n') {
    throw new Error(`unexpected draft after promote gray: ${JSON.stringify(draftAfterPromote)}`);
  }

  const activeAfterPromote = await request(
    'GET',
    `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );
  if (activeAfterPromote.data.name !== releaseV3) {
    throw new Error(`promote gray must not publish normal release automatically: ${JSON.stringify(activeAfterPromote)}`);
  }

  await request('POST', '/config/v1/files/release', {
    namespace,
    group,
    file_name: fileName,
    name: promotedRelease,
    release_description: 'promoted gray as normal',
    release_type: 'normal',
  });

  const promotedActive = await request(
    'GET',
    `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );
  if (promotedActive.data.name !== promotedRelease || promotedActive.data.content !== 'a: gray-a\n') {
    throw new Error(`unexpected promoted normal release: ${JSON.stringify(promotedActive)}`);
  }

  const stopGrayResp = await request('POST', '/config/v1/files/releases/stopbeta', [{
    namespace,
    group,
    file_name: fileName,
    name: grayA,
    release_type: 'gray',
  }]);
  assertBatchSuccess('stop one gray release', stopGrayResp);

  const releasesAfterStopGray = await request(
    'GET',
    `/config/v1/files/releases?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}&offset=0&limit=10`,
  );
  const stoppedGray = releaseItems(releasesAfterStopGray).find((item) => item.name === grayA && item.release_type === 'gray');
  const activeGrayB = releaseItems(releasesAfterStopGray).find((item) => item.name === grayB && item.release_type === 'gray');
  if (!stoppedGray || stoppedGray.active || !activeGrayB || !activeGrayB.active) {
    throw new Error(`unexpected gray status after stopping one gray: ${JSON.stringify(releasesAfterStopGray)}`);
  }

  const rollbackResp = await request('PUT', '/config/v1/files/releases/rollback', [{
    namespace,
    group,
    file_name: fileName,
    name: releaseV1,
  }]);
  assertBatchSuccess('rollback release', rollbackResp);

  const rolledBackRelease = await request(
    'GET',
    `/config/v1/files/release?namespace=${namespace}&group=${group}&file_name=${encodeURIComponent(fileName)}`,
  );
  if (rolledBackRelease.data.name !== releaseV1 || rolledBackRelease.data.content !== 'a: 2\n') {
    throw new Error(`unexpected release after rollback: ${JSON.stringify(rolledBackRelease)}`);
  }

  const deleteReleaseResp = await request('POST', '/config/v1/files/releases/delete', [{
    namespace,
    group,
    file_name: fileName,
    name: releaseV2,
    release_type: 'normal',
  }]);
  assertBatchSuccess('delete inactive release', deleteReleaseResp);

  const deleteFileResp = await request('POST', '/config/v1/files/delete', [{ namespace, group, name: fileName }]);
  assertBatchSuccess('delete file', deleteFileResp);
  cleanupFile = false;

  const filesAfterDelete = await request(
    'GET',
    `/config/v1/files/search?namespace=${namespace}&group=${group}&name=${encodeURIComponent(fileName)}&offset=0&limit=10`,
  );
  assertAmount('files after delete', filesAfterDelete, 0);

  const deleteGroupResp = await request('POST', '/config/v1/groups/delette', [{ namespace, name: group }]);
  assertBatchSuccess('delete group', deleteGroupResp);
  cleanupGroup = false;

  const groupsAfterDelete = await request(
    'GET',
    `/config/v1/groups?namespace=${namespace}&name=${encodeURIComponent(group)}&offset=0&limit=10`,
  );
  assertAmount('groups after delete', groupsAfterDelete, 0);

  console.log(`configuration flow smoke passed: ${namespace}/${group}/${fileName}`);
}

main()
  .catch(async (error) => {
    console.error(error);
    await cleanup();
    process.exit(1);
  });
