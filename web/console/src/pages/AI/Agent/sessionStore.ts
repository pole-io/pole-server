export type AgentMessageRole = 'agent' | 'user';
export type AgentToolStatus = 'running' | 'success' | 'failed';

export interface AgentToolTrace {
  name: string;
  summary: string;
  detail?: string;
  status: AgentToolStatus;
}

export interface AgentChatMessage {
  id: string;
  role: AgentMessageRole;
  title?: string;
  content: string;
  createdAt: number;
  tool?: AgentToolTrace;
}

export interface AgentResourceContext {
  kind: 'config.file';
  namespace: string;
  group: string;
  name: string;
  path: string;
}

export interface AgentMemorySettings {
  enabled: boolean;
  maxTurns: 4 | 10 | 20;
}

export interface AgentLocalSession {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  messages: AgentChatMessage[];
  draft: string;
  memory: AgentMemorySettings;
  namespaceScope?: AgentNamespaceScope;
  resourceContext?: AgentResourceContext;
  proposal?: ConfigFileProposal;
  receipt?: DraftReceipt;
}

const databaseName = 'pole-agent-workbench';
const databaseVersion = 1;
const sessionStoreName = 'sessions';
const metadataStoreName = 'metadata';
const activeSessionKey = 'active-session';

let databasePromise: Promise<IDBDatabase> | undefined;

const requestValue = <T>(request: IDBRequest<T>) => new Promise<T>((resolve, reject) => {
  request.onsuccess = () => resolve(request.result);
  request.onerror = () => reject(request.error || new Error('IndexedDB request failed'));
});

const transactionDone = (transaction: IDBTransaction) => new Promise<void>((resolve, reject) => {
  transaction.oncomplete = () => resolve();
  transaction.onerror = () => reject(transaction.error || new Error('IndexedDB transaction failed'));
  transaction.onabort = () => reject(transaction.error || new Error('IndexedDB transaction aborted'));
});

const openDatabase = () => {
  if (!databasePromise) {
    databasePromise = new Promise((resolve, reject) => {
      const request = indexedDB.open(databaseName, databaseVersion);
      request.onupgradeneeded = () => {
        const database = request.result;
        if (!database.objectStoreNames.contains(sessionStoreName)) {
          const sessions = database.createObjectStore(sessionStoreName, { keyPath: 'id' });
          sessions.createIndex('updatedAt', 'updatedAt');
        }
        if (!database.objectStoreNames.contains(metadataStoreName)) {
          database.createObjectStore(metadataStoreName, { keyPath: 'key' });
        }
      };
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error || new Error('Unable to open IndexedDB'));
    });
  }
  return databasePromise;
};

export async function listAgentSessions() {
  const database = await openDatabase();
  const transaction = database.transaction(sessionStoreName, 'readonly');
  const done = transactionDone(transaction);
  const sessions = await requestValue(transaction.objectStore(sessionStoreName).getAll() as IDBRequest<AgentLocalSession[]>);
  await done;
  return sessions.sort((left, right) => right.updatedAt - left.updatedAt);
}

export async function saveAgentSession(session: AgentLocalSession) {
  const database = await openDatabase();
  const transaction = database.transaction(sessionStoreName, 'readwrite');
  const done = transactionDone(transaction);
  transaction.objectStore(sessionStoreName).put(session);
  await done;
}

export async function deleteAgentSession(sessionID: string) {
  const database = await openDatabase();
  const transaction = database.transaction(sessionStoreName, 'readwrite');
  const done = transactionDone(transaction);
  transaction.objectStore(sessionStoreName).delete(sessionID);
  await done;
}

export async function getActiveAgentSessionID() {
  const database = await openDatabase();
  const transaction = database.transaction(metadataStoreName, 'readonly');
  const done = transactionDone(transaction);
  const metadata = await requestValue(transaction.objectStore(metadataStoreName).get(activeSessionKey) as IDBRequest<{ key: string; value: string } | undefined>);
  await done;
  return metadata?.value;
}

export async function saveActiveAgentSessionID(sessionID: string) {
  const database = await openDatabase();
  const transaction = database.transaction(metadataStoreName, 'readwrite');
  const done = transactionDone(transaction);
  transaction.objectStore(metadataStoreName).put({ key: activeSessionKey, value: sessionID });
  await done;
}
import type { AgentNamespaceScope, ConfigFileProposal, DraftReceipt } from 'services/agent';
