import React from 'react';
import { Button, Input, Select, Switch } from 'components/Fluent';
import { AddIcon, DeleteIcon } from 'components/Fluent/icons';
import {
  ConfigTemplateParameterSchema,
  TemplateParameterType,
} from 'services/config_templates';
import styles from './index.module.less';

const typeOptions: Array<{ label: string; value: TemplateParameterType }> = [
  { label: 'String', value: 'TEMPLATE_PARAMETER_STRING' },
  { label: 'Boolean', value: 'TEMPLATE_PARAMETER_BOOLEAN' },
  { label: 'Integer', value: 'TEMPLATE_PARAMETER_INTEGER' },
  { label: 'Decimal', value: 'TEMPLATE_PARAMETER_DECIMAL' },
];

export const emptySchemaParameter = (): ConfigTemplateParameterSchema => ({
  name: '',
  type: 'TEMPLATE_PARAMETER_STRING',
  required: true,
  sensitive: false,
  description: '',
});

interface SchemaEditorProps {
  value: ConfigTemplateParameterSchema[];
  editable: boolean;
  onChange: (value: ConfigTemplateParameterSchema[]) => void;
}

const SchemaEditor: React.FC<SchemaEditorProps> = ({ value, editable, onChange }) => {
  const update = (index: number, patch: Partial<ConfigTemplateParameterSchema>) => {
    onChange(value.map((item, itemIndex) => itemIndex === index ? { ...item, ...patch } : item));
  };

  return (
    <section className={styles.schemaSection}>
      <div className={styles.sectionHeading}>
        <div>
          <strong>参数 Schema</strong>
          <span>参数名使用 dotted name；发布后由各 Namespace 分别维护类型化 Value。</span>
        </div>
        {editable && (
          <Button
            variant="outline"
            size="small"
            icon={<AddIcon />}
            onClick={() => onChange([...value, emptySchemaParameter()])}
          >
            添加参数
          </Button>
        )}
      </div>
      <div className={styles.schemaTable} role="table" aria-label="模板参数 Schema">
        <div className={styles.schemaHeader} role="row">
          <span>参数名</span>
          <span>类型</span>
          <span>必填</span>
          <span>敏感</span>
          <span>说明</span>
          <span>操作</span>
        </div>
        {value.map((item, index) => (
          <div className={styles.schemaRow} role="row" key={`${item.name}-${index}`}>
            {editable
              ? <Input value={item.name} placeholder="database.host" onChange={(name) => update(index, { name })} />
              : <code>{item.name}</code>}
            {editable
              ? <Select
                value={item.type}
                options={typeOptions}
                onChange={(type: TemplateParameterType) => update(index, { type })}
              />
              : <span>{typeOptions.find(option => option.value === item.type)?.label || item.type}</span>}
            <Switch
              checked={item.required}
              disabled={!editable}
              aria-label={`${item.name || `参数 ${index + 1}`}必填`}
              onChange={(required: boolean) => update(index, { required })}
            />
            <Switch
              checked={item.sensitive}
              disabled={!editable}
              aria-label={`${item.name || `参数 ${index + 1}`}敏感`}
              onChange={(sensitive: boolean) => update(index, { sensitive })}
            />
            {editable
              ? <Input value={item.description || ''} placeholder="参数用途" onChange={(description) => update(index, { description })} />
              : <span>{item.description || '-'}</span>}
            {editable ? (
              <Button
                variant="text"
                shape="square"
                icon={<DeleteIcon />}
                aria-label={`删除参数 ${item.name || index + 1}`}
                onClick={() => onChange(value.filter((_, itemIndex) => itemIndex !== index))}
              />
            ) : <span />}
          </div>
        ))}
        {value.length === 0 && (
          <div className={styles.emptyState}>尚未声明参数。无参数模板仍可发布，但不能接收额外 Value。</div>
        )}
      </div>
    </section>
  );
};

export default React.memo(SchemaEditor);
