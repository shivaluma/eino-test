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
import { cn, createDebounce, generateUUID } from "lib/utils";

import { useShallow } from "zustand/shallow";

import { safe } from "ts-safe";

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
    error,
    stop,
    reload,
  } = useChat({
    api: "/api/messages",
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
    },
    prepareRequestBody: ({ message, conversationId }) => ({
      message,
      conversation_id: conversationId,
      stream: true,
      model: selectedModel || "gpt-4",
      metadata: { timestamp: Date.now() },
    }),
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !isLoading) {
      await sendMessage(input);
      setInput("");
    }
  };

  const scrollToBottom = useCallback(() => {
    containerRef.current?.scrollTo({
      top: containerRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, []);

  const [showParticles, setShowParticles] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);

  const handleFocus = useCallback(() => {
    setShowParticles(false);
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
    if (!showParticles) return;
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
              className="flex-1 overflow-y-auto"
            >
              <div className="flex flex-col gap-4">
                {messages.map((message, index) => (
                  <div key={index}>{message.content}</div>
                ))}
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

          {/* <PromptInput
            input={input}
            threadId={threadId}
            append={append}
            thinking={thinking}
            setInput={setInput}
            onThinkingChange={handleThinkingChange}
            isLoading={isLoading || isPendingToolCall}
            onStop={stop}
            onFocus={isFirstTime ? undefined : handleFocus}
          /> */}
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
