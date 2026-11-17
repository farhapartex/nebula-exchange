"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";

import { CoinPackDialog } from "@/features/coin-packs/coin-pack-dialog";

type CoinPackDialogControls = {
  openCoinPackDialog: (missingCoins?: bigint) => void;
};

const CoinPackDialogContext = createContext<CoinPackDialogControls | null>(null);

export function CoinPackDialogProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const [missingCoins, setMissingCoins] = useState<bigint | null>(null);

  const openCoinPackDialog = useCallback((neededCoins?: bigint) => {
    setMissingCoins(neededCoins ?? null);
    setIsOpen(true);
  }, []);
  const controls = useMemo(() => ({ openCoinPackDialog }), [openCoinPackDialog]);

  return (
    <CoinPackDialogContext.Provider value={controls}>
      {children}
      <CoinPackDialog isOpen={isOpen} missingCoins={missingCoins} onOpenChange={setIsOpen} />
    </CoinPackDialogContext.Provider>
  );
}

export function useCoinPackDialog(): CoinPackDialogControls {
  const controls = useContext(CoinPackDialogContext);
  if (!controls) {
    throw new Error("useCoinPackDialog must be used inside CoinPackDialogProvider");
  }
  return controls;
}
