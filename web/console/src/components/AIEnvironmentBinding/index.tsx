import React from 'react';
import { Button, Dialog, Input, Select, Space } from 'components/Fluent';

export interface AILogicalDefinitionOption {
  id: string;
  name: string;
}

interface AIEnvironmentBindingProps {
  resourceLabel: string;
  suggestedName: string;
  loadDefinitions: () => Promise<AILogicalDefinitionOption[]>;
  createDefinition: (name: string) => Promise<AILogicalDefinitionOption>;
  bindDefinition: (definitionId: string) => Promise<void>;
  onBound: () => void | Promise<void>;
}

const AIEnvironmentBinding: React.FC<AIEnvironmentBindingProps> = ({
  resourceLabel,
  suggestedName,
  loadDefinitions,
  createDefinition,
  bindDefinition,
  onBound,
}) => {
  const [visible, setVisible] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [definitions, setDefinitions] = React.useState<AILogicalDefinitionOption[]>([]);
  const [definitionId, setDefinitionId] = React.useState('');
  const [newName, setNewName] = React.useState(suggestedName);
  const [error, setError] = React.useState('');

  const open = async () => {
    setVisible(true);
    setError('');
    setNewName(suggestedName);
    setLoading(true);
    try {
      setDefinitions(await loadDefinitions());
    } catch (cause) {
      setError((cause as Error).message || '逻辑定义加载失败');
    } finally {
      setLoading(false);
    }
  };

  const confirm = async () => {
    setLoading(true);
    setError('');
    try {
      let targetId = definitionId;
      if (!targetId) {
        const definition = await createDefinition(newName.trim());
        targetId = definition.id;
      }
      await bindDefinition(targetId);
      setVisible(false);
      setDefinitionId('');
      await onBound();
    } catch (cause) {
      setError((cause as Error).message || '环境关联失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button size="small" variant="outline" onClick={open}>关联跨环境定义</Button>
      <Dialog
        visible={visible}
        onClose={() => setVisible(false)}
        onConfirm={confirm}
        confirmBtn={loading ? '处理中...' : '确认关联'}
        header={`关联 ${resourceLabel} 逻辑定义`}
        width={520}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <p>选择已有定义即可把不同 Namespace 中的实例明确归为同一逻辑资源；系统不会按名称自动合并。</p>
          <Select
            clearable
            filterable
            loading={loading}
            value={definitionId || undefined}
            options={definitions.map((definition) => ({ label: definition.name, value: definition.id }))}
            placeholder="选择已有逻辑定义"
            onChange={(value) => setDefinitionId(String(value || ''))}
          />
          {!definitionId && (
            <Input
              value={newName}
              placeholder="或输入新逻辑定义名称"
              onChange={(value: unknown) => setNewName(String(value ?? ''))}
            />
          )}
          {error && <span role="alert">{error}</span>}
        </Space>
      </Dialog>
    </>
  );
};

export default React.memo(AIEnvironmentBinding);
