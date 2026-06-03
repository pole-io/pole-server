import React from 'react';
import { AddIcon, CloseIcon } from 'tdesign-icons-react';
import { Button, Col, Form, Input, Popup, RangeInput, Row, Select, Table, TagInput } from 'tdesign-react';
import type { CustomValidator, FieldData, InternalFormInstance, NamePath, PrimaryTableProps, TableRowData } from 'tdesign-react';

import { MatchType, MatchTypeMap, MatchTypeOption, MatchValueType } from 'services/types';
import { RouteArgumentTextMap, RoutingArgumentsType, RoutingArgumentsTypeOptions, RoutingSourceArgument } from 'services/router';
import Text from 'components/Text';

const { FormItem } = Form;

const defaultMatchArgs: () => RoutingSourceArgument = () => ({
    type: RoutingArgumentsType.HEADER,
    key: '',
    value: {
        type: MatchType.EXACT,
        value: '',
        value_type: MatchValueType.TEXT
    }
});

interface IMatchInputProps {
    form: InternalFormInstance;
    name: NamePath;
    label: string;
    editable: boolean;
}

const MatchInput: React.FC<IMatchInputProps> = ({ form, name, label, editable }) => {

    const trafficArgs: RoutingSourceArgument[] = Form.useWatch(name, form);
    const [loaded, setLoaded] = React.useState(false);

    // 创建新的 rules 对象
    const [rules, setRules] = React.useState<RoutingSourceArgument[]>([defaultMatchArgs()]);

    const labelsValidator: CustomValidator = (val) => {
        if (!form) {
            return { result: true, message: '' };
        }
        const keys = rules.map(label => label.type + '/' + label.key);
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

    React.useEffect(() => {
        if (!trafficArgs || loaded) {
            return;
        }
        console.log('initialRules changed', trafficArgs);
        setLoaded(true);
        setRules(trafficArgs);
    }, [trafficArgs]);

    const addMatch = () => {
        const newRules = [...rules];
        newRules.push(defaultMatchArgs());
        setRules(newRules);
    };


    const delMatchArgs = (idx: number, callback: (field: FieldData[]) => void) => {
        const newRules = [...rules];
        newRules.splice(idx, 1);
        setRules(newRules);
        callback([{
            name: name,
            value: newRules,
        }]);
    };

    const updateMatchArgs = (idx: number, args: RoutingSourceArgument, callback: (fields: FieldData[]) => void) => {
        const newRules = [...rules];
        if (idx < newRules.length) {
            newRules[idx] = args;
        } else {
            newRules.push(args);
        }
        setRules(newRules);
        callback([{
            name: name,
            value: newRules,
        }]);
    };

    // 将 MatchValueInput 用 React.forwardRef 包裹，支持 ref 透传
    const MatchValueInput = React.forwardRef<any,
        {
            value: any;
            onChange: (value: any) => void;
            matchType: MatchType;
        }
    >(({ value, onChange, matchType }, ref) => {
        switch (matchType) {
            case MatchType.RANGE:
                return (
                    <RangeInput
                        ref={ref}
                        value={value ? value.split(',').map(Number) : [0, 0]}
                        onChange={(val) => onChange(val.join(','))}
                        placeholder={['最小值', '最大值']}
                    />
                );
            case MatchType.IN:
            case MatchType.NOT_IN:
                return (
                    <TagInput
                        ref={ref}
                        value={value ? value.split(',') : []}
                        onChange={(val) => onChange(val.join(','))}
                        placeholder="请输入多个值，回车分隔"
                    />
                );
            case MatchType.EXACT:
            case MatchType.REGEX:
            case MatchType.NOT_EQUALS:
                return (
                    <Input
                        ref={ref}
                        value={value}
                        onChange={(val) => onChange(val)}
                        placeholder="请输入匹配值"
                    />
                );
            default:
                return null;
        }
    });

    const trafficTableColumns = (setFields: (fields: FieldData[]) => void): PrimaryTableProps['columns'] => [
        {
            colKey: 'type',
            title: '参数类型',
            edit: {
                keepEditMode: editable,
                showEditIcon: editable,
                component: Select,
                abortEditOnEvent: ['onChange'],
                props: {
                    clearable: true,
                    options: RoutingArgumentsTypeOptions,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(context.rowIndex, context.newRowData as RoutingSourceArgument, setFields);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{RouteArgumentTextMap[row.type]}</Text>
        },
        {
            colKey: 'key',
            title: '参数键',
            edit: {
                keepEditMode: editable,
                showEditIcon: editable,
                component: Input,
                abortEditOnEvent: ['onChange'],
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(context.rowIndex, context.newRowData as RoutingSourceArgument, setFields);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
        },
        {
            colKey: 'value.type',
            title: '匹配类型',
            edit: {
                keepEditMode: editable,
                showEditIcon: editable,
                component: Select,
                abortEditOnEvent: ['onChange'],
                props: {
                    clearable: true,
                    options: MatchTypeOption,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(context.rowIndex, context.newRowData as RoutingSourceArgument, setFields);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
            cell: ({ row }) => <Text>{MatchTypeMap[row.value.type as MatchType]}</Text>
        },
        {
            colKey: 'value.value',
            title: '匹配值',
            edit: {
                keepEditMode: editable,
                showEditIcon: editable,
                component: MatchValueInput,
                abortEditOnEvent: ['onChange'],
                props: (context: { row: RoutingSourceArgument }) => ({
                    matchType: context.row.value.type,
                }),
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateMatchArgs(context.rowIndex, context.newRowData as RoutingSourceArgument, setFields);
                },
                // 校验规则，此处同 Form 表单
                validateTrigger: 'change',
            },
        },
        {
            colKey: 'action',
            title: '操作',
            cell: ({ row, rowIndex }) => (
                <Popup trigger="hover" content="删除参数">
                    <Button
                        shape="circle"
                        variant="text"
                        onClick={() => {
                            delMatchArgs(rowIndex, setFields);
                        }}>
                        <CloseIcon />
                    </Button>
                </Popup>

            ),
        }
    ]

    return (
        <div style={{ marginTop: 20 }}>
            <Row align="middle" style={{ marginBottom: 8 }}>
                <FormItem name={name}>
                    {({ getFieldValue, setFields }) => {
                        return (
                            <FormItem label={label} name={name} rules={[{ validator: labelsValidator }]}>
                                <Table
                                    rowKey="key"
                                    data={rules || []}
                                    columns={editable ? trafficTableColumns(setFields) : trafficTableColumns(setFields)?.filter(col => col.colKey !== 'action')}
                                />
                            </FormItem>
                        )
                    }}
                </FormItem>
                <FormItem label=" ">
                    <Col span={2}>
                        {editable && (
                            <Button variant="text" onClick={() => addMatch()} icon={<AddIcon />}>添加</Button>
                        )}
                    </Col>
                </FormItem>
            </Row>
        </div>
    );
};

export default React.memo(MatchInput);
