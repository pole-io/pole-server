import { createElement } from 'react';
import { Button } from 'components/Fluent';
import { CopyIcon } from 'components/Fluent/icons';
import i18n from '../i18n';
import { showFluentToast } from 'components/Fluent/toast';
import { RequestErrorPayload } from './request';
import { copyToClipboard } from './sys';

const requestIdPattern = /[，,]\s*RequestId:\s*([^\s，,]+)/i;
const requestCodePattern = /请求失败（(\d+)）：\s*/;

type ErrorNotification = RequestErrorPayload | string | Error | null | undefined;

const isRequestErrorPayload = (value: unknown): value is RequestErrorPayload => (
    Boolean(value)
    && typeof value === 'object'
    && typeof (value as RequestErrorPayload).message === 'string'
    && ('code' in (value as Record<string, unknown>) || 'requestId' in (value as Record<string, unknown>))
);

export const splitRequestId = (description: string) => {
    const match = description.match(requestIdPattern);
    if (!match) {
        return { body: description, requestId: undefined };
    }
    return {
        body: description.replace(requestIdPattern, '').trim(),
        requestId: match[1],
    };
};

const normalizeError = (description: ErrorNotification): RequestErrorPayload | undefined => {
    if (isRequestErrorPayload(description)) return description;
    const text = description instanceof Error ? description.message : typeof description === 'string' ? description : '';
    const { body, requestId } = splitRequestId(text);
    const codeMatch = body.match(requestCodePattern);
    if (!codeMatch && !requestId) return undefined;
    return {
        code: codeMatch ? Number(codeMatch[1]) : undefined,
        message: body.replace(requestCodePattern, '').trim() || i18n.t('requestError.fallbackMessage'),
        requestId,
    };
};

const localizedMessage = ({ code, message }: RequestErrorPayload) => {
    const key = code === undefined ? '' : `requestError.code.${code}`;
    return key && i18n.exists(key) ? i18n.t(key) : message;
};

const errorDetail = ({ code, message, requestId }: RequestErrorPayload) => createElement(
    'div',
    { style: { display: 'grid', gap: '6px' } },
    createElement('div', null,
        createElement('strong', null, `${i18n.t('requestError.codeLabel')}：`),
        createElement('code', null, String(code ?? '-')),
    ),
    createElement('div', null,
        createElement('strong', null, `${i18n.t('requestError.messageLabel')}：`),
        message,
    ),
    requestId ? createElement('div', null,
        createElement('strong', null, `${i18n.t('requestError.requestIdLabel')}：`),
        createElement('code', null, requestId),
    ) : null,
);

export const openErrNotification = (title: string, description: ErrorNotification) => {
    const structured = normalizeError(description);
    if (!structured) {
        showFluentToast({ title, body: description instanceof Error ? description.message : description || i18n.t('requestError.fallbackMessage'), intent: 'error', timeout: 10000 });
        return;
    }
    const message = localizedMessage(structured);
    const payload: RequestErrorPayload = { ...structured, message };
    showFluentToast({
        title,
        body: errorDetail(payload),
        action: createElement(
            Button,
            {
                size: 'small',
                variant: 'text',
                'aria-label': i18n.t('requestError.copy'),
                icon: createElement(CopyIcon),
                onClick: () => copyToClipboard(JSON.stringify(payload, null, 2), i18n.t('requestError.errorJsonCopied')),
            },
            i18n.t('requestError.copy'),
        ),
        intent: 'error',
        timeout: 10000,
    });
};

export const openInfoNotification = (title: string, description: string) => {
    showFluentToast({ title, body: description, intent: 'success', timeout: 10000 });
};
