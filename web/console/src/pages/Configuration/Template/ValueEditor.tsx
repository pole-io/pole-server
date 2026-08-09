import React from 'react';
import { Button, Input, Select } from 'components/Fluent';
import { AddIcon } from 'components/Fluent/icons';
import { ConfigTemplateParameterSchema, ConfigTemplateValue } from 'services/config_templates';
import styles from './index.module.less';

const canonicalIntegerPattern = /^-?(?:0|[1-9][0-9]*)$/;
const canonicalDecimalPattern = /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*[1-9])?$/;

export const valueToText = (value?: ConfigTemplateValue) => {
  if (!value) return '';
  if ('stringValue' in value) return value.stringValue;
  if ('booleanValue' in value) return value.booleanValue ? 'true' : 'false';
  if ('integerValue' in value) return String(value.integerValue);
  return value.decimalValue;
};

export const textToValue = (type: ConfigTemplateParameterSchema['type'], value: string): ConfigTemplateValue => {
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

export const validateTemplateValue = (
  parameter: ConfigTemplateParameterSchema,
  value?: ConfigTemplateValue
): string => {
  const text = valueToText(value);
  if (!value || text === '') return parameter.required ? '请输入必填 Value' : '';
  if (
    parameter.type === 'TEMPLATE_PARAMETER_INTEGER' &&
    (!canonicalIntegerPattern.test(text) || text === '-0')
  ) {
    return '请输入规范整数，例如 0、12 或 -12（不允许前导零）';
  }
  if (
    parameter.type === 'TEMPLATE_PARAMETER_DECIMAL' &&
    (!canonicalDecimalPattern.test(text) || text === '-0')
  ) {
    return '请输入规范小数，例如 0、1.25 或 -0.5（不允许前导零或末尾零）';
  }
  return '';
};

export const validateTemplateValues = (
  schema: ConfigTemplateParameterSchema[],
  values: Record<string, ConfigTemplateValue>
) =>
  Object.fromEntries(
    schema
      .map((parameter) => [parameter.name, validateTemplateValue(parameter, values[parameter.name])] as const)
      .filter(([, error]) => Boolean(error))
  );

interface ValueEditorProps {
  schema: ConfigTemplateParameterSchema[];
  values: Record<string, ConfigTemplateValue>;
  onChange: (values: Record<string, ConfigTemplateValue>) => void;
  onCreateSchema: () => void;
}

const ValueEditor: React.FC<ValueEditorProps> = ({ schema, values, onChange, onCreateSchema }) => (
  <div className={styles.valueGrid}>
    {schema.map((parameter, index) => {
      const text = valueToText(values[parameter.name]);
      const error = validateTemplateValue(parameter, values[parameter.name]);
      const errorId = `template-value-error-${index}`;
      const update = (next: string) =>
        onChange({
          ...values,
          [parameter.name]: textToValue(parameter.type, next),
        });
      return (
        <label className={styles.valueField} key={parameter.name}>
          <span>
            <code>{parameter.name}</code>
            {parameter.required && <em>必填</em>}
            {parameter.sensitive && <em>加密存储</em>}
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
              status={error ? 'error' : 'default'}
              aria-invalid={Boolean(error)}
              aria-describedby={error ? errorId : undefined}
              onChange={update}
            />
          )}
          <small>{parameter.description || parameter.type.replace('TEMPLATE_PARAMETER_', '')}</small>
          {error && (
            <small id={errorId} className={styles.valueError} role="alert">
              {error}
            </small>
          )}
        </label>
      );
    })}
    {schema.length === 0 && (
      <div className={styles.actionEmptyState}>
        <strong>先创建参数 Schema</strong>
        <span>环境 Value 会根据 Schema 自动生成输入项。</span>
        <Button theme="primary" icon={<AddIcon />} onClick={onCreateSchema}>
          添加 Schema 参数
        </Button>
      </div>
    )}
  </div>
);

export default React.memo(ValueEditor);
