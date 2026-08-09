import React from 'react';
import { Button, PrimaryTableProps, Table, Tag } from 'components/Fluent';
import { ConfigTemplateRelease, NamespaceTemplateValueRelease, RenderPreview } from 'services/config_templates';
import RenderPreviewPanel from './RenderPreviewPanel';
import styles from './index.module.less';

interface VersionManagerProps {
  namespace: string;
  releases: ConfigTemplateRelease[];
  valueReleases: NamespaceTemplateValueRelease[];
  templateReleaseId: string;
  valueReleaseId: string;
  preview?: RenderPreview;
  onTemplateReleaseChange: (value: string) => void;
  onValueReleaseChange: (value: string) => void;
  onPreview: () => void;
}

const VersionManager: React.FC<VersionManagerProps> = ({
  namespace,
  releases,
  valueReleases,
  templateReleaseId,
  valueReleaseId,
  preview,
  onTemplateReleaseChange,
  onValueReleaseChange,
  onPreview,
}) => {
  const templateVersionById = React.useMemo(
    () => new Map(releases.map((release) => [release.id, release.version])),
    [releases]
  );

  const selectionColumn = (selectedId: string): PrimaryTableProps['columns'][number] => ({
    colKey: 'selection',
    title: '选择',
    width: 54,
    ellipsis: false,
    cell: ({ row }) => {
      const selected = row.id === selectedId;
      return (
        <span
          aria-label={selected ? '已选择' : '未选择'}
          className={selected ? styles.versionSelectMarkActive : styles.versionSelectMark}
          data-version-selected={selected ? 'true' : undefined}
        />
      );
    },
  });

  const templateColumns: PrimaryTableProps['columns'] = [
    selectionColumn(templateReleaseId),
    { colKey: 'version', title: '版本', width: 76, cell: ({ row }) => `v${row.version}` },
    { colKey: 'format', title: '格式', width: 78, cell: ({ row }) => String(row.format).toUpperCase() },
    { colKey: 'contentSha256', title: '内容 SHA-256', cell: ({ row }) => row.contentSha256 || '-' },
    { colKey: 'id', title: 'Release ID' },
  ];

  const valueColumns: PrimaryTableProps['columns'] = [
    selectionColumn(valueReleaseId),
    { colKey: 'version', title: '版本', width: 76, cell: ({ row }) => `v${row.version}` },
    {
      colKey: 'releaseType',
      title: '类型',
      width: 86,
      cell: ({ row }) => (row.releaseType === 'TEMPLATE_VALUE_RELEASE_GRAY' ? '灰度' : '全量'),
    },
    {
      colKey: 'templateReleaseId',
      title: '模板版本',
      width: 100,
      cell: ({ row }) => `v${templateVersionById.get(row.templateReleaseId) || '-'}`,
    },
    {
      colKey: 'active',
      title: '状态',
      width: 82,
      cell: ({ row }) => <Tag theme={row.active ? 'success' : 'default'}>{row.active ? '生效' : '停止'}</Tag>,
    },
    { colKey: 'id', title: 'Value Release ID' },
  ];

  return (
    <div className={styles.versionPane}>
      <div className={styles.versionActions}>
        <Button theme="primary" disabled={!templateReleaseId || !valueReleaseId} onClick={onPreview}>
          渲染预览
        </Button>
      </div>

      <div className={styles.versionHistoryGrid}>
        <section className={styles.versionHistory}>
          <div className={styles.versionHistoryHeading}>
            <div>
              <strong>模板版本</strong>
              <span>不可变的模板内容、格式和参数 Schema 快照</span>
            </div>
            <Tag>{releases.length}</Tag>
          </div>
          <Table
            aria-label="模板版本，点击一行进行选择"
            data={releases}
            columns={templateColumns}
            rowKey="id"
            pagination={false}
            onRowClick={({ row }) => onTemplateReleaseChange(row.id)}
          />
        </section>
        <section className={styles.versionHistory}>
          <div className={styles.versionHistoryHeading}>
            <div>
              <strong>Value 版本</strong>
              <span>{namespace ? `环境空间：${namespace}` : '请选择环境空间'}</span>
            </div>
            <Tag>{valueReleases.length}</Tag>
          </div>
          <Table
            aria-label="环境 Value 版本，点击一行进行选择"
            data={valueReleases}
            columns={valueColumns}
            rowKey="id"
            pagination={false}
            onRowClick={({ row }) => onValueReleaseChange(row.id)}
          />
        </section>
      </div>

      <RenderPreviewPanel preview={preview} emptyMessage="选择模板版本和 Value 版本后查看最终配置。" />
    </div>
  );
};

export default React.memo(VersionManager);
