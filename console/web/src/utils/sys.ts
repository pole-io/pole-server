import { MessagePlugin } from 'components/Fluent';

export async function copyToClipboard(text: string, successMessage = '文本复制成功') {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const copied = copyWithLegacyCommand(text);
      if (!copied) throw new Error('浏览器拒绝了剪贴板写入');
    }
    MessagePlugin.success(successMessage);
    return true;
  } catch {
    MessagePlugin.error('复制失败，请检查浏览器剪贴板权限后重试');
    return false;
  }
}

function copyWithLegacyCommand(text: string) {
    const textarea = document.createElement('textarea');
    textarea.style.position = 'fixed';
    textarea.style.clip = 'rect(0 0 0 0)';
    textarea.style.top = '10px';
    textarea.value = text;
    document.body.appendChild(textarea);
    try {
      textarea.select();
      return document.execCommand('copy');
    } finally {
      document.body.removeChild(textarea);
    }
  }
