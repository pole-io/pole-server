import React from 'react';
import { Tag } from 'components/Fluent';
import CodeEditor from 'components/CodeEditor';
import { RenderPreview } from 'services/config_templates';
import styles from './index.module.less';

interface RenderPreviewPanelProps {
  preview?: RenderPreview;
  title?: string;
  emptyMessage: string;
  compact?: boolean;
}

const RenderPreviewPanel: React.FC<RenderPreviewPanelProps> = ({
  preview,
  title = '格式化渲染结果',
  emptyMessage,
  compact = false,
}) => (
  <section className={`${styles.versionPreview} ${compact ? styles.valuePreview : ''}`}>
    <div className={styles.versionPreviewOutput}>
      <div className={styles.versionHistoryHeading}>
        <div>
          <strong>{title}</strong>
          <span>
            {preview ? `${String(preview.format).toUpperCase()} · ${preview.renderedSha256}` : emptyMessage}
          </span>
        </div>
      </div>
      {preview ? (
        <div className={styles.versionPreviewEditor}>
          <CodeEditor
            readonly
            allowFullScreen
            height="100%"
            language={preview.format}
            value={preview.renderedContent}
          />
        </div>
      ) : (
        <div className={styles.versionPreviewEmpty}>{emptyMessage}</div>
      )}
    </div>
    <aside className={styles.versionDiagnostics}>
      <div className={styles.versionHistoryHeading}>
        <div>
          <strong>渲染诊断</strong>
          <span>{preview?.diagnostics.length || 0} 项</span>
        </div>
      </div>
      {(preview?.diagnostics || []).map((diagnostic, index) => (
        <div className={styles.diagnosticItem} key={`${diagnostic.code}-${index}`}>
          <Tag theme={diagnostic.severity === 'DIAGNOSTIC_ERROR' ? 'danger' : 'warning'}>{diagnostic.code}</Tag>
          <p>{diagnostic.message}</p>
          {diagnostic.parameter && <code>{diagnostic.parameter}</code>}
        </div>
      ))}
      {!preview?.diagnostics.length && <span className={styles.muted}>暂无 diagnostics</span>}
    </aside>
  </section>
);

export default React.memo(RenderPreviewPanel);
