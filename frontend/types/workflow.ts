export interface WorkflowSummary {
  id: string;
  name: string;
  description?: string;
  icon?: {
    style?: any;
    value?: string;
  };
}