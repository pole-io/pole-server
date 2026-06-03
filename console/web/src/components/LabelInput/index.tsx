import React, { } from 'react';
import { Label } from 'services/types';
import { AddIcon, CloseIcon } from 'tdesign-icons-react';
import { Button, Col, Form, Input, Popup, Row, Table } from 'tdesign-react';
import type { CustomValidator, Data, FieldData, InternalFormInstance, NamePath, PrimaryTableProps, TableRowData } from 'tdesign-react';

const { FormItem, FormList } = Form;

interface ILabelInputProps {
    form: InternalFormInstance;
    name: NamePath;
    label: string;
    editable?: boolean;
}

const LabelInput: React.FC<ILabelInputProps> = ({ form, name, label, editable }) => {
    const [labels, setLabels] = React.useState<Label[]>([]);

    const resLabels: Label[] = Form.useWatch(name, form);
    const [loaded, setLoaded] = React.useState(false);

    React.useEffect(() => {
        if (!resLabels || loaded) {
            return;
        }
        setLoaded(true);
        const initialLabels = form.getFieldValue(name) || [];
        setLabels(initialLabels as Label[]);
    }, [resLabels]);

    const labelsValidator: CustomValidator = (val) => {
        if (!form) {
            return { result: true, message: '' };
        }
        const labels = form.getFieldValue(name) as Label[];
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

    const addLabel = (callback: (fields: FieldData[]) => void) => {
        const newLabels = [...labels];
        newLabels.push({ key: '', value: '' });
        setLabels(newLabels);
        callback([{
            name: name,
            value: newLabels,
        }]);
    };


    const delLabel = (idx: number, callback: (fields: FieldData[]) => void) => {
        const newLabels = [...labels];
        newLabels.splice(idx, 1);
        setLabels(newLabels);
        callback([{
            name: name,
            value: newLabels,
        }]);
    };

    const updateLabel = (idx: number, args: Label, callback: (fields: FieldData[]) => void) => {
        const newLabels = [...labels];
        if (idx < newLabels.length) {
            newLabels[idx] = {
                ...newLabels[idx],
                ...Object.fromEntries(Object.entries(args).filter(([_, value]) => value !== undefined && value !== null)),
            };
        } else {
            newLabels.push(args);
        }
        setLabels(newLabels);
        callback([{
            name: name,
            value: newLabels,
        }]);
    };

    const labelTableColumns = (setFields: (fields: FieldData[]) => void): PrimaryTableProps['columns'] => [
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
                    updateLabel(context.rowIndex, context.newRowData as Label, setFields);
                },
            },
        },
        {
            colKey: 'value',
            title: '匹配值',
            edit: {
                keepEditMode: editable,
                showEditIcon: editable,
                component: Input,
                abortEditOnEvent: ['onChange'],
                props: {
                    clearable: true,
                },
                onEdited: (context: { rowIndex: number; newRowData: TableRowData }) => {
                    updateLabel(context.rowIndex, context.newRowData as Label, setFields);
                },
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
                            delLabel(rowIndex, setFields);
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
                <FormItem>
                    {({ getFieldValue, setFieldsValue, setFields }) => {
                        return (
                            <FormItem label={label} name={name} rules={[{ validator: labelsValidator }]}>
                                <Table
                                    rowKey="key"
                                    data={labels || []}
                                    columns={editable ? labelTableColumns(setFields) : labelTableColumns(setFields)?.filter(col => col.colKey !== 'action')}
                                />
                            </FormItem>
                        )
                    }}
                </FormItem>
                <FormItem>
                    {({ getFieldValue, setFieldsValue, setFields }) => {
                        return (
                            <FormItem label=" ">
                                <Col span={2}>
                                    {editable && (
                                        <Button variant="text" onClick={() => addLabel(setFields)} icon={<AddIcon />}>添加</Button>
                                    )}
                                </Col>
                            </FormItem>
                        )
                    }}
                </FormItem>
            </Row>
        </div>
    );
};

export default React.memo(LabelInput);
