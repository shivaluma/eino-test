export interface ChatModel {
  model: string;
  provider: 'openai' | 'anthropic' | 'google' | 'xai';
}

export interface ChatMention {
  type: 'workflow' | 'agent' | 'mcpServer' | 'mcpTool' | 'defaultTool';
  name: string;
  label?: string;
  icon?: {
    style?: any;
    value?: string;
  };
  workflowId?: string;
  agentId?: string;
  serverId?: string;
  serverName?: string;
  toolCount?: number;
  description?: string;
}