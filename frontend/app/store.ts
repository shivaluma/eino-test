import { create } from 'zustand';
import { ChatModel, ChatMention } from '@/types/chat';

interface AppState {
  chatModel: ChatModel;
  threadMentions: Record<string, ChatMention[]>;
  voiceChat: {
    isOpen: boolean;
    agentId?: string;
  };
  mutate: (updates: Partial<AppState> | ((state: AppState) => Partial<AppState>)) => void;
}

export const appStore = create<AppState>((set) => ({
  chatModel: {
    model: 'gpt-4',
    provider: 'openai',
  },
  threadMentions: {},
  voiceChat: {
    isOpen: false,
  },
  mutate: (updates) => {
    set((state) => {
      const newState = typeof updates === 'function' ? updates(state) : updates;
      return { ...state, ...newState };
    });
  },
}));