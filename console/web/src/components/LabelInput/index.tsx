import React from 'react';
import { Label } from 'services/types';
import { AddIcon, DeleteIcon } from 'tdesign-icons-react';
import { Button, Form, Input, Popup, Space, Tag } from 'tdesign-react';
import type { CustomValidator, FieldData, InternalFormInstance, NamePath } from 'tdesign-react';

import style from './index.module.less';

const { FormItem } = Form;

interface ILabelInputProps {
    form?: InternalFormInstance;
    name: NamePath;
    label: string;
    editable?: boolean;
    disabled?: boolean;
    keyPlaceholder?: string;
    valuePlaceholder?: string;
    hideLabel?: boolean;
    editorId?: string;
    emptyId?: string;
    countId?: string;
}

const normalizeLabels = (value: unknown): Label[] => {
    if (!Array.isArray(value)) {
        return [];
    }
    return value.map((item) => ({
        key: item?.key ?? '',
        value: item?.value ?? '',
    }));
};

const labelsEqual = (left: Label[], right: Label[]) => {
    if (left.length !== right.length) {
        return false;
    }
    return left.every((item, index) => item.key === right[index]?.key && item.value === right[index]?.value);
};

const LabelInput: React.FC<ILabelInputProps> = ({
    form,
    name,
    label,
    editable,
    disabled,
    keyPlaceholder = '标签键',
    valuePlaceholder = '标签值',
    hideLabel = false,
    editorId,
    emptyId,
    countId,
}) => {
    const watchedLabels = Form.useWatch(name, form);
    const [labels, setLabels] = React.useState<Label[]>(() => normalizeLabels(form?.getFieldValue(name)));
    const canEdit = editable !== undefined ? editable : !disabled;
    const nameKey = JSON.stringify(name);

    React.useEffect(() => {
        const nextLabels = normalizeLabels(watchedLabels ?? form?.getFieldValue(name));
        setLabels(prev => labelsEqual(prev, nextLabels) ? prev : nextLabels);
    }, [watchedLabels, form, nameKey]);

    const labelsValidator: CustomValidator = () => {
        const currentLabels = normalizeLabels(form?.getFieldValue(name) ?? labels);
        const missingKey = currentLabels.some((item) => !item.key.trim() && item.value.trim());
        if (missingKey) {
            return {
                result: false,
                type: 'error',
                message: '标签键不能为空',
            };
        }
        const filledKeys = currentLabels.map((item) => item.key.trim()).filter(Boolean);
        const hasDuplicate = filledKeys.length !== new Set(filledKeys).size;
        if (hasDuplicate) {
            return {
                result: false,
                type: 'error',
                message: '标签 key 不能重复',
            };
        }
        return { result: true, message: '' };
    };

    const syncLabels = (nextLabels: Label[], setFields?: (fields: FieldData[]) => void) => {
        setLabels(nextLabels);
        setFields?.([{
            name,
            value: nextLabels,
        }]);
    };

    const addLabel = (setFields?: (fields: FieldData[]) => void) => {
        syncLabels([...labels, { key: '', value: '' }], setFields);
    };

    const removeLabel = (idx: number, setFields?: (fields: FieldData[]) => void) => {
        syncLabels(labels.filter((_, index) => index !== idx), setFields);
    };

    const updateLabel = (idx: number, field: keyof Label, value: string, setFields?: (fields: FieldData[]) => void) => {
        const nextLabels = labels.map((item, index) => index === idx ? { ...item, [field]: value } : item);
        syncLabels(nextLabels, setFields);
    };

    const keyCount = labels.reduce<Record<string, number>>((acc, item) => {
        const key = item.key.trim();
        if (key) {
            acc[key] = (acc[key] || 0) + 1;
        }
        return acc;
    }, {});

    const renderReadOnly = () => {
        if (!labels.length) {
            return <div className={style.empty}>暂无标签</div>;
        }
        return (
            <Space breakLine>
                {labels.map((item, index) => (
                    <Tag key={`${item.key}-${index}`} theme="primary" variant="light">
                        {item.key}: {item.value}
                    </Tag>
                ))}
            </Space>
        );
    };

    const renderEditor = (setFields?: (fields: FieldData[]) => void) => (
        <div id={editorId} className={style.editor}>
            <div className={style.editorHeader}>
                <span>键</span>
                <span>值</span>
                <span>操作</span>
            </div>
            {labels.length === 0 ? (
                <div id={emptyId} className={style.emptyEditor}>
                    <div>暂无标签</div>
                </div>
            ) : (
                <div className={style.rows}>
                    {labels.map((item, index) => {
                        const duplicated = item.key.trim() !== '' && keyCount[item.key.trim()] > 1;
                        const missingKey = !item.key.trim() && item.value.trim();
                        const rowError = missingKey ? '标签键不能为空' : duplicated ? '标签 key 不能重复' : '';
                        return (
                            <div key={`${index}-${item.key}`} className={`${style.row} ${rowError ? style.rowError : ''}`}>
                                <Input
                                    value={item.key}
                                    clearable
                                    placeholder={keyPlaceholder}
                                    onChange={(value) => updateLabel(index, 'key', value, setFields)}
                                />
                                <Input
                                    value={item.value}
                                    clearable
                                    placeholder={valuePlaceholder}
                                    onChange={(value) => updateLabel(index, 'value', value, setFields)}
                                />
                                <Popup trigger="hover" content="删除标签">
                                    <Button
                                        shape="square"
                                        variant="text"
                                        onClick={() => removeLabel(index, setFields)}
                                >
                                    <DeleteIcon />
                                </Button>
                                </Popup>
                                {rowError && (
                                    <div className={style.rowErrorText}>{rowError}</div>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}
            <div className={style.footer}>
                <Button size="small" variant="text" icon={<AddIcon />} onClick={() => addLabel(setFields)}>
                    添加标签
                </Button>
                <span id={countId}>{labels.length} 个标签</span>
            </div>
        </div>
    );

    return (
        <FormItem>
            {({ setFields }) => (
                <FormItem label={hideLabel ? undefined : label} name={name} rules={[{ validator: labelsValidator }]}>
                    {canEdit ? renderEditor(setFields) : renderReadOnly()}
                </FormItem>
            )}
        </FormItem>
    );
};

export default React.memo(LabelInput);
