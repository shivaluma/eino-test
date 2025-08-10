"use client";

import {
  getShortcutKeyList,
  isShortcutEvent,
  Shortcuts,
} from "lib/keyboard-shortcuts";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { useTranslations } from "next-intl";
import { useAppStore } from "@/lib/store/store-provider";
import { useEffect } from "react";

export function KeyboardShortcutsPopup() {
  const keyboardShortcutsOpen = useAppStore((state) => state.keyboardShortcutsOpen);
  const toggleKeyboardShortcuts = useAppStore((state) => state.toggleKeyboardShortcuts);
  const t = useTranslations("KeyboardShortcuts");

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isShortcutEvent(e, Shortcuts.openShortcutsPopup)) {
        e.preventDefault();
        e.stopPropagation();
        toggleKeyboardShortcuts();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleKeyboardShortcuts]);

  return (
    <Dialog
      open={keyboardShortcutsOpen}
      onOpenChange={toggleKeyboardShortcuts}
    >
      <DialogContent className="md:max-w-3xl">
        <DialogTitle>{t("title")}</DialogTitle>
        <DialogDescription />
        <div className="grid grid-cols-2 gap-5">
          {Object.entries(Shortcuts).map(([key, shortcut]) => (
            <div
              key={key}
              className="flex items-center gap-2 w-full text-sm px-2"
            >
              <p>{t(shortcut.description ?? "")}</p>
              <div className="flex-1" />
              {getShortcutKeyList(shortcut).map((key) => {
                return (
                  <div
                    key={key}
                    className="p-1.5 text-xs border min-w-8 min-h-8 flex items-center justify-center rounded-md bg-muted"
                  >
                    <span>{key}</span>
                  </div>
                );
              })}
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
