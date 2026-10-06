"use client";

import { useEffect, useState } from "react";

const copiedConfirmationInMilliseconds = 2000;

export function useCopyToClipboard() {
  const [hasCopied, setHasCopied] = useState(false);

  useEffect(() => {
    if (!hasCopied) {
      return;
    }
    const resetTimer = window.setTimeout(() => setHasCopied(false), copiedConfirmationInMilliseconds);
    return () => window.clearTimeout(resetTimer);
  }, [hasCopied]);

  async function copyText(text: string): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(text);
      setHasCopied(true);
      return true;
    } catch {
      return false;
    }
  }

  return { copyText, hasCopied };
}
