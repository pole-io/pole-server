import React, { } from 'react';
import { Button, Col, Link, Popconfirm, PrimaryTableProps, Row, Space, Table, TableRowData, Tooltip } from 'tdesign-react';
import { CreditcardIcon, DeleteIcon, EditIcon, RefreshIcon } from 'tdesign-icons-react';

import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { useNavigate } from 'react-router-dom';
import ConfigGroupEditor from './ConfigGroupEditor';
import Search from 'components/Search';

import style from './index.module.less';
import { cleanConfigGroupPage, editorConfigGroup, listConfigGroups, removeConfigGroups, resetConfigGroup, selectConfigGroup } from 'modules/configuration/group';
import { ConfigFileGroup, ConfigFileGroupView } from 'services/config_group';
import AuthorizeInput from 'components/Authorize';
import { PolicySourceType } from 'services/auth_policy';
import { Op } from 'services/types';
import { openErrNotification, openInfoNotification } from 'utils/notifition';

interface IConfigGroupTableProps {
}

const columns = (operateConfigGroup: (op: Op, row: TableRowData) => void, redirect: (row: TableRowData) => void): PrimaryTableProps['columns'] => [
	{
		colKey: 'name',
		title: '分组名',
		cell: ({ row }) => <Link
			theme="primary"
			onClick={() => { redirect(row) }}
		>{row.name}</Link>,
	},
	{
		colKey: 'namespace',
		title: '命名空间',
		cell: ({ row: { namespace } }) => (<Text>{namespace}</Text>),
	},
	{
		colKey: 'department',
		title: '部门',
		cell: ({ row: { department } }) => <Text>{department || '-'}</Text>,
	},
	{
		colKey: 'business',
		title: '业务',
		cell: ({ row: { business } }) => <Text>{business || '-'}</Text>,
	},
	{
		colKey: 'total',
		title: '文件数',
		cell: ({ row: { fileCount } }) => (<Text>{fileCount}</Text>),
	},
	{
		colKey: 'commnet',
		title: '描述',
		ellipsis: true,
		cell: ({ row: { comment } }: TableRowData) => (<Text>{comment || '-'}</Text>),
	},
	{
		colKey: 'time',
		title: '操作时间',
		cell: ({ row: { createTime, modifyTime } }: TableRowData) => <Text>修改: {modifyTime}<br />创建: {createTime}</Text>,
	},
	{
		colKey: 'action',
		title: '操作',
		cell: ({ row }) => {
			return (
				<Space>
					<Tooltip content={row.editable === false ? '无权限操作' : '编辑'}>
						<Button
							shape="square"
							variant="text"
							disabled={row.editable === false}
							onClick={() => operateConfigGroup('edit', row)}>
							<EditIcon />
						</Button>
					</Tooltip>
					<Tooltip content={row.editable === false ? '无权限操作' : '授权'}>
						<Button
							shape="square"
							variant="text"
							disabled={row.editable === false}
							onClick={() => operateConfigGroup('authorize', row)}>
							<CreditcardIcon />
						</Button>
					</Tooltip>
					<Tooltip content={row.deleteable === false ? '无权限操作' : '删除'}>
						<Popconfirm
							content="确认删除吗"
							destroyOnClose
							placement="top"
							showArrow
							theme="default"
							onConfirm={() => {
								operateConfigGroup('delete', row);
							}}
						>
							<Button shape="square" variant="text" disabled={row.deleteable === false}>
								<DeleteIcon />
							</Button>
						</Popconfirm>
					</Tooltip>
				</Space>
			)
		},
	},
];


const ConfigGroupTable: React.FC<IConfigGroupTableProps> = ({ }) => {
	const dispatch = useAppDispatch();
	const navigate = useNavigate()

	const groupState = useAppSelector(selectConfigGroup);
	const { datas, total, page, limit, loading } = groupState;

	// 合并编辑相关状态
	const [editorState, setEditorState] = React.useState<{
		visible: boolean;
		mode: 'create' | 'edit' | 'delete';
		data?: TableRowData;
		authorizeVisible: boolean;
	}>({ visible: false, mode: 'create', data: undefined, authorizeVisible: false });


	// 编辑、新建事件
	const operateConfigGroup = (op: Op, row?: TableRowData) => {
		switch (op) {
			case 'create':
				dispatch(resetConfigGroup());
				setEditorState(prev => ({ ...prev, visible: true, mode: 'create', data: undefined }));
				return;
			case 'edit':
				dispatch(editorConfigGroup(row as ConfigFileGroupView));
				setEditorState(prev => ({ ...prev, visible: true, mode: 'edit', data: { ...row } }));
				return;
			case 'delete':
				dispatch(removeConfigGroups({ param: { id: row?.id } }))
				.then((res) => {
					if (res.meta.requestStatus === 'rejected') {
						// 删除失败
						openErrNotification('请求失败', `删除分组 ${row?.name} 失败, ${res.payload as string}`);
					} else {
						// 删除成功
						openInfoNotification('请求成功', `删除分组 ${row?.name} 成功`);
						refreshTable();
					}
				});
				return;
			case 'authorize':
				setEditorState(prev => ({ ...prev, authorizeVisible: true, data: { ...row } }));
				return;
		}
	}

	React.useEffect(() => {
		refreshTable();
		return () => {
			dispatch(cleanConfigGroupPage());
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	const refreshTable = (page = 1, limit = 10, query = '') => {
		dispatch(listConfigGroups({
			param: {
				offset: (page - 1) * limit,
				limit: limit,
				namespace: query,
				group: query,
			}
		}))
	}

	const table = (
		<>
			<Row justify='space-between' className={style.toolBar}>
				<Col>
					<Row gutter={8} align='middle'>
						<Col>
							<Button onClick={() => operateConfigGroup('create')}>新建</Button>
						</Col>
					</Row>
				</Col>
				<Col>
					<Space>
						<Search
							onChange={(value: string) => {
								refreshTable(1, limit, value);
							}}
						/>
						<Tooltip content="刷新">
							<RefreshIcon onClick={() => refreshTable(1, limit)} />
						</Tooltip>
					</Space>
				</Col>
			</Row>
			{editorState.visible && (
				<ConfigGroupEditor
					key={editorState.mode + (editorState.data?.name || 'new') + (editorState.visible ? '1' : '0')}
					op={editorState.mode}
					visible={editorState.visible}
					closeDrawer={() => {
						// 关闭后重置编辑器状态
						dispatch(resetConfigGroup());
						setEditorState(s => ({ ...s, visible: false }));
						refreshTable();
					}}
				/>
			)}
			{editorState.authorizeVisible && (
				<AuthorizeInput
					resource_type={PolicySourceType.ConfigGroups}
					resource_id={editorState.data?.id}
					resource_name={`${editorState.data?.namespace}/${editorState.data?.name}`}
					visible={editorState.authorizeVisible}
					onClose={() => {
						setEditorState(s => ({ ...s, authorizeVisible: false }));
					}}
				/>
			)}
			<Table
				data={datas}
				columns={columns(operateConfigGroup, (row: TableRowData) => {
					// 发送行信息给到配置文件列表页面，需要限制某些操作
					dispatch(editorConfigGroup({ ...row as ConfigFileGroupView }));
					navigate(`files?namespace=${row.namespace}&group=${row.name}`);
				})}
				loading={loading}
				rowKey="id"
				size={"large"}
				tableLayout={'auto'}
				cellEmptyContent={'-'}
				pagination={{
					current: page,
					pageSize: limit,
					total: total,
					showJumper: true,
					onChange(pageInfo) {
						refreshTable(pageInfo.current, pageInfo.pageSize);
					},
				}}
				onPageChange={(pageInfo) => {
					refreshTable(pageInfo.current, pageInfo.pageSize);
				}}
			/>
		</>
	)

	return (
		<>
			{table}
		</>
	);
}

export default React.memo(ConfigGroupTable);
