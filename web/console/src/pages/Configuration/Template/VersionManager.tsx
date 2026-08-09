import React from 'react';
import { Button, PrimaryTableProps, Table, Tag } from 'components/Fluent';
import {
  ConfigFileTemplate,
  ConfigTemplateRelease,
  ConfigTemplateValue,
  NamespaceTemplateValueRelease,
  RenderPreview,
} from 'services/config_templates';
import RenderPreviewPanel from './RenderPreviewPanel';
import CodeDiffEditor from 'components/CodeDiffEditor';
import styles from './index.module.less';

interface VersionManagerProps {
  namespace: string;
  templateDraft: ConfigFileTemplate;
  valueDraft: Record<string, ConfigTemplateValue>;
  templateSnapshots: ConfigTemplateRelease[];
  releases: NamespaceTemplateValueRelease[];
  releaseIds: string[];
  preview?: RenderPreview;
  comparison?: {
    before: RenderPreview;
    after: RenderPreview;
    beforeLabel: string;
    afterLabel: string;
  };
  onReleaseChange: (value: string) => void;
  onPreview: () => void;
}

export const ENVIRONMENT_DRAFT_ROW_ID = '__environment_draft__';

type EnvironmentVersionRow = NamespaceTemplateValueRelease & {
  selectionId: string;
  draftRow?: boolean;
};

const VersionManager: React.FC<VersionManagerProps> = ({
  namespace,
  templateDraft,
  valueDraft,
  templateSnapshots,
  releases,
  releaseIds,
  preview,
  comparison,
  onReleaseChange,
  onPreview,
}) => {
  const rows = React.useMemo<EnvironmentVersionRow[]>(
    () => [
      {
        id: ENVIRONMENT_DRAFT_ROW_ID,
        valuesId: '',
        namespace,
        templateId: templateDraft.id,
        templateReleaseId: '',
        values: valueDraft,
        releaseType: 'TEMPLATE_VALUE_RELEASE_NORMAL',
        version: '草稿',
        active: false,
        selectionId: ENVIRONMENT_DRAFT_ROW_ID,
        draftRow: true,
      },
      ...releases.map((release) => ({ ...release, selectionId: release.id })),
    ],
    [namespace, releases, templateDraft.id, valueDraft]
  );
  const templateSnapshotById = React.useMemo(
    () => new Map(templateSnapshots.map((release) => [release.id, release])),
    [templateSnapshots]
  );

  const columns: PrimaryTableProps['columns'] = [
    {
      colKey: 'selection',
      title: '选择',
      width: 54,
      ellipsis: false,
      cell: ({ row }) => {
        const selected = releaseIds.includes(row.selectionId);
        return (
          <span
            aria-label={selected ? '已选择' : '未选择'}
            className={selected ? styles.versionSelectMarkActive : styles.versionSelectMark}
            data-version-selected={selected ? 'true' : undefined}
          />
        );
      },
    },
    {
      colKey: 'version',
      title: '环境版本',
      width: 96,
      cell: ({ row }) => (row.draftRow ? <Tag theme="primary">当前草稿组合</Tag> : `v${row.version}`),
    },
    {
      colKey: 'templateReleaseId',
      title: '模板快照',
      cell: ({ row }) => {
        if (row.draftRow) return '当前模板草稿';
        const snapshot = templateSnapshotById.get(row.templateReleaseId);
        return snapshot ? `v${snapshot.version} · ${snapshot.id}` : row.templateReleaseId || '-';
      },
    },
    {
      colKey: 'id',
      title: 'Value 快照',
      cell: ({ row }) => (row.draftRow ? '当前 Value 草稿' : row.id),
    },
    {
      colKey: 'releaseType',
      title: '发布类型',
      width: 92,
      cell: ({ row }) =>
        row.draftRow ? '草稿' : row.releaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? '灰度' : '全量',
    },
    {
      colKey: 'active',
      title: '状态',
      width: 82,
      cell: ({ row }) =>
        row.draftRow ? (
          <Tag theme="primary">待发布</Tag>
        ) : (
          <Tag theme={row.active ? 'success' : 'default'}>{row.active ? '生效' : '停止'}</Tag>
        ),
    },
    { colKey: 'comment', title: '发布说明', cell: ({ row }) => (row.draftRow ? '-' : row.comment || '-') },
  ];

  return (
    <div className={styles.versionPane}>
      <div className={styles.versionActions}>
        <Button theme="primary" disabled={!namespace} onClick={onPreview}>
          渲染预览
        </Button>
      </div>

      <div className={styles.versionHistoryGrid}>
        <section className={styles.versionHistory}>
          <div className={styles.versionHistoryHeading}>
            <div>
              <strong>环境配置版本</strong>
              <span>模板快照与 Value 快照原子绑定 · 环境空间：{namespace || '-'}</span>
            </div>
            <Tag>草稿 + {releases.length}</Tag>
          </div>
          <Table
            aria-label="环境配置版本，点击一行进行选择"
            data={rows}
            columns={columns}
            rowKey="id"
            pagination={false}
            onRowClick={({ row }) => onReleaseChange(row.selectionId)}
          />
        </section>
      </div>

      {comparison ? (
        <section className={styles.versionComparison}>
          <div className={styles.versionHistoryHeading}>
            <div>
              <strong>格式化版本对比</strong>
              <span>{comparison.beforeLabel} → {comparison.afterLabel}</span>
            </div>
            <Tag>{comparison.before.diagnostics.length + comparison.after.diagnostics.length} 项 diagnostics</Tag>
          </div>
          <CodeDiffEditor
            namespace={namespace}
            group="config-template"
            filename={String(templateDraft.name || templateDraft.id)}
            readonly
            allowFullScreen
            height="100%"
            language={comparison.after.format || comparison.before.format}
            curValue={comparison.before.renderedContent}
            nextValue={comparison.after.renderedContent}
          />
        </section>
      ) : (
        <RenderPreviewPanel preview={preview} emptyMessage="默认预览当前模板草稿与当前环境 Value 草稿，也可选择历史环境版本。" />
      )}
    </div>
  );
};

export default React.memo(VersionManager);
