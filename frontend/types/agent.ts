export interface Agent {
  id: string;
  name: string;
  description?: string;
  icon?: {
    style?: any;
    value?: string;
  };
  createdAt?: string;
  updatedAt?: string;
  instructions?: string;
}