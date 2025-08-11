'use client';

import {
  CornerRightUp,
  Square,
  PlusIcon,
  Mic,
  AudioLines,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState, KeyboardEvent } from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface PromptInputProps {
  placeholder?: string;
  input: string;
  setInput: (value: string) => void;
  onSubmit: () => void;
  onStop: () => void;
  isLoading?: boolean;
  disabled?: boolean;
  className?: string;
  onFocus?: () => void;
  onBlur?: () => void;
}

export default function PromptInput({
  placeholder = 'Type a message...',
  input,
  setInput,
  onSubmit,
  onStop,
  isLoading = false,
  disabled = false,
  className,
  onFocus,
  onBlur,
}: PromptInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [isFocused, setIsFocused] = useState(false);
  const [isMultiline, setIsMultiline] = useState(false);

  // Auto-resize textarea and detect multiline
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      const scrollHeight = textareaRef.current.scrollHeight;
      textareaRef.current.style.height = `${scrollHeight}px`;
      
      // Detect if multiline (assuming single line is ~24px)
      const newIsMultiline = scrollHeight > 30;
      if (newIsMultiline !== isMultiline) {
        setIsMultiline(newIsMultiline);
      }
    }
  }, [input, isMultiline]);

  // Handle keyboard shortcuts
  const handleKeyDown = useCallback((e: KeyboardEvent<HTMLTextAreaElement>) => {
    // Submit on Enter (without Shift)
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (!isLoading && input.trim()) {
        onSubmit();
      }
    }
    
    // Stop on Escape
    if (e.key === 'Escape' && isLoading) {
      e.preventDefault();
      onStop();
    }
  }, [isLoading, input, onSubmit, onStop]);

  const handleSubmitClick = useCallback(() => {
    if (isLoading) {
      onStop();
    } else if (input.trim()) {
      onSubmit();
    }
  }, [isLoading, input, onSubmit, onStop]);

  const handleFocus = useCallback(() => {
    setIsFocused(true);
    onFocus?.();
  }, [onFocus]);

  const handleBlur = useCallback(() => {
    setIsFocused(false);
    onBlur?.();
  }, [onBlur]);

  // Focus textarea on mount
  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  return (
    <div className={cn('max-w-3xl mx-auto fade-in animate-in px-4', className)}>
      <div className="relative flex min-h-14 w-full items-end">
        {/* Main input container with background */}
        <div 
          className={cn(
            'relative flex w-full flex-auto transition-all duration-300 ease-in-out',
            'bg-zinc-800/90 backdrop-blur-sm border border-zinc-700/50',
            'rounded-3xl shadow-xl',
            // Dynamic layout based on multiline
            isMultiline ? 'flex-col' : 'flex-row items-center',
            // Focus states
            isFocused && 'ring-2 ring-zinc-600/30 bg-zinc-800 border-zinc-600/50',
            !isFocused && 'hover:bg-zinc-800/95 hover:border-zinc-600/50',
          )}
        >
          {/* Plus button - positioned dynamically */}
          <div 
            className={cn(
              'flex items-center transition-all duration-300 ease-in-out',
              isMultiline ? 'absolute start-2.5 bottom-2.5' : 'pl-4'
            )}
          >
            <Button
              variant="ghost"
              size="sm"
              className={cn(
                'h-9 w-9 rounded-full p-0 transition-all duration-200',
                'hover:bg-zinc-700/50 text-zinc-400 hover:text-white'
              )}
              onClick={() => {
                console.log('Attachment clicked');
              }}
              disabled={disabled}
              type="button"
            >
              <PlusIcon className="h-5 w-5" />
            </Button>
          </div>

          {/* Text input - single textarea with dynamic wrapper */}
          <div 
            className={cn(
              'flex-1 transition-all duration-300 ease-in-out',
              isMultiline ? 'px-5.5 pt-3 pb-16' : 'h-6 my-3 mx-2'
            )}
          >
            <div className={isMultiline ? 'max-h-[max(35svh,5rem)] max-h-52 overflow-y-auto' : ''}>
              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                onFocus={handleFocus}
                onBlur={handleBlur}
                placeholder={placeholder}
                disabled={disabled || isLoading}
                className={cn(
                  'w-full resize-none bg-transparent outline-none border-none',
                  'text-white placeholder:text-zinc-400 text-base leading-6',
                  'overflow-hidden transition-all duration-300 ease-in-out',
                  isMultiline ? 'min-h-6' : '',
                )}
                rows={1}
                style={{ fieldSizing: 'content' } as any}
              />
            </div>
          </div>

          {/* Right actions - positioned dynamically */}
          <div 
            className={cn(
              'flex items-center gap-1.5 transition-all duration-300 ease-in-out',
              isMultiline ? 'absolute end-2.5 bottom-2.5' : 'pr-4'
            )}
          >
            <Button
              variant="ghost"
              size="sm"
              className={cn(
                'h-9 w-9 rounded-full p-0 transition-all duration-200',
                'hover:bg-zinc-700/50 text-zinc-400 hover:text-white'
              )}
              onClick={() => {
                console.log('Microphone clicked');
              }}
              disabled={disabled}
              type="button"
            >
              <Mic className="h-5 w-5" />
            </Button>

            {!isMultiline && (
              <Button
                variant="ghost"
                size="sm"
                className={cn(
                  'h-9 w-9 rounded-full p-0 transition-all duration-200',
                  'hover:bg-zinc-700/50 text-zinc-400 hover:text-white'
                )}
                onClick={() => {
                  console.log('Audio lines clicked');
                }}
                disabled={disabled}
                type="button"
              >
                <AudioLines className="h-5 w-5" />
              </Button>
            )}

            {(input.trim() || isLoading) && (
              <Button
                onClick={handleSubmitClick}
                className={cn(
                  'h-9 w-9 rounded-full p-0 transition-all duration-200 ml-1',
                  isLoading 
                    ? 'bg-red-600 hover:bg-red-700 text-white' 
                    : 'bg-white hover:bg-zinc-100 text-black',
                )}
                disabled={disabled && !isLoading}
              >
                {isLoading ? (
                  <Square size={20} className="fill-current" />
                ) : (
                  <CornerRightUp size={20} />
                )}
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}