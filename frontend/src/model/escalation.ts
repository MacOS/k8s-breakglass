export interface BreakglassEscalationSpec {
  cluster: string;
  username: string;
  escalatedGroup: string;
  allowedGroups: Array<string>;
  approvers: Array<BreakglassEscalationApprovers>;
}

interface BreakglassEscalationApprovers {
  users: Array<string>;
  groups: Array<string>;
}
