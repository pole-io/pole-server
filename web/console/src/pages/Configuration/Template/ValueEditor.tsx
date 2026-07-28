import React from 'react';
import { Input, Select } from 'components/Fluent';
import {
  ConfigTemplateParameterSchema,
  ConfigTemplateValue,
} from 'services/config_templates';
import styles from './index.module.less';

export const valueToText = (value?: ConfigTemplateValue) => {
  if (!value) return '';
  if ('stringValue' in value) return value.stringValue;
  if ('booleanValue' in value) return value.booleanValue ? 'true' : 'false';
  if ('integerValue' in value) return String(value.integerValue);
  return value.decimalValue;
};

export const textToValue = (
  type: ConfigTemplateParameterSchema['type'],
  value: string,
): ConfigTemplateValue => {
  switch (type) {
    case 'TEMPLATE_PARAMETER_BOOLEAN':
      return { booleanValue: value === 'true' };
    case 'TEMPLATE_PARAMETER_INTEGER':
      return { integerValue: value };
    case 'TEMPLATE_PARAMETER_DECIMAL':
      return { decimalValue: value };
    default:
      return { stringValue: value };
  }
};

interface ValueEditorProps {
  schema: ConfigTemplateParameterSchema[];
  values: Record<string, ConfigTemplateValue>;
  onChange: (values: Record<string, ConfigTemplateValue>) => void;
}

const ValueEditor: React.FC<ValueEditorProps> = ({ schema, values, onChange }) => (
  <div className={styles.valueGrid}>
    {schema.map((parameter) => {
      const text = valueToText(values[parameter.name]);
      const update = (next: string) => onChange({
        ...values,
        [parameter.name]: textToValue(parameter.type, next),
      });
      return (
        <label className={styles.valueField} key={parameter.name}>
          <span>
            <code>{parameter.name}</code>
            {parameter.required && <em>必填</em>}
            {parameter.sensitive && <em>敏感</em>}
          </span>
          {parameter.type === 'TEMPLATE_PARAMETER_BOOLEAN' ? (
            <Select
              value={text || 'false'}
              options={[
                { label: 'true', value: 'true' },
                { label: 'false', value: 'false' },
              ]}
              onChange={update}
            />
          ) : (
            <Input
              value={text}
              type={parameter.sensitive ? 'password' : 'text'}
              placeholder={parameter.type === 'TEMPLATE_PARAMETER_DECIMAL' ? '1.25' : '请输入 Value'}
              onChange={update}
            />
          )}
          <small>{parameter.description || parameter.type.replace('TEMPLATE_PARAMETER_', '')}</small>
        </label>
      );
    })}
    {schema.length === 0 && <div className={styles.emptyState}>当前模板没有参数，无需维护 Namespace Value。</div>}
  </div>
);

export default React.memo(ValueEditor);
