import React from 'react';
import {
  Combobox,
  Option,
  Popover,
  PopoverSurface,
  PopoverTrigger,
  Tag as FluentTag,
  TagGroup,
} from '@fluentui/react-components';
import { Button, Input, Select } from 'components/Fluent';
import { FilterIcon, SearchIcon } from 'components/Fluent/icons';

import style from './index.module.less';

export type QueryValue = string | number | boolean | Array<string | number> | null | undefined;

export interface QueryOption {
  label: React.ReactNode;
  value: string | number | boolean;
  disabled?: boolean;
}

interface QueryFieldBase {
  key: string;
  label: string;
  placeholder?: string;
  formatValue?: (value: QueryValue, field: QueryField) => React.ReactNode;
}

export interface QueryTextField extends QueryFieldBase {
  type: 'text';
}

export interface QuerySelectField extends QueryFieldBase {
  type: 'select' | 'multiselect';
  options?: QueryOption[];
  filterable?: boolean;
}

export interface QueryBooleanField extends QueryFieldBase {
  type: 'boolean';
  trueLabel?: string;
  falseLabel?: string;
}

export interface QueryCustomField extends QueryFieldBase {
  type: 'custom';
  render?: (value: QueryValue, onChange: (value: QueryValue) => void) => React.ReactNode;
}

export type QueryField = QueryTextField | QuerySelectField | QueryBooleanField | QueryCustomField;

export type QueryValues = Record<string, QueryValue>;

export interface QuerySnapshot {
  keyword: string;
  values: QueryValues;
}

interface QueryComposerBaseProps {
  keyword: string;
  onKeywordChange: (value: string) => void;
  onReset: () => void;
  fields?: QueryField[];
  values?: QueryValues;
  onValuesChange?: (values: QueryValues) => void;
  keywordPlaceholder?: string;
  suggestions?: Array<string | { label: string; value: string }>;
  timeRange?: React.ReactNode;
  actions?: React.ReactNode;
  loading?: boolean;
  className?: string;
}

export type QueryComposerProps = QueryComposerBaseProps & (
  | {
    mode?: 'explicit';
    onSubmit: (snapshot: QuerySnapshot) => void;
  }
  | {
    mode: 'instant';
    onSubmit?: never;
  }
);

const booleanOptions = (field: QueryBooleanField): QueryOption[] => [
  { label: field.trueLabel || '是', value: 'true' },
  { label: field.falseLabel || '否', value: 'false' },
];

const getFieldOptions = (field: QueryField): QueryOption[] => {
  if (field.type === 'boolean') return booleanOptions(field);
  if (field.type === 'select' || field.type === 'multiselect') return field.options || [];
  return [];
};

const hasValue = (value: QueryValue) => (
  Array.isArray(value)
    ? value.length > 0
    : value !== undefined && value !== null && value !== ''
);

const getOptionLabel = (field: QueryField, value: string | number | boolean) => {
  const option = getFieldOptions(field).find((item) => String(item.value) === String(value));
  return option?.label ?? String(value);
};

const formatFieldValue = (field: QueryField, value: QueryValue) => {
  if (field.formatValue) return field.formatValue(value, field);
  if (Array.isArray(value)) {
    return value.map((item) => getOptionLabel(field, item)).reduce<React.ReactNode[]>((result, item, index) => {
      if (index > 0) result.push('、');
      result.push(item);
      return result;
    }, []);
  }
  return getOptionLabel(field, value as string | number | boolean);
};

export const QueryComposer: React.FC<QueryComposerProps> = ({
  keyword,
  onKeywordChange,
  onSubmit,
  onReset,
  fields = [],
  values = {},
  onValuesChange,
  keywordPlaceholder = '搜索名称、标识或关键字',
  suggestions = [],
  timeRange,
  actions,
  loading,
  className,
  mode = 'explicit',
}) => {
  const [advancedOpen, setAdvancedOpen] = React.useState(false);
  const [suggestionsOpen, setSuggestionsOpen] = React.useState(false);
  const [draftKeyword, setDraftKeyword] = React.useState(keyword);
  const [draftValues, setDraftValues] = React.useState<QueryValues>(values);
  const latestValues = React.useRef(values);
  latestValues.current = values;
  const valuesSignature = JSON.stringify(values);

  React.useEffect(() => {
    setDraftKeyword(keyword);
  }, [keyword]);

  React.useEffect(() => {
    setDraftValues(latestValues.current);
  }, [valuesSignature]);

  const normalizedSuggestions = React.useMemo(
    () => Array.from(new Map(
      suggestions
        .map((item) => typeof item === 'string' ? { label: item, value: item } : item)
        .map((item) => [item.value, item]),
    ).values()),
    [suggestions],
  );
  const filteredSuggestions = React.useMemo(() => {
    const normalizedKeyword = draftKeyword.trim().toLocaleLowerCase();
    if (!normalizedKeyword) return normalizedSuggestions.slice(0, 8);
    return normalizedSuggestions
      .filter((item) => `${item.label} ${item.value}`.toLocaleLowerCase().includes(normalizedKeyword))
      .slice(0, 8);
  }, [draftKeyword, normalizedSuggestions]);
  const activeFields = fields.filter((field) => hasValue(draftValues[field.key]));
  const hasConditions = Boolean(draftKeyword.trim()) || activeFields.length > 0 || Boolean(timeRange);

  const updateDraftValue = (key: string, value: QueryValue) => {
    setDraftValues((current) => ({ ...current, [key]: value }));
  };

  const removeField = (field: QueryField) => {
    const nextValues = {
      ...draftValues,
      [field.key]: Array.isArray(draftValues[field.key]) ? [] : '',
    };
    setDraftValues(nextValues);
    if (mode === 'instant') onValuesChange?.(nextValues);
  };

  const updateDraftKeyword = (value: string) => {
    setDraftKeyword(value);
    if (mode === 'instant') onKeywordChange(value);
  };

  const submit = () => {
    const snapshot = { keyword: draftKeyword, values: draftValues };
    onKeywordChange(draftKeyword);
    onValuesChange?.(draftValues);
    onSubmit?.(snapshot);
  };

  const reset = () => {
    setDraftKeyword('');
    setDraftValues(fields.reduce<QueryValues>((nextValues, field) => ({
      ...nextValues,
      [field.key]: Array.isArray(draftValues[field.key]) ? [] : '',
    }), { ...draftValues }));
    onReset();
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (
      event.key === 'Enter'
      && !event.nativeEvent.isComposing
      && !(suggestionsOpen && filteredSuggestions.length > 0)
    ) {
      event.preventDefault();
      submit();
    }
  };

  return (
    <div className={`${style.composer} ${className || ''}`} data-query-composer>
      <div className={style.primaryRow}>
        <div className={style.keywordWrap}>
          <span className={style.searchIcon} aria-hidden><SearchIcon /></span>
          <Combobox
            className={style.keyword}
            aria-label="全局搜索"
            freeform
            clearable
            inlinePopup
            placeholder={keywordPlaceholder}
            value={draftKeyword}
            onChange={(event) => updateDraftKeyword(event.currentTarget.value)}
            onKeyDownCapture={(event) => {
              if (event.key === 'Enter' && event.nativeEvent.isComposing) {
                event.stopPropagation();
              }
            }}
            onKeyDown={handleKeyDown}
            onOpenChange={(_, data) => setSuggestionsOpen(data.open)}
            selectedOptions={draftKeyword ? [draftKeyword] : []}
            onOptionSelect={(_, data) => {
              updateDraftKeyword(data.optionValue || '');
              setSuggestionsOpen(false);
            }}
          >
            {filteredSuggestions.map((item) => (
              <Option key={item.value} value={item.value} text={item.label}>
                {item.label}
              </Option>
            ))}
          </Combobox>
        </div>

        {timeRange && <div className={style.timeRange} data-query-time-range>{timeRange}</div>}

        {fields.length > 0 && (
          <Popover
            open={advancedOpen}
            onOpenChange={(_, data) => {
              if (!data.open) setDraftValues(values);
              setAdvancedOpen(data.open);
            }}
            positioning="below-end"
            trapFocus
          >
            <PopoverTrigger disableButtonEnhancement>
              <Button
                variant="outline"
                icon={<FilterIcon />}
                aria-label={`高级筛选，已选 ${activeFields.length} 项`}
                aria-expanded={advancedOpen}
              >
                高级筛选{activeFields.length > 0 ? ` ${activeFields.length}` : ''}
              </Button>
            </PopoverTrigger>
            <PopoverSurface className={style.advancedSurface} aria-label="高级筛选条件">
              <div className={style.advancedHeader}>
                <div>
                  <strong>高级筛选</strong>
                  <span>组合多个维度缩小结果范围</span>
                </div>
                {activeFields.length > 0 && (
                  <Button
                    size="small"
                    variant="text"
                    onClick={() => setDraftValues(
                      fields.reduce<QueryValues>((nextValues, field) => ({
                        ...nextValues,
                        [field.key]: Array.isArray(draftValues[field.key]) ? [] : '',
                      }), { ...draftValues }),
                    )}
                  >
                    清空条件
                  </Button>
                )}
              </div>
              <div className={style.advancedFields}>
                {fields.map((field) => (
                  <label className={style.field} key={field.key}>
                    <span>{field.label}</span>
                    {field.type === 'custom'
                      ? field.render?.(draftValues[field.key], (value) => updateDraftValue(field.key, value))
                      : (
                      field.type === 'select' || field.type === 'multiselect' || field.type === 'boolean'
                        ? (
                          <Select
                            clearable={field.type !== 'multiselect'}
                            filterable={'filterable' in field ? field.filterable : false}
                            inlinePopup
                            multiple={field.type === 'multiselect'}
                            options={getFieldOptions(field)}
                            placeholder={field.placeholder || `请选择${field.label}`}
                            value={draftValues[field.key]}
                            onChange={(value: QueryValue) => updateDraftValue(field.key, value)}
                          />
                        )
                        : (
                          <Input
                            clearable
                            placeholder={field.placeholder || `请输入${field.label}`}
                            value={String(draftValues[field.key] ?? '')}
                            onChange={(value) => updateDraftValue(field.key, value)}
                          />
                        )
                    )}
                  </label>
                ))}
              </div>
              <div className={style.advancedActions}>
                <Button
                  variant="text"
                  onClick={() => {
                    setDraftValues(values);
                    setAdvancedOpen(false);
                  }}
                >
                  取消
                </Button>
                <Button
                  theme="primary"
                  onClick={() => {
                    if (mode === 'instant') onValuesChange?.(draftValues);
                    setAdvancedOpen(false);
                  }}
                >
                  完成
                </Button>
              </div>
            </PopoverSurface>
          </Popover>
        )}

        {mode === 'explicit' && (
          <Button theme="primary" icon={<SearchIcon />} loading={loading} onClick={submit}>
            查询
          </Button>
        )}

        {hasConditions && (
          <Button variant="text" onClick={reset}>重置</Button>
        )}

        {actions}
      </div>

      {activeFields.length > 0 && (
        <div className={style.activeConditions} aria-label="已选筛选条件">
          <span className={style.activeLabel}>已选条件</span>
          <TagGroup
            className={style.tags}
            aria-label="已选筛选条件"
            dismissible
            onDismiss={(_, data) => {
              const field = fields.find((item) => item.key === data.value);
              if (!field) return;
              removeField(field);
            }}
          >
            {activeFields.map((field) => (
              <FluentTag
                key={field.key}
                value={field.key}
                appearance="outline"
                dismissible
                dismissIcon={{ 'aria-label': `移除${field.label}条件` }}
              >
                <span className={style.tagLabel}>{field.label}</span>
                <span className={style.tagValue}>{formatFieldValue(field, draftValues[field.key])}</span>
              </FluentTag>
            ))}
          </TagGroup>
        </div>
      )}
    </div>
  );
};

export default QueryComposer;
