"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LedgerHistoryTable } from "@/features/history/ledger-history-table";
import { PaymentsHistoryTable } from "@/features/history/payments-history-table";
import { MissionHistoryList } from "@/features/missions/mission-history-list";

const upcomingHistoryTabs = [
  { value: "trades", label: "Trades" },
  { value: "crafts", label: "Crafts" },
  { value: "auctions", label: "Auctions" },
];

export function HistoryTabs() {
  return (
    <Tabs defaultValue="ledger">
      <TabsList>
        <TabsTrigger value="ledger">Ledger</TabsTrigger>
        <TabsTrigger value="payments">Payments</TabsTrigger>
        <TabsTrigger value="missions">Missions</TabsTrigger>
        {upcomingHistoryTabs.map((upcomingTab) => (
          <TabsTrigger key={upcomingTab.value} value={upcomingTab.value} disabled>
            {upcomingTab.label}
          </TabsTrigger>
        ))}
      </TabsList>
      <TabsContent value="ledger">
        <LedgerHistoryTable />
      </TabsContent>
      <TabsContent value="payments">
        <PaymentsHistoryTable />
      </TabsContent>
      <TabsContent value="missions">
        <MissionHistoryList />
      </TabsContent>
    </Tabs>
  );
}
