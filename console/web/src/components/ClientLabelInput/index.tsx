import React, { useState } from 'react';
import { AddIcon, DeleteIcon } from 'tdesign-icons-react';
import { Button, Form, Input, InputAdornment, RangeInput, Select, Space, TagInput } from 'tdesign-react';
import type { CustomValidator, InternalFormInstance } from 'tdesign-react';

import { ClientLabelType, ClientLabelTypeOption, MatcheLabel, MatchType, MatchTypeOption, MatchValueType, MatchValueTypeOption } from 'services/types';
import { get, set } from 'lodash';
import { current } from '@reduxjs/toolkit';
import Text from 'components/Text';

const { FormItem, FormList } = Form;

interface IClientLabelInputProps {
    form?: InternalFormInstance;
    name: string;
    label: string;
    disabled: boolean;
}

const ClientLabelInput: React.FC<IClientLabelInputProps> = (props) => {

    const labelsValidator: CustomValidator = (val) => {
        if (!props.form) {
            return { result: true, message: '' };
        }
        const labels = props.form.getFieldValue(props.name) as MatcheLabel[];
        const keys = labels.map(label => label.key);
        const hasDuplicate = keys.length !== new Set(keys).size;
        if (hasDuplicate) {
            return {
                result: false,
                type: 'error',
                message: '标签 key 不能重复',
            };
        }
        return { result: true, message: '' };
    };

    return (
        <div style={{ marginTop: 20 }}>
            <FormList
                name={[props.name]}
                rules={[{ validator: labelsValidator }]}
                initialData={[
                    {
                        key: ClientLabelType.CLIENT_IP,
                        value: {
                            type: MatchType.EXACT,
                            value_type: MatchValueType.TEXT,
                            value: ''
                        }
                    }
                ]}
            >
                {(fields, { add, remove }) => (
                    <>
                        {!fields || fields.length === 0 && (
                            <>
                                <FormItem label={`${props.label}`}>
                                    {props.disabled ? null : (
                                        <Space>
                                            <AddIcon onClick={() => add()} />
                                        </Space>
                                    )}
                                </FormItem>
                            </>
                        )}
                        {fields.map(({ key, name, ...restField }, index) => (
                            <FormItem key={key} label={index === 0 ? `${props.label}` : ' '}>
                                <FormItem
                                    key={'key-' + key}
                                    {...restField}
                                    name={[name, 'key']}
                                    rules={[
                                        { required: true, message: '标签不能为空' },
                                        { max: 128, message: '长度不超过128个字符' },
                                        { validator: labelsValidator },
                                    ]}
                                >
                                    <Select
                                        options={ClientLabelTypeOption}
                                        filterable
                                        creatable
                                    />
                                </FormItem>
                                <FormItem
                                    key={'type-' + key}
                                    {...restField}
                                    name={[name, 'value', 'type']}
                                >
                                    <Select autoWidth={true} options={MatchTypeOption} />
                                </FormItem>
                                <FormItem
                                    key={'value_type-' + key}
                                    {...restField}
                                    name={[name, 'value', 'value_type']}
                                >
                                    <Select autoWidth={true} options={MatchValueTypeOption} />
                                </FormItem>
                                <FormItem
                                    shouldUpdate={(prev, next) => {
                                        const prevType = prev[props.name]?.[name]?.value?.type;
                                        const nextType = next[props.name]?.[name]?.value?.type;
                                        return prevType !== nextType;
                                    }}
                                >
                                    {({ getFieldValue }) => {
                                        const matchType = (getFieldValue([props.name]) as MatcheLabel[] || [])[name]?.value?.type;
                                        if (matchType === MatchType.RANGE) {
                                            return (
                                                <FormItem key={'value-' + key}
                                                    {...restField}
                                                    name={[name, 'value', 'value']}
                                                    rules={[
                                                        { required: true, message: '值不能为空' },
                                                        { max: 4096, message: '长度不超过4096个字符' },
                                                    ]}>
                                                    <RangeInput readonly={props.disabled} />
                                                </FormItem>
                                            )
                                        } else if (matchType === MatchType.IN || matchType === MatchType.NOT_IN) {
                                            return (
                                                <FormItem key={'value-' + key}
                                                    {...restField}
                                                    name={[name, 'value', 'value']}
                                                    rules={[
                                                        { required: true, message: '值不能为空' },
                                                        { max: 4096, message: '长度不超过4096个字符' },
                                                    ]}>
                                                    <TagInput readonly={props.disabled} />
                                                </FormItem>
                                            )
                                        } else {
                                            return (
                                                <FormItem key={'value-' + key}
                                                    {...restField}
                                                    name={[name, 'value', 'value']}
                                                    rules={[
                                                        { required: true, message: '值不能为空' },
                                                        { max: 4096, message: '长度不超过4096个字符' },
                                                    ]}>
                                                    <Input readonly={props.disabled} />
                                                </FormItem>
                                            )
                                        }
                                    }}
                                </FormItem>
                                {props.disabled ? null : (
                                    <FormItem>
                                        <Button size='small' variant='text' icon={<DeleteIcon />} onClick={() => remove(name)} disabled={props.disabled} />
                                    </FormItem>
                                )}
                                {props.disabled ? null : (
                                    <Space>
                                        <Button size='small' variant='text' icon={<AddIcon />} onClick={() => add({
                                            key: ClientLabelType.CLIENT_IP,
                                            value: {
                                                type: MatchType.EXACT,
                                                value_type: MatchValueType.TEXT,
                                                value: ''
                                            }
                                        })} disabled={props.disabled} />
                                    </Space>
                                )}
                            </FormItem>
                        ))}
                    </>
                )}
            </FormList>
        </div>
    );
};

export default React.memo(ClientLabelInput);
