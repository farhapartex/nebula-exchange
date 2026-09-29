import { ContentSection } from "@/components/layout/content-section";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const historyTabs = [
  { value: "ledger", label: "Ledger", body: "Every NC and item movement appears here." },
  { value: "payments", label: "Payments", body: "Card and USDC payments with their status." },
  { value: "trades", label: "Trades", body: "Your exchange fills with fees." },
  { value: "missions", label: "Missions", body: "Past runs and the loot they brought back." },
];

export function TabsShowcase() {
  return (
    <ContentSection title="Tabs" description="Keyboard accessible with arrow keys. Scrolls sideways on small screens.">
      <Tabs defaultValue="ledger">
        <TabsList>
          {historyTabs.map((historyTab) => (
            <TabsTrigger key={historyTab.value} value={historyTab.value}>
              {historyTab.label}
            </TabsTrigger>
          ))}
          <TabsTrigger value="auctions" disabled>
            Auctions
          </TabsTrigger>
        </TabsList>
        {historyTabs.map((historyTab) => (
          <TabsContent key={historyTab.value} value={historyTab.value} className="text-sm text-muted">
            {historyTab.body}
          </TabsContent>
        ))}
      </Tabs>
    </ContentSection>
  );
}
