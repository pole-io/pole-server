import { set } from "lodash";
import React from "react";
import { authorizeResources, describeResourcePrincipals, PolicySourceType } from "services/auth_policy";
import { describeAllRoles } from "services/role";
import { describeAllUserGroups } from "services/user_group";
import { describeAllUsers } from "services/users";
import { Drawer, Form, FormProps, Input, Loading, Select } from "tdesign-react";
import { openErrNotification, openInfoNotification } from "utils/notifition";

const { FormItem } = Form;

export interface IAuthorizeInputProps {
    resource_type: PolicySourceType;
    resource_name: string;
    resource_id: string;
    visible: boolean;
    onClose: () => void;
}

const AuthorizeInput: React.FC<IAuthorizeInputProps> = (props) => {
    const [form] = Form.useForm();

    // 下拉/Transfer数据
    const [state, setState] = React.useState<{
        loading: boolean;
        userOpts: { value: string; label: string }[];
        groupOpts: { value: string; label: string }[];
        roleOpts: { value: string; label: string }[];
        checkedUsers: string[];
        checkedGroups: string[];
        checkedRoles: string[];
    }>({
        loading: false,
        userOpts: [],
        groupOpts: [],
        roleOpts: [],
        checkedUsers: [],
        checkedGroups: [],
        checkedRoles: []
    });

    React.useEffect(() => {
        loadOptions();
    }, []);

    // 拉取所有下拉选项的异步函数
    async function loadOptions() {
        setState(pre => ({ ...pre, loading: true }));
        try {
            const [users, groups, rolesRes, principals] = await Promise.all([
                describeAllUsers(),
                describeAllUserGroups(),
                describeAllRoles(),
                describeResourcePrincipals({
                    res_type: props.resource_type,
                    res_id: props.resource_id,
                    action: "ALLOW"
                })
            ]);

            setState(pre => ({
                ...pre,
                loading: false,
                userOpts: users.map((item) => ({
                    value: item.id,
                    label: item.name,
                    // 管理员账户不需要可选
                    disabled: item.user_type === 'main'
                })),
                groupOpts: groups.map((item) => ({
                    value: item.id,
                    label: item.name
                })),
                roleOpts: rolesRes.map((item) => ({
                    value: item.id,
                    label: item.name
                })),
                checkedUsers: principals?.data?.users?.map((item) => item.id) || [],
                checkedGroups: principals?.data?.groups?.map((item) => item.id) || [],
                checkedRoles: principals?.data?.roles?.map((item) => item.id) || [],
            }));
        } catch (err) {
            openErrNotification("获取资源授权失败", err as string);
        }
    }

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (!e.validateResult) {
            return;
        }
        const { users, groups, roles } = e.fields;

        const newData = {
            resource_id: props.resource_id,
            resource_type: props.resource_type,
            resource_name: props.resource_name,
            principals: {
                users: (users as string[]).map((item) => ({ id: item })),
                groups: (groups as string[]).map((item) => ({ id: item })),
                roles: (roles as string[]).map((item) => ({ id: item }))
            }
        };

        try {
            const ret = await authorizeResources([newData]);
            openInfoNotification("请求成功", "授权成功");
            // 关闭弹窗
            setTimeout(() => {
                props.onClose();
            }, 2000)
        } catch (err) {
            openErrNotification("授权失败", err as string);
        }
    }

    return (
        <>
            <Drawer
                header={'授权 (只会影响成员的默认策略)'}
                visible={props.visible}
                size={'large'}
                onClose={() => {
                    props.onClose();
                }}
                onConfirm={() => {
                    form.submit();
                }}
            >
                <Form
                    form={form}
                    labelWidth={100}
                    onSubmit={onSubmit}
                    layout={'vertical'}
                    style={{ width: '100%' }}
                >
                    <FormItem label="资源类型" name="resource_type" initialData={props.resource_type}>
                        <Input readonly={true} />
                    </FormItem>
                    <FormItem label="资源ID" name="resource_id" initialData={props.resource_id}>
                        <Input readonly={true} />
                    </FormItem>
                    <FormItem label="资源名称" name="resource_name" initialData={props.resource_name}>
                        <Input readonly={true} />
                    </FormItem>
                    <Loading loading={state.loading}>
                        <FormItem label="授权用户" name="users" initialData={state.checkedUsers}>
                            <Select multiple={true} options={state.userOpts} />
                        </FormItem>
                        <FormItem label="授权用户组" name="groups" initialData={state.checkedGroups}>
                            <Select multiple={true} options={state.groupOpts} />
                        </FormItem>
                        <FormItem label="授权角色" name="roles" initialData={state.checkedRoles}>
                            <Select multiple={true} options={state.roleOpts} />
                        </FormItem>
                    </Loading>
                </Form>
            </Drawer>
        </>
    )
}

export default React.memo(AuthorizeInput);
