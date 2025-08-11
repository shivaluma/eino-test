"use client";

import { toast } from "sonner";
import {
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import clsx from "clsx";
import { cn, generateUUID } from "lib/utils";
import { useShallow } from "zustand/shallow";
import { Button } from "@/components/ui/button";
import { ArrowDown, Loader } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AnimatePresence, motion } from "framer-motion";
import dynamic from "next/dynamic";
import { useChat } from "@/hooks/use-chat";
import { ChatGreeting } from "./chat-greeting";
import PromptInput from "./prompt-input";
import { appStore } from "@/app/store";
import { createDebounce } from "@/lib/utils";

type ChatBotProps = {
  threadId: string;
  initialMessages: any[];
  selectedModel?: string;
  slots?: {
    emptySlot?: ReactNode;
    inputBottomSlot?: ReactNode;
  };
};

const LightRays = dynamic(() => import("@/components/ui/light-rays"), {
  ssr: false,
});

const Particles = dynamic(() => import("@/components/ui/particles"), {
  ssr: false,
});

const debounce = createDebounce();

export function ChatBot({
  initialMessages,
  threadId,
  selectedModel,
  slots,
}: ChatBotProps) {
  const {
    messages,
    input,
    setInput,
    sendMessage,
    isLoading,
    stop,
  } = useChat({
    api: process.env.NEXT_PUBLIC_API_URL + "/api/v1/messages",
    id: threadId,
    initialMessages: initialMessages,
    onChunk: (chunk) => {
      console.log("Received chunk:", chunk);
    },
    onFinish: (message) => {
      console.log("Message completed:", message);
    },
    onError: (error) => {
      console.error("Chat error:", error);
      toast.error(error.message || "An error occurred");
    },
    prepareRequestBody: ({ message }) => ({
      message,
      conversation_id: threadId,
      stream: true,
      model: selectedModel || "gpt-4",
      metadata: { timestamp: Date.now() },
    }),
  });

  const scrollToBottom = useCallback(() => {
    containerRef.current?.scrollTo({
      top: containerRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, []);

  const [showParticles, setShowParticles] = useState(true);
  const [_thinking, setThinking] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);

  const handleFocus = useCallback(() => {
    // setShowParticles(false);
    debounce(() => setShowParticles(true), 30000);
  }, []);

  const handleScroll = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    const { scrollTop, scrollHeight, clientHeight } = container;
    const isScrollAtBottom = scrollHeight - scrollTop - clientHeight < 50;

    setIsAtBottom(isScrollAtBottom);
    handleFocus();
  }, [handleFocus]);

  const particle = useMemo(() => {
    if (!showParticles) return null;
    return (
      <>
        <div className="absolute top-0 left-0 w-full h-full z-10 fade-in animate-in duration-5000">
          <LightRays />
        </div>
        <div className="absolute top-0 left-0 w-full h-full z-10 fade-in animate-in duration-5000">
          <Particles particleCount={400} particleBaseSize={10} />
        </div>

        <div className="absolute top-0 left-0 w-full h-full z-10 fade-in animate-in duration-5000">
          <div className="w-full h-full bg-gradient-to-t from-background to-50% to-transparent z-20" />
        </div>
        <div className="absolute top-0 left-0 w-full h-full z-10 fade-in animate-in duration-5000">
          <div className="w-full h-full bg-gradient-to-l from-background to-20% to-transparent z-20" />
        </div>
        <div className="absolute top-0 left-0 w-full h-full z-10 fade-in animate-in duration-5000">
          <div className="w-full h-full bg-gradient-to-r from-background to-20% to-transparent z-20" />
        </div>
      </>
    );
  }, [showParticles]);

  const emptyMessage = useMemo(() => {
    return messages.length === 0;
  }, [messages]);

  const _handleThinkingChange = useCallback((newThinking: boolean) => {
    setThinking(newThinking);
  }, []);

  const _append = useCallback(async (message: any) => {
    await sendMessage(message.parts?.[0]?.text || message.content || "");
  }, [sendMessage]);

  return (
    <>
      {particle}

      <div
        className={cn(
          emptyMessage && "justify-center pb-24",
          "flex flex-col min-w-0 relative h-full z-40",
        )}
      >
        {emptyMessage ? (
          slots?.emptySlot ? (
            slots.emptySlot
          ) : (
            <ChatGreeting />
          )
        ) : (
          <>
            <div
              ref={containerRef}
              onScroll={handleScroll}
              className="flex-1 overflow-y-auto py-6"
            >
              <div className="max-w-3xl mx-auto px-4">
                <div className="flex flex-col gap-4">
                  {messages.map((message, index) => (
                    <div
                      key={message.id || index}
                      className={cn(
                        "flex",
                        message.role === "user" ? "justify-end" : "justify-start"
                      )}
                    >
                      <div
                        className={cn(
                          "max-w-[80%] rounded-lg px-4 py-2",
                          message.role === "user"
                            ? "bg-primary text-primary-foreground"
                            : "bg-muted"
                        )}
                      >
                        <div className="text-xs font-medium mb-1 opacity-70">
                          {message.role === "user" ? "You" : "Assistant"}
                        </div>
                        <div className="whitespace-pre-wrap break-words">
                          {message.content}
                        </div>
                      </div>
                    </div>
                  ))}
                  
                  {isLoading && messages[messages.length - 1]?.role === "user" && (
                    <div className="flex justify-start">
                      <div className="max-w-[80%] rounded-lg px-4 py-2 bg-muted">
                        <div className="text-xs font-medium mb-1 opacity-70">
                          Assistant
                        </div>
                        <div className="flex items-center space-x-2">
                          <div className="w-2 h-2 bg-current rounded-full animate-bounce" />
                          <div className="w-2 h-2 bg-current rounded-full animate-bounce delay-100" />
                          <div className="w-2 h-2 bg-current rounded-full animate-bounce delay-200" />
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </>
        )}

        <div
          className={clsx(
            messages.length && "absolute bottom-14",
            "w-full z-10",
          )}
        >
          <div className="max-w-3xl mx-auto relative flex justify-center items-center -top-2">
            <ScrollToBottomButton
              show={!isAtBottom && messages.length > 0}
              onClick={scrollToBottom}
              className=""
            />
          </div>

          <PromptInput
            placeholder="Ask anything"
            input={input}
            setInput={setInput}
            onSubmit={async () => {
              if (input.trim()) {
                await sendMessage(input.trim());
              }
            }}
            onStop={stop}
            isLoading={isLoading}
            onFocus={handleFocus}
          />
          {slots?.inputBottomSlot}
        </div>
      </div>
    </>
  );
}

interface ScrollToBottomButtonProps {
  show: boolean;
  onClick: () => void;
  className?: string;
}

function ScrollToBottomButton({
  show,
  onClick,
  className,
}: ScrollToBottomButtonProps) {
  return (
    <AnimatePresence>
      {show && (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.8 }}
          transition={{ duration: 0.2, ease: "easeInOut" }}
          className={className}
        >
          <Button
            onClick={onClick}
            className="shadow-lg backdrop-blur-sm border transition-colors"
            size="icon"
            variant="ghost"
          >
            <ArrowDown />
          </Button>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
