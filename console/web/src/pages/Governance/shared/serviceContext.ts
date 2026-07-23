export type GovernanceServiceRole = 'caller' | 'callee';

export interface GovernanceServiceContext {
    namespace: string;
    service: string;
    role: GovernanceServiceRole;
}
