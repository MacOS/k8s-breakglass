export interface Breakglass extends AvailableBreakglass, ActiveBreakglass {}

export interface AvailableBreakglass {
  from: string;
  to: string;
  duration: number;
  selfApproval: boolean;
  approvalGroups: [string];
}

export interface ActiveBreakglass {
  group: string;
  expiry: number;
}
